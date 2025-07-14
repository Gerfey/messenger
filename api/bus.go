package api

import (
	"context"
)

type MessageBus interface {
	Dispatch(context.Context, any, ...Stamp) (Envelope, error)
}

type BusLocator interface {
	Register(string, MessageBus) error
	Get(string) (MessageBus, bool)
	GetAll() []MessageBus
}
