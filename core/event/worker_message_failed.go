package event

import (
	"context"

	"github.com/gerfey/messenger/api"
)

type WorkerMessageFailedEvent struct {
	Ctx           context.Context
	Envelope      api.Envelope
	TransportName string
	Error         error
}
