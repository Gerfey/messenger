package amqp

import (
	"context"

	"github.com/gerfey/messenger/api"
)

type Publisher struct {
	*BasePublisher
}

func NewPublisher(conn *Connection, cfg TransportConfig, serializer api.Serializer) *Publisher {
	return &Publisher{
		BasePublisher: NewBasePublisher(conn, cfg, serializer),
	}
}

func (p *Publisher) Publish(ctx context.Context, env api.Envelope) error {
	return p.PublishMessage(ctx, env)
}
