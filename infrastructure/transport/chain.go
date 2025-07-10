package transport

import (
	"fmt"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/config"
)

type FactoryChain struct {
	factories []api.TransportFactory
}

func NewFactoryChain(factories ...api.TransportFactory) *FactoryChain {
	return &FactoryChain{factories: factories}
}

func (c *FactoryChain) CreateTransport(name string, config config.TransportConfig) (api.Transport, error) {
	for _, factory := range c.factories {
		if factory.Supports(config.DSN) {
			return factory.Create(name, config.DSN, config.Options)
		}
	}
	return nil, fmt.Errorf("no factory supports DSN: %s", config.DSN)
}

func (c *FactoryChain) Factories() []api.TransportFactory {
	return c.factories
}
