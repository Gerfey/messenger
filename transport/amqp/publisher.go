package amqp

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gerfey/messenger/envelope"
	"github.com/gerfey/messenger/message"
	"github.com/gerfey/messenger/transport"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn       *Connection
	cfg        TransportConfig
	serializer transport.Serializer
}

func NewPublisher(conn *Connection, cfg TransportConfig, serializer transport.Serializer) *Publisher {
	return &Publisher{
		conn:       conn,
		cfg:        cfg,
		serializer: serializer,
	}
}

func (p *Publisher) Publish(ctx context.Context, env *envelope.Envelope) error {
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
		return fmt.Errorf("failed to get channel: %w", err)
	}
	defer ch.Close()

	if p.cfg.Options.AutoSetup {
		err := ch.ExchangeDeclare(
			p.cfg.Options.Exchange.Name,
			p.cfg.Options.Exchange.Type,
			p.cfg.Options.Exchange.Durable,
			p.cfg.Options.Exchange.AutoDelete,
			p.cfg.Options.Exchange.Internal,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange: %w", err)
		}
	}

	routingKey := getRoutingKey(env.Message())

	return ch.PublishWithContext(ctx,
		p.cfg.Options.Exchange.Name,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Headers:     headers,
		})
}

func getRoutingKey(msg any) string {
	var routingKey string
	if rk, ok := msg.(message.RoutedMessage); ok {
		routingKey = rk.RoutingKey()
	} else {
		routingKey = reflect.TypeOf(msg).String()
	}

	return routingKey
}
