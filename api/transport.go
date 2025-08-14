package api

import (
	"context"
	"reflect"
)

type Transport interface {
	Sender
	Receiver
	Closer
}

type Sender interface {
	Name() string
	Send(context.Context, Envelope) error
}

type Receiver interface {
	Receive(context.Context, func(context.Context, Envelope) error) error
}

type Closer interface {
	Close() error
}

type Producer interface {
	Send(context.Context, Envelope) error
	Close() error
}

type Consumer interface {
	Consume(context.Context, func(context.Context, Envelope) error) error
	Close() error
}

type RetryableTransport interface {
	Transport
	Retry(context.Context, Envelope) error
}

type SetupableTransport interface {
	Transport
	Setup(ctx context.Context) error
}

type SenderLocator interface {
	Register(string, Sender) error
	GetSenders(Envelope) []Sender
	RegisterMessageType(reflect.Type, []string)
	SetFallback([]string)
}

type TransportFactory interface {
	Supports(string) bool
	Create(string, string, []byte, Serializer) (Transport, error)
}

type RoutedMessage interface {
	RoutingKey() string
}
