package implementation

import (
	"context"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/event"
	"github.com/gerfey/messenger/core/stamps"
)

type SendMessageMiddleware struct {
	router           api.Router
	transportLocator api.TransportLocator
	eventDispatcher  api.EventDispatcher
}

func NewSendMessageMiddleware(
	router api.Router,
	transportLocator api.TransportLocator,
	eventDispatcher api.EventDispatcher,
) api.Middleware {
	return &SendMessageMiddleware{
		router:           router,
		transportLocator: transportLocator,
		eventDispatcher:  eventDispatcher,
	}
}

func (m *SendMessageMiddleware) Handle(ctx context.Context, env api.Envelope, next api.NextFunc) (api.Envelope, error) {
	if _, ok := envelope.LastStampOf[stamps.ReceivedStamp](env); ok {
		return next(ctx, env)
	}

	msg := env.Message()
	transportNames := m.router.GetTransportFor(msg)

	if len(transportNames) == 0 {
		return next(ctx, env)
	}

	errDispatcher := m.eventDispatcher.Dispatch(ctx, &event.SendMessageToTransportsEvent{
		Ctx:            ctx,
		Envelope:       env,
		TransportNames: transportNames,
	})
	if errDispatcher != nil {
		return nil, errDispatcher
	}

	for _, name := range transportNames {
		sender := m.transportLocator.GetTransport(name)

		err := sender.Send(ctx, env)
		if err != nil {
			return nil, err
		}
		env = env.WithStamp(stamps.SentStamp{Transport: name})
	}

	return env, nil
}
