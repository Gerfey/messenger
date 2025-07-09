package transport

import "reflect"

type TypeResolver interface {
	ResolveMessageType(typeName string) (reflect.Type, error)
	ResolveStampType(typeName string) (reflect.Type, error)
}
