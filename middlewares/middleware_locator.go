package middlewares

import (
	"fmt"
)

type MiddlewareLocator struct {
	middlewares map[string]Middleware
}

func NewMiddlewareLocator() *MiddlewareLocator {
	return &MiddlewareLocator{
		middlewares: make(map[string]Middleware),
	}
}

func (m *MiddlewareLocator) Register(name string, middleware Middleware) {
	m.middlewares[name] = middleware
}

func (m *MiddlewareLocator) GetAll() []Middleware {
	var all []Middleware
	for _, middleware := range m.middlewares {
		all = append(all, middleware)
	}

	return all
}

func (m *MiddlewareLocator) Get(name string) (Middleware, error) {
	mw, ok := m.middlewares[name]
	if !ok {
		return nil, fmt.Errorf("no middleware with name %s found", name)
	}

	return mw, nil
}
