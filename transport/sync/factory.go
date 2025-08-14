package sync

import (
	"strings"

	"github.com/gerfey/messenger/api"
)

type TransportFactory struct {
	locator api.BusLocator
}

func NewTransportFactory(locator api.BusLocator) api.TransportFactory {
	return &TransportFactory{
		locator: locator,
	}
}

func (f *TransportFactory) Supports(dsn string) bool {
	return strings.HasPrefix(dsn, "sync://")
}

func (f *TransportFactory) Create(_ string, _ string, _ []byte, _ api.Serializer) (api.Transport, error) {
	return NewTransport(f.locator), nil
}
