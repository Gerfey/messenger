package middleware

import (
	"fmt"

	"github.com/gerfey/messenger/api"
)

type Locator struct {
	middlewares map[string]api.Middleware
}

func NewMiddlewareLocator() api.MiddlewareLocator {
	return &Locator{
		middlewares: make(map[string]api.Middleware),
	}
}

func (m *Locator) Register(name string, middleware api.Middleware) {
	m.middlewares[name] = middleware
}

func (m *Locator) GetAll() []api.Middleware {
	all := make([]api.Middleware, 0)
	for _, middleware := range m.middlewares {
		all = append(all, middleware)
	}

	return all
}

func (m *Locator) Get(name string) (api.Middleware, error) {
	mw, ok := m.middlewares[name]
	if !ok {
		return nil, fmt.Errorf("no middleware with name %s found", name)
	}

	return mw, nil
}
