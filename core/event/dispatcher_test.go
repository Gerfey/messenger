package event_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/gerfey/messenger/core/event"
	"github.com/gerfey/messenger/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestNewDispatcher(t *testing.T) {
	logger, _ := helpers.NewFakeLogger()

	d := event.NewEventDispatcher(logger)

	assert.NotNil(t, d)
}

func TestDispatcher_AddListener(t *testing.T) {
	t.Run("add function listener", func(t *testing.T) {
		logger, handler := helpers.NewFakeLogger()
		dispatcher := event.NewEventDispatcher(logger)

		testEvent := &helpers.TestEvent{}
		dispatcher.AddListener(testEvent, helpers.SimpleEventListener)

		assert.True(t, handler.HasMessage(slog.LevelDebug, "event listener added"))
	})

	t.Run("add struct listener", func(t *testing.T) {
		logger, handler := helpers.NewFakeLogger()
		dispatcher := event.NewEventDispatcher(logger)

		testEvent := &helpers.TestEvent{}
		eventHandler := &helpers.TestEventHandler{}
		dispatcher.AddListener(testEvent, eventHandler)

		assert.True(t, handler.HasMessage(slog.LevelDebug, "event listener added"))
	})

	t.Run("add multiple listeners for same event", func(t *testing.T) {
		logger, handler := helpers.NewFakeLogger()
		dispatcher := event.NewEventDispatcher(logger)

		testEvent := &helpers.TestEvent{}
		handler1 := &helpers.TestEventHandler{}
		handler2 := &helpers.TestEventHandlerWithContext{}

		dispatcher.AddListener(testEvent, handler1)
		dispatcher.AddListener(testEvent, handler2)

		assert.Equal(t, 2, handler.CountByLevel(slog.LevelDebug))
	})

	t.Run("add listeners for different events", func(t *testing.T) {
		logger, handler := helpers.NewFakeLogger()
		dispatcher := event.NewEventDispatcher(logger)

		testEvent := &helpers.TestEvent{}
		anotherEvent := &helpers.AnotherTestEvent{}
		eventHandler := &helpers.TestEventHandler{}

		dispatcher.AddListener(testEvent, eventHandler)
		dispatcher.AddListener(anotherEvent, helpers.SimpleEventListener)

		assert.Equal(t, 2, handler.CountByLevel(slog.LevelDebug))
	})
}

func TestDispatcher_Dispatch(t *testing.T) {
	ctx := context.Background()

	t.Run("dispatch to function listener", func(t *testing.T) {
		logger, handler := helpers.NewFakeLogger()
		dispatcher := event.NewEventDispatcher(logger)

		testEvent := &helpers.TestEvent{ID: "test", Message: "hello"}
		dispatcher.AddListener(testEvent, helpers.SimpleEventListener)

		err := dispatcher.Dispatch(ctx, testEvent)

		assert.NoError(t, err)
		assert.True(t, handler.HasMessage(slog.LevelDebug, "dispatching event"))
		assert.True(t, handler.HasMessage(slog.LevelDebug, "event handled successfully"))
	})

	t.Run("dispatch to function listener with context", func(t *testing.T) {
		logger, handler := helpers.NewFakeLogger()
		dispatcher := event.NewEventDispatcher(logger)

		testEvent := &helpers.TestEvent{ID: "test", Message: "hello"}
		dispatcher.AddListener(testEvent, helpers.TestEventListenerWithContext)

		err := dispatcher.Dispatch(ctx, testEvent)

		assert.NoError(t, err)
		assert.True(t, handler.HasMessage(slog.LevelDebug, "dispatching event"))
		assert.True(t, handler.HasMessage(slog.LevelDebug, "event handled successfully"))
	})

	t.Run("dispatch to struct listener", func(t *testing.T) {
		logger, handler := helpers.NewFakeLogger()
		dispatcher := event.NewEventDispatcher(logger)

		testEvent := &helpers.TestEvent{ID: "test", Message: "hello"}
		eventHandler := &helpers.TestEventHandler{}
		dispatcher.AddListener(testEvent, eventHandler)

		err := dispatcher.Dispatch(ctx, testEvent)

		assert.NoError(t, err)
		assert.Equal(t, 1, eventHandler.CallCount)
		assert.True(t, handler.HasMessage(slog.LevelDebug, "event handled successfully"))
	})

	t.Run("dispatch to struct listener with context", func(t *testing.T) {
		logger, handler := helpers.NewFakeLogger()
		dispatcher := event.NewEventDispatcher(logger)

		testEvent := &helpers.TestEvent{ID: "test", Message: "hello"}
		eventHandler := &helpers.TestEventHandlerWithContext{}
		dispatcher.AddListener(testEvent, eventHandler)

		err := dispatcher.Dispatch(ctx, testEvent)

		assert.NoError(t, err)
		assert.Equal(t, 1, eventHandler.CallCount)
		assert.True(t, handler.HasMessage(slog.LevelDebug, "event handled successfully"))
	})

	t.Run("dispatch to multiple listeners", func(t *testing.T) {
		logger, handler := helpers.NewFakeLogger()
		dispatcher := event.NewEventDispatcher(logger)

		testEvent := &helpers.TestEvent{ID: "test", Message: "hello"}
		handler1 := &helpers.TestEventHandler{}
		handler2 := &helpers.TestEventHandlerWithContext{}

		dispatcher.AddListener(testEvent, handler1)
		dispatcher.AddListener(testEvent, handler2)

		err := dispatcher.Dispatch(ctx, testEvent)

		assert.NoError(t, err)
		assert.Equal(t, 1, handler1.CallCount)
		assert.Equal(t, 1, handler2.CallCount)

		successEntries := handler.GetEntriesByLevel(slog.LevelDebug)
		successCount := 0
		for _, entry := range successEntries {
			if entry.Message == "event handled successfully" {
				successCount++
			}
		}
		assert.Equal(t, 2, successCount)
	})

	t.Run("dispatch with no listeners", func(t *testing.T) {
		logger, handler := helpers.NewFakeLogger()
		dispatcher := event.NewEventDispatcher(logger)

		testEvent := &helpers.TestEvent{ID: "test", Message: "hello"}

		err := dispatcher.Dispatch(ctx, testEvent)

		assert.NoError(t, err)
		assert.True(t, handler.HasMessage(slog.LevelDebug, "no listeners found for event"))
	})

	t.Run("dispatch with listener error", func(t *testing.T) {
		logger, handler := helpers.NewFakeLogger()
		dispatcher := event.NewEventDispatcher(logger)

		errorEvent := &helpers.ErrorEvent{ShouldFail: true}
		dispatcher.AddListener(errorEvent, helpers.ErrorEventListener)

		err := dispatcher.Dispatch(ctx, errorEvent)

		assert.Error(t, err)
		assert.Equal(t, "listener error", err.Error())
		assert.True(t, handler.HasMessage(slog.LevelError, "event handler failed"))
	})

	t.Run("dispatch with struct listener error", func(t *testing.T) {
		logger, handler := helpers.NewFakeLogger()
		dispatcher := event.NewEventDispatcher(logger)

		errorEvent := &helpers.ErrorEvent{ShouldFail: true}
		eventHandler := &helpers.ErrorEventHandler{ShouldFail: true}
		dispatcher.AddListener(errorEvent, eventHandler)

		err := dispatcher.Dispatch(ctx, errorEvent)

		assert.Error(t, err)
		assert.Equal(t, "handler error", err.Error())
		assert.True(t, handler.HasMessage(slog.LevelError, "event handler failed"))
	})
}

