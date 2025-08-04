package redis

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/stamps"
)

const (
	defaultBatchSize = 10
)

type Consumer struct {
	cfg        TransportConfig
	serializer api.Serializer
	conn       *Connection
	logger     *slog.Logger
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

func NewConsumer(cfg TransportConfig, ser api.Serializer, conn *Connection, logger *slog.Logger) *Consumer {
	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		cfg:        cfg,
		serializer: ser,
		conn:       conn,
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (c *Consumer) Consume(ctx context.Context, handler func(context.Context, api.Envelope) error) error {
	group := c.cfg.Options.Group
	stream := c.cfg.Options.Stream

	_ = c.conn.Client().XGroupCreateMkStream(ctx, stream, group, "$")

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.consumeLoop(handler)
	}()

	<-ctx.Done()
	c.cancel()
	c.wg.Wait()

	return ctx.Err()
}

func (c *Consumer) consumeLoop(handler func(context.Context, api.Envelope) error) {
	rdb := c.conn.Client()
	stream := c.cfg.Options.Stream
	group := c.cfg.Options.Group
	consumer := c.cfg.Options.Consumer

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			streams, err := rdb.XReadGroup(c.ctx, &redis.XReadGroupArgs{
				Group:    group,
				Consumer: consumer,
				Streams:  []string{stream, ">"},
				Count:    defaultBatchSize,
				Block:    time.Second,
			}).Result()

			if err != nil && !errors.Is(err, redis.Nil) {
				c.logger.Error("XREADGROUP error", "error", err)

				continue
			}

			for _, s := range streams {
				for _, msg := range s.Messages {
					c.handleMessage(msg, handler)
				}
			}
		}
	}
}

func (c *Consumer) handleMessage(msg redis.XMessage, handler func(context.Context, api.Envelope) error) {
	bodyRaw, ok := msg.Values["body"]
	if !ok {
		c.logger.Warn("missing 'body' field in message", "id", msg.ID)

		return
	}

	bodyBytes, ok := bodyRaw.(string)
	if !ok {
		c.logger.Warn("invalid body format", "id", msg.ID)

		return
	}

	headers := make(map[string]string)
	for k, v := range msg.Values {
		if strings.HasPrefix(k, "header_") {
			if str, isString := v.(string); isString {
				headers[strings.TrimPrefix(k, "header_")] = str
			}
		}
	}

	env, errUnmarshal := c.serializer.Unmarshal([]byte(bodyBytes), headers)
	if errUnmarshal != nil {
		c.logger.Error("failed to unmarshal", "error", errUnmarshal)

		return
	}

	env = env.WithStamp(stamps.ReceivedStamp{Transport: c.cfg.Name})

	if errHandler := handler(c.ctx, env); errHandler != nil {
		c.logger.Error("handler failed", "error", errHandler)

		return
	}

	if err := c.conn.Client().XAck(c.ctx, c.cfg.Options.Stream, c.cfg.Options.Group, msg.ID).Err(); err != nil {
		c.logger.Error("XACK failed", "id", msg.ID, "error", err)
	}
}
