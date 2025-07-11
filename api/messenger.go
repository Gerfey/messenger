package api

import "context"

type Messenger interface {
	Run(context.Context) error
	GetBus() (MessageBus, error)
	GetBusWith(string) (MessageBus, error)
}
