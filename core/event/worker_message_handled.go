package event

import (
	"context"

	"github.com/gerfey/messenger/api"
)

type WorkerMessageHandledEvent struct {
	Ctx           context.Context
	Envelope      api.Envelope
	TransportName string
}
