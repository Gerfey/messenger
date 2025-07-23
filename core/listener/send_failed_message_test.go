package listener_test

import (
	"context"
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
			mockTransport,
			mockFailureTransport,
			mockStrategy,
			logger,
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
			mockTransport,
			nil,
			mockStrategy,
			logger,
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
			mockTransport,
			nil,
			mockStrategy,
			logger,
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

		ctx := context.Background()
		l.Handle(ctx, evt)

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
			mockTransport,
			mockFailureTransport,
			mockStrategy,
			logger,
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

		ctx := context.Background()
		l.Handle(ctx, evt)

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
			mockTransport,
			mockFailureTransport,
			mockStrategy,
			logger,
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

		ctx := context.Background()
		l.Handle(ctx, evt)

		assert.True(t, fakeLogger.HasMessage(slog.LevelError, "failed to send message to failure transport"))
	})

	t.Run("handles retry with existing RedeliveryStamp", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTransport := mocks.NewMockRetryableTransport(ctrl)
		mockStrategy := mocks.NewMockStrategy(ctrl)
		logger, _ := helpers.NewFakeLogger()

		l := listener.NewSendFailedMessageForRetryListener(
			mockTransport,
			nil,
			mockStrategy,
			logger,
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

		ctx := context.Background()
		l.Handle(ctx, evt)

		time.Sleep(10 * time.Millisecond)
	})

	t.Run("ignores event without ReceivedStamp", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTransport := mocks.NewMockRetryableTransport(ctrl)
		mockStrategy := mocks.NewMockStrategy(ctrl)
		logger, _ := helpers.NewFakeLogger()

		l := listener.NewSendFailedMessageForRetryListener(
			mockTransport,
			nil,
			mockStrategy,
			logger,
		)

		msg := &helpers.TestMessage{ID: "123", Content: "test"}
		env := envelope.NewEnvelope(msg)

		evt := event.SendFailedMessageEvent{
			Envelope:      env,
			TransportName: "test-transport",
			Error:         errors.New("send failed"),
		}

		ctx := context.Background()
		l.Handle(ctx, evt)
	})

	t.Run("ignores event with mismatched transport name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTransport := mocks.NewMockRetryableTransport(ctrl)
		mockStrategy := mocks.NewMockStrategy(ctrl)
		logger, _ := helpers.NewFakeLogger()

		l := listener.NewSendFailedMessageForRetryListener(
			mockTransport,
			nil,
			mockStrategy,
			logger,
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

		ctx := context.Background()
		l.Handle(ctx, evt)
	})

	t.Run("does nothing when retry not allowed and no failure transport", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTransport := mocks.NewMockRetryableTransport(ctrl)
		mockStrategy := mocks.NewMockStrategy(ctrl)
		logger, _ := helpers.NewFakeLogger()

		l := listener.NewSendFailedMessageForRetryListener(
			mockTransport,
			nil,
			mockStrategy,
			logger,
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

		ctx := context.Background()
		l.Handle(ctx, evt)
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
			mockTransport,
			nil,
			strategy,
			logger,
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

		ctx := context.Background()
		l.Handle(ctx, evt)

		time.Sleep(50 * time.Millisecond)

		assert.False(t, fakeLogger.HasMessage(slog.LevelError, "retry dispatch failed"))
	})
}
