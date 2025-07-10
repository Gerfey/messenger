package transport

type TransportLocator struct {
	transports map[string]Transport
}

func NewTransportLocator() *TransportLocator {
	return &TransportLocator{
		transports: make(map[string]Transport),
	}
}

func (r *TransportLocator) Register(name string, transport Transport) error {
	r.transports[name] = transport

	return nil
}

func (r *TransportLocator) GetAllTransports() []Transport {
	var all []Transport
	for _, t := range r.transports {
		all = append(all, t)
	}

	return all
}

func (r *TransportLocator) GetTransport(name string) Transport {
	return r.transports[name]
}
