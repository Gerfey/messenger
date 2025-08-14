package inmemory

import (
	"strings"

	"github.com/gerfey/messenger/api"
)

type TransportFactory struct{}

func NewTransportFactory() api.TransportFactory {
	return &TransportFactory{}
}

func (f *TransportFactory) Supports(dsn string) bool {
	return strings.HasPrefix(dsn, "in-memory://")
}

func (f *TransportFactory) Create(name string, _ string, _ []byte, _ api.Serializer) (api.Transport, error) {
	return NewTransport(name), nil
}
