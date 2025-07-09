package factory

import (
	"fmt"

	"github.com/gerfey/messenger/config"
	"github.com/gerfey/messenger/transport"
)

type TransportFactoryChain struct {
	factories []TransportFactory
}

func NewChain(factories ...TransportFactory) *TransportFactoryChain {
	return &TransportFactoryChain{factories: factories}
}

func (c *TransportFactoryChain) CreateTransport(name string, config config.TransportConfig) (transport.Transport, error) {
	for _, factory := range c.factories {
		if factory.Supports(config.DSN) {
			return factory.Create(name, config.DSN, config.Options)
		}
	}
	return nil, fmt.Errorf("no factory supports DSN: %s", config.DSN)
}

func (c *TransportFactoryChain) Factories() []TransportFactory {
	return c.factories
}
