package redis

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gerfey/messenger/api"
)

type Transport struct {
	cfg        TransportConfig
	producer   *Producer
	consumer   *Consumer
	serializer api.Serializer
	logger     *slog.Logger
	conn       *Connection
}

func NewTransport(cfg TransportConfig, logger *slog.Logger, ser api.Serializer) (api.Transport, error) {
	conn, err := NewConnection(cfg.DSN)
	if err != nil {
		logger.Error("failed to connect", "error", err)

		return nil, err
	}

	producer := NewProducer(cfg, ser, conn, logger)
	consumer := NewConsumer(cfg, ser, conn, logger)

	return &Transport{
		cfg:        cfg,
		producer:   producer,
		consumer:   consumer,
		serializer: ser,
		logger:     logger,
		conn:       conn,
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

	_, err := t.conn.Client().XGroupCreateMkStream(ctx, stream, group, "$").Result()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	return nil
}
