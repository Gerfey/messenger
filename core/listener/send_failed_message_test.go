package listener_test

import (
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/event"
	"github.com/gerfey/messenger/core/listener"
	"github.com/gerfey/messenger/core/retry"
	"github.com/gerfey/messenger/core/stamps"
	"github.com/gerfey/messenger/tests/helpers"
	"github.com/gerfey/messenger/tests/mocks"
)

func TestNewSendFailedMessageForRetryListener(t *testing.T) {
	t.Run("creates listener with all dependencies", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTransport := mocks.NewMockRetryableTransport(ctrl)
		mockFailureTransport := mocks.NewMockTransport(ctrl)
		mockStrategy := mocks.NewMockStrategy(ctrl)
		logger, _ := helpers.NewFakeLogger()

		l := listener.NewSendFailedMessageForRetryListener(
			logger,
			mockTransport,
			mockFailureTransport,
			mockStrategy,
		)

		require.NotNil(t, l)
	})

	t.Run("creates listener with nil failure transport", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTransport := mocks.NewMockRetryableTransport(ctrl)
		mockStrategy := mocks.NewMockStrategy(ctrl)
		logger, _ := helpers.NewFakeLogger()

		l := listener.NewSendFailedMessageForRetryListener(
			logger,
			mockTransport,
			nil,
			mockStrategy,
		)

		require.NotNil(t, l)
	})
}

