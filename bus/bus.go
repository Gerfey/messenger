package bus

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gerfey/messenger/envelope"
	"github.com/gerfey/messenger/middlewares"
	"github.com/gerfey/messenger/stamps"
)

type Bus struct {
	name            string
	middlewareChain []middlewares.Middleware
}

func NewBus(name string, middleware ...middlewares.Middleware) *Bus {
	return &Bus{
		name:            name,
		middlewareChain: middleware,
	}
}

func (b *Bus) Dispatch(ctx context.Context, msg any, st ...envelope.Stamp) (*envelope.Envelope, error) {
	if _, ok := msg.(*envelope.Envelope); ok {
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

func (b *Bus) DispatchWithEnvelope(ctx context.Context, env *envelope.Envelope) (*envelope.Envelope, error) {
	return b.buildChain()(ctx, env)
}

func (b *Bus) buildChain() middlewares.NextFunc {
	handler := func(ctx context.Context, env *envelope.Envelope) (*envelope.Envelope, error) {
		return env, nil
	}

	for i := len(b.middlewareChain) - 1; i >= 0; i-- {
		handler = b.createMiddlewareHandler(b.middlewareChain[i], handler)
	}

	return handler
}

func (b *Bus) createMiddlewareHandler(m middlewares.Middleware, next middlewares.NextFunc) middlewares.NextFunc {
	return func(ctx context.Context, env *envelope.Envelope) (*envelope.Envelope, error) {
		return m.Handle(ctx, env, next)
	}
}
