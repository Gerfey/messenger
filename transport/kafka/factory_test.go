package kafka_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gopkg.in/yaml.v3"

	"github.com/gerfey/messenger/core/serializer"

	"github.com/gerfey/messenger/tests/mocks"
	"github.com/gerfey/messenger/transport/kafka"
)

func TestNewTransportFactory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()

	factory := kafka.NewTransportFactory(logger)

	assert.NotNil(t, factory)
	assert.IsType(t, &kafka.TransportFactory{}, factory)
}

func TestTransportFactory_Supports(t *testing.T) {
	testCases := []struct {
		name     string
		dsn      string
		expected bool
	}{
		{
			name:     "supports kafka dsn",
			dsn:      "kafka://localhost:9092",
			expected: true,
		},
		{
			name:     "does not support amqp dsn",
			dsn:      "amqp://guest:guest@localhost:5672/",
			expected: false,
		},
		{
			name:     "does not support in-memory dsn",
			dsn:      "in-memory://",
			expected: false,
		},
		{
			name:     "does not support sync dsn",
			dsn:      "sync://",
			expected: false,
		},
		{
			name:     "does not support empty dsn",
			dsn:      "",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := slog.Default()
			factory := kafka.NewTransportFactory(logger)

			result := factory.Supports(tc.dsn)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTransportFactory_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	mockResolver := mocks.NewMockTypeResolver(ctrl)
	factory := kafka.NewTransportFactory(logger)

	name := "test-kafka"
	dsn := "kafka://non-existent-host:9092"
	options := kafka.OptionsConfig{
		Topics: []string{"test-topic"},
		Group:  "test-group",
	}
	ser := serializer.NewSerializer(mockResolver)

	optionsBytes, err := yaml.Marshal(options)
	require.NoError(t, err)

	transport, err := factory.Create(name, dsn, optionsBytes, ser)

	require.Error(t, err)
	assert.Nil(t, transport)
}
