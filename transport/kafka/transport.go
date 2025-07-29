package kafka

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/serializer"
)

type Transport struct {
	cfg        TransportConfig
	producer   *Producer
	consumer   *Consumer
	serializer api.Serializer
	logger     *slog.Logger
}

func NewTransport(cfg TransportConfig, resolver api.TypeResolver, logger *slog.Logger) (api.Transport, error) {
	ser := serializer.NewSerializer(resolver)

	producer, err := NewProducer(cfg, ser)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	consumer := NewConsumer(cfg, ser)

	return &Transport{
		cfg:        cfg,
		producer:   producer,
		consumer:   consumer,
		serializer: ser,
		logger:     logger,
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
