package implementation

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/event"
	"github.com/gerfey/messenger/core/stamps"
)

type SendMessageMiddleware struct {
	logger          *slog.Logger
	senderLocator   api.SenderLocator
	eventDispatcher api.EventDispatcher
}

func NewSendMessageMiddleware(
	logger *slog.Logger,
	senderLocator api.SenderLocator,
	eventDispatcher api.EventDispatcher,
) api.Middleware {
	return &SendMessageMiddleware{
		logger:          logger,
		senderLocator:   senderLocator,
		eventDispatcher: eventDispatcher,
	}
}

func (m *SendMessageMiddleware) Handle(ctx context.Context, env api.Envelope, next api.NextFunc) (api.Envelope, error) {
	if _, ok := envelope.LastStampOf[stamps.ReceivedStamp](env); ok {
		m.logger.DebugContext(ctx, "received message from transport")

		return next(ctx, env)
	}

	msg := env.Message()

	senders := m.senderLocator.GetSenders(env)

	if len(senders) == 0 {
		m.logger.WarnContext(ctx, "no senders configured for message", "message_type", msg)

		return env, fmt.Errorf("no senders configured for message %T", msg)
	}

	var isSent = false

	for _, sender := range senders {
		m.logger.DebugContext(ctx, "sending message to sender",
			"message_type", msg,
			"sender", sender.Name())

		errDispatcher := m.eventDispatcher.Dispatch(ctx, &event.SendMessageToTransportsEvent{
			Ctx:      ctx,
			Envelope: env,
			Senders:  senders,
		})
		if errDispatcher != nil {
			m.logger.ErrorContext(ctx, "failed to dispatch send event", "error", errDispatcher)

			return nil, errDispatcher
		}

		env = env.WithStamp(stamps.SentStamp{SenderName: sender.Name()})

		err := sender.Send(ctx, env)
		if err != nil {
			m.logger.ErrorContext(ctx, "failed to send message to sender",
				"sender", sender.Name(),
				"error", err)

			return nil, err
		}

		isSent = true

		m.logger.DebugContext(ctx, "message sent successfully", "sender", sender.Name())
	}

	if !isSent {
		return next(ctx, env)
	}

	return env, nil
}
