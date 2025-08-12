package amqp

import (
	"context"
	"fmt"
	"reflect"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/gerfey/messenger/api"
)

type ConnectionAMQP interface {
	api.Connection
	Channel() (*amqp.Channel, error)
}

type Transport struct {
	config     TransportConfig
	producer   api.Producer
	consumer   api.Consumer
	connection ConnectionAMQP
	serializer api.Serializer
}

func NewTransport(
	config TransportConfig,
	serializer api.Serializer,
) (api.Transport, error) {
	connection, errConnection := NewConnection(config.DSN)
	if errConnection != nil {
		return nil, fmt.Errorf("failed to create connection: %w", errConnection)
	}

	producer, errProducer := NewProducer(config, connection, serializer)
	if errProducer != nil {
		return nil, fmt.Errorf("failed to create producer: %w", errProducer)
	}

	consumer, errConsumer := NewConsumer(config, connection, serializer)
	if errConsumer != nil {
		return nil, fmt.Errorf("failed to create producer: %w", errConsumer)
	}

	return &Transport{
		config:     config,
		producer:   producer,
		consumer:   consumer,
		connection: connection,
		serializer: serializer,
	}, nil
}

func (t *Transport) Name() string {
	return t.config.Name
}

func (t *Transport) Send(ctx context.Context, env api.Envelope) error {
	return t.producer.Send(ctx, env)
}

func (t *Transport) Receive(ctx context.Context, handler func(context.Context, api.Envelope) error) error {
	return t.consumer.Consume(ctx, handler)
}

func (t *Transport) Retry(ctx context.Context, env api.Envelope) error {
	return t.producer.Send(ctx, env)
}

func (t *Transport) Setup(_ context.Context) error {
	if !t.config.Options.AutoSetup {
		return nil
	}

	ch, err := t.connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer func() {
		_ = ch.Close()
	}()

	err = ch.ExchangeDeclare(
		t.config.Options.Exchange.Name,
		t.config.Options.Exchange.Type,
		t.config.Options.Exchange.Durable,
		t.config.Options.Exchange.AutoDelete,
		t.config.Options.Exchange.Internal,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	for queueName, queueCfg := range t.config.Options.Queues {
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

		bindingKeys := queueCfg.BindingKeys

		if len(bindingKeys) == 0 {
			bindingKeys = []string{""}
		}

		for _, bindingKey := range bindingKeys {
			bindErr := ch.QueueBind(
				queueName,
				bindingKey,
				t.config.Options.Exchange.Name,
				false,
				nil,
			)
			if bindErr != nil {
				return fmt.Errorf("bind queue: %w", bindErr)
			}
		}
	}

	return nil
}

func (t *Transport) Close() error {
	if err := t.connection.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	return nil
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
