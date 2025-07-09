package inmemory

import (
	"strings"

	"github.com/gerfey/messenger/config"
	"github.com/gerfey/messenger/transport"
)

type InMemoryTransportFactory struct {
	resolver transport.TypeResolver
}

func NewInMemoryTransportFactory(resolver transport.TypeResolver) *InMemoryTransportFactory {
	return &InMemoryTransportFactory{resolver: resolver}
}

func (f *InMemoryTransportFactory) Supports(dsn string) bool {
	return strings.HasPrefix(dsn, "in-memory://")
}

func (f *InMemoryTransportFactory) Create(name string, dsn string, options config.OptionsConfig) (transport.Transport, error) {
	return New(), nil
}
