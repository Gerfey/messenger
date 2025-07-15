package implementation

import (
	"context"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/stamps"
)

type SendMessageMiddleware struct {
	router           api.Router
	transportLocator api.TransportLocator
}

func NewSendMessageMiddleware(
	router api.Router,
	transportLocator api.TransportLocator,
) api.Middleware {
	return &SendMessageMiddleware{
		router:           router,
		transportLocator: transportLocator,
	}
}

func (m *SendMessageMiddleware) Handle(ctx context.Context, env api.Envelope, next api.NextFunc) (api.Envelope, error) {
	receivedStamp := envelope.HasStampOf[stamps.ReceivedStamp](env)
	if receivedStamp {
		return next(ctx, env)
	}

	msg := env.Message()
	transportNames := m.router.GetTransportFor(msg)

	if len(transportNames) == 0 {
		return next(ctx, env)
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
