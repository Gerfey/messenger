package sync_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/gerfey/messenger/config"
	"github.com/gerfey/messenger/tests/mocks"
	"github.com/gerfey/messenger/transport/sync"
)

func TestNewTransportFactory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	mockLocator := mocks.NewMockBusLocator(ctrl)

	factory := sync.NewTransportFactory(logger, mockLocator)

	assert.NotNil(t, factory)
	assert.IsType(t, &sync.Factory{}, factory)
}

func TestFactory_Supports(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	mockLocator := mocks.NewMockBusLocator(ctrl)
	factory := sync.NewTransportFactory(logger, mockLocator)

	testCases := []struct {
		name     string
		dsn      string
		expected bool
	}{
		{
			name:     "supports sync dsn",
			dsn:      "sync://",
			expected: true,
		},
		{
			name:     "does not support amqp dsn",
			dsn:      "amqp://guest:guest@localhost:5672/",
			expected: false,
		},
		{
			name:     "does not support in-memory dsn",
			dsn:      "in-memory://test",
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
			result := factory.Supports(tc.dsn)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFactory_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := slog.Default()
	mockLocator := mocks.NewMockBusLocator(ctrl)
	factory := sync.NewTransportFactory(logger, mockLocator)

	name := "test-sync"
	dsn := "sync://"
	options := config.OptionsConfig{}

	transport, err := factory.Create(name, dsn, options)

	require.NoError(t, err)
	assert.NotNil(t, transport)
	assert.IsType(t, &sync.Transport{}, transport)
}
