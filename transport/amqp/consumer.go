package amqp

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/stamps"
)

const (
	defaultPoolSize         = 10
	workerPoolCheckInterval = 30 * time.Second
	workerBatchSize         = 5
)

type Consumer struct {
	conn       *Connection
	cfg        TransportConfig
	serializer api.Serializer
	wg         sync.WaitGroup
	logger     *slog.Logger
}

func NewConsumer(conn *Connection, cfg TransportConfig, serializer api.Serializer) *Consumer {
	return &Consumer{
		conn:       conn,
		cfg:        cfg,
		serializer: serializer,
		logger:     slog.Default(),
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
	close(jobs)
	c.wg.Wait()

	return ctx.Err()
}

func (c *Consumer) startWorkerPool(
	ctx context.Context,
	jobs chan job,
	handler func(context.Context, api.Envelope) error,
) {
	poolSize := c.cfg.Options.Pool.Size
	if poolSize <= 0 {
		poolSize = defaultPoolSize
	}

	for range poolSize {
		c.wg.Add(1)
		go c.startWorker(ctx, jobs, handler)
	}

	if c.cfg.Options.Pool.Dynamic {
		go c.manageWorkerPool(ctx, jobs, handler)
	}
}

func (c *Consumer) manageWorkerPool(
	ctx context.Context,
	jobs chan job,
	handler func(context.Context, api.Envelope) error,
) {
	ticker := time.NewTicker(workerPoolCheckInterval)
	defer ticker.Stop()

	currentSize := c.cfg.Options.Pool.Size
	minSize := c.cfg.Options.Pool.MinSize
	maxSize := c.cfg.Options.Pool.MaxSize

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if len(jobs) > currentSize && currentSize < maxSize {
				toAdd := min(maxSize-currentSize, workerBatchSize)

				for range toAdd {
					c.wg.Add(1)
					go c.startWorker(ctx, jobs, handler)
				}

				currentSize += toAdd
				c.logger.DebugContext(ctx, "Increased worker pool size", "new_size", currentSize)
			} else if len(jobs) == 0 && currentSize > minSize {
				currentSize = max(currentSize-workerBatchSize, minSize)
				c.logger.DebugContext(ctx, "Decreased worker pool size", "new_size", currentSize)
			}
		}
	}
}

func (c *Consumer) startWorker(
	ctx context.Context,
	jobs chan job,
	handler func(context.Context, api.Envelope) error,
) {
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case j, ok := <-jobs:
			if !ok {
				return
			}
			c.handleDelivery(ctx, j.d, handler)
		}
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
