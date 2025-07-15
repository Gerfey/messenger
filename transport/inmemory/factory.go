package inmemory

import (
	"strings"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/config"
)

type TransportFactory struct {
	resolver api.TypeResolver
}

func NewTransportFactory(resolver api.TypeResolver) api.TransportFactory {
	return &TransportFactory{resolver: resolver}
}

func (f *TransportFactory) Supports(dsn string) bool {
	return strings.HasPrefix(dsn, "in-memory://")
}

func (f *TransportFactory) Create(name string, dsn string, options config.OptionsConfig) (api.Transport, error) {
	return NewTransport(), nil
}
