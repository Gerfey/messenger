package amqp

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/gerfey/messenger/api"
)

type ConnectionAdapter struct {
	conn *amqp.Connection
}

func NewConnectionAdapter(conn *amqp.Connection) ConnectionAMQP {
	return &ConnectionAdapter{conn: conn}
}

func (c *ConnectionAdapter) Channel() (ChannelAMQP, error) {
	ch, err := c.conn.Channel()
	if err != nil {
		return nil, err
	}

	return NewChannelAdapter(ch), nil
}

func (c *ConnectionAdapter) IsClosed() bool {
	return c.conn.IsClosed()
}

func (c *ConnectionAdapter) Close() error {
	return c.conn.Close()
}

type ChannelAdapter struct {
	ch *amqp.Channel
}

func NewChannelAdapter(ch *amqp.Channel) ChannelAMQP {
	return &ChannelAdapter{ch: ch}
}

func (c *ChannelAdapter) ExchangeDeclare(
	name, kind string,
	durable, autoDelete, internal, noWait bool,
	args amqp.Table,
) error {
	return c.ch.ExchangeDeclare(
		name, kind,
		durable, autoDelete, internal, noWait,
		args,
	)
}

func (c *ChannelAdapter) QueueDeclare(
	name string,
	durable, autoDelete, exclusive, noWait bool,
	args amqp.Table,
) (amqp.Queue, error) {
	return c.ch.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
}

func (c *ChannelAdapter) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	return c.ch.QueueBind(name, key, exchange, noWait, args)
}

func (c *ChannelAdapter) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return c.ch.Publish(exchange, key, mandatory, immediate, msg)
}

func (c *ChannelAdapter) Consume(
	queue, consumer string,
	autoAck, exclusive, noLocal, noWait bool,
	args amqp.Table,
) (<-chan amqp.Delivery, error) {
	return c.ch.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
}

func (c *ChannelAdapter) Close() error {
	return c.ch.Close()
}

type PublisherAdapter struct {
	publisher *Publisher
}

func NewPublisherAdapter(publisher *Publisher) PublisherAMQP {
	return &PublisherAdapter{publisher: publisher}
}

func (p *PublisherAdapter) Publish(ctx context.Context, env api.Envelope) error {
	return p.publisher.Publish(ctx, env)
}

type ConsumerAdapter struct {
	consumer *Consumer
}

func NewConsumerAdapter(consumer *Consumer) ConsumerAMQP {
	return &ConsumerAdapter{consumer: consumer}
}

func (c *ConsumerAdapter) Consume(ctx context.Context, handler func(context.Context, api.Envelope) error) error {
	return c.consumer.Consume(ctx, handler)
}

type RetryAdapter struct {
	retry *Retry
}

func NewRetryAdapter(retry *Retry) RetryAMQP {
	return &RetryAdapter{retry: retry}
}

func (r *RetryAdapter) Retry(ctx context.Context, env api.Envelope) error {
	return r.retry.Retry(ctx, env)
}
