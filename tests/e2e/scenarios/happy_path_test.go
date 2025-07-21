package scenarios

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/builder"
	"github.com/gerfey/messenger/config"
	"github.com/gerfey/messenger/tests/e2e/fixtures/handlers"
	"github.com/gerfey/messenger/tests/e2e/helpers"
	testHelpers "github.com/gerfey/messenger/tests/helpers"
)

func TestE2E_HappyPath_SingleHandlerSingleTransport(t *testing.T) {
	ctx := context.Background()

	logger, fakeHandler := testHelpers.NewFakeLogger()

	cfg, err := config.LoadConfig("../config/e2e.yaml")
	require.NoError(t, err)

	b := builder.NewBuilder(cfg, logger)

	testHandler := handlers.NewE2ETestHandler()
	err = b.RegisterHandler(testHandler)
	require.NoError(t, err)

	b.RegisterMiddleware("debug", helpers.NewDebugMiddleware("debug", logger))

	messenger, err := b.Build()
	require.NoError(t, err)

	go func() {
		if runErr := messenger.Run(ctx); runErr != nil {
			t.Logf("Messenger run error: %v", runErr)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	testMessage := &testHelpers.TestMessage{Content: "Hello E2E World!"}

	bus, err := messenger.GetDefaultBus()
	require.NoError(t, err)

	env, err := bus.Dispatch(ctx, testMessage)

	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, err)
	assert.NotNil(t, env)

	assert.Equal(t, int64(1), testHandler.GetCallCount())
	assert.Equal(t, testMessage, testHandler.GetLastMessage())

	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "sending message to transports"))
	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message sent successfully"))
	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "processing message"))
	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message handled successfully"))
}

func TestE2E_HappyPath_MultipleHandlers(t *testing.T) {
	ctx := context.Background()

	logger, fakeHandler := testHelpers.NewFakeLogger()

	cfg, err := config.LoadConfig("../config/e2e.yaml")
	require.NoError(t, err)

	b := builder.NewBuilder(cfg, logger)

	handler1 := handlers.NewE2ETestHandler()
	handler2 := handlers.NewE2ETestHandler()
	
	err = b.RegisterHandler(handler1)
	require.NoError(t, err)
	err = b.RegisterHandler(handler2)
	require.NoError(t, err)

	b.RegisterMiddleware("debug", helpers.NewDebugMiddleware("debug", logger))

	messenger, err := b.Build()
	require.NoError(t, err)

	go func() {
		if runErr := messenger.Run(ctx); runErr != nil {
			t.Logf("Messenger run error: %v", runErr)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	testMessage := &testHelpers.TestMessage{Content: "Multiple handlers test"}

	bus, err := messenger.GetDefaultBus()
	require.NoError(t, err)

	env, err := bus.Dispatch(ctx, testMessage)

	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, err)
	assert.NotNil(t, env)

	assert.Equal(t, int64(1), handler1.GetCallCount())
	assert.Equal(t, int64(1), handler2.GetCallCount())

	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "sending message to transports"))
	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message sent successfully"))
	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "processing message"))
	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message handled successfully"))
}

func TestE2E_HappyPath_MultipleTransports(t *testing.T) {
	ctx := context.Background()

	logger, fakeHandler := testHelpers.NewFakeLogger()

	cfg, err := config.LoadConfig("../config/multiple_transports.yaml")
	require.NoError(t, err)

	b := builder.NewBuilder(cfg, logger)

	testHandler := handlers.NewE2ETestHandler()
	err = b.RegisterHandler(testHandler)
	require.NoError(t, err)

	b.RegisterMiddleware("debug", helpers.NewDebugMiddleware("debug", logger))

	messenger, err := b.Build()
	require.NoError(t, err)

	go func() {
		if runErr := messenger.Run(ctx); runErr != nil {
			t.Logf("Messenger run error: %v", runErr)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	testMessage := &testHelpers.TestMessage{Content: "Multiple transports test"}

	bus, err := messenger.GetDefaultBus()
	require.NoError(t, err)

	env, err := bus.Dispatch(ctx, testMessage)

	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, err)
	assert.NotNil(t, env)

	assert.Equal(t, int64(1), testHandler.GetCallCount())

	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "sending message to transports"))
	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message sent successfully"))
	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "processing message"))
	assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message handled successfully"))
}
