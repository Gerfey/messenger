package implementation_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/gerfey/messenger/core/middleware/implementation"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/handler"
	"github.com/gerfey/messenger/core/stamps"
	"github.com/gerfey/messenger/tests/helpers"
)

func TestNewHandleMessageMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	locator := handler.NewHandlerLocator()
	logger, _ := helpers.NewFakeLogger()

	middleware := implementation.NewHandleMessageMiddleware(logger, locator)

	require.NotNil(t, middleware)
	require.IsType(t, &implementation.HandleMessageMiddleware{}, middleware)
}

func TestHandleMessageMiddleware_Handle(t *testing.T) {
	t.Run("process message even if envelope has SentStamp", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		locator := handler.NewHandlerLocator()
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewHandleMessageMiddleware(logger, locator)

		testHandler := &helpers.TestMessageHandler{}
		err := locator.Register(testHandler)
		require.NoError(t, err)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg).WithStamp(stamps.SentStamp{SenderName: "test"})

		nextCalled := false
		next := func(_ context.Context, env api.Envelope) (api.Envelope, error) {
			nextCalled = true

			return env, nil
		}

		result, err := middleware.Handle(t.Context(), env, next)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.True(t, nextCalled)
		require.Equal(t, 1, testHandler.CallCount)
		require.True(t, fakeHandler.HasMessage(slog.LevelDebug, "processing message"))
	})

	t.Run("return error when no handlers registered", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		locator := handler.NewHandlerLocator()
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewHandleMessageMiddleware(logger, locator)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)

		next := func(_ context.Context, env api.Envelope) (api.Envelope, error) {
			return env, nil
		}

		result, err := middleware.Handle(t.Context(), env, next)

		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "no handlers registered for message type")

		require.True(t, fakeHandler.HasMessage(slog.LevelWarn, "no handlers registered for message type"))
	})

	t.Run("successfully handle message with single handler", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		locator := handler.NewHandlerLocator()
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewHandleMessageMiddleware(logger, locator)

		testHandler := &helpers.TestMessageHandler{}
		err := locator.Register(testHandler)
		require.NoError(t, err)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)

		nextCalled := false
		next := func(_ context.Context, env api.Envelope) (api.Envelope, error) {
			nextCalled = true

			return env, nil
		}

		result, err := middleware.Handle(t.Context(), env, next)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.True(t, nextCalled)
		require.Equal(t, 1, testHandler.CallCount)
		require.True(t, fakeHandler.HasMessage(slog.LevelDebug, "processing message"))
		require.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message handled successfully"))
	})

	t.Run("successfully handle message with single handler", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		locator := handler.NewHandlerLocator()
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewHandleMessageMiddleware(logger, locator)

		testHandler := &helpers.TestMessageHandler{}
		err := locator.Register(testHandler)
		require.NoError(t, err)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)

		nextCalled := false
		next := func(_ context.Context, env api.Envelope) (api.Envelope, error) {
			nextCalled = true

			return env, nil
		}

		result, err := middleware.Handle(t.Context(), env, next)

		require.NoError(t, err)
		require.NotNil(t, result)
		handledStamps := envelope.StampsOf[stamps.HandledStamp](result)
		require.NotEmpty(t, handledStamps)
		require.True(t, nextCalled)
		require.Equal(t, 1, testHandler.CallCount)
		require.True(t, fakeHandler.HasMessage(slog.LevelDebug, "processing message"))
		require.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message handled successfully"))
	})

	t.Run("successfully handle message with single handler", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		locator := handler.NewHandlerLocator()
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewHandleMessageMiddleware(logger, locator)

		testHandler := &helpers.TestMessageHandler{}
		err := locator.Register(testHandler)
		require.NoError(t, err)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)

		nextCalled := false
		next := func(_ context.Context, env api.Envelope) (api.Envelope, error) {
			nextCalled = true

			return env, nil
		}

		result, err := middleware.Handle(t.Context(), env, next)

		require.NoError(t, err)
		require.NotNil(t, result)
		handledStamps := envelope.StampsOf[stamps.HandledStamp](result)
		require.NotEmpty(t, handledStamps)
		require.True(t, nextCalled)
		require.Equal(t, 1, testHandler.CallCount)

		require.True(t, fakeHandler.HasMessage(slog.LevelDebug, "processing message"))
		require.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message handled successfully"))
	})

	t.Run("successfully handle message with multiple handlers", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		locator := handler.NewHandlerLocator()
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewHandleMessageMiddleware(logger, locator)

		testHandler1 := &helpers.TestMessageHandler{}
		err := locator.Register(testHandler1)
		require.NoError(t, err)

		testHandler2 := &helpers.TestMessageHandler{}
		err = locator.Register(testHandler2)
		require.NoError(t, err)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)

		nextCalled := false
		next := func(_ context.Context, env api.Envelope) (api.Envelope, error) {
			nextCalled = true

			return env, nil
		}

		result, err := middleware.Handle(t.Context(), env, next)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.True(t, nextCalled)
		require.Equal(t, 1, testHandler1.CallCount)
		require.Equal(t, 1, testHandler2.CallCount)
		require.True(t, fakeHandler.HasMessage(slog.LevelDebug, "processing message"))
		require.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message handled successfully"))
	})

	t.Run("handle error from handler", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		locator := handler.NewHandlerLocator()
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewHandleMessageMiddleware(logger, locator)

		errorHandler := &helpers.ErrorTestMessageHandler{
			Error: errors.New("handler error"),
		}
		err := locator.Register(errorHandler)
		require.NoError(t, err)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)

		next := func(_ context.Context, env api.Envelope) (api.Envelope, error) {
			return env, nil
		}

		result, err := middleware.Handle(t.Context(), env, next)

		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "handler error")
		require.Equal(t, 1, errorHandler.CallCount)

		require.True(t, fakeHandler.HasMessage(slog.LevelError, "handler failed"))
	})

	t.Run("handle handler with result", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		locator := handler.NewHandlerLocator()
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewHandleMessageMiddleware(logger, locator)

		resultHandler := &helpers.ResultTestMessageHandler{
			Result: "test result",
		}
		err := locator.Register(resultHandler)
		require.NoError(t, err)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)

		nextCalled := false
		next := func(_ context.Context, env api.Envelope) (api.Envelope, error) {
			nextCalled = true

			return env, nil
		}

		result, err := middleware.Handle(t.Context(), env, next)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.True(t, nextCalled)
		require.Equal(t, 1, resultHandler.CallCount)

		handledStamps := envelope.StampsOf[stamps.HandledStamp](result)
		require.Len(t, handledStamps, 1)
		require.Equal(t, "test result", handledStamps[0].Result)
		require.NotEmpty(t, handledStamps[0].Handler)

		require.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message handled successfully"))
	})
}
