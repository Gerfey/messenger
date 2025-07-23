package amqp

import (
	"log/slog"
	"strings"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/config"
)

type TransportFactory struct {
	resolver api.TypeResolver
	logger   *slog.Logger
}

func NewTransportFactory(resolver api.TypeResolver, logger *slog.Logger) api.TransportFactory {
	return &TransportFactory{
		resolver: resolver,
		logger:   logger,
	}
}

func (f *TransportFactory) Supports(dsn string) bool {
	return strings.HasPrefix(dsn, "amqp://")
}

func (f *TransportFactory) Create(name string, dsn string, options config.OptionsConfig) (api.Transport, error) {
	cfg := TransportConfig{
		Name:    name,
		DSN:     dsn,
		Options: options,
	}

	return NewTransport(cfg, f.resolver, f.logger)
}
