package middlewares

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gerfey/messenger/core"
	"github.com/gerfey/messenger/envelope"
	"github.com/gerfey/messenger/stamps"
)

type HandleMessageMiddleware struct {
	handlers *core.HandlersRegistry
}

func NewHandleMessageMiddleware(reg *core.HandlersRegistry) *HandleMessageMiddleware {
	return &HandleMessageMiddleware{
		handlers: reg,
	}
}

func (h *HandleMessageMiddleware) Handle(ctx context.Context, env *envelope.Envelope, next core.NextFunc) (*envelope.Envelope, error) {
	if env.LastStampOfType(reflect.TypeOf(stamps.SentStamp{})) != nil {
		return env, nil
	}

	msg := env.Message()

	handlers := h.handlers.GetHandlers(msg)

	if len(handlers) == 0 {
		return nil, fmt.Errorf("no handlers for message %T", msg)
	}

	for _, handler := range handlers {
		results := handler.Fn.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(msg)})

		var result any
		var err error

		if len(results) == 1 {
			if e, ok := results[0].Interface().(error); ok {
				err = e
			} else {
				result = results[0].Interface()
			}
		} else if len(results) == 2 {
			result = results[0].Interface()
			if e, ok := results[1].Interface().(error); ok {
				err = e
			}
		}

		if err != nil {
			return nil, fmt.Errorf("handler %s failed: %w", handler.HandlerStr, err)
		}

		env = env.WithStamp(stamps.HandledStamp{
			Handler:    handler.HandlerStr,
			Result:     result,
			ResultType: reflect.TypeOf(result),
		})
	}

	return next(ctx, env)
}
