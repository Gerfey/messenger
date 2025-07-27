package implementation

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/stamps"
)

const (
	expectedResultsWithError = 2
)

type HandleMessageMiddleware struct {
	logger          *slog.Logger
	handlersLocator api.HandlerLocator
}

func NewHandleMessageMiddleware(logger *slog.Logger, handlersLocator api.HandlerLocator) api.Middleware {
	return &HandleMessageMiddleware{
		logger:          logger,
		handlersLocator: handlersLocator,
	}
}

func (h *HandleMessageMiddleware) Handle(
	ctx context.Context,
	env api.Envelope,
	next api.NextFunc,
) (api.Envelope, error) {
	msg := env.Message()
	msgType := reflect.TypeOf(msg)

	handlers := h.handlersLocator.Get(msg)

	if len(handlers) == 0 {
		h.logger.WarnContext(ctx, "no handlers registered for message type", "message_type", msgType.String())

		return nil, fmt.Errorf("no handlers registered for message type %T", msg)
	}

	h.logger.DebugContext(
		ctx,
		"processing message",
		"message_type",
		msgType.String(),
		"handlers_count",
		len(handlers),
	)

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
		} else if len(results) == expectedResultsWithError {
			result = results[0].Interface()
			if e, ok := results[1].Interface().(error); ok {
				err = e
			}
		}

		if err != nil {
			h.logger.ErrorContext(ctx, "handler failed",
				"handler", handlerFunc.HandlerStr,
				"message_type", msgType.String(),
				"error", err)

			return nil, fmt.Errorf("handler %s failed for message type %T: %w", handlerFunc.HandlerStr, msg, err)
		}

		env = env.WithStamp(stamps.HandledStamp{
			Handler:    handlerFunc.HandlerStr,
			Result:     result,
			ResultType: reflect.TypeOf(result),
		})

		h.logger.DebugContext(ctx, "message handled successfully",
			"handler", handlerFunc.HandlerStr,
			"message_type", msgType.String())
	}

	return next(ctx, env)
}