func TestDispatcher_Dispatch_InvalidListeners(t *testing.T) {
	ctx := context.Background()

	t.Run("listener without Handle method", func(t *testing.T) {
		logger, handler := helpers.NewFakeLogger()
		dispatcher := event.NewEventDispatcher(logger)

		testEvent := &helpers.TestEvent{ID: "test", Message: "hello"}
		invalidHandler := &helpers.InvalidEventHandler{}
		dispatcher.AddListener(testEvent, invalidHandler)

		err := dispatcher.Dispatch(ctx, testEvent)

		assert.NoError(t, err)
		assert.True(t, handler.HasMessage(slog.LevelError, "listener does not have Handle method"))
	})

	t.Run("listener with wrong signature", func(t *testing.T) {
		logger, handler := helpers.NewFakeLogger()
		dispatcher := event.NewEventDispatcher(logger)

		testEvent := &helpers.TestEvent{ID: "test", Message: "hello"}
		invalidHandler := &helpers.InvalidEventHandlerWrongSignature{}
		dispatcher.AddListener(testEvent, invalidHandler)

		err := dispatcher.Dispatch(ctx, testEvent)

		assert.NoError(t, err)
		assert.True(t, handler.HasMessage(slog.LevelError, "invalid handler signature"))
	})

	t.Run("listener with too many parameters", func(t *testing.T) {
		logger, handler := helpers.NewFakeLogger()
		dispatcher := event.NewEventDispatcher(logger)

		testEvent := &helpers.TestEvent{ID: "test", Message: "hello"}
		invalidHandler := &helpers.InvalidEventHandlerTooManyParams{}
		dispatcher.AddListener(testEvent, invalidHandler)

		err := dispatcher.Dispatch(ctx, testEvent)

		assert.NoError(t, err)
		assert.True(t, handler.HasMessage(slog.LevelError, "invalid handler signature"))
	})
}

func TestDispatcher_Dispatch_EventTypeResolution(t *testing.T) {
	ctx := context.Background()

	t.Run("dispatch with pointer event", func(t *testing.T) {
		logger, _ := helpers.NewFakeLogger()
		dispatcher := event.NewEventDispatcher(logger)

		testEvent := &helpers.TestEvent{ID: "test", Message: "hello"}
		eventHandler := &helpers.TestEventHandler{}

		dispatcher.AddListener(testEvent, eventHandler)

		err := dispatcher.Dispatch(ctx, testEvent)

		assert.NoError(t, err)
		assert.Equal(t, 1, eventHandler.CallCount)
	})
}
