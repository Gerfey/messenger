package implementation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/middleware/implementation"
	"github.com/gerfey/messenger/core/stamps"
	"github.com/gerfey/messenger/tests/helpers"
	"github.com/gerfey/messenger/tests/mocks"
)

func TestNewSendMessageMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTransportLocator := mocks.NewMockSenderLocator(ctrl)
	mockEventDispatcher := mocks.NewMockEventDispatcher(ctrl)
	logger, _ := helpers.NewFakeLogger()

	middleware := implementation.NewSendMessageMiddleware(logger, mockTransportLocator, mockEventDispatcher)

	require.NotNil(t, middleware)
	require.IsType(t, &implementation.SendMessageMiddleware{}, middleware)
}

func TestSendMessageMiddleware_Handle(t *testing.T) {
	t.Run("skip processing if envelope has ReceivedStamp", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTransportLocator := mocks.NewMockSenderLocator(ctrl)
		mockEventDispatcher := mocks.NewMockEventDispatcher(ctrl)
		logger, _ := helpers.NewFakeLogger()
		middleware := implementation.NewSendMessageMiddleware(
			logger,
			mockTransportLocator,
			mockEventDispatcher,
		)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg).WithStamp(stamps.ReceivedStamp{})

		nextCalled := false
		var nextEnv api.Envelope
		next := func(_ context.Context, env api.Envelope) (api.Envelope, error) {
			nextCalled = true
			nextEnv = env

			return env, nil
		}

		result, err := middleware.Handle(t.Context(), env, next)

		require.NoError(t, err)
		assert.True(t, nextCalled)
		assert.Equal(t, result, nextEnv)
	})

	t.Run("return error when no senders configured", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTransportLocator := mocks.NewMockSenderLocator(ctrl)
		mockEventDispatcher := mocks.NewMockEventDispatcher(ctrl)
		logger, _ := helpers.NewFakeLogger()
		middleware := implementation.NewSendMessageMiddleware(
			logger,
			mockTransportLocator,
			mockEventDispatcher,
		)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)

		mockTransportLocator.EXPECT().GetSenders(env).Return([]api.Sender{})

		nextCalled := false
		next := func(_ context.Context, env api.Envelope) (api.Envelope, error) {
			nextCalled = true

			return env, nil
		}

		_, err := middleware.Handle(t.Context(), env, next)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no senders configured for message")
		assert.False(t, nextCalled)
	})

	t.Run("successfully send message to single sender", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTransportLocator := mocks.NewMockSenderLocator(ctrl)
		mockEventDispatcher := mocks.NewMockEventDispatcher(ctrl)
		mockTransport := mocks.NewMockTransport(ctrl)
		logger, _ := helpers.NewFakeLogger()
		middleware := implementation.NewSendMessageMiddleware(
			logger,
			mockTransportLocator,
			mockEventDispatcher,
		)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)

		mockTransportLocator.EXPECT().GetSenders(env).Return([]api.Sender{mockTransport})
		mockTransport.EXPECT().Name().Return("test-transport").Times(3)
		mockEventDispatcher.EXPECT().Dispatch(t.Context(), gomock.Any()).Return(nil)
		mockTransport.EXPECT().Send(t.Context(), gomock.Any()).Return(nil)

		nextCalled := false
		next := func(_ context.Context, env api.Envelope) (api.Envelope, error) {
			nextCalled = true

			return env, nil
		}

		result, err := middleware.Handle(t.Context(), env, next)

		require.NoError(t, err)
		assert.False(t, nextCalled)
		sentStamp, ok := envelope.LastStampOf[stamps.SentStamp](result)
		assert.True(t, ok)
		assert.Equal(t, "test-transport", sentStamp.SenderName)
	})

	t.Run("successfully send message to multiple senders", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTransportLocator := mocks.NewMockSenderLocator(ctrl)
		mockEventDispatcher := mocks.NewMockEventDispatcher(ctrl)
		mockTransport1 := mocks.NewMockTransport(ctrl)
		mockTransport2 := mocks.NewMockTransport(ctrl)
		logger, _ := helpers.NewFakeLogger()
		middleware := implementation.NewSendMessageMiddleware(
			logger,
			mockTransportLocator,
			mockEventDispatcher,
		)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)

		mockTransportLocator.EXPECT().GetSenders(env).Return([]api.Sender{mockTransport1, mockTransport2})

		mockTransport1.EXPECT().Name().Return("sender1").Times(3)
		mockEventDispatcher.EXPECT().Dispatch(t.Context(), gomock.Any()).Return(nil)
		mockTransport1.EXPECT().Send(t.Context(), gomock.Any()).Return(nil)

		mockTransport2.EXPECT().Name().Return("sender2").Times(3)
		mockEventDispatcher.EXPECT().Dispatch(t.Context(), gomock.Any()).Return(nil)
		mockTransport2.EXPECT().Send(t.Context(), gomock.Any()).Return(nil)

		nextCalled := false
		next := func(_ context.Context, env api.Envelope) (api.Envelope, error) {
			nextCalled = true

			return env, nil
		}

		result, err := middleware.Handle(t.Context(), env, next)

		require.NoError(t, err)
		assert.False(t, nextCalled)

		var sentStamps []stamps.SentStamp
		for _, stamp := range result.Stamps() {
			if sentStamp, ok := stamp.(stamps.SentStamp); ok {
				sentStamps = append(sentStamps, sentStamp)
			}
		}

		require.Len(t, sentStamps, 2)

		senderNames := make([]string, 0, len(sentStamps))
		for _, stamp := range sentStamps {
			senderNames = append(senderNames, stamp.SenderName)
		}
		require.Contains(t, senderNames, "sender1")
		require.Contains(t, senderNames, "sender2")
	})

	t.Run("handle event dispatcher error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTransportLocator := mocks.NewMockSenderLocator(ctrl)
		mockEventDispatcher := mocks.NewMockEventDispatcher(ctrl)
		mockTransport := mocks.NewMockTransport(ctrl)
		logger, _ := helpers.NewFakeLogger()
		middleware := implementation.NewSendMessageMiddleware(
			logger,
			mockTransportLocator,
			mockEventDispatcher,
		)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)
		dispatchError := errors.New("dispatch error")

		mockTransportLocator.EXPECT().GetSenders(env).Return([]api.Sender{mockTransport})
		mockTransport.EXPECT().Name().Return("test-transport").Times(1)
		mockEventDispatcher.EXPECT().Dispatch(t.Context(), gomock.Any()).Return(dispatchError)

		nextCalled := false
		next := func(_ context.Context, env api.Envelope) (api.Envelope, error) {
			nextCalled = true

			return env, nil
		}

		_, err := middleware.Handle(t.Context(), env, next)

		require.Error(t, err)
		assert.Equal(t, dispatchError, err)
		assert.False(t, nextCalled)
	})

	t.Run("handle sender send error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTransportLocator := mocks.NewMockSenderLocator(ctrl)
		mockEventDispatcher := mocks.NewMockEventDispatcher(ctrl)
		mockTransport := mocks.NewMockTransport(ctrl)
		logger, _ := helpers.NewFakeLogger()
		middleware := implementation.NewSendMessageMiddleware(
			logger,
			mockTransportLocator,
			mockEventDispatcher,
		)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)
		sendError := errors.New("send error")

		mockTransportLocator.EXPECT().GetSenders(env).Return([]api.Sender{mockTransport})
		mockTransport.EXPECT().Name().Return("test-transport").Times(3)
		mockEventDispatcher.EXPECT().Dispatch(t.Context(), gomock.Any()).Return(nil)
		mockTransport.EXPECT().Send(t.Context(), gomock.Any()).Return(sendError)

		nextCalled := false
		next := func(_ context.Context, env api.Envelope) (api.Envelope, error) {
			nextCalled = true

			return env, nil
		}

		_, err := middleware.Handle(t.Context(), env, next)

		require.Error(t, err)
		assert.Equal(t, sendError, err)
		assert.False(t, nextCalled)
	})
}
