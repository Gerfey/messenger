package event

import (
	"context"

	"github.com/gerfey/messenger/api"
)

type SendMessageToTransportsEvent struct {
	Ctx            context.Context
	Envelope       api.Envelope
	TransportNames []string
}
