package amqp

import (
	"strings"

	messenger2 "github.com/gerfey/messenger/config"
	"github.com/gerfey/messenger/transport"
)

type AMQPTransportFactory struct {
	resolver transport.TypeResolver
}

func NewAMQPTransportFactory(resolver transport.TypeResolver) *AMQPTransportFactory {
	return &AMQPTransportFactory{resolver: resolver}
}

func (f *AMQPTransportFactory) Supports(dsn string) bool {
	return strings.HasPrefix(dsn, "amqp://")
}

func (f *AMQPTransportFactory) Create(name string, dsn string, options messenger2.OptionsConfig) (transport.Transport, error) {
	cfg := TransportConfig{
		Name:    name,
		DSN:     dsn,
		Options: options,
	}

	return New(cfg, f.resolver)
}
