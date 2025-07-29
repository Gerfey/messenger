package sync

import (
	"log/slog"
	"strings"

	"github.com/gerfey/messenger/api"
)

type Factory struct {
	logger  *slog.Logger
	locator api.BusLocator
}

func NewTransportFactory(logger *slog.Logger, locator api.BusLocator) api.TransportFactory {
	return &Factory{
		logger:  logger,
		locator: locator,
	}
}

func (f *Factory) Supports(dsn string) bool {
	return strings.HasPrefix(dsn, "sync://")
}

func (f *Factory) Create(_ string, _ string, _ []byte) (api.Transport, error) {
	return NewTransport(f.locator), nil
}
