package amqp

import (
	"context"

	"github.com/gerfey/messenger/envelope"
	"github.com/gerfey/messenger/transport"
)

type AMQPTransport struct {
	cfg        TransportConfig
	publisher  *Publisher
	consumer   *Consumer
	serializer transport.Serializer
}

func New(cfg TransportConfig, resolver transport.TypeResolver) (*AMQPTransport, error) {
	conn, err := NewConnection(cfg.DSN)
	if err != nil {
		return nil, err
	}

	serializer := NewSerializer(resolver)

	pub := NewPublisher(conn, cfg, serializer)
	cons := NewConsumer(conn, cfg, serializer)

	return &AMQPTransport{
		cfg:        cfg,
		publisher:  pub,
		consumer:   cons,
		serializer: serializer,
	}, nil
}

func (t *AMQPTransport) Send(ctx context.Context, env *envelope.Envelope) error {
	return t.publisher.Publish(ctx, env)
}

func (t *AMQPTransport) Receive(ctx context.Context, handler func(context.Context, *envelope.Envelope) error) error {
	return t.consumer.Consume(ctx, handler)
}
