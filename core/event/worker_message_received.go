package event

import (
	"context"

	"github.com/gerfey/messenger/api"
)

type WorkerMessageReceivedEvent struct {
	Ctx           context.Context
	Envelope      api.Envelope
	TransportName string
}
