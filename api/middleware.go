package api

import "context"

type Middleware interface {
	Handle(context.Context, Envelope, NextFunc) (Envelope, error)
}

type NextFunc func(ctx context.Context, env Envelope) (Envelope, error)

type MiddlewareLocator interface {
	Register(string, Middleware)
	GetAll() []Middleware
	Get(name string) (Middleware, error)
}
