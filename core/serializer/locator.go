package serializer

import (
	"fmt"

	"github.com/gerfey/messenger/api"
)

type Locator struct {
	serializers map[string]api.Serializer
}

func NewSerializerLocator() api.SerializerLocator {
	return &Locator{
		serializers: make(map[string]api.Serializer),
	}
}

func (m *Locator) Register(name string, serializer api.Serializer) {
	m.serializers[name] = serializer
}

func (m *Locator) GetAll() []api.Serializer {
	all := make([]api.Serializer, 0)
	for _, serializer := range m.serializers {
		all = append(all, serializer)
	}

	return all
}

func (m *Locator) Get(name string) (api.Serializer, error) {
	mw, ok := m.serializers[name]
	if !ok {
		return nil, fmt.Errorf("no serializer with name %s found", name)
	}

	return mw, nil
}
