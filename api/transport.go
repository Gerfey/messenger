package api

import (
	"context"

	"github.com/gerfey/messenger/config"
)

type Transport interface {
	Name() string
	Send(context.Context, Envelope) error
	Receive(context.Context, func(context.Context, Envelope) error) error
}

type RetryableTransport interface {
	Transport
	Retry(context.Context, Envelope) error
}

type Sender interface {
	Send(context.Context, Envelope) error
}

type Receiver interface {
	Receive(context.Context, func(context.Context, Envelope) error) error
}

type TransportLocator interface {
	Register(string, Transport) error
	GetAllTransports() []Transport
	GetTransport(string) Transport
}

type TransportFactory interface {
	Supports(string) bool
	Create(string, string, config.OptionsConfig) (Transport, error)
}

type RoutedMessage interface {
	RoutingKey() string
}
