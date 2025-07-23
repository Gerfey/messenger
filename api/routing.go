package api

import "reflect"

type Router interface {
	RouteMessageTo(any, ...string)
	RouteTypeTo(reflect.Type, ...string)
	GetTransportFor(any) []string
	GetUsedTransports() []string
}