func TestSendFailedMessageForRetryListener_Handle(t *testing.T) {
	t.Run("handles retry when within retry limits", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTransport := mocks.NewMockRetryableTransport(ctrl)
		mockStrategy := mocks.NewMockStrategy(ctrl)
		logger, fakeLogger := helpers.NewFakeLogger()

		l := listener.NewSendFailedMessageForRetryListener(
			logger,
			mockTransport,
			nil,
			mockStrategy,
		)

		msg := &helpers.TestMessage{ID: "123", Content: "test"}
		env := envelope.NewEnvelope(msg).WithStamp(stamps.ReceivedStamp{
			Transport: "test-transport",
		})

		evt := event.SendFailedMessageEvent{
			Envelope:      env,
			TransportName: "test-transport",
			Error:         errors.New("send failed"),
		}

		mockStrategy.EXPECT().ShouldRetry(uint(0)).Return(100*time.Millisecond, true)

		mockTransport.EXPECT().Retry(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		l.Handle(t.Context(), evt)

		time.Sleep(10 * time.Millisecond)

		assert.False(t, fakeLogger.HasMessage(slog.LevelError, "retry dispatch failed"))
	})

	t.Run("sends to failure transport when retry limit exceeded", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTransport := mocks.NewMockRetryableTransport(ctrl)
		mockFailureTransport := mocks.NewMockTransport(ctrl)
		mockStrategy := mocks.NewMockStrategy(ctrl)
		logger, fakeLogger := helpers.NewFakeLogger()

		l := listener.NewSendFailedMessageForRetryListener(
			logger,
			mockTransport,
			mockFailureTransport,
			mockStrategy,
		)

		msg := &helpers.TestMessage{ID: "123", Content: "test"}
		env := envelope.NewEnvelope(msg).WithStamp(stamps.ReceivedStamp{
			Transport: "test-transport",
		})

		evt := event.SendFailedMessageEvent{
			Envelope:      env,
			TransportName: "test-transport",
			Error:         errors.New("send failed"),
		}

		mockStrategy.EXPECT().ShouldRetry(uint(0)).Return(time.Duration(0), false)

		mockFailureTransport.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

		l.Handle(t.Context(), evt)

		assert.False(t, fakeLogger.HasMessage(slog.LevelError, "failed to send message to failure transport"))
	})

	t.Run("logs error when failure transport fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTransport := mocks.NewMockRetryableTransport(ctrl)
		mockFailureTransport := mocks.NewMockTransport(ctrl)
		mockStrategy := mocks.NewMockStrategy(ctrl)
		logger, fakeLogger := helpers.NewFakeLogger()

		l := listener.NewSendFailedMessageForRetryListener(
			logger,
			mockTransport,
			mockFailureTransport,
			mockStrategy,
		)

		msg := &helpers.TestMessage{ID: "123", Content: "test"}
		env := envelope.NewEnvelope(msg).WithStamp(stamps.ReceivedStamp{
			Transport: "test-transport",
		})

		evt := event.SendFailedMessageEvent{
			Envelope:      env,
			TransportName: "test-transport",
			Error:         errors.New("send failed"),
		}

		mockStrategy.EXPECT().ShouldRetry(uint(0)).Return(time.Duration(0), false)

		expectedErr := errors.New("failure transport error")
		mockFailureTransport.EXPECT().Send(gomock.Any(), gomock.Any()).Return(expectedErr)

		l.Handle(t.Context(), evt)

		assert.True(t, fakeLogger.HasMessage(slog.LevelError, "failed to send message to failure transport"))
	})

	t.Run("handles retry with existing RedeliveryStamp", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTransport := mocks.NewMockRetryableTransport(ctrl)
		mockStrategy := mocks.NewMockStrategy(ctrl)
		logger, _ := helpers.NewFakeLogger()

		l := listener.NewSendFailedMessageForRetryListener(
			logger,
			mockTransport,
			nil,
			mockStrategy,
		)

		msg := &helpers.TestMessage{ID: "123", Content: "test"}
		env := envelope.NewEnvelope(msg).
			WithStamp(stamps.ReceivedStamp{
				Transport: "test-transport",
			}).
			WithStamp(stamps.RedeliveryStamp{
				RetryCount: 2,
			})

		evt := event.SendFailedMessageEvent{
			Envelope:      env,
			TransportName: "test-transport",
			Error:         errors.New("send failed"),
		}

		mockStrategy.EXPECT().ShouldRetry(uint(3)).Return(100*time.Millisecond, true)

		mockTransport.EXPECT().Retry(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		l.Handle(t.Context(), evt)

		time.Sleep(10 * time.Millisecond)
	})

	t.Run("ignores event without ReceivedStamp", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTransport := mocks.NewMockRetryableTransport(ctrl)
		mockStrategy := mocks.NewMockStrategy(ctrl)
		logger, _ := helpers.NewFakeLogger()

		l := listener.NewSendFailedMessageForRetryListener(
			logger,
			mockTransport,
			nil,
			mockStrategy,
		)

		msg := &helpers.TestMessage{ID: "123", Content: "test"}
		env := envelope.NewEnvelope(msg)

		evt := event.SendFailedMessageEvent{
			Envelope:      env,
			TransportName: "test-transport",
			Error:         errors.New("send failed"),
		}

		l.Handle(t.Context(), evt)
	})

	t.Run("ignores event with mismatched transport name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTransport := mocks.NewMockRetryableTransport(ctrl)
		mockStrategy := mocks.NewMockStrategy(ctrl)
		logger, _ := helpers.NewFakeLogger()

		l := listener.NewSendFailedMessageForRetryListener(
			logger,
			mockTransport,
			nil,
			mockStrategy,
		)

		msg := &helpers.TestMessage{ID: "123", Content: "test"}
		env := envelope.NewEnvelope(msg).WithStamp(stamps.ReceivedStamp{
			Transport: "other-transport",
		})

		evt := event.SendFailedMessageEvent{
			Envelope:      env,
			TransportName: "test-transport",
			Error:         errors.New("send failed"),
		}

		l.Handle(t.Context(), evt)
	})

	t.Run("does nothing when retry not allowed and no failure transport", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTransport := mocks.NewMockRetryableTransport(ctrl)
		mockStrategy := mocks.NewMockStrategy(ctrl)
		logger, _ := helpers.NewFakeLogger()

		l := listener.NewSendFailedMessageForRetryListener(
			logger,
			mockTransport,
			nil,
			mockStrategy,
		)

		msg := &helpers.TestMessage{ID: "123", Content: "test"}
		env := envelope.NewEnvelope(msg).WithStamp(stamps.ReceivedStamp{
			Transport: "test-transport",
		})

		evt := event.SendFailedMessageEvent{
			Envelope:      env,
			TransportName: "test-transport",
			Error:         errors.New("send failed"),
		}

		mockStrategy.EXPECT().ShouldRetry(uint(0)).Return(time.Duration(0), false)

		l.Handle(t.Context(), evt)
	})
}

func TestSendFailedMessageForRetryListener_Integration(t *testing.T) {
	t.Run("full retry flow with real strategy", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTransport := mocks.NewMockRetryableTransport(ctrl)
		logger, fakeLogger := helpers.NewFakeLogger()

		strategy := retry.NewMultiplierRetryStrategy(3, 10*time.Millisecond, 2.0, 100*time.Millisecond)

		l := listener.NewSendFailedMessageForRetryListener(
			logger,
			mockTransport,
			nil,
			strategy,
		)

		msg := &helpers.TestMessage{ID: "123", Content: "test"}
		env := envelope.NewEnvelope(msg).WithStamp(stamps.ReceivedStamp{
			Transport: "test-transport",
		})

		evt := event.SendFailedMessageEvent{
			Envelope:      env,
			TransportName: "test-transport",
			Error:         errors.New("send failed"),
		}

		mockTransport.EXPECT().Retry(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		l.Handle(t.Context(), evt)

		time.Sleep(50 * time.Millisecond)

		assert.False(t, fakeLogger.HasMessage(slog.LevelError, "retry dispatch failed"))
	})
}
