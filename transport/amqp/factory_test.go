package amqp_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gopkg.in/yaml.v3"

	"github.com/gerfey/messenger/tests/mocks"
	"github.com/gerfey/messenger/transport/amqp"
)

func TestNewTransportFactory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	mockResolver := mocks.NewMockTypeResolver(ctrl)

	factory := amqp.NewTransportFactory(logger, mockResolver)

	assert.NotNil(t, factory)
	assert.IsType(t, &amqp.TransportFactory{}, factory)
}

func TestTransportFactory_Supports(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
		want bool
	}{
		{
			name: "supports amqp dsn",
			dsn:  "amqp://guest:guest@localhost:5672/",
			want: true,
		},
		{
			name: "does not support amqps dsn",
			dsn:  "amqps://guest:guest@localhost:5672/",
			want: false,
		},
		{
			name: "does not support in-memory dsn",
			dsn:  "in-memory://",
			want: false,
		},
		{
			name: "does not support sync dsn",
			dsn:  "sync://",
			want: false,
		},
		{
			name: "does not support empty dsn",
			dsn:  "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := slog.Default()
			mockResolver := mocks.NewMockTypeResolver(ctrl)
			factory := amqp.NewTransportFactory(logger, mockResolver)

			got := factory.Supports(tt.dsn)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTransportFactory_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	mockResolver := mocks.NewMockTypeResolver(ctrl)
	factory := amqp.NewTransportFactory(logger, mockResolver)

	name := "test-amqp"

	dsn := "amqp://guest:guest@non-existent-host:5672/"
	options := amqp.OptionsConfig{
		AutoSetup: true,
		Exchange: amqp.ExchangeConfig{
			Name:       "test-exchange",
			Type:       "direct",
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
		},
		Queues: map[string]amqp.Queue{
			"test-queue": {
				BindingKeys: []string{"test-key"},
				Durable:     true,
				AutoDelete:  false,
				Exclusive:   false,
			},
		},
	}

	optionsBytes, err := yaml.Marshal(options)
	require.NoError(t, err)

	_, err = factory.Create(name, dsn, optionsBytes)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect")
}
