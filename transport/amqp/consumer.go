package amqp

import (
	"context"
	"fmt"

	"github.com/gerfey/messenger/envelope"
	"github.com/gerfey/messenger/stamps"
	"github.com/gerfey/messenger/transport"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn       *Connection
	cfg        TransportConfig
	serializer transport.Serializer
}

func NewConsumer(conn *Connection, cfg TransportConfig, serializer transport.Serializer) *Consumer {
	return &Consumer{
		conn:       conn,
		cfg:        cfg,
		serializer: serializer,
	}
}

func (c *Consumer) Consume(ctx context.Context, handler func(context.Context, *envelope.Envelope) error) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	if c.cfg.Options.AutoSetup {
		err := ch.ExchangeDeclare(
			c.cfg.Options.Exchange.Name,
			c.cfg.Options.Exchange.Type,
			c.cfg.Options.Exchange.Durable,
			c.cfg.Options.Exchange.AutoDelete,
			c.cfg.Options.Exchange.Internal,
			false, nil,
		)
		if err != nil {
			return fmt.Errorf("declare exchange: %w", err)
		}

		for queueName, queueCfg := range c.cfg.Options.Queues {
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

			for _, bindingKey := range queueCfg.BindingKeys {
				err := ch.QueueBind(
					queueName,
					bindingKey,
					c.cfg.Options.Exchange.Name,
					false,
					nil,
				)
				if err != nil {
					return fmt.Errorf("bind queue: %w", err)
				}
			}
		}
	}

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

func (c *Consumer) handleDelivery(ctx context.Context, d amqp.Delivery, handler func(context.Context, *envelope.Envelope) error) {
	headersMap := map[string]string{}
	for k, v := range d.Headers {
		if s, ok := v.(string); ok {
			headersMap[k] = s
		}
	}

	env, err := c.serializer.Unmarshal(d.Body, headersMap)
	if err != nil {
		_ = d.Nack(false, true)

		fmt.Printf("[AMQP] Failed to unmarshal message: %v\n", err)

		return
	}

	env = env.WithStamp(stamps.ReceivedStamp{
		Transport: c.cfg.Name,
	})

	err = handler(ctx, env)
	if err != nil {
		_ = d.Nack(false, true)

		fmt.Printf("[AMQP] Handler error: %v\n", err)

		return
	}

	_ = d.Ack(false)
}
