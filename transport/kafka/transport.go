package kafka

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/segmentio/kafka-go"

	"github.com/gerfey/messenger/api"
)

type ConnectionKafka interface {
	CreateReader(kafka.ReaderConfig) *kafka.Reader
	CreateWriter(string, ProducerOptionsConfig, bool, kafka.Balancer) *kafka.Writer
}

type Transport struct {
	cfg        TransportConfig
	producer   api.Producer
	consumer   api.Consumer
	connection ConnectionKafka
}

func NewTransport(cfg TransportConfig, serializer api.Serializer) (api.Transport, error) {
	u, err := url.Parse(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
	}

	brokers := strings.Split(u.Host, ",")

	connection, errConnection := NewConnection(brokers)
	if errConnection != nil {
		return nil, fmt.Errorf("failed to create connection: %w", errConnection)
	}

	producer, errProducer := NewProducer(cfg, connection, serializer)
	if errProducer != nil {
		return nil, fmt.Errorf("failed to create producer: %w", errProducer)
	}

	consumer, errConsumer := NewConsumer(cfg, connection, serializer)
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

func (t *Transport) Close() error {
	return t.producer.Close()
}

func (t *Transport) Retry(ctx context.Context, env api.Envelope) error {
	return t.producer.Send(ctx, env)
}
