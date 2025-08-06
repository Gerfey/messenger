package transport_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/builder"
	"github.com/gerfey/messenger/core/serializer"

	"github.com/gerfey/messenger/transport"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/config"
	"github.com/gerfey/messenger/tests/helpers"
)

type mockFactory struct {
	supportedDSN     string
	createdTransport api.Transport
	createError      error
}

func (f *mockFactory) Supports(dsn string) bool {
	return dsn == f.supportedDSN
}

func (f *mockFactory) Create(name, _ string, _ []byte, _ api.Serializer) (api.Transport, error) {
	if f.createError != nil {
		return nil, f.createError
	}

	if f.createdTransport != nil {
		return f.createdTransport, nil
	}

	return &helpers.TestTransport{TransportName: name}, nil
}

func TestNewFactoryChain(t *testing.T) {
	t.Run("create chain without factories", func(t *testing.T) {
		chain := transport.NewFactoryChain()

		require.NotNil(t, chain)
		assert.IsType(t, &transport.FactoryChain{}, chain)
		assert.Empty(t, chain.Factories())
	})

	t.Run("create chain with single factory", func(t *testing.T) {
		factory := &mockFactory{supportedDSN: "test://"}
		chain := transport.NewFactoryChain(factory)

		require.NotNil(t, chain)
		factories := chain.Factories()
		assert.Len(t, factories, 1)
		assert.Same(t, factory, factories[0])
	})

	t.Run("create chain with multiple factories", func(t *testing.T) {
		factory1 := &mockFactory{supportedDSN: "test1://"}
		factory2 := &mockFactory{supportedDSN: "test2://"}
		factory3 := &mockFactory{supportedDSN: "test3://"}

		chain := transport.NewFactoryChain(factory1, factory2, factory3)

		require.NotNil(t, chain)
		factories := chain.Factories()
		assert.Len(t, factories, 3)
		assert.Same(t, factory1, factories[0])
		assert.Same(t, factory2, factories[1])
		assert.Same(t, factory3, factories[2])
	})
}

func TestFactoryChain_CreateTransport(t *testing.T) {
	t.Run("create transport with matching factory", func(t *testing.T) {
		expectedTransport := &helpers.TestTransport{TransportName: "test-transport"}
		factory := &mockFactory{
			supportedDSN:     "test://localhost",
			createdTransport: expectedTransport,
		}
		chain := transport.NewFactoryChain(factory)

		transportConfig := config.TransportConfig{
			DSN:     "test://localhost",
			Options: map[string]any{},
		}

		resolver := builder.NewResolver()
		ser := serializer.NewSerializer(resolver)

		tr, err := chain.CreateTransport("test-transport", transportConfig, ser)

		require.NoError(t, err)
		assert.Same(t, expectedTransport, tr)
	})

	t.Run("create transport with first matching factory", func(t *testing.T) {
		expectedTransport := &helpers.TestTransport{TransportName: "first-transport"}
		factory1 := &mockFactory{
			supportedDSN:     "test://localhost",
			createdTransport: expectedTransport,
		}
		factory2 := &mockFactory{
			supportedDSN: "test://localhost",
		}
		chain := transport.NewFactoryChain(factory1, factory2)

		transportConfig := config.TransportConfig{
			DSN:     "test://localhost",
			Options: map[string]any{},
		}

		resolver := builder.NewResolver()
		ser := serializer.NewSerializer(resolver)

		tr, err := chain.CreateTransport("first-transport", transportConfig, ser)

		require.NoError(t, err)
		assert.Same(t, expectedTransport, tr)
	})

	t.Run("create transport with no matching factory", func(t *testing.T) {
		factory1 := &mockFactory{supportedDSN: "amqp://"}
		factory2 := &mockFactory{supportedDSN: "redis://"}
		chain := transport.NewFactoryChain(factory1, factory2)

		transportConfig := config.TransportConfig{
			DSN:     "unknown://localhost",
			Options: map[string]any{},
		}

		resolver := builder.NewResolver()
		ser := serializer.NewSerializer(resolver)

		tr, err := chain.CreateTransport("unknown-transport", transportConfig, ser)

		require.Error(t, err)
		assert.Nil(t, tr)
		assert.Contains(t, err.Error(), "no transport factory supports DSN")
		assert.Contains(t, err.Error(), "unknown://localhost")
		assert.Contains(t, err.Error(), "unknown-transport")
	})

	t.Run("create transport with factory error", func(t *testing.T) {
		expectedError := assert.AnError
		factory := &mockFactory{
			supportedDSN: "test://localhost",
			createError:  expectedError,
		}
		chain := transport.NewFactoryChain(factory)

		transportConfig := config.TransportConfig{
			DSN:     "test://localhost",
			Options: map[string]any{},
		}

		resolver := builder.NewResolver()
		ser := serializer.NewSerializer(resolver)

		tr, err := chain.CreateTransport("error-transport", transportConfig, ser)

		require.Error(t, err)
		assert.Same(t, expectedError, err)
		assert.Nil(t, tr)
	})

	t.Run("create transport with empty chain", func(t *testing.T) {
		chain := transport.NewFactoryChain()

		transportConfig := config.TransportConfig{
			DSN:     "test://localhost",
			Options: map[string]any{},
		}

		resolver := builder.NewResolver()
		ser := serializer.NewSerializer(resolver)

		tr, err := chain.CreateTransport("test-transport", transportConfig, ser)

		require.Error(t, err)
		assert.Nil(t, tr)
		assert.Contains(t, err.Error(), "no transport factory supports DSN")
	})

	t.Run("create transport with multiple factories - second matches", func(t *testing.T) {
		expectedTransport := &helpers.TestTransport{TransportName: "second-transport"}
		factory1 := &mockFactory{supportedDSN: "amqp://"}
		factory2 := &mockFactory{
			supportedDSN:     "redis://localhost",
			createdTransport: expectedTransport,
		}
		factory3 := &mockFactory{supportedDSN: "kafka://"}

		chain := transport.NewFactoryChain(factory1, factory2, factory3)

		transportConfig := config.TransportConfig{
			DSN:     "redis://localhost",
			Options: map[string]any{},
		}

		resolver := builder.NewResolver()
		ser := serializer.NewSerializer(resolver)

		tr, err := chain.CreateTransport("second-transport", transportConfig, ser)

		require.NoError(t, err)
		assert.Same(t, expectedTransport, tr)
	})
}

