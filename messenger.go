package messenger

import (
	"context"
	"fmt"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/infrastructure/transport"
)

type Messenger struct {
	defaultBus       api.MessageBus
	busLocator       api.BusLocator
	transportManager *transport.Manager
}

func NewMessenger(defaultBus api.MessageBus, manager *transport.Manager, busLocator api.BusLocator) api.Messenger {
	return &Messenger{
		defaultBus:       defaultBus,
		busLocator:       busLocator,
		transportManager: manager,
	}
}

func (m *Messenger) Run(ctx context.Context) error {
	m.transportManager.Start(ctx)
	<-ctx.Done()
	m.transportManager.Stop()

	return ctx.Err()
}

func (m *Messenger) GetBus() (api.MessageBus, error) {
	return m.defaultBus, nil
}

func (m *Messenger) GetMessageBus() (api.MessageBus, error) {
	messageBus, ok := m.busLocator.Get("message.bus")
	if !ok {
		return nil, fmt.Errorf("message bus not found")
	}

	return messageBus, nil
}

func (m *Messenger) GetCommandBus() (api.MessageBus, error) {
	commandBus, ok := m.busLocator.Get("command.bus")
	if !ok {
		return nil, fmt.Errorf("message bus not found")
	}

	return commandBus, nil
}

func (m *Messenger) GetQueryBus() (api.MessageBus, error) {
	queueBus, ok := m.busLocator.Get("query.bus")
	if !ok {
		return nil, fmt.Errorf("message bus not found")
	}

	return queueBus, nil
}
