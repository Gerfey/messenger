package implementation

import (
	"context"
	"log/slog"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/event"
	"github.com/gerfey/messenger/core/stamps"
)

type SendMessageMiddleware struct {
	router           api.Router
	transportLocator api.TransportLocator
	eventDispatcher  api.EventDispatcher
	logger           *slog.Logger
}

func NewSendMessageMiddleware(
	router api.Router,
	transportLocator api.TransportLocator,
	eventDispatcher api.EventDispatcher,
	logger *slog.Logger,
) api.Middleware {
	return &SendMessageMiddleware{
		router:           router,
		transportLocator: transportLocator,
		eventDispatcher:  eventDispatcher,
		logger:           logger,
	}
}

func (m *SendMessageMiddleware) Handle(ctx context.Context, env api.Envelope, next api.NextFunc) (api.Envelope, error) {
	if _, ok := envelope.LastStampOf[stamps.ReceivedStamp](env); ok {
		return next(ctx, env)
	}

	msg := env.Message()
	transportNames := m.router.GetTransportFor(msg)

	if len(transportNames) == 0 {
		m.logger.WarnContext(ctx, "no transports configured for message", "message_type", msg)

		return next(ctx, env)
	}

	m.logger.DebugContext(ctx, "sending message to transports",
		"message_type", msg,
		"transports", transportNames)

	errDispatcher := m.eventDispatcher.Dispatch(ctx, &event.SendMessageToTransportsEvent{
		Ctx:            ctx,
		Envelope:       env,
		TransportNames: transportNames,
	})
	if errDispatcher != nil {
		m.logger.ErrorContext(ctx, "failed to dispatch send event", "error", errDispatcher)

		return nil, errDispatcher
	}

	for _, name := range transportNames {
		sender := m.transportLocator.GetTransport(name)
		if sender == nil {
			m.logger.ErrorContext(ctx, "transport not found", "transport", name)

			continue
		}

		err := sender.Send(ctx, env)
		if err != nil {
			m.logger.ErrorContext(ctx, "failed to send message to transport",
				"transport", name,
				"error", err)

			return nil, err
		}
		env = env.WithStamp(stamps.SentStamp{Transport: name})

		m.logger.DebugContext(ctx, "message sent successfully", "transport", name)
	}

	return next(ctx, env)
}
