package factory

import (
	"github.com/gerfey/messenger/config"
	"github.com/gerfey/messenger/transport"
)

type TransportFactory interface {
	Supports(dsn string) bool
	Create(name string, dsn string, options config.OptionsConfig) (transport.Transport, error)
}
