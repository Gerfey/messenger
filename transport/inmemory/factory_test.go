package inmemory_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/gerfey/messenger/config"
	"github.com/gerfey/messenger/tests/mocks"
	"github.com/gerfey/messenger/transport/inmemory"
)

func TestNewTransportFactory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	mockResolver := mocks.NewMockTypeResolver(ctrl)

	factory := inmemory.NewTransportFactory(logger, mockResolver)

	assert.NotNil(t, factory)
	assert.IsType(t, &inmemory.TransportFactory{}, factory)
}

func TestTransportFactory_Supports(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	mockResolver := mocks.NewMockTypeResolver(ctrl)
	factory := inmemory.NewTransportFactory(logger, mockResolver)

	testCases := []struct {
		name     string
		dsn      string
		expected bool
	}{
		{
			name:     "supports in-memory dsn",
			dsn:      "in-memory://test",
			expected: true,
		},
		{
			name:     "does not support amqp dsn",
			dsn:      "amqp://guest:guest@localhost:5672/",
			expected: false,
		},
		{
			name:     "does not support empty dsn",
			dsn:      "",
			expected: false,
		},
		{
			name:     "does not support sync dsn",
			dsn:      "sync://",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
	factory := inmemory.NewTransportFactory(logger, mockResolver)

	name := "test-inmemory"
	dsn := "in-memory://test"
	options := config.OptionsConfig{}

	transport, err := factory.Create(name, dsn, options)

	require.NoError(t, err)
	assert.NotNil(t, transport)
	assert.IsType(t, &inmemory.Transport{}, transport)
}
