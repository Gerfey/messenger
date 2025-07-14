package bus

import (
	"context"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/envelope"
)

type Bus struct {
	middlewareChain []api.Middleware
}

func NewBus(middleware ...api.Middleware) api.MessageBus {
	return &Bus{
		middlewareChain: middleware,
	}
}

func (b *Bus) Dispatch(ctx context.Context, msg any, st ...api.Stamp) (api.Envelope, error) {
	env, ok := msg.(api.Envelope)
	if !ok {
		env = envelope.NewEnvelope(msg)
	}

	for _, s := range st {
		env = env.WithStamp(s)
	}

	return b.buildChain()(ctx, env)
}

func (b *Bus) buildChain() api.NextFunc {
	handler := func(ctx context.Context, env api.Envelope) (api.Envelope, error) {
		return env, nil
	}

	for i := len(b.middlewareChain) - 1; i >= 0; i-- {
		handler = b.createMiddlewareHandler(b.middlewareChain[i], handler)
	}

	return handler
}

func (b *Bus) createMiddlewareHandler(m api.Middleware, next api.NextFunc) api.NextFunc {
	return func(ctx context.Context, env api.Envelope) (api.Envelope, error) {
		return m.Handle(ctx, env, next)
	}
}
