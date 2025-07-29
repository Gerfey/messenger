package kafka

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"

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

func (t *TransportFactory) Supports(dsn string) bool {
	return strings.HasPrefix(dsn, "kafka://")
}

func (t *TransportFactory) Create(name string, dsn string, options []byte) (api.Transport, error) {
	var optsConfig OptionsConfig
	if err := defaults.Set(&optsConfig); err != nil {
		return nil, fmt.Errorf("kafka: set defaults: %w", err)
	}

	if err := yaml.Unmarshal(options, &optsConfig); err != nil {
		return nil, fmt.Errorf("kafka: unmarshal options: %w", err)
	}

	tCfg := TransportConfig{
		Name:    name,
		DSN:     dsn,
		Options: optsConfig,
	}

	return NewTransport(tCfg, t.resolver, t.logger)
}
