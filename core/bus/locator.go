package bus

import "github.com/gerfey/messenger/api"

type Locator struct {
	busses map[string]api.MessageBus
}

func NewLocator() api.BusLocator {
	return &Locator{
		busses: make(map[string]api.MessageBus),
	}
}

func (b *Locator) Register(name string, bus api.MessageBus) error {
	b.busses[name] = bus

	return nil
}

func (b *Locator) GetAll() []api.MessageBus {
	all := make([]api.MessageBus, 0)
	for _, bus := range b.busses {
		all = append(all, bus)
	}

	return all
}

func (b *Locator) Get(name string) (api.MessageBus, bool) {
	bus, ok := b.busses[name]
	if !ok {
		return nil, false
	}

	return bus, true
}
