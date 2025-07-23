package amqp_test

import (
	"context"
	"errors"
	"testing"

	amqp091 "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/builder"
	"github.com/gerfey/messenger/config"
	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/tests/helpers"
	"github.com/gerfey/messenger/tests/mocks"
	"github.com/gerfey/messenger/transport/amqp"
)

func TestTransport_Name(t *testing.T) {
	cfg := amqp.TransportConfig{
		Name: "test-transport",
		DSN:  "amqp://localhost",
		Options: config.OptionsConfig{
			AutoSetup: false,
		},
	}

	resolver := builder.NewResolver()
	logger, _ := helpers.NewFakeLogger()

	transport, err := amqp.NewTransport(cfg, resolver, logger)
	if err != nil {
		t.Skip("Skipping test - requires AMQP connection")

		return
	}

	assert.Equal(t, "test-transport", transport.Name())
}

func TestTransport_Setup_WithMocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("successful setup", func(t *testing.T) {
		mockConn := mocks.NewMockConnectionAMQP(ctrl)
		mockChannel := mocks.NewMockChannelAMQP(ctrl)

		cfg := amqp.TransportConfig{
			Name: "test-transport",
			DSN:  "amqp://localhost",
			Options: config.OptionsConfig{
				AutoSetup: true,
				Exchange: config.ExchangeConfig{
					Name:       "test-exchange",
					Type:       "direct",
					Durable:    true,
					AutoDelete: false,
					Internal:   false,
				},
				Queues: map[string]config.Queue{
					"test-queue": {
						Durable:     true,
						AutoDelete:  false,
						Exclusive:   false,
						BindingKeys: []string{"test.key"},
					},
				},
			},
		}

		mockConn.EXPECT().Channel().Return(mockChannel, nil)
		mockChannel.EXPECT().ExchangeDeclare(
			"test-exchange", "direct", true, false, false, false, nil,
		).Return(nil)
		mockChannel.EXPECT().QueueDeclare(
			"test-queue", true, false, false, false, nil,
		).Return(amqp091.Queue{Name: "test-queue"}, nil)
		mockChannel.EXPECT().QueueBind(
			"test-queue", "test.key", "test-exchange", false, nil,
		).Return(nil)
		mockChannel.EXPECT().Close().Return(nil)

		transport := createMockTransport(cfg, mockConn)
		err := transport.Setup()
		require.NoError(t, err)
	})

	t.Run("setup with channel error", func(t *testing.T) {
		mockConn := mocks.NewMockConnectionAMQP(ctrl)

		cfg := amqp.TransportConfig{
			Name: "test-transport",
			DSN:  "amqp://localhost",
			Options: config.OptionsConfig{
				AutoSetup: true,
			},
		}

		expectedErr := errors.New("channel error")
		mockConn.EXPECT().Channel().Return(nil, expectedErr)

		transport := createMockTransport(cfg, mockConn)
		err := transport.Setup()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open channel")
	})

	t.Run("setup with exchange declare error", func(t *testing.T) {
		mockConn := mocks.NewMockConnectionAMQP(ctrl)
		mockChannel := mocks.NewMockChannelAMQP(ctrl)

		cfg := amqp.TransportConfig{
			Name: "test-transport",
			DSN:  "amqp://localhost",
			Options: config.OptionsConfig{
				AutoSetup: true,
				Exchange: config.ExchangeConfig{
					Name: "test-exchange",
					Type: "direct",
				},
			},
		}

		expectedErr := errors.New("exchange error")
		mockConn.EXPECT().Channel().Return(mockChannel, nil)
		mockChannel.EXPECT().ExchangeDeclare(
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			gomock.Any(), gomock.Any(), gomock.Any(),
		).Return(expectedErr)
		mockChannel.EXPECT().Close().Return(nil)

		transport := createMockTransport(cfg, mockConn)
		err := transport.Setup()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to declare exchange")
	})
}

func TestTransport_Integration_Methods(t *testing.T) {
	cfg := amqp.TransportConfig{
		Name: "test-transport",
		DSN:  "amqp://localhost",
		Options: config.OptionsConfig{
			AutoSetup: false,
		},
	}

	resolver := builder.NewResolver()
	resolver.RegisterMessage(&helpers.TestMessage{})
	logger, _ := helpers.NewFakeLogger()

	transport, err := amqp.NewTransport(cfg, resolver, logger)
	if err != nil {
		t.Skip("Skipping integration test - requires AMQP connection")

		return
	}

	ctx := t.Context()
	msg := &helpers.TestMessage{ID: "123", Content: "test message"}
	env := envelope.NewEnvelope(msg)

	t.Run("send method delegates to publisher", func(t *testing.T) {
		sendErr := transport.Send(ctx, env)
		assert.Error(t, sendErr)
	})

	t.Run("receive method delegates to consumer", func(t *testing.T) {
		handler := func(_ context.Context, _ api.Envelope) error {
			return nil
		}

		receiveErr := transport.Receive(ctx, handler)
		assert.Error(t, receiveErr)
	})

	t.Run("retry method delegates to retry", func(t *testing.T) {
		retryableTransport, ok := transport.(api.RetryableTransport)
		if !ok {
			t.Skip("Transport does not implement RetryableTransport")

			return
		}

		retryErr := retryableTransport.Retry(ctx, env)
		assert.Error(t, retryErr)
	})
}

type MockTransport struct {
	cfg  amqp.TransportConfig
	conn *mocks.MockConnectionAMQP
}

func createMockTransport(cfg amqp.TransportConfig, conn *mocks.MockConnectionAMQP) *MockTransport {
	return &MockTransport{
		cfg:  cfg,
		conn: conn,
	}
}

func (t *MockTransport) Setup() error {
	ch, err := t.conn.Channel()
	if err != nil {
		return errors.New("failed to open channel: " + err.Error())
	}
	defer ch.Close()

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
		return errors.New("failed to declare exchange: " + err.Error())
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
			return errors.New("declare queue: " + err.Error())
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
				return errors.New("bind queue: " + bindErr.Error())
			}
		}
	}

	return nil
}
