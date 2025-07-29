package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/stamps"
)

type Consumer struct {
	cfg        TransportConfig
	serializer api.Serializer
}

func NewConsumer(cfg TransportConfig, ser api.Serializer) *Consumer {
	return &Consumer{
		cfg:        cfg,
		serializer: ser,
	}
}

func (c *Consumer) Consume(ctx context.Context, handler func(context.Context, api.Envelope) error) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:         c.cfg.Options.Brokers,
		GroupID:         c.cfg.Options.Group,
		Topic:           c.cfg.Options.Topic,
		StartOffset:     c.startOffset(c.cfg.Options.Offset),
		CommitInterval:  c.cfg.Options.CommitInterval,
		MinBytes:        10e3, // 10KB
		MaxBytes:        10e6, // 10MB
		ReadLagInterval: -1,

		SessionTimeout:    10 * time.Second,
		RebalanceTimeout:  5 * time.Second,
		HeartbeatInterval: 2 * time.Second,
		MaxWait:           1 * time.Second,
	})
	defer r.Close()

	jobs := make(chan job)
	c.startWorkerPool(ctx, jobs, handler)

	go c.fetchMessages(ctx, r, jobs)

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

	for i := 0; i < poolSize; i++ {
		go func(workerID int) {
			for j := range jobs {
				c.handleMessage(ctx, j.r, j.msg, handler)
			}
		}(i)
	}
}

func (c *Consumer) fetchMessages(ctx context.Context, r *kafka.Reader, jobs chan job) {
	for {
		select {
		case <-ctx.Done():
			close(jobs)

			return
		default:
			msg, err := r.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}

				continue
			}

			jobs <- job{r: r, msg: msg}
		}
	}
}

func (c *Consumer) handleMessage(
	ctx context.Context,
	r *kafka.Reader,
	msg kafka.Message,
	handler func(context.Context, api.Envelope) error,
) {
	env, err := c.serializer.Unmarshal(msg.Value, c.headerMap(msg.Headers))
	if err != nil {
		_ = r.CommitMessages(ctx, msg)

		return
	}

	env = env.WithStamp(stamps.ReceivedStamp{Transport: c.cfg.Name})

	if err := handler(ctx, env); err != nil {
		_ = r.CommitMessages(ctx, msg)

		return
	}

	_ = r.CommitMessages(ctx, msg)
}

type job struct {
	r   *kafka.Reader
	msg kafka.Message
}

func (c *Consumer) startOffset(offset string) int64 {
	if offset == "earliest" {
		return kafka.FirstOffset
	}
	return kafka.LastOffset
}

func (c *Consumer) headerMap(hdrs []kafka.Header) map[string]string {
	m := make(map[string]string, len(hdrs))
	for _, h := range hdrs {
		m[h.Key] = string(h.Value)
	}
	return m
}
