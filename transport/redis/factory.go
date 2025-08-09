package redis

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"

	"github.com/gerfey/messenger/api"
)

type TransportFactory struct {
	logger *slog.Logger
}

func NewTransportFactory(logger *slog.Logger) api.TransportFactory {
	return &TransportFactory{
		logger: logger,
	}
}

func (t *TransportFactory) Supports(dsn string) bool {
	return strings.HasPrefix(dsn, "redis://")
}

func (t *TransportFactory) Create(name string, dsn string, options []byte, ser api.Serializer) (api.Transport, error) {
	var optsConfig OptionsConfig
	if err := defaults.Set(&optsConfig); err != nil {
		return nil, fmt.Errorf("set defaults: %w", err)
	}

	if err := yaml.Unmarshal(options, &optsConfig); err != nil {
		return nil, fmt.Errorf("unmarshal options: %w", err)
	}

	tCfg := TransportConfig{
		Name:    name,
		DSN:     dsn,
		Options: optsConfig,
	}

	return NewTransport(tCfg, t.logger, ser)
}
