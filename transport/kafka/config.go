package kafka

import (
	"time"
)

type TransportConfig struct {
	Name    string
	DSN     string
	Options OptionsConfig
}

type OptionsConfig struct {
	ConsumerPoolSize int           `yaml:"consumer_pool_size" default:"10"`
	CommitInterval   time.Duration `yaml:"commit_interval"`
	Offset           string        `yaml:"offset"             default:"latest"`
	Group            string        `yaml:"group"              default:"group"`
	Topic            string        `yaml:"topic"              default:"topic"`
	Brokers          []string
}
