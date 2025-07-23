package e2e_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/tests/fixtures/handlers"

	"github.com/gerfey/messenger/builder"
	"github.com/gerfey/messenger/config"
	testHelpers "github.com/gerfey/messenger/tests/helpers"
)

func TestE2E_HappyPath_SingleHandlerSingleTransport(t *testing.T) {
	logger, fakeHandler := testHelpers.NewFakeLogger()

	cfg, err := config.LoadConfig("../fixtures/configs/e2e.yaml")
	require.NoError(t, err)

	b := builder.NewBuilder(cfg, logger)

	testHandler := handlers.NewE2ETestHandler()
	err = b.RegisterHandler(testHandler)
	require.NoError(t, err)

	b.RegisterMiddleware("debug", testHelpers.NewDebugMiddleware("debug", logger))

	messenger, err := b.Build()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	go func() {
		if runErr := messenger.Run(ctx); runErr != nil && !errors.Is(runErr, context.Canceled) {
			t.Logf("Messenger run error: %v", runErr)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	testMessage := &testHelpers.TestMessage{Content: "Hello E2E World!"}

	bus, err := messenger.GetDefaultBus()
	require.NoError(t, err)

	env, err := bus.Dispatch(t.Context(), testMessage)

	time.Sleep(100 * time.Millisecond)

	require.NoError(t, err)
	assert.NotNil(t, env)

	assert.Equal(t, int64(1), testHandler.GetCallCount())
	assert.Equal(t, testMessage, testHandler.GetLastMessage())

	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "sending message to transports"))
	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message sent successfully"))
	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "processing message"))
	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message handled successfully"))
}

func TestE2E_HappyPath_MultipleHandlers(t *testing.T) {
	logger, fakeHandler := testHelpers.NewFakeLogger()

	cfg, err := config.LoadConfig("../fixtures/configs/e2e.yaml")
	require.NoError(t, err)

	b := builder.NewBuilder(cfg, logger)

	handler1 := handlers.NewE2ETestHandler()
	handler2 := handlers.NewE2ETestHandler()

	err = b.RegisterHandler(handler1)
	require.NoError(t, err)
	err = b.RegisterHandler(handler2)
	require.NoError(t, err)

	b.RegisterMiddleware("debug", testHelpers.NewDebugMiddleware("debug", logger))

	messenger, err := b.Build()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	go func() {
		if runErr := messenger.Run(ctx); runErr != nil && !errors.Is(runErr, context.Canceled) {
			t.Logf("Messenger run error: %v", runErr)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	testMessage := &testHelpers.TestMessage{Content: "Multiple handlers test"}

	bus, err := messenger.GetDefaultBus()
	require.NoError(t, err)

	env, err := bus.Dispatch(t.Context(), testMessage)

	time.Sleep(100 * time.Millisecond)

	require.NoError(t, err)
	assert.NotNil(t, env)

	assert.Equal(t, int64(1), handler1.GetCallCount())
	assert.Equal(t, int64(1), handler2.GetCallCount())

	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "sending message to transports"))
	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message sent successfully"))
	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "processing message"))
	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message handled successfully"))
}

func TestE2E_HappyPath_MultipleTransports(t *testing.T) {
	logger, fakeHandler := testHelpers.NewFakeLogger()

	cfg, err := config.LoadConfig("../fixtures/configs/multiple_transports.yaml")
	require.NoError(t, err)

	b := builder.NewBuilder(cfg, logger)

	testHandler := handlers.NewE2ETestHandler()
	err = b.RegisterHandler(testHandler)
	require.NoError(t, err)

	b.RegisterMiddleware("debug", testHelpers.NewDebugMiddleware("debug", logger))

	messenger, err := b.Build()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	go func() {
		if runErr := messenger.Run(ctx); runErr != nil && !errors.Is(runErr, context.Canceled) {
			t.Logf("Messenger run error: %v", runErr)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	testMessage := &testHelpers.TestMessage{Content: "Multiple transports test"}

	bus, err := messenger.GetDefaultBus()
	require.NoError(t, err)

	env, err := bus.Dispatch(t.Context(), testMessage)

	time.Sleep(100 * time.Millisecond)

	require.NoError(t, err)
	assert.NotNil(t, env)

	assert.Equal(t, int64(1), testHandler.GetCallCount())

	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "sending message to transports"))
	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message sent successfully"))
	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "processing message"))
	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message handled successfully"))
}
