package api

import "reflect"

type TypeResolver interface {
	Register(string, reflect.Type)
	RegisterMessage(any)
	RegisterStamp(any)
	ResolveMessageType(string) (reflect.Type, error)
	ResolveStampType(string) (reflect.Type, error)
}
