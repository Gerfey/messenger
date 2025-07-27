package api

import (
	"context"
	"reflect"

	"github.com/gerfey/messenger/config"
)

type Transport interface {
	Sender
	Receiver
}

type Sender interface {
	Name() string
	Send(context.Context, Envelope) error
}

type Receiver interface {
	Receive(context.Context, func(context.Context, Envelope) error) error
}

type RetryableTransport interface {
	Transport
	Retry(context.Context, Envelope) error
}

type SenderLocator interface {
	Register(string, Sender) error
	GetSenders(Envelope) []Sender
	RegisterMessageType(reflect.Type, []string)
	SetFallback([]string)
}

type TransportFactory interface {
	Supports(string) bool
	Create(string, string, config.OptionsConfig) (Transport, error)
}

type RoutedMessage interface {
	RoutingKey() string
}
