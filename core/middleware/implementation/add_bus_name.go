package implementation

import (
	"context"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/stamps"
)

type AddBusNameMiddleware struct {
	busName string
}

func NewAddBusNameMiddleware(busName string) api.Middleware {
	return &AddBusNameMiddleware{
		busName: busName,
	}
}

func (h *AddBusNameMiddleware) Handle(ctx context.Context, env api.Envelope, next api.NextFunc) (api.Envelope, error) {
	if _, ok := envelope.LastStampOf[stamps.BusNameStamp](env); !ok {
		env = env.WithStamp(stamps.BusNameStamp{Name: h.busName})
	}

	return next(ctx, env)
}
