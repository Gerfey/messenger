package api

import (
	"reflect"
)

type HandlerLocator interface {
	Register(any) error
	GetAll() []HandlerFunc
	Get(any) []HandlerFunc
	ResolveMessageType(string) (reflect.Type, error)
}

type HandlerFunc struct {
	Fn         reflect.Value
	InputType  reflect.Type
	HandlerStr string
	BusName    string
}

type MessageHandlerType interface {
	GetBusName() string
}
