package kafka

import (
	"fmt"
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

func (t *TransportFactory) Supports(dsn string) bool {
	return strings.HasPrefix(dsn, "kafka://")
}

func (t *TransportFactory) Create(name string, dsn string, options config.OptionsConfig) (api.Transport, error) {
	cfg, err := NewConfig(name, dsn, options)
	if err != nil {
		return nil, fmt.Errorf("create config kafka: %w", err)
	}

	transport, err := NewTransport(cfg, t.resolver, t.logger)
	if err != nil {
		return nil, fmt.Errorf("create kafka transport: %w", err)
	}

	return transport, nil
}
