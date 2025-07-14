package amqp

import (
	"context"
	"fmt"

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

	transport := &Transport{
		cfg:        cfg,
		serializer: serializer,
		publisher:  pub,
		consumer:   cons,
	}

	if cfg.Options.AutoSetup {
		err := transport.setup(conn)
		if err != nil {
			return nil, err
		}
	}

	return transport, nil
}

func (t *Transport) Send(ctx context.Context, env api.Envelope) error {
	return t.publisher.Publish(ctx, env)
}

func (t *Transport) Receive(ctx context.Context, handler func(context.Context, api.Envelope) error) error {
	return t.consumer.Consume(ctx, handler)
}

func (t *Transport) setup(conn *Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		t.cfg.Options.Exchange.Name,
		t.cfg.Options.Exchange.Type,
		t.cfg.Options.Exchange.Durable,
		t.cfg.Options.Exchange.AutoDelete,
		t.cfg.Options.Exchange.Internal,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	for queueName, queueCfg := range t.cfg.Options.Queues {
		_, err = ch.QueueDeclare(
			queueName,
			queueCfg.Durable,
			queueCfg.AutoDelete,
			queueCfg.Exclusive,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("declare queue: %w", err)
		}

		for _, bindingKey := range queueCfg.BindingKeys {
			err := ch.QueueBind(
				queueName,
				bindingKey,
				t.cfg.Options.Exchange.Name,
				false,
				nil,
			)
			if err != nil {
				return fmt.Errorf("bind queue: %w", err)
			}
		}
	}

	return nil
}
