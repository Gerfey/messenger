package sync_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gopkg.in/yaml.v3"

	"github.com/gerfey/messenger/core/serializer"

	"github.com/gerfey/messenger/tests/mocks"
	"github.com/gerfey/messenger/transport/sync"
)

func TestNewTransportFactory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLocator := mocks.NewMockBusLocator(ctrl)

	factory := sync.NewTransportFactory(mockLocator)

	assert.NotNil(t, factory)
	assert.IsType(t, &sync.TransportFactory{}, factory)
}

func TestFactory_Supports(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLocator := mocks.NewMockBusLocator(ctrl)
	factory := sync.NewTransportFactory(mockLocator)

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

	mockLocator := mocks.NewMockBusLocator(ctrl)
	factory := sync.NewTransportFactory(mockLocator)

	name := "test-sync"
	dsn := "sync://"
	options := map[string]any{}
	mockResolver := mocks.NewMockTypeResolver(ctrl)
	ser := serializer.NewSerializer(mockResolver)

	optionsBytes, err := yaml.Marshal(options)
	require.NoError(t, err)

	transport, err := factory.Create(name, dsn, optionsBytes, ser)

	require.NoError(t, err)
	assert.NotNil(t, transport)
	assert.IsType(t, &sync.Transport{}, transport)
}
