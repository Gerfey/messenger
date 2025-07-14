package amqp

import (
	"context"
	"fmt"
	"reflect"

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
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer func() {
		_ = ch.Close()
	}()

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
	if rk, ok := msg.(api.RoutedMessage); ok {
		routingKey = rk.RoutingKey()
	} else {
		routingKey = reflect.TypeOf(msg).String()
	}

	return routingKey
}
