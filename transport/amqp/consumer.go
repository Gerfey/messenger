package amqp

import (
	"context"
	"fmt"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/stamps"
	amqp "github.com/rabbitmq/amqp091-go"
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
		return err
	}
	defer func() {
		_ = ch.Close()
	}()

	for queueName := range c.cfg.Options.Queues {
		msgs, err := ch.ConsumeWithContext(
			ctx,
			queueName,
			"",
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("consume from queue %s failed: %w", queueName, err)
		}

		go func(queue string, messages <-chan amqp.Delivery) {
			for {
				select {
				case <-ctx.Done():
					return
				case d, ok := <-messages:
					if !ok {
						return
					}
					go c.handleDelivery(ctx, d, handler)
				}
			}
		}(queueName, msgs)
	}

	<-ctx.Done()

	return ctx.Err()
}

func (c *Consumer) handleDelivery(ctx context.Context, d amqp.Delivery, handler func(context.Context, api.Envelope) error) {
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
