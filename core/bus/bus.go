package bus

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/stamps"
)

type Bus struct {
	name            string
	middlewareChain []api.Middleware
}

func NewBus(name string, middleware ...api.Middleware) api.MessageBus {
	return &Bus{
		name:            name,
		middlewareChain: middleware,
	}
}

func (b *Bus) Dispatch(ctx context.Context, msg any, st ...api.Stamp) (api.Envelope, error) {
	if _, ok := msg.(api.Envelope); ok {
		return nil, fmt.Errorf("message type must not be %v", reflect.TypeOf(envelope.Envelope{}))
	}

	env := envelope.NewEnvelope(msg)
	for _, s := range st {
		env = env.WithStamp(s)
	}

	if env.LastStampOfType(reflect.TypeOf(stamps.BusNameStamp{})) == nil {
		env = env.WithStamp(stamps.BusNameStamp{
			Name: b.name,
		})
	}

	return b.buildChain()(ctx, env)
}

func (b *Bus) DispatchWithEnvelope(ctx context.Context, env api.Envelope) (api.Envelope, error) {
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
