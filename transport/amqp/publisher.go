package amqp

import (
	"context"
	"fmt"

	"github.com/gerfey/messenger/api"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn       *Connection
	cfg        TransportConfig
	serializer api.Serializer
}

func NewPublisher(conn *Connection, cfg TransportConfig, serializer api.Serializer) *Publisher {
	return &Publisher{
		conn:       conn,
		cfg:        cfg,
		serializer: serializer,
	}
}

func (p *Publisher) Publish(ctx context.Context, env api.Envelope) error {
	body, headersMap, err := p.serializer.Marshal(env)
	if err != nil {
		return fmt.Errorf("failed to marshal envelope: %w", err)
	}

	headers := amqp.Table{}
	for k, v := range headersMap {
		headers[k] = v
	}

	ch, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to create AMQP channel for publisher: %w", err)
	}
	defer func() {
		_ = ch.Close()
	}()

	routingKey := getRoutingKey(env.Message())

	err = ch.PublishWithContext(ctx,
		p.cfg.Options.Exchange.Name,
		routingKey,
		false,
		false,
		amqp.Publishing{
			Headers:     headers,
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("failed to publish message to exchange '%s' with routing key '%s': %w", p.cfg.Options.Exchange.Name, routingKey, err)
	}

	return nil
}
