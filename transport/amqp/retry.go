package amqp

import (
	"context"
	"fmt"

	"github.com/gerfey/messenger/api"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Retry struct {
	conn       *Connection
	cfg        TransportConfig
	serializer api.Serializer
}

func NewRetry(conn *Connection, cfg TransportConfig, serializer api.Serializer) *Retry {
	return &Retry{
		conn:       conn,
		cfg:        cfg,
		serializer: serializer,
	}
}

func (r *Retry) Retry(ctx context.Context, env api.Envelope) error {
	body, headersMap, err := r.serializer.Marshal(env)
	if err != nil {
		return fmt.Errorf("failed to marshal envelope: %w", err)
	}

	headers := amqp.Table{}
	for k, v := range headersMap {
		headers[k] = v
	}

	ch, err := r.conn.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get AMQP channel for retry: %w", err)
	}
	defer r.conn.PutChannel(ch)

	routingKey := getRoutingKey(env.Message())

	err = ch.PublishWithContext(ctx,
		r.cfg.Options.Exchange.Name,
		routingKey,
		false,
		false,
		amqp.Publishing{
			Headers:     headers,
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("failed to retry message to exchange '%s' with routing key '%s': %w", r.cfg.Options.Exchange.Name, routingKey, err)
	}

	return nil
}
