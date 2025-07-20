package event

import (
	"context"
	"errors"
	"reflect"
	"sync"

	"github.com/gerfey/messenger/api"
)

const (
	handlerParamsWithEvent        = 1
	handlerParamsWithContextEvent = 2
)

type Dispatcher struct {
	mu        sync.RWMutex
	listeners map[reflect.Type][]any
}

func NewEventDispatcher() api.EventDispatcher {
	return &Dispatcher{
		listeners: make(map[reflect.Type][]any),
	}
}

func (d *Dispatcher) AddListener(event any, listener any) {
	d.mu.Lock()
	defer d.mu.Unlock()

	t := reflect.TypeOf(event)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	d.listeners[t] = append(d.listeners[t], listener)
}

func (d *Dispatcher) Dispatch(ctx context.Context, event any) error {
	if event == nil {
		return errors.New("event cannot be nil")
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
		case handlerParamsWithEvent: // Handle(event)

			if mType.In(0) == reflect.TypeOf(event) {
				method.Call([]reflect.Value{reflect.ValueOf(event)})

				continue
			}
		case handlerParamsWithContextEvent: // Handle(context.Context, event)
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
