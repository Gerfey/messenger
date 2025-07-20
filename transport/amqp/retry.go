package amqp

import (
	"context"

	"github.com/gerfey/messenger/api"
)

type Retry struct {
	*BasePublisher
}

func NewRetry(conn *Connection, cfg TransportConfig, serializer api.Serializer) *Retry {
	return &Retry{
		BasePublisher: NewBasePublisher(conn, cfg, serializer),
	}
}

func (r *Retry) Retry(ctx context.Context, env api.Envelope) error {
	return r.PublishMessage(ctx, env)
}
