package transport

import (
	"context"

	"github.com/gerfey/messenger/envelope"
)

type Transport interface {
	Send(ctx context.Context, env *envelope.Envelope) error
	Receive(ctx context.Context, handler func(context.Context, *envelope.Envelope) error) error
}

type Sender interface {
	Send(ctx context.Context, env *envelope.Envelope) error
}

type Receiver interface {
	Receive(ctx context.Context, handler func(context.Context, *envelope.Envelope) error) error
}