func TestFactoryChain_Factories(t *testing.T) {
	t.Run("get factories from empty chain", func(t *testing.T) {
		chain := transport.NewFactoryChain()

		factories := chain.Factories()
		assert.Empty(t, factories)
	})

	t.Run("get factories from chain with factories", func(t *testing.T) {
		factory1 := &mockFactory{supportedDSN: "test1://"}
		factory2 := &mockFactory{supportedDSN: "test2://"}

		chain := transport.NewFactoryChain(factory1, factory2)

		factories := chain.Factories()
		assert.Len(t, factories, 2)
		assert.Same(t, factory1, factories[0])
		assert.Same(t, factory2, factories[1])
	})

	t.Run("factories returns slice copy", func(t *testing.T) {
		factory := &mockFactory{supportedDSN: "test://"}
		chain := transport.NewFactoryChain(factory)

		factories1 := chain.Factories()
		factories2 := chain.Factories()

		assert.ElementsMatch(t, factories1, factories2)

		assert.NotSame(t, &factories1, &factories2)
	})
}

func TestFactoryChain_Integration(t *testing.T) {
	t.Run("full workflow with multiple factories and transports", func(t *testing.T) {
		amqpFactory := &mockFactory{supportedDSN: "amqp://localhost"}
		redisFactory := &mockFactory{supportedDSN: "redis://localhost"}
		inMemoryFactory := &mockFactory{supportedDSN: "memory://"}

		chain := transport.NewFactoryChain(amqpFactory, redisFactory, inMemoryFactory)

		amqpConfig := config.TransportConfig{DSN: "amqp://localhost"}
		redisConfig := config.TransportConfig{DSN: "redis://localhost"}
		memoryConfig := config.TransportConfig{DSN: "memory://"}
		unknownConfig := config.TransportConfig{DSN: "unknown://localhost"}

		resolver := builder.NewResolver()
		ser := serializer.NewSerializer(resolver)

		amqpTransport, err := chain.CreateTransport("amqp", amqpConfig, ser)
		require.NoError(t, err)
		require.NotNil(t, amqpTransport)
		assert.Equal(t, "amqp", amqpTransport.Name())

		redisTransport, err := chain.CreateTransport("redis", redisConfig, ser)
		require.NoError(t, err)
		require.NotNil(t, redisTransport)
		assert.Equal(t, "redis", redisTransport.Name())

		memoryTransport, err := chain.CreateTransport("memory", memoryConfig, ser)
		require.NoError(t, err)
		require.NotNil(t, memoryTransport)
		assert.Equal(t, "memory", memoryTransport.Name())

		unknownTransport, err := chain.CreateTransport("unknown", unknownConfig, ser)
		require.Error(t, err)
		assert.Nil(t, unknownTransport)

		assert.NotSame(t, amqpTransport, redisTransport)
		assert.NotSame(t, redisTransport, memoryTransport)
		assert.NotSame(t, amqpTransport, memoryTransport)
	})
}
