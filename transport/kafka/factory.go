package kafka

import (
	"fmt"
	"strings"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"

	"github.com/gerfey/messenger/api"
)

type TransportFactory struct{}

func NewTransportFactory() api.TransportFactory {
	return &TransportFactory{}
}

func (t *TransportFactory) Supports(dsn string) bool {
	return strings.HasPrefix(dsn, "kafka://")
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

	return NewTransport(tCfg, ser)
}
