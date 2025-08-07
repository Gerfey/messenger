package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
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
	writers    map[string]*kafka.Writer
	mu         sync.RWMutex
}

func NewProducer(cfg TransportConfig, ser api.Serializer, conn *Connection, logger *slog.Logger) (*Producer, error) {
	p := &Producer{
		cfg:        cfg,
		serializer: ser,
		conn:       conn,
		logger:     logger,
		writers:    make(map[string]*kafka.Writer),
	}

	if len(cfg.Options.Topics) == 0 {
		return nil, errors.New("no topics configured for kafka transport")
	}

	var balancer kafka.Balancer = &kafka.LeastBytes{}
	switch cfg.Options.Producer.Balancer {
	case "hash":
		balancer = &kafka.Hash{}
	case "round_robin":
		balancer = &kafka.RoundRobin{}
	case "least_bytes":
		balancer = &kafka.LeastBytes{}
	}

	if cfg.Options.Key.Strategy != "none" {
		balancer = &kafka.Hash{}
	}

	for _, topic := range cfg.Options.Topics {
		writer := conn.CreateWriter(
			topic,
			cfg.Options.Producer,
			cfg.Options.Producer.Async,
			balancer,
		)

		p.writers[topic] = writer
	}

	return p, nil
}

func (p *Producer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errs []error
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close writer for topic %s: %w", topic, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing writers: %v", errs)
	}

	return nil
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

	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, topic := range p.cfg.Options.Topics {
		writer, exists := p.writers[topic]
		if !exists {
			return fmt.Errorf("writer for topic %s not found", topic)
		}

		if writeErr := writer.WriteMessages(ctx, msg); writeErr != nil {
			return fmt.Errorf("producer failed to write messages to topic %s: %w", topic, writeErr)
		}

		p.logger.DebugContext(ctx, "message sent to kafka topic",
			slog.String("topic", topic),
			slog.String("message_type", fmt.Sprintf("%T", env.Message())))
	}

	return nil
}

func (p *Producer) extractMessageKey(env api.Envelope) ([]byte, error) {
	if p.cfg.Options.Key.Strategy != "message_id" {
		return nil, nil
	}

	for _, s := range env.Stamps() {
		if msgIDStamp, ok := s.(stamps.MessageIDStamp); ok {
			return []byte(msgIDStamp.MessageID), nil
		}
	}

	return nil, errors.New("message_id stamp not found")
}
