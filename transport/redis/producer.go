package redis

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/redis/go-redis/v9"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/stamps"
)

type Producer struct {
	config     TransportConfig
	serializer api.Serializer
	connection ConnectionRedis
}

func NewProducer(config TransportConfig, serializer api.Serializer, connection ConnectionRedis) (api.Producer, error) {
	return &Producer{
		config:     config,
		serializer: serializer,
		connection: connection,
	}, nil
}

func (p *Producer) Send(ctx context.Context, env api.Envelope) error {
	payload, headers, err := p.serializer.Marshal(env)
	if err != nil {
		return fmt.Errorf("redis: marshal envelope failed: %w", err)
	}

	data := map[string]any{
		"body": payload,
	}

	for k, v := range headers {
		data["header_"+k] = v
	}

	stream := p.config.Options.Stream
	if stream == "" {
		return errors.New("redis: stream name is not configured")
	}

	id := "*"
	if stamp, ok := envelope.LastStampOf[stamps.MessageIDStamp](env); ok {
		if p.isValidRedisStreamID(stamp.MessageID) {
			id = stamp.MessageID
		}
	}

	_, err = p.connection.Client().XAdd(ctx, &redis.XAddArgs{
		ID:     id,
		Stream: stream,
		Values: data,
	}).Result()
	if err != nil {
		return fmt.Errorf("redis: XADD failed: %w", err)
	}

	return nil
}

func (p *Producer) Close() error {
	return nil
}

func (p *Producer) isValidRedisStreamID(id string) bool {
	return regexp.MustCompile(`^\d+-\d+$`).MatchString(id)
}
