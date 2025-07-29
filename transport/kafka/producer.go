package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/gerfey/messenger/api"
)

type Producer struct {
	writer     *kafka.Writer
	cfg        TransportConfig
	serializer api.Serializer
}

func NewProducer(cfg TransportConfig, ser api.Serializer, conn *Connection) (*Producer, error) {
	return &Producer{
		cfg:        cfg,
		serializer: ser,
		writer:     conn.CreateWriter(cfg.Options.Topic),
	}, nil
}

func (p *Producer) Send(ctx context.Context, env api.Envelope) error {
	payload, headers, err := p.serializer.Marshal(env)
	if err != nil {
		return fmt.Errorf("serializer envelope failed: %w", err)
	}

	kHeaders := make([]kafka.Header, 0, len(headers))
	for k, v := range headers {
		kHeaders = append(kHeaders, kafka.Header{Key: k, Value: []byte(v)})
	}

	msg := kafka.Message{
		Headers: kHeaders,
		Value:   payload,
		Time:    time.Now(),
	}

	if writeErr := p.writer.WriteMessages(ctx, msg); writeErr != nil {
		return fmt.Errorf("producer failed to write messages: %w", writeErr)
	}

	return nil
}
