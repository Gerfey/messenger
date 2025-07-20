package event

import (
	"context"
	"log/slog"
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
	logger    *slog.Logger
}

func NewEventDispatcher(logger *slog.Logger) api.EventDispatcher {
	return &Dispatcher{
		listeners: make(map[reflect.Type][]any),
		logger:    logger,
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

	d.logger.Debug("event listener added", "event_type", t.String(), "listener", reflect.TypeOf(listener).String())
}

func (d *Dispatcher) Dispatch(ctx context.Context, event any) error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	t := reflect.TypeOf(event)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	listeners, exists := d.listeners[t]
	if !exists {
		d.logger.DebugContext(ctx, "no listeners found for event", "event_type", t.String())

		return nil
	}

	d.logger.DebugContext(ctx, "dispatching event", "event_type", t.String(), "listeners_count", len(listeners))

	for _, listener := range listeners {
		listenerValue := reflect.ValueOf(listener)
		listenerType := reflect.TypeOf(listener)

		var method reflect.Value
		if listenerValue.Kind() == reflect.Func {
			method = listenerValue
		} else {
			method = listenerValue.MethodByName("Handle")
			if !method.IsValid() {
				d.logger.ErrorContext(ctx, "listener does not have Handle method", "listener", listenerType.String())

				continue
			}
		}

		methodType := method.Type()
		numIn := methodType.NumIn()

		var args []reflect.Value
		switch numIn {
		case handlerParamsWithEvent:
			args = []reflect.Value{reflect.ValueOf(event)}
		case handlerParamsWithContextEvent:
			args = []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(event)}
		default:
			d.logger.ErrorContext(ctx, "invalid handler signature",
				"listener", listenerType.String(),
				"expected_params", "1 or 2",
				"actual_params", numIn)

			continue
		}

		results := method.Call(args)
		if len(results) > 0 && !results[0].IsNil() {
			if err, isError := results[0].Interface().(error); isError {
				d.logger.ErrorContext(ctx, "event handler failed",
					"event_type", t.String(),
					"listener", listenerType.String(),
					"error", err)

				return err
			}
		}

		d.logger.DebugContext(ctx, "event handled successfully",
			"event_type", t.String(),
			"listener", listenerType.String())
	}

	return nil
}
