package routing

import (
	"reflect"

	"github.com/gerfey/messenger/api"
)

type Router struct {
	routes map[reflect.Type][]string
}

func NewRouter() api.Router {
	return &Router{
		routes: make(map[reflect.Type][]string),
	}
}

func (r *Router) RouteMessageTo(msg any, transports ...string) {
	t := reflect.TypeOf(msg)
	r.routes[t] = transports
}

func (r *Router) GetTransportFor(msg any) []string {
	t := reflect.TypeOf(msg)

	return r.routes[t]
}

func (r *Router) RouteTypeTo(t reflect.Type, transports ...string) {
	r.routes[t] = transports
}
