package amqp

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/gerfey/messenger/api"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=interfaces.go -destination=../../tests/mocks/mock_amqp.go -package=mocks

type ConnectionAMQP interface {
	Channel() (ChannelAMQP, error)
	IsClosed() bool
	Close() error
}

type ChannelAMQP interface {
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Consume(
		queue, consumer string,
		autoAck, exclusive, noLocal, noWait bool,
		args amqp.Table,
	) (<-chan amqp.Delivery, error)
	Close() error
}

type PublisherAMQP interface {
	Publish(ctx context.Context, env api.Envelope) error
}

type ConsumerAMQP interface {
	Consume(ctx context.Context, handler func(context.Context, api.Envelope) error) error
}

type RetryAMQP interface {
	Retry(ctx context.Context, env api.Envelope) error
}
