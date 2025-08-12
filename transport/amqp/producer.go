package amqp

import (
	"context"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/gerfey/messenger/api"
)

type Producer struct {
	config     TransportConfig
	connection ConnectionAMQP
	serializer api.Serializer
	lock       sync.Mutex
}

func NewProducer(config TransportConfig, connection ConnectionAMQP, serializer api.Serializer) (api.Producer, error) {
	return &Producer{
		config:     config,
		connection: connection,
		serializer: serializer,
	}, nil
}

func (p *Producer) Send(ctx context.Context, env api.Envelope) error {
	if !p.connection.IsConnect() {
		return fmt.Errorf("amqp connection is not available")
	}

	body, headersMap, err := p.serializer.Marshal(env)
	if err != nil {
		return fmt.Errorf("failed to marshal envelope: %w", err)
	}

	headers := amqp.Table{}
	for k, v := range headersMap {
		headers[k] = v
	}

	ch, err := p.connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to create AMQP channel: %w", err)
	}
	defer func() {
		_ = ch.Close()
	}()

	routingKey := getRoutingKey(env.Message())

	err = ch.PublishWithContext(ctx,
		p.config.Options.Exchange.Name,
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
			p.config.Options.Exchange.Name,
			routingKey,
			err,
		)
	}

	return nil
}
