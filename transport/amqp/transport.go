package amqp

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/serializer"
)

type Transport struct {
	cfg        TransportConfig
	publisher  *Publisher
	consumer   *Consumer
	retry      *Retry
	serializer api.Serializer
	conn       *Connection
	logger     *slog.Logger
}

func NewTransport(cfg TransportConfig, resolver api.TypeResolver, logger *slog.Logger) (api.Transport, error) {
	conn, err := NewConnection(cfg.DSN)
	if err != nil {
		logger.Error("failed to create AMQP connection", "dsn", cfg.DSN, "error", err)

		return nil, err
	}

	ser := serializer.NewSerializer(resolver)

	pub := NewPublisher(conn, cfg, ser)
	cons := NewConsumer(conn, cfg, ser)
	ret := NewRetry(conn, cfg, ser)

	transport := &Transport{
		cfg:        cfg,
		publisher:  pub,
		consumer:   cons,
		retry:      ret,
		serializer: ser,
		conn:       conn,
		logger:     logger,
	}

	if cfg.Options.AutoSetup {
		if setupErr := transport.setup(); setupErr != nil {
			logger.Error("failed to setup AMQP transport", "transport", cfg.Name, "error", setupErr)

			return nil, setupErr
		}

		logger.Info("AMQP transport setup completed", "transport", cfg.Name)
	}

	return transport, nil
}

func (t *Transport) Name() string {
	return t.cfg.Name
}

func (t *Transport) Send(ctx context.Context, env api.Envelope) error {
	return t.publisher.Publish(ctx, env)
}

func (t *Transport) Receive(ctx context.Context, handler func(context.Context, api.Envelope) error) error {
	return t.consumer.Consume(ctx, handler)
}

func (t *Transport) Retry(ctx context.Context, env api.Envelope) error {
	return t.retry.Retry(ctx, env)
}

func (t *Transport) setup() error {
	ch, err := t.conn.Channel()
	if err != nil {
		t.logger.Error("failed to open channel", "error", err)

		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer func() {
		_ = ch.Close()
	}()

	err = ch.ExchangeDeclare(
		t.cfg.Options.Exchange.Name,
		t.cfg.Options.Exchange.Type,
		t.cfg.Options.Exchange.Durable,
		t.cfg.Options.Exchange.AutoDelete,
		t.cfg.Options.Exchange.Internal,
		false,
		nil,
	)
	if err != nil {
		t.logger.Error("failed to declare exchange", "exchange", t.cfg.Options.Exchange.Name, "error", err)

		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	for queueName, queueCfg := range t.cfg.Options.Queues {
		_, err = ch.QueueDeclare(
			queueName,
			queueCfg.Durable,
			queueCfg.AutoDelete,
			queueCfg.Exclusive,
			false,
			nil,
		)
		if err != nil {
			t.logger.Error("declare queue", "queue", queueName, "error", err)

			return fmt.Errorf("declare queue: %w", err)
		}

		for _, bindingKey := range queueCfg.BindingKeys {
			bindErr := ch.QueueBind(
				queueName,
				bindingKey,
				t.cfg.Options.Exchange.Name,
				false,
				nil,
			)
			if bindErr != nil {
				t.logger.Error("bind queue", "queue", queueName, "binding_key", bindingKey, "error", bindErr)

				return fmt.Errorf("bind queue: %w", bindErr)
			}
		}
	}

	return nil
}

func getRoutingKey(msg any) string {
	var routingKey string
	if rk, ok := msg.(api.RoutedMessage); ok {
		routingKey = rk.RoutingKey()
	} else {
		routingKey = reflect.TypeOf(msg).String()
	}

	return routingKey
}
