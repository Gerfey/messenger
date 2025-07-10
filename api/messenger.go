package api

import "context"

type Messenger interface {
	Run(context.Context) error
	GetBus() (MessageBus, error)
	GetMessageBus() (MessageBus, error)
	GetCommandBus() (MessageBus, error)
	GetQueryBus() (MessageBus, error)
}
