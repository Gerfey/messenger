package amqp

import (
	"github.com/gerfey/messenger/config"
)

type TransportConfig struct {
	Name    string
	DSN     string
	Options config.OptionsConfig
}
