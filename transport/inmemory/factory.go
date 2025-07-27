package inmemory

import (
	"log/slog"
	"strings"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/config"
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

func (f *TransportFactory) Create(name string, dsn string, options config.OptionsConfig) (api.Transport, error) {
	cfg := TransportConfig{
		Name:    name,
		DSN:     dsn,
		Options: options,
	}

	return NewTransport(cfg), nil
}
