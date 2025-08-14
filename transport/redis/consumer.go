package redis

import (
	"context"
	"errors"
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
	config     TransportConfig
	serializer api.Serializer
	connection ConnectionRedis
	wg         sync.WaitGroup
}

func NewConsumer(config TransportConfig, serializer api.Serializer, connection ConnectionRedis) (api.Consumer, error) {
	return &Consumer{
		config:     config,
		serializer: serializer,
		connection: connection,
	}, nil
}

func (c *Consumer) Consume(ctx context.Context, handler func(context.Context, api.Envelope) error) error {
	group := c.config.Options.Group
	stream := c.config.Options.Stream

	_ = c.connection.Client().XGroupCreateMkStream(ctx, stream, group, "$")

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.consumeLoop(ctx, handler)
	}()

	<-ctx.Done()

	c.wg.Wait()

	return ctx.Err()
}

func (c *Consumer) Close() error {
	return nil
}

func (c *Consumer) consumeLoop(ctx context.Context, handler func(context.Context, api.Envelope) error) {
	rdb := c.connection.Client()
	stream := c.config.Options.Stream
	group := c.config.Options.Group
	consumer := c.config.Options.Consumer

	for {
		select {
		case <-ctx.Done():
			return
		default:
			streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    group,
				Consumer: consumer,
				Streams:  []string{stream, ">"},
				Count:    defaultBatchSize,
				Block:    time.Second,
			}).Result()

			if err != nil && !errors.Is(err, redis.Nil) {
				continue
			}

			for _, s := range streams {
				for _, msg := range s.Messages {
					c.handleMessage(ctx, msg, handler)
				}
			}
		}
	}
}

func (c *Consumer) handleMessage(
	ctx context.Context,
	msg redis.XMessage,
	handler func(context.Context, api.Envelope) error,
) {
	bodyRaw, ok := msg.Values["body"]
	if !ok {
		return
	}

	bodyBytes, ok := bodyRaw.(string)
	if !ok {
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
		return
	}

	env = env.WithStamp(stamps.ReceivedStamp{Transport: c.config.Name})

	if errHandler := handler(ctx, env); errHandler != nil {
		return
	}

	if err := c.connection.Client().XAck(ctx, c.config.Options.Stream, c.config.Options.Group, msg.ID).Err(); err != nil {
		return
	}
}
