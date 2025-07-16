package event

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/gerfey/messenger/api"
)

type EventDispatcher struct {
	mu        sync.RWMutex
	listeners map[reflect.Type][]any
}

func NewEventDispatcher() api.EventDispatcher {
	return &EventDispatcher{
		listeners: make(map[reflect.Type][]any),
	}
}

func (d *EventDispatcher) AddListener(event any, listener any) {
	d.mu.Lock()
	defer d.mu.Unlock()

	t := reflect.TypeOf(event)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	d.listeners[t] = append(d.listeners[t], listener)
}

func (d *EventDispatcher) Dispatch(ctx context.Context, event any) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	eventType := reflect.TypeOf(event)
	if eventType.Kind() == reflect.Ptr {
		eventType = eventType.Elem()
	}

	d.mu.RLock()
	listeners := d.listeners[eventType]
	d.mu.RUnlock()

	if len(listeners) == 0 {
		return nil
	}

	for _, listener := range listeners {
		v := reflect.ValueOf(listener)
		method := v.MethodByName("Handle")
		if !method.IsValid() {
			continue
		}

		mType := method.Type()

		switch mType.NumIn() {
		case 1: // Handle(event)
			if mType.In(0) == reflect.TypeOf(event) {
				method.Call([]reflect.Value{reflect.ValueOf(event)})
				continue
			}
		case 2: // Handle(context.Context, event)
			if mType.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) &&
				mType.In(1) == reflect.TypeOf(event) {

				method.Call([]reflect.Value{
					reflect.ValueOf(ctx),
					reflect.ValueOf(event),
				})
				continue
			}
		}
	}

	return nil
}
