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
	"github.com/gerfey/messenger/core/event"
	"github.com/gerfey/messenger/core/stamps"
	"github.com/gerfey/messenger/tests/helpers"
	"github.com/gerfey/messenger/tests/mocks"
)

func TestNewSendMessageMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRouter := mocks.NewMockRouter(ctrl)
	mockTransportLocator := mocks.NewMockTransportLocator(ctrl)
	mockEventDispatcher := mocks.NewMockEventDispatcher(ctrl)
	logger, _ := helpers.NewFakeLogger()

	middleware := implementation.NewSendMessageMiddleware(mockRouter, mockTransportLocator, mockEventDispatcher, logger)

	assert.NotNil(t, middleware)
	assert.IsType(t, &implementation.SendMessageMiddleware{}, middleware)
}

func TestSendMessageMiddleware_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("skip processing if envelope has ReceivedStamp", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRouter := mocks.NewMockRouter(ctrl)
		mockTransportLocator := mocks.NewMockTransportLocator(ctrl)
		mockEventDispatcher := mocks.NewMockEventDispatcher(ctrl)
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewSendMessageMiddleware(mockRouter, mockTransportLocator, mockEventDispatcher, logger)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg).WithStamp(stamps.ReceivedStamp{Transport: "test"})

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
		assert.Equal(t, 0, fakeHandler.Count())
	})

	t.Run("continue processing when no transports configured", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRouter := mocks.NewMockRouter(ctrl)
		mockTransportLocator := mocks.NewMockTransportLocator(ctrl)
		mockEventDispatcher := mocks.NewMockEventDispatcher(ctrl)
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewSendMessageMiddleware(mockRouter, mockTransportLocator, mockEventDispatcher, logger)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)

		mockRouter.EXPECT().GetTransportFor(msg).Return([]string{})

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
		assert.True(t, fakeHandler.HasMessage(slog.LevelWarn, "no transports configured for message"))
	})

	t.Run("successfully send message to single transport", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRouter := mocks.NewMockRouter(ctrl)
		mockTransportLocator := mocks.NewMockTransportLocator(ctrl)
		mockEventDispatcher := mocks.NewMockEventDispatcher(ctrl)
		mockTransport := mocks.NewMockTransport(ctrl)
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewSendMessageMiddleware(mockRouter, mockTransportLocator, mockEventDispatcher, logger)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)
		transportNames := []string{"test-transport"}

		mockRouter.EXPECT().GetTransportFor(msg).Return(transportNames)
		mockEventDispatcher.EXPECT().Dispatch(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, event *event.SendMessageToTransportsEvent) error {
				assert.Equal(t, ctx, event.Ctx)
				assert.Equal(t, env, event.Envelope)
				assert.Equal(t, transportNames, event.TransportNames)
				return nil
			})
		mockTransportLocator.EXPECT().GetTransport("test-transport").Return(mockTransport)
		mockTransport.EXPECT().Send(ctx, env).Return(nil)

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

		sentStamps := envelope.StampsOf[stamps.SentStamp](result)
		assert.Len(t, sentStamps, 1)
		assert.Equal(t, "test-transport", sentStamps[0].Transport)

		assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "sending message to transports"))
		assert.True(t, fakeHandler.HasMessage(slog.LevelDebug, "message sent successfully"))
	})

	t.Run("successfully send message to multiple transports", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRouter := mocks.NewMockRouter(ctrl)
		mockTransportLocator := mocks.NewMockTransportLocator(ctrl)
		mockEventDispatcher := mocks.NewMockEventDispatcher(ctrl)
		mockTransport1 := mocks.NewMockTransport(ctrl)
		mockTransport2 := mocks.NewMockTransport(ctrl)
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewSendMessageMiddleware(mockRouter, mockTransportLocator, mockEventDispatcher, logger)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)
		transportNames := []string{"transport-1", "transport-2"}

		mockRouter.EXPECT().GetTransportFor(msg).Return(transportNames)
		mockEventDispatcher.EXPECT().Dispatch(ctx, gomock.Any()).Return(nil)
		mockTransportLocator.EXPECT().GetTransport("transport-1").Return(mockTransport1)
		mockTransportLocator.EXPECT().GetTransport("transport-2").Return(mockTransport2)
		mockTransport1.EXPECT().Send(ctx, gomock.Any()).Return(nil)
		mockTransport2.EXPECT().Send(ctx, gomock.Any()).Return(nil)

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

		sentStamps := envelope.StampsOf[stamps.SentStamp](result)
		assert.Len(t, sentStamps, 2)

		transportsSent := make(map[string]bool)
		for _, stamp := range sentStamps {
			transportsSent[stamp.Transport] = true
		}
		assert.True(t, transportsSent["transport-1"])
		assert.True(t, transportsSent["transport-2"])

		successCount := 0
		entries := fakeHandler.GetEntriesByLevel(slog.LevelDebug)
		for _, entry := range entries {
			if entry.Message == "message sent successfully" {
				successCount++
			}
		}
		assert.Equal(t, 2, successCount)
	})

	t.Run("handle event dispatcher error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRouter := mocks.NewMockRouter(ctrl)
		mockTransportLocator := mocks.NewMockTransportLocator(ctrl)
		mockEventDispatcher := mocks.NewMockEventDispatcher(ctrl)
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewSendMessageMiddleware(mockRouter, mockTransportLocator, mockEventDispatcher, logger)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)
		transportNames := []string{"test-transport"}
		dispatcherError := errors.New("dispatcher error")

		mockRouter.EXPECT().GetTransportFor(msg).Return(transportNames)
		mockEventDispatcher.EXPECT().Dispatch(ctx, gomock.Any()).Return(dispatcherError)

		next := func(ctx context.Context, env api.Envelope) (api.Envelope, error) {
			return env, nil
		}

		result, err := middleware.Handle(ctx, env, next)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, dispatcherError, err)
		assert.True(t, fakeHandler.HasMessage(slog.LevelError, "failed to dispatch send event"))
	})

	t.Run("handle transport not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRouter := mocks.NewMockRouter(ctrl)
		mockTransportLocator := mocks.NewMockTransportLocator(ctrl)
		mockEventDispatcher := mocks.NewMockEventDispatcher(ctrl)
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewSendMessageMiddleware(mockRouter, mockTransportLocator, mockEventDispatcher, logger)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)
		transportNames := []string{"missing-transport"}

		mockRouter.EXPECT().GetTransportFor(msg).Return(transportNames)
		mockEventDispatcher.EXPECT().Dispatch(ctx, gomock.Any()).Return(nil)
		mockTransportLocator.EXPECT().GetTransport("missing-transport").Return(nil)

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
		assert.True(t, fakeHandler.HasMessage(slog.LevelError, "transport not found"))
	})

	t.Run("handle transport send error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRouter := mocks.NewMockRouter(ctrl)
		mockTransportLocator := mocks.NewMockTransportLocator(ctrl)
		mockEventDispatcher := mocks.NewMockEventDispatcher(ctrl)
		mockTransport := mocks.NewMockTransport(ctrl)
		logger, fakeHandler := helpers.NewFakeLogger()
		middleware := implementation.NewSendMessageMiddleware(mockRouter, mockTransportLocator, mockEventDispatcher, logger)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)
		transportNames := []string{"test-transport"}
		sendError := errors.New("send error")

		mockRouter.EXPECT().GetTransportFor(msg).Return(transportNames)
		mockEventDispatcher.EXPECT().Dispatch(ctx, gomock.Any()).Return(nil)
		mockTransportLocator.EXPECT().GetTransport("test-transport").Return(mockTransport)
		mockTransport.EXPECT().Send(ctx, env).Return(sendError)

		next := func(ctx context.Context, env api.Envelope) (api.Envelope, error) {
			return env, nil
		}

		result, err := middleware.Handle(ctx, env, next)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, sendError, err)
		assert.True(t, fakeHandler.HasMessage(slog.LevelError, "failed to send message to transport"))
	})
}
