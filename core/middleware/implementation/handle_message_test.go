package implementation_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/gerfey/messenger/core/middleware/implementation"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

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

	middleware := implementation.NewHandleMessageMiddleware(locator, logger)

	assert.NotNil(t, middleware)
	assert.IsType(t, &implementation.HandleMessageMiddleware{}, middleware)
}

func TestHandleMessageMiddleware_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("skip processing if envelope has SentStamp", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		locator := handler.NewHandlerLocator()
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewHandleMessageMiddleware(locator, logger)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg).WithStamp(stamps.SentStamp{Transport: "test"})

		nextCalled := false
		next := func(ctx context.Context, env api.Envelope) (api.Envelope, error) {
			nextCalled = true
			return env, nil
		}

		result, err := middleware.Handle(ctx, env, next)

		assert.NoError(t, err)
		assert.Equal(t, env, result)
		assert.False(t, nextCalled)
		assert.Equal(t, 0, fakeHandler.Count())
	})

	t.Run("return error when no handlers registered", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		locator := handler.NewHandlerLocator()
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewHandleMessageMiddleware(locator, logger)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)

		next := func(ctx context.Context, env api.Envelope) (api.Envelope, error) {
			return env, nil
		}

		result, err := middleware.Handle(ctx, env, next)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "no handlers registered for message type")

		assert.True(t, fakeHandler.HasMessage(slog.LevelWarn, "no handlers registered for message type"))
	})

	t.Run("successfully handle message with single handler", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		locator := handler.NewHandlerLocator()
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewHandleMessageMiddleware(locator, logger)

		testHandler := &helpers.TestMessageHandler{}
		err := locator.Register(testHandler)
		assert.NoError(t, err)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)

		nextCalled := false
		var nextEnv api.Envelope
		next := func(ctx context.Context, env api.Envelope) (api.Envelope, error) {
			nextCalled = true
			nextEnv = env
			return env, nil
		}

		result, err := middleware.Handle(ctx, env, next)

		assert.NoError(t, err)
		assert.Equal(t, result, nextEnv)
		assert.True(t, nextCalled)
		assert.Equal(t, 1, testHandler.CallCount)

		handledStamps := envelope.StampsOf[stamps.HandledStamp](result)
		assert.Len(t, handledStamps, 1)
		assert.NotEmpty(t, handledStamps[0].Handler)

		assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "processing message"))
		assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message handled successfully"))
	})

	t.Run("successfully handle message with multiple handlers", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		locator := handler.NewHandlerLocator()
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewHandleMessageMiddleware(locator, logger)

		handler1 := &helpers.TestMessageHandler{}
		handler2 := &helpers.AnotherTestMessageHandler{}

		err := locator.Register(handler1)
		assert.NoError(t, err)
		err = locator.Register(handler2)
		assert.NoError(t, err)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)

		nextCalled := false
		var nextEnv api.Envelope
		next := func(ctx context.Context, env api.Envelope) (api.Envelope, error) {
			nextCalled = true
			nextEnv = env
			return env, nil
		}

		result, err := middleware.Handle(ctx, env, next)

		assert.NoError(t, err)
		assert.Equal(t, result, nextEnv)
		assert.True(t, nextCalled)
		assert.Equal(t, 1, handler1.CallCount)
		assert.Equal(t, 1, handler2.CallCount)

		handledStamps := envelope.StampsOf[stamps.HandledStamp](result)
		assert.Len(t, handledStamps, 2)

		assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "processing message"))
		successCount := 0
		entries := fakeHandler.GetEntriesByLevel(slog.LevelDebug)
		for _, entry := range entries {
			if entry.Message == "message handled successfully" {
				successCount++
			}
		}
		assert.Equal(t, 2, successCount)
	})

	t.Run("handle error from handler", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		locator := handler.NewHandlerLocator()
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewHandleMessageMiddleware(locator, logger)

		errorHandler := &helpers.ErrorTestMessageHandler{
			Error: errors.New("handler error"),
		}
		err := locator.Register(errorHandler)
		assert.NoError(t, err)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)

		next := func(ctx context.Context, env api.Envelope) (api.Envelope, error) {
			return env, nil
		}

		result, err := middleware.Handle(ctx, env, next)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "handler error")
		assert.Equal(t, 1, errorHandler.CallCount)

		assert.True(t, fakeHandler.HasMessage(slog.LevelError, "handler failed"))
	})

	t.Run("handle handler with result", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		locator := handler.NewHandlerLocator()
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewHandleMessageMiddleware(locator, logger)

		resultHandler := &helpers.ResultTestMessageHandler{
			Result: "test result",
		}
		err := locator.Register(resultHandler)
		assert.NoError(t, err)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)

		nextCalled := false
		var nextEnv api.Envelope
		next := func(ctx context.Context, env api.Envelope) (api.Envelope, error) {
			nextCalled = true
			nextEnv = env
			return env, nil
		}

		result, err := middleware.Handle(ctx, env, next)

		assert.NoError(t, err)
		assert.Equal(t, result, nextEnv)
		assert.True(t, nextCalled)
		assert.Equal(t, 1, resultHandler.CallCount)

		handledStamps := envelope.StampsOf[stamps.HandledStamp](result)
		assert.Len(t, handledStamps, 1)
		assert.Equal(t, "test result", handledStamps[0].Result)
		assert.NotEmpty(t, handledStamps[0].Handler)

		assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message handled successfully"))
	})
}
