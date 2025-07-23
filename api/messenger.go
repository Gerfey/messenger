package api

import "context"

type Messenger interface {
	Run(context.Context) error
	GetDefaultBus() (MessageBus, error)
	GetBusWith(string) (MessageBus, error)
}
