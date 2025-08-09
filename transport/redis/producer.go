package redis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/redis/go-redis/v9"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/stamps"
)

type Producer struct {
	cfg        TransportConfig
	serializer api.Serializer
	conn       *Connection
	logger     *slog.Logger
}

func NewProducer(cfg TransportConfig, ser api.Serializer, conn *Connection, logger *slog.Logger) *Producer {
	return &Producer{
		cfg:        cfg,
		serializer: ser,
		conn:       conn,
		logger:     logger,
	}
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

	stream := p.cfg.Options.Stream
	if stream == "" {
		return errors.New("redis: stream name is not configured")
	}

	id := "*"
	if stamp, ok := envelope.LastStampOf[stamps.MessageIDStamp](env); ok {
		if p.isValidRedisStreamID(stamp.MessageID) {
			id = stamp.MessageID
		}
	}

	_, err = p.conn.Client().XAdd(ctx, &redis.XAddArgs{
		ID:     id,
		Stream: stream,
		Values: data,
	}).Result()
	if err != nil {
		return fmt.Errorf("redis: XADD failed: %w", err)
	}

	p.logger.DebugContext(ctx, "message sent to redis stream",
		slog.String("stream", stream),
		slog.String("message_type", fmt.Sprintf("%T", env.Message())))

	return nil
}

func (p *Producer) isValidRedisStreamID(id string) bool {
	return regexp.MustCompile(`^\d+-\d+$`).MatchString(id)
}
