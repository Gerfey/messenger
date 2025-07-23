package amqp

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/gerfey/messenger/api"
)

type BasePublisher struct {
	conn       *Connection
	cfg        TransportConfig
	serializer api.Serializer
}

func NewBasePublisher(conn *Connection, cfg TransportConfig, serializer api.Serializer) *BasePublisher {
	return &BasePublisher{
		conn:       conn,
		cfg:        cfg,
		serializer: serializer,
	}
}

func (bp *BasePublisher) PublishMessage(ctx context.Context, env api.Envelope) error {
	body, headersMap, err := bp.serializer.Marshal(env)
	if err != nil {
		return fmt.Errorf("failed to marshal envelope: %w", err)
	}

	headers := amqp.Table{}
	for k, v := range headersMap {
		headers[k] = v
	}

	ch, err := bp.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to create AMQP channel: %w", err)
	}
	defer func() {
		_ = ch.Close()
	}()

	routingKey := getRoutingKey(env.Message())

	err = ch.PublishWithContext(ctx,
		bp.cfg.Options.Exchange.Name,
		routingKey,
		false,
		false,
		amqp.Publishing{
			Headers:     headers,
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf(
			"failed to publish message to exchange '%s' with routing key '%s': %w",
			bp.cfg.Options.Exchange.Name,
			routingKey,
			err,
		)
	}

	return nil
}
