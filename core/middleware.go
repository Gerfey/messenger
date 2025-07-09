package core

import (
	"context"

	"github.com/gerfey/messenger/envelope"
)

type NextFunc func(ctx context.Context, env *envelope.Envelope) (*envelope.Envelope, error)

type Middleware interface {
	Handle(ctx context.Context, env *envelope.Envelope, next NextFunc) (*envelope.Envelope, error)
}
