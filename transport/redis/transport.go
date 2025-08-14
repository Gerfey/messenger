package redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"

	"github.com/gerfey/messenger/api"
)

type ConnectionRedis interface {
	Client() *redis.Client
}

type Transport struct {
	cfg        TransportConfig
	producer   api.Producer
	consumer   api.Consumer
	connection ConnectionRedis
}

func NewTransport(cfg TransportConfig, serializer api.Serializer) (api.Transport, error) {
	connection, errConnection := NewConnection(cfg.DSN)
	if errConnection != nil {
		return nil, fmt.Errorf("failed to create connection: %w", errConnection)
	}

	producer, errProducer := NewProducer(cfg, serializer, connection)
	if errProducer != nil {
		return nil, fmt.Errorf("failed to create producer: %w", errProducer)
	}

	consumer, errConsumer := NewConsumer(cfg, serializer, connection)
	if errConsumer != nil {
		return nil, fmt.Errorf("failed to create producer: %w", errConsumer)
	}

	return &Transport{
		cfg:        cfg,
		producer:   producer,
		consumer:   consumer,
		connection: connection,
	}, nil
}

func (t *Transport) Name() string {
	return t.cfg.Name
}

func (t *Transport) Send(ctx context.Context, env api.Envelope) error {
	return t.producer.Send(ctx, env)
}

func (t *Transport) Receive(ctx context.Context, handler func(context.Context, api.Envelope) error) error {
	return t.consumer.Consume(ctx, handler)
}

func (t *Transport) Retry(ctx context.Context, env api.Envelope) error {
	return t.producer.Send(ctx, env)
}

func (t *Transport) Setup(ctx context.Context) error {
	if !t.cfg.Options.AutoSetup {
		return nil
	}

	stream := t.cfg.Options.Stream
	group := t.cfg.Options.Group

	_, err := t.connection.Client().XGroupCreateMkStream(ctx, stream, group, "$").Result()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	return nil
}

func (t *Transport) Close() error {
	return nil
}
