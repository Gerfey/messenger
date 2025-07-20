package config

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	defaultProcessorsCount = 2
)

type MessengerConfig struct {
	DefaultBus       string                     `yaml:"default_bus"       default:"default"`
	FailureTransport string                     `yaml:"failure_transport"`
	Buses            map[string]BusConfig       `yaml:"buses"`
	Transports       map[string]TransportConfig `yaml:"transports"`
	Routing          map[string]string          `yaml:"routing"`
}

type BusConfig struct {
	Middleware []string `yaml:"middleware"`
}

type TransportConfig struct {
	DSN           string               `yaml:"dsn"`
	RetryStrategy *RetryStrategyConfig `yaml:"retry_strategy"`
	Options       OptionsConfig        `yaml:"options"`
}

type RetryStrategyConfig struct {
	MaxRetries uint          `yaml:"max_retries"`
	Delay      time.Duration `yaml:"delay"`
	Multiplier float64       `yaml:"multiplier"`
	MaxDelay   time.Duration `yaml:"max_delay"`
}

type OptionsConfig struct {
	AutoSetup        bool             `yaml:"auto_setup"         default:"true"`
	ConsumerPoolSize int              `yaml:"consumer_pool_size" default:"10"`
	Exchange         ExchangeConfig   `yaml:"exchange"`
	Queues           map[string]Queue `yaml:"queues"`
}

type ExchangeConfig struct {
	Name       string `yaml:"name"`
	Type       string `yaml:"type"        default:"topic"` // topic, direct, fanout
	Durable    bool   `yaml:"durable"     default:"true"`
	AutoDelete bool   `yaml:"auto_delete" default:"false"`
	Internal   bool   `yaml:"internal"    default:"false"`
}

type Queue struct {
	BindingKeys []string `yaml:"binding_keys"`
	Durable     bool     `yaml:"durable"      default:"true"`
	Exclusive   bool     `yaml:"exclusive"    default:"false"`
	AutoDelete  bool     `yaml:"auto_delete"  default:"false"`
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

func LoadConfig(path string, processors ...Processor) (*MessengerConfig, error) {
	var cfg MessengerConfig

	reader := &FileReader{}
	parser := &YAMLParser{}

	allProcessors := make([]Processor, 0, len(processors)+defaultProcessorsCount)
	allProcessors = append(allProcessors, &EnvVarProcessor{})
	allProcessors = append(allProcessors, processors...)

	content, err := reader.Read(path, allProcessors...)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", path, err)
	}

	if parseErr := parser.Parse(content, &cfg); parseErr != nil {
		return nil, fmt.Errorf("failed to parse config file '%s': %w", path, parseErr)
	}

	return &cfg, nil
}
