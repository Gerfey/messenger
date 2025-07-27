package transport_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/transport"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/tests/helpers"
)

func TestNewManager(t *testing.T) {
	t.Run("create manager with all parameters", func(t *testing.T) {
		handler := func(_ context.Context, _ api.Envelope) error { return nil }
		logger := slog.Default()

		manager := transport.NewManager(logger, handler, nil)

		require.NotNil(t, manager)
		assert.IsType(t, &transport.Manager{}, manager)
	})

	t.Run("create manager with nil parameters", func(t *testing.T) {
		manager := transport.NewManager(nil, nil, nil)

		require.NotNil(t, manager)
		assert.IsType(t, &transport.Manager{}, manager)
	})
}

func TestManager_HasTransport(t *testing.T) {
	t.Run("has transport by name", func(t *testing.T) {
		handler := func(_ context.Context, _ api.Envelope) error { return nil }
		logger := slog.Default()
		manager := transport.NewManager(logger, handler, nil)

		tr := &helpers.TestTransport{TransportName: "test-transport"}
		manager.AddTransport(tr)

		assert.True(t, manager.HasTransport("test-transport"))
		assert.False(t, manager.HasTransport("non-existing"))
		assert.False(t, manager.HasTransport(""))
	})

	t.Run("has transport with multiple transports", func(t *testing.T) {
		handler := func(_ context.Context, _ api.Envelope) error { return nil }
		logger := slog.Default()
		manager := transport.NewManager(logger, handler, nil)

		transport1 := &helpers.TestTransport{TransportName: "transport1"}
		transport2 := &helpers.TestTransport{TransportName: "transport2"}
		transport3 := &helpers.TestTransport{TransportName: "transport3"}

		manager.AddTransport(transport1)
		manager.AddTransport(transport2)
		manager.AddTransport(transport3)

		assert.True(t, manager.HasTransport("transport1"))
		assert.True(t, manager.HasTransport("transport2"))
		assert.True(t, manager.HasTransport("transport3"))
		assert.False(t, manager.HasTransport("transport4"))
	})

	t.Run("has transport with no transports", func(t *testing.T) {
		handler := func(_ context.Context, _ api.Envelope) error { return nil }
		logger := slog.Default()
		manager := transport.NewManager(logger, handler, nil)

		assert.False(t, manager.HasTransport("any-transport"))
	})

	t.Run("has transport with nil transport", func(t *testing.T) {
		handler := func(_ context.Context, _ api.Envelope) error { return nil }
		logger := slog.Default()
		manager := transport.NewManager(logger, handler, nil)

		manager.AddTransport(nil)

		assert.Panics(t, func() {
			manager.HasTransport("any-name")
		})
	})
}

func TestManager_Integration(t *testing.T) {
	t.Run("manager with handler error", func(t *testing.T) {
		expectedError := errors.New("handler error")
		handler := func(_ context.Context, _ api.Envelope) error {
			return expectedError
		}

		logger := slog.Default()
		manager := transport.NewManager(logger, handler, nil)

		tr := &helpers.TestTransport{
			TransportName: "error-transport",
		}
		manager.AddTransport(tr)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		go manager.Start(ctx, nil)
		time.Sleep(10 * time.Millisecond)

		manager.Stop()
	})
}
