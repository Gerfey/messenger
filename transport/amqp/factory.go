package amqp

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

func (f *TransportFactory) Supports(dsn string) bool {
	return strings.HasPrefix(dsn, "amqp://")
}

func (f *TransportFactory) Create(
	name string,
	dsn string,
	options []byte,
	serializer api.Serializer,
) (api.Transport, error) {
	var opts OptionsConfig
	if err := defaults.Set(&opts); err != nil {
		return nil, fmt.Errorf("set defaults: %w", err)
	}

	if err := yaml.Unmarshal(options, &opts); err != nil {
		return nil, fmt.Errorf("unmarshal options: %w", err)
	}

	cfg := TransportConfig{
		Name:    name,
		DSN:     dsn,
		Options: opts,
	}

	return NewTransport(cfg, serializer)
}
