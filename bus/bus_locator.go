package bus

type BusLocator struct {
	busses map[string]*Bus
}

func NewBusLocator() *BusLocator {
	return &BusLocator{
		busses: make(map[string]*Bus),
	}
}

func (b *BusLocator) Register(name string, bus *Bus) error {
	b.busses[name] = bus

	return nil
}

func (b *BusLocator) GetAll() []*Bus {
	var all []*Bus
	for _, bus := range b.busses {
		all = append(all, bus)
	}

	return all
}

func (b *BusLocator) Get(name string) (*Bus, bool) {
	bus, ok := b.busses[name]
	if !ok {
		return nil, false
	}

	return bus, true
}
