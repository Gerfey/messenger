package inmemory

import (
	"log/slog"
	"strings"

	"github.com/gerfey/messenger/api"
)

type TransportFactory struct {
	logger   *slog.Logger
	resolver api.TypeResolver
}

func NewTransportFactory(logger *slog.Logger, resolver api.TypeResolver) api.TransportFactory {
	return &TransportFactory{
		logger:   logger,
		resolver: resolver,
	}
}

func (f *TransportFactory) Supports(dsn string) bool {
	return strings.HasPrefix(dsn, "in-memory://")
}

func (f *TransportFactory) Create(name string, _ string, _ []byte) (api.Transport, error) {
	return NewTransport(name), nil
}
