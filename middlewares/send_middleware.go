package middlewares

import (
	"context"
	"reflect"

	"github.com/gerfey/messenger/envelope"
	"github.com/gerfey/messenger/routing"
	"github.com/gerfey/messenger/stamps"
	"github.com/gerfey/messenger/transport"
)

type SendMessageMiddleware struct {
	router           *routing.Router
	transportLocator *transport.TransportLocator
}

func NewSendMessageMiddleware(
	router *routing.Router,
	transportLocator *transport.TransportLocator,
) *SendMessageMiddleware {
	return &SendMessageMiddleware{
		router:           router,
		transportLocator: transportLocator,
	}
}

func (m *SendMessageMiddleware) Handle(ctx context.Context, env *envelope.Envelope, next NextFunc) (*envelope.Envelope, error) {
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
