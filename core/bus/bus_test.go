package bus_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/core/bus"
	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/stamps"
	"github.com/gerfey/messenger/tests/helpers"
)

func TestNewBus(t *testing.T) {
	t.Run("create bus without middleware", func(t *testing.T) {
		messageBus := bus.NewBus()

		require.NotNil(t, messageBus)
		assert.IsType(t, &bus.Bus{}, messageBus)
	})

	t.Run("create bus with single middleware", func(t *testing.T) {
		middleware := &helpers.TestMiddleware{}
		messageBus := bus.NewBus(middleware)

		require.NotNil(t, messageBus)
		assert.NotNil(t, messageBus)
	})

	t.Run("create bus with multiple middleware", func(t *testing.T) {
		middleware1 := &helpers.TestMiddleware{}
		middleware2 := &helpers.TestMiddleware{}
		messageBus := bus.NewBus(middleware1, middleware2)

		require.NotNil(t, messageBus)
		assert.NotNil(t, messageBus)
	})
}

func TestBus_Dispatch(t *testing.T) {
	t.Run("dispatch simple message without middleware", func(t *testing.T) {
		messageBus := bus.NewBus()
		msg := &helpers.TestMessage{Content: "test"}

		env, err := messageBus.Dispatch(t.Context(), msg)

		require.NoError(t, err)
		assert.NotNil(t, env)
		assert.Equal(t, msg, env.Message())
	})

	t.Run("dispatch envelope message without middleware", func(t *testing.T) {
		messageBus := bus.NewBus()
		originalMsg := &helpers.TestMessage{Content: "test"}
		originalEnv := envelope.NewEnvelope(originalMsg)

		env, err := messageBus.Dispatch(t.Context(), originalEnv)

		require.NoError(t, err)
		assert.NotNil(t, env)
		assert.Equal(t, originalMsg, env.Message())
	})

	t.Run("dispatch message with stamps", func(t *testing.T) {
		messageBus := bus.NewBus()
		msg := &helpers.TestMessage{Content: "test"}
		busNameStamp := &stamps.BusNameStamp{Name: "test-bus"}
		sentStamp := &stamps.SentStamp{Transport: "test-transport"}

		env, err := messageBus.Dispatch(t.Context(), msg, busNameStamp, sentStamp)

		require.NoError(t, err)
		assert.NotNil(t, env)
		assert.Equal(t, msg, env.Message())

		allStamps := env.Stamps()
		assert.Len(t, allStamps, 2)

		var foundBusNameStamp bool
		var foundSentStamp bool

		for _, s := range allStamps {
			switch stamp := s.(type) {
			case *stamps.BusNameStamp:
				assert.Equal(t, "test-bus", stamp.Name)
				foundBusNameStamp = true
			case *stamps.SentStamp:
				assert.Equal(t, "test-transport", stamp.Transport)
				foundSentStamp = true
			}
		}

		assert.True(t, foundBusNameStamp, "BusNameStamp не найден")
		assert.True(t, foundSentStamp, "SentStamp не найден")
	})

	t.Run("dispatch message with single middleware", func(t *testing.T) {
		middleware := &helpers.TestMiddleware{}
		messageBus := bus.NewBus(middleware)
		msg := &helpers.TestMessage{Content: "test"}

		env, err := messageBus.Dispatch(t.Context(), msg)

		require.NoError(t, err)
		assert.NotNil(t, env)
		assert.Equal(t, msg, env.Message())
		assert.True(t, middleware.Called)
	})

	t.Run("dispatch message with multiple middleware", func(t *testing.T) {
		middleware1 := &helpers.TestMiddleware{}
		middleware2 := &helpers.TestMiddleware{}
		messageBus := bus.NewBus(middleware1, middleware2)
		msg := &helpers.TestMessage{Content: "test"}

		env, err := messageBus.Dispatch(t.Context(), msg)

		require.NoError(t, err)
		assert.NotNil(t, env)
		assert.Equal(t, msg, env.Message())
		assert.True(t, middleware1.Called)
		assert.True(t, middleware2.Called)
	})

	t.Run("dispatch message with middleware error", func(t *testing.T) {
		expectedErr := errors.New("middleware error")
		middleware := &helpers.ErrorMiddleware{Error: expectedErr}
		messageBus := bus.NewBus(middleware)
		msg := &helpers.TestMessage{Content: "test"}

		env, err := messageBus.Dispatch(t.Context(), msg)

		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, env)
	})

	t.Run("dispatch with context cancellation", func(t *testing.T) {
		middleware := &helpers.ContextMiddleware{}
		messageBus := bus.NewBus(middleware)
		msg := &helpers.TestMessage{Content: "test"}

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		env, err := messageBus.Dispatch(ctx, msg)

		require.Error(t, err)
		assert.Equal(t, context.Canceled, err)
		assert.Nil(t, env)
	})
}

func TestMiddlewareExecution(t *testing.T) {
	t.Run("middleware execution order", func(t *testing.T) {
		var executionOrder []string
		middleware1 := &helpers.OrderedMiddleware{Name: "first", ExecutionOrder: &executionOrder}
		middleware2 := &helpers.OrderedMiddleware{Name: "second", ExecutionOrder: &executionOrder}
		middleware3 := &helpers.OrderedMiddleware{Name: "third", ExecutionOrder: &executionOrder}

		messageBus := bus.NewBus(middleware1, middleware2, middleware3)
		msg := &helpers.TestMessage{Content: "test"}

		_, err := messageBus.Dispatch(t.Context(), msg)
		require.NoError(t, err)

		assert.Equal(t, []string{"first", "second", "third"}, executionOrder)
	})
}
