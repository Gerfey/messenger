package transport

import "github.com/gerfey/messenger/api"

type Locator struct {
	transports map[string]api.Transport
}

func NewLocator() api.TransportLocator {
	return &Locator{
		transports: make(map[string]api.Transport),
	}
}

func (r *Locator) Register(name string, transport api.Transport) error {
	r.transports[name] = transport

	return nil
}

func (r *Locator) GetAllTransports() []api.Transport {
	var all []api.Transport
	for _, t := range r.transports {
		all = append(all, t)
	}

	return all
}

func (r *Locator) GetTransport(name string) api.Transport {
	return r.transports[name]
}
