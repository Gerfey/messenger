package amqp

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/stamps"
)

type Consumer struct {
	conn       *Connection
	cfg        TransportConfig
	serializer api.Serializer
}

func NewConsumer(conn *Connection, cfg TransportConfig, serializer api.Serializer) *Consumer {
	return &Consumer{
		conn:       conn,
		cfg:        cfg,
		serializer: serializer,
	}
}

func (c *Consumer) Consume(ctx context.Context, handler func(context.Context, api.Envelope) error) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to create AMQP channel for consumer: %w", err)
	}
	defer func() {
		_ = ch.Close()
	}()

	jobs := make(chan job)
	c.startWorkerPool(ctx, jobs, handler)

	err = c.startQueueConsumers(ctx, ch, jobs)
	if err != nil {
		return err
	}

	<-ctx.Done()

	return ctx.Err()
}

func (c *Consumer) startWorkerPool(
	ctx context.Context,
	jobs chan job,
	handler func(context.Context, api.Envelope) error,
) {
	poolSize := c.cfg.Options.ConsumerPoolSize
	if poolSize <= 0 {
		poolSize = 10
	}

	for range poolSize {
		go func() {
			for j := range jobs {
				c.handleDelivery(ctx, j.d, handler)
			}
		}()
	}
}

func (c *Consumer) startQueueConsumers(ctx context.Context, ch *amqp.Channel, jobs chan job) error {
	for queueName := range c.cfg.Options.Queues {
		msgs, consumeErr := ch.ConsumeWithContext(
			ctx,
			queueName,
			"",
			false,
			false,
			false,
			false,
			nil,
		)
		if consumeErr != nil {
			return fmt.Errorf("failed to start consuming from queue '%s': %w", queueName, consumeErr)
		}

		go c.processQueueMessages(ctx, jobs, msgs)
	}

	return nil
}

func (c *Consumer) processQueueMessages(ctx context.Context, jobs chan job, messages <-chan amqp.Delivery) {
	for {
		select {
		case <-ctx.Done():
			close(jobs)

			return
		case d, ok := <-messages:
			if !ok {
				return
			}
			jobs <- job{d: d}
		}
	}
}

func (c *Consumer) handleDelivery(
	ctx context.Context,
	d amqp.Delivery,
	handler func(context.Context, api.Envelope) error,
) {
	headersMap := map[string]string{}
	for k, v := range d.Headers {
		if s, ok := v.(string); ok {
			headersMap[k] = s
		}
	}

	env, err := c.serializer.Unmarshal(d.Body, headersMap)
	if err != nil {
		_ = d.Nack(false, false)

		return
	}

	env = env.WithStamp(stamps.ReceivedStamp{
		Transport: c.cfg.Name,
	})

	err = handler(ctx, env)
	if err != nil {
		_ = d.Nack(false, false)

		return
	}

	_ = d.Ack(false)
}

type job struct {
	d amqp.Delivery
}
