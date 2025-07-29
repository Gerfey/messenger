package amqp

import (
	"time"
)

type TransportConfig struct {
	Name    string
	DSN     string
	Options OptionsConfig
}

type OptionsConfig struct {
	AutoSetup        bool             `yaml:"auto_setup"         default:"false"`
	ConsumerPoolSize int              `yaml:"consumer_pool_size" default:"10"`
	CommitInterval   time.Duration    `yaml:"commit_interval"    default:"10"`
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
