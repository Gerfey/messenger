package transport

import (
	"fmt"

	"gopkg.in/yaml.v3"

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
			rawOptions, errOptions := yaml.Marshal(config.Options)
			if errOptions != nil {
				return nil, fmt.Errorf("%s: marshal options map: %w", name, errOptions)
			}

			return factory.Create(name, config.DSN, rawOptions)
		}
	}

	return nil, fmt.Errorf("no transport factory supports DSN '%s' for transport '%s'", config.DSN, name)
}

func (c *FactoryChain) Factories() []api.TransportFactory {
	return c.factories
}
