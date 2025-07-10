package implementation

import (
	"context"
	"reflect"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/routing"
	"github.com/gerfey/messenger/stamps"
)

type SendMessageMiddleware struct {
	router           *routing.Router
	transportLocator api.TransportLocator
}

func NewSendMessageMiddleware(
	router *routing.Router,
	transportLocator api.TransportLocator,
) api.Middleware {
	return &SendMessageMiddleware{
		router:           router,
		transportLocator: transportLocator,
	}
}

func (m *SendMessageMiddleware) Handle(ctx context.Context, env api.Envelope, next api.NextFunc) (api.Envelope, error) {
	if env.LastStampOfType(reflect.TypeOf(stamps.ReceivedStamp{})) != nil {
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
