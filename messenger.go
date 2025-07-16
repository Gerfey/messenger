package messenger

import (
	"context"
	"fmt"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/transport"
)

type Messenger struct {
	defaultBusName   string
	busLocator       api.BusLocator
	transportManager *transport.Manager
	routing          api.Router
}

func NewMessenger(defaultBusName string, manager *transport.Manager, busLocator api.BusLocator, routing api.Router) api.Messenger {
	return &Messenger{
		defaultBusName:   defaultBusName,
		busLocator:       busLocator,
		transportManager: manager,
		routing:          routing,
	}
}

func (m *Messenger) Run(ctx context.Context) error {
	usedTransports := m.routing.GetUsedTransports()
	m.transportManager.Start(ctx, usedTransports)

	<-ctx.Done()
	m.transportManager.Stop()

	return ctx.Err()
}

func (m *Messenger) GetDefaultBus() (api.MessageBus, error) {
	bus, ok := m.busLocator.Get(m.defaultBusName)
	if !ok {
		return nil, fmt.Errorf("bus not found")
	}

	return bus, nil
}

func (m *Messenger) GetBusWith(name string) (api.MessageBus, error) {
	bus, ok := m.busLocator.Get(name)
	if !ok {
		return nil, fmt.Errorf("bus not found")
	}

	return bus, nil
}
