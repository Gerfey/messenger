package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
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
	u, err := url.Parse(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
	}

	brokers := strings.Split(u.Host, ",")

	conn, err := NewConnection(brokers)
	if err != nil {
		logger.Error("failed to connect to Kafka brokers", "error", err)

		return nil, err
	}

	producer, err := NewProducer(cfg, ser, conn, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

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

func (t *Transport) Close() error {
	return t.producer.Close()
}

func (t *Transport) Retry(ctx context.Context, env api.Envelope) error {
	return t.producer.Send(ctx, env)
}
