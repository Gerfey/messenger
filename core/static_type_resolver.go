package core

import (
	"fmt"
	"reflect"
)

type StaticTypeResolver struct {
	messageTypes map[string]reflect.Type
	stampTypes   map[string]reflect.Type
}

func NewStaticTypeResolver() *StaticTypeResolver {
	return &StaticTypeResolver{
		messageTypes: make(map[string]reflect.Type),
		stampTypes:   make(map[string]reflect.Type),
	}
}

func (r *StaticTypeResolver) Register(typeStr string, t reflect.Type) {
	r.messageTypes[typeStr] = t
}

func (r *StaticTypeResolver) RegisterMessage(message any) {
	t := reflect.TypeOf(message)
	r.messageTypes[t.String()] = t
}

func (r *StaticTypeResolver) RegisterStamp(stamp any) {
	t := reflect.TypeOf(stamp)
	r.stampTypes[t.String()] = t
}

func (r *StaticTypeResolver) ResolveMessageType(name string) (reflect.Type, error) {
	t, ok := r.messageTypes[name]
	if !ok {
		return nil, fmt.Errorf("unknown message type: %s", name)
	}
	return t, nil
}

func (r *StaticTypeResolver) ResolveStampType(name string) (reflect.Type, error) {
	t, ok := r.stampTypes[name]
	if !ok {
		return nil, fmt.Errorf("unknown stamp type: %s", name)
	}
	return t, nil
}
