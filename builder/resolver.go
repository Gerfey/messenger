package builder

import (
	"fmt"
	"reflect"

	"github.com/gerfey/messenger/api"
)

type Resolver struct {
	messageTypes map[string]reflect.Type
	stampTypes   map[string]reflect.Type
}

func NewStaticTypeResolver() api.TypeResolver {
	return &Resolver{
		messageTypes: make(map[string]reflect.Type),
		stampTypes:   make(map[string]reflect.Type),
	}
}

func (r *Resolver) Register(typeStr string, t reflect.Type) {
	r.messageTypes[typeStr] = t
}

func (r *Resolver) RegisterMessage(message any) {
	t := reflect.TypeOf(message)
	r.messageTypes[t.String()] = t
}

func (r *Resolver) RegisterStamp(stamp any) {
	t := reflect.TypeOf(stamp)
	r.stampTypes[t.String()] = t
}

func (r *Resolver) ResolveMessageType(name string) (reflect.Type, error) {
	t, ok := r.messageTypes[name]
	if !ok {
		return nil, fmt.Errorf("unknown message type: %s", name)
	}

	return t, nil
}

func (r *Resolver) ResolveStampType(name string) (reflect.Type, error) {
	t, ok := r.stampTypes[name]
	if !ok {
		return nil, fmt.Errorf("unknown stamp type: %s", name)
	}

	return t, nil
}
