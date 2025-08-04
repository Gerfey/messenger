package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/stamps"
)

type Producer struct {
	cfg        TransportConfig
	serializer api.Serializer
	conn       *Connection
	logger     *slog.Logger
}

func NewProducer(cfg TransportConfig, ser api.Serializer, conn *Connection, logger *slog.Logger) (*Producer, error) {
	return &Producer{
		cfg:        cfg,
		serializer: ser,
		conn:       conn,
		logger:     logger,
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

	key, keyErr := p.extractMessageKey(env)
	if keyErr == nil && len(key) > 0 {
		msg.Key = key
	}

	topics := p.cfg.Options.Topics

	if len(topics) == 0 {
		return errors.New("no topics configured for kafka transport")
	}

	for _, topic := range topics {
		writer := p.conn.CreateWriter(topic)

		if p.cfg.Options.Key.Strategy != "none" {
			writer.Balancer = &kafka.Hash{}
		}

		if writeErr := writer.WriteMessages(ctx, msg); writeErr != nil {
			return fmt.Errorf("producer failed to write messages: %w", writeErr)
		}

		p.logger.DebugContext(ctx, "message sent to kafka topic",
			slog.String("topic", topic),
			slog.String("message_type", fmt.Sprintf("%T", env.Message())))

		closeErr := writer.Close()
		if closeErr != nil {
			return closeErr
		}
	}

	return nil
}

func (p *Producer) extractMessageKey(env api.Envelope) ([]byte, error) {
	switch p.cfg.Options.Key.Strategy {
	case "none":
		return nil, nil
	case "message_id":
		for _, s := range env.Stamps() {
			if msgIDStamp, ok := s.(stamps.MessageIDStamp); ok {
				return []byte(msgIDStamp.MessageID), nil
			}
		}

		return nil, errors.New("message_id stamp not found")
	default:
		return nil, fmt.Errorf("unknown key strategy: %s", p.cfg.Options.Key.Strategy)
	}
}
