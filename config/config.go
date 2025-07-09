package config

import (
	"encoding/json"
)

type MessengerConfig struct {
	DefaultBus string                     `yaml:"default_bus" default:"default"`
	Buses      map[string]BusConfig       `yaml:"buses"`
	Transports map[string]TransportConfig `yaml:"transports"`
	Routing    map[string]string          `yaml:"routing"`
}

type BusConfig struct {
	Middleware []string `yaml:"middleware"`
}

type TransportConfig struct {
	DSN     string        `yaml:"dsn"`
	Options OptionsConfig `yaml:"options"`
}

type OptionsConfig struct {
	AutoSetup bool             `yaml:"auto_setup" default:"true"`
	Exchange  ExchangeConfig   `yaml:"exchange"`
	Queues    map[string]Queue `yaml:"queues"`
}

type ExchangeConfig struct {
	Name       string `yaml:"name"`
	Type       string `yaml:"type" default:"topic"` // topic, direct, fanout
	Durable    bool   `yaml:"durable" default:"true"`
	AutoDelete bool   `yaml:"auto_delete" default:"false"`
	Internal   bool   `yaml:"internal" default:"false"`
}

type Queue struct {
	BindingKeys []string `yaml:"binding_keys"`
	Durable     bool     `yaml:"durable" default:"true"`
	Exclusive   bool     `yaml:"exclusive" default:"false"`
	AutoDelete  bool     `yaml:"auto_delete" default:"false"`
}

type SerializedEnvelope struct {
	Message     any               `json:"message"`
	MessageType string            `json:"type"`
	Stamps      []SerializedStamp `json:"stamps"`
}

type SerializedStamp struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func LoadConfig(path string, processors ...ConfigProcessor) (*MessengerConfig, error) {
	var cfg MessengerConfig

	reader := &FileConfigReader{}
	parser := &YAMLParser{}

	allProcessors := make([]ConfigProcessor, 0, len(processors)+2)
	allProcessors = append(allProcessors, &EnvVarProcessor{})
	allProcessors = append(allProcessors, processors...)

	content, err := reader.Read(path, allProcessors...)
	if err != nil {
		return nil, err
	}

	if err := parser.Parse(content, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
