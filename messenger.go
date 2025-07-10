package messenger

import (
	"context"
	"fmt"

	"github.com/gerfey/messenger/bus"
	"github.com/gerfey/messenger/transport"
)

type Messenger struct {
	defaultBus       *bus.Bus
	busLocator       *bus.BusLocator
	transportManager *transport.Manager
}

func NewMessenger(defaultBus *bus.Bus, manager *transport.Manager, busLocator *bus.BusLocator) *Messenger {
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

func (m *Messenger) GetBus() (*bus.Bus, error) {
	return m.defaultBus, nil
}

func (m *Messenger) GetMessageBus() (*bus.Bus, error) {
	messageBus, ok := m.busLocator.Get("message.bus")
	if !ok {
		return nil, fmt.Errorf("message bus not found")
	}

	return messageBus, nil
}

func (m *Messenger) GetCommandBus() (*bus.Bus, error) {
	commandBus, ok := m.busLocator.Get("command.bus")
	if !ok {
		return nil, fmt.Errorf("message bus not found")
	}

	return commandBus, nil
}

func (m *Messenger) GetQueryBus() (*bus.Bus, error) {
	queueBus, ok := m.busLocator.Get("query.bus")
	if !ok {
		return nil, fmt.Errorf("message bus not found")
	}

	return queueBus, nil
}
