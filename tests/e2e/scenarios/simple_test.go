package scenarios_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/builder"
	"github.com/gerfey/messenger/config"
	"github.com/gerfey/messenger/tests/e2e/fixtures/handlers"
	"github.com/gerfey/messenger/tests/e2e/helpers"
	testHelpers "github.com/gerfey/messenger/tests/helpers"
)

func TestE2E_Simple_HandlerOnly(t *testing.T) {
	logger, fakeHandler := testHelpers.NewFakeLogger()

	cfg, err := config.LoadConfig("../config/handler_only.yaml")
	require.NoError(t, err)

	b := builder.NewBuilder(cfg, logger)

	testHandler := handlers.NewE2ETestHandler()
	err = b.RegisterHandler(testHandler)
	require.NoError(t, err)

	b.RegisterMiddleware("debug", helpers.NewDebugMiddleware("debug", logger))

	messenger, err := b.Build()
	require.NoError(t, err)

	bus, err := messenger.GetDefaultBus()
	require.NoError(t, err)

	env, err := bus.Dispatch(t.Context(), &testHelpers.TestMessage{Content: "test"})
	require.NoError(t, err)
	require.NotNil(t, env)

	require.Equal(t, int64(1), testHandler.GetCallCount())
	require.True(t, fakeHandler.HasMessage(slog.LevelDebug, "processing message"))
	require.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message handled successfully"))
	require.True(t, fakeHandler.HasMessage(slog.LevelWarn, "no transports configured for message"))
}
