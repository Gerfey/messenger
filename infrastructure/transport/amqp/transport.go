package amqp

import (
	"context"

	"github.com/gerfey/messenger/api"
)

type Transport struct {
	cfg        TransportConfig
	publisher  *Publisher
	consumer   *Consumer
	serializer api.Serializer
}

func NewTransport(cfg TransportConfig, resolver api.TypeResolver) (api.Transport, error) {
	conn, err := NewConnection(cfg.DSN)
	if err != nil {
		return nil, err
	}

	serializer := NewSerializer(resolver)

	pub := NewPublisher(conn, cfg, serializer)
	cons := NewConsumer(conn, cfg, serializer)

	return &Transport{
		cfg:        cfg,
		publisher:  pub,
		consumer:   cons,
		serializer: serializer,
	}, nil
}

func (t *Transport) Send(ctx context.Context, env api.Envelope) error {
	return t.publisher.Publish(ctx, env)
}

func (t *Transport) Receive(ctx context.Context, handler func(context.Context, api.Envelope) error) error {
	return t.consumer.Consume(ctx, handler)
}
