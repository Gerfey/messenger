package messenger

import (
	"context"
	"fmt"

	"github.com/gerfey/messenger/bus"
	"github.com/gerfey/messenger/transport"
)

type Messenger struct {
	defaultBus       *bus.Bus
	buses            map[string]*bus.Bus
	transportManager *transport.Manager
	transports       map[string]transport.Transport
}

func NewMessenger(defaultBus *bus.Bus, manager *transport.Manager, transports map[string]transport.Transport, buses map[string]*bus.Bus) *Messenger {
	return &Messenger{
		defaultBus:       defaultBus,
		buses:            buses,
		transportManager: manager,
		transports:       transports,
	}
}

func (m *Messenger) Run(ctx context.Context) error {
	m.transportManager.Start(ctx)
	<-ctx.Done()
	m.transportManager.Stop()

	return ctx.Err()
}

func (m *Messenger) GetMessageBus() (*bus.Bus, error) {
	messageBus, ok := m.getBus("message.bus")
	if !ok {
		return nil, fmt.Errorf("message bus not found")
	}

	return messageBus, nil
}

func (m *Messenger) GetCommandBus() (*bus.Bus, error) {
	commandBus, ok := m.getBus("command.bus")
	if !ok {
		return nil, fmt.Errorf("message bus not found")
	}

	return commandBus, nil
}

func (m *Messenger) GetQueryBus() (*bus.Bus, error) {
	queueBus, ok := m.getBus("query.bus")
	if !ok {
		return nil, fmt.Errorf("message bus not found")
	}

	return queueBus, nil
}

func (m *Messenger) getBus(name string) (*bus.Bus, bool) {
	b, ok := m.buses[name]

	return b, ok
}

func (m *Messenger) GetDefaultBus() *bus.Bus {
	return m.defaultBus
}

func (m *Messenger) GetTransport(name string) (transport.Transport, error) {
	t, ok := m.transports[name]
	if !ok {
		return nil, fmt.Errorf("transport %q not found", name)
	}

	return t, nil
}
