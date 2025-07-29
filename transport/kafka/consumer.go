package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/stamps"
)

const (
	minBytes          = 10e3 // 10KB
	maxBytes          = 10e6 // 10MB
	sessionTimeout    = 10 * time.Second
	rebalanceTimeout  = 5 * time.Second
	heartbeatInterval = 2 * time.Second
	defaultPoolSize   = 10
	readLagInterval   = -1
)

type Consumer struct {
	cfg        TransportConfig
	serializer api.Serializer
	conn       *Connection
}

func NewConsumer(cfg TransportConfig, ser api.Serializer, conn *Connection) *Consumer {
	return &Consumer{
		cfg:        cfg,
		serializer: ser,
		conn:       conn,
	}
}

func (c *Consumer) Consume(ctx context.Context, handler func(context.Context, api.Envelope) error) error {
	readerConfig := kafka.ReaderConfig{
		GroupID:         c.cfg.Options.Group,
		Topic:           c.cfg.Options.Topic,
		StartOffset:     c.startOffset(c.cfg.Options.Offset),
		CommitInterval:  c.cfg.Options.CommitInterval,
		MinBytes:        minBytes,
		MaxBytes:        maxBytes,
		ReadLagInterval: readLagInterval,

		SessionTimeout:    sessionTimeout,
		RebalanceTimeout:  rebalanceTimeout,
		HeartbeatInterval: heartbeatInterval,
		MaxWait:           time.Second,
	}

	r := c.conn.CreateReader(readerConfig)
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
		poolSize = defaultPoolSize
	}

	for i := range make([]struct{}, poolSize) {
		go func(_ int) {
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

	if handlerErr := handler(ctx, env); handlerErr != nil {
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

func (c *Consumer) headerMap(headers []kafka.Header) map[string]string {
	m := make(map[string]string, len(headers))
	for _, h := range headers {
		m[h.Key] = string(h.Value)
	}

	return m
}
