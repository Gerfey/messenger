package api

import (
	"context"
)

type EventDispatcher interface {
	Dispatch(context.Context, any) error
	AddListener(any, any)
}
