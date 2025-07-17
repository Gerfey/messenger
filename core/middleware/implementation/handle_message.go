package implementation

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/stamps"
)

type HandleMessageMiddleware struct {
	handlersLocator api.HandlerLocator
}

func NewHandleMessageMiddleware(handlersLocator api.HandlerLocator) api.Middleware {
	return &HandleMessageMiddleware{
		handlersLocator: handlersLocator,
	}
}

func (h *HandleMessageMiddleware) Handle(ctx context.Context, env api.Envelope, next api.NextFunc) (api.Envelope, error) {
	if _, ok := envelope.LastStampOf[stamps.SentStamp](env); ok {
		return env, nil
	}

	msg := env.Message()

	handlers := h.handlersLocator.Get(msg)

	if len(handlers) == 0 {
		return nil, fmt.Errorf("no handlers registered for message type %T", msg)
	}

	for _, handlerFunc := range handlers {
		results := handlerFunc.Fn.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(msg)})

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
			return nil, fmt.Errorf("handler %s failed for message type %T: %w", handlerFunc.HandlerStr, msg, err)
		}

		env = env.WithStamp(stamps.HandledStamp{
			Handler:    handlerFunc.HandlerStr,
			Result:     result,
			ResultType: reflect.TypeOf(result),
		})
	}

	return next(ctx, env)
}
