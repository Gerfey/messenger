package messenger_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/gerfey/messenger"

	"github.com/gerfey/messenger/core/bus"
	"github.com/gerfey/messenger/tests/helpers"
	"github.com/gerfey/messenger/tests/mocks"
	"github.com/gerfey/messenger/transport"
)

func TestNewMessenger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("create messenger with all components", func(t *testing.T) {
		defaultBusName := "default"
		mockManager := transport.NewManager(nil, nil, nil)
		mockBusLocator := mocks.NewMockBusLocator(ctrl)
		mockRouter := mocks.NewMockRouter(ctrl)

		m := messenger.NewMessenger(defaultBusName, mockManager, mockBusLocator, mockRouter)

		require.NotNil(t, m)
		assert.IsType(t, &messenger.Messenger{}, m)
	})
}

func TestMessenger_GetDefaultBus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("get existing default bus", func(t *testing.T) {
		defaultBusName := "default"
		mockBusLocator := mocks.NewMockBusLocator(ctrl)
		mockBus := bus.NewBus()

		mockBusLocator.EXPECT().Get(defaultBusName).Return(mockBus, true)

		m := messenger.NewMessenger(defaultBusName, nil, mockBusLocator, nil)

		result, err := m.GetDefaultBus()

		require.NoError(t, err)
		assert.Equal(t, mockBus, result)
	})

	t.Run("get non-existing default bus", func(t *testing.T) {
		defaultBusName := "non-existing"
		mockBusLocator := mocks.NewMockBusLocator(ctrl)

		mockBusLocator.EXPECT().Get(defaultBusName).Return(nil, false)

		m := messenger.NewMessenger(defaultBusName, nil, mockBusLocator, nil)

		result, err := m.GetDefaultBus()

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "default bus 'non-existing' not found")
	})

	t.Run("get default bus with empty name", func(t *testing.T) {
		mockBusLocator := mocks.NewMockBusLocator(ctrl)

		mockBusLocator.EXPECT().Get("").Return(nil, false)

		m := messenger.NewMessenger("", nil, mockBusLocator, nil)

		result, err := m.GetDefaultBus()

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "default bus '' not found")
	})
}

func TestMessenger_GetBusWith(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("get existing bus by name", func(t *testing.T) {
		busName := "async"
		mockBusLocator := mocks.NewMockBusLocator(ctrl)
		mockBus := bus.NewBus(&helpers.TestMiddleware{})

		mockBusLocator.EXPECT().Get(busName).Return(mockBus, true)

		m := messenger.NewMessenger("default", nil, mockBusLocator, nil)

		result, err := m.GetBusWith(busName)

		require.NoError(t, err)
		assert.Equal(t, mockBus, result)
	})

	t.Run("get non-existing bus by name", func(t *testing.T) {
		busName := "non-existing"
		mockBusLocator := mocks.NewMockBusLocator(ctrl)

		mockBusLocator.EXPECT().Get(busName).Return(nil, false)

		m := messenger.NewMessenger("default", nil, mockBusLocator, nil)

		result, err := m.GetBusWith(busName)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "bus 'non-existing' not found")
	})

	t.Run("get bus with empty name", func(t *testing.T) {
		mockBusLocator := mocks.NewMockBusLocator(ctrl)

		mockBusLocator.EXPECT().Get("").Return(nil, false)

		m := messenger.NewMessenger("default", nil, mockBusLocator, nil)

		result, err := m.GetBusWith("")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "bus '' not found")
	})

	t.Run("get multiple different buses", func(t *testing.T) {
		mockBusLocator := mocks.NewMockBusLocator(ctrl)
		bus1 := bus.NewBus()
		bus2 := bus.NewBus(&helpers.TestMiddleware{})

		mockBusLocator.EXPECT().Get("bus1").Return(bus1, true)
		mockBusLocator.EXPECT().Get("bus2").Return(bus2, true)

		m := messenger.NewMessenger("default", nil, mockBusLocator, nil)

		result1, err1 := m.GetBusWith("bus1")
		result2, err2 := m.GetBusWith("bus2")

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.Equal(t, bus1, result1)
		assert.Equal(t, bus2, result2)
		assert.NotEqual(t, result1, result2)
	})
}

func TestMessenger_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("run messenger with context cancellation", func(t *testing.T) {
		mockManager := transport.NewManager(nil, nil, nil)
		mockRouter := mocks.NewMockRouter(ctrl)
		usedTransports := []string{"amqp", "inmemory"}

		mockRouter.EXPECT().GetUsedTransports().Return(usedTransports)

		m := messenger.NewMessenger("default", mockManager, nil, mockRouter)

		ctx, cancel := context.WithCancel(t.Context())

		errChan := make(chan error, 1)
		go func() {
			errChan <- m.Run(ctx)
		}()

		time.Sleep(10 * time.Millisecond)

		cancel()

		select {
		case runErr := <-errChan:
			assert.Equal(t, context.Canceled, runErr)
		case <-time.After(1 * time.Second):
			t.Fatal("Run did not complete within timeout")
		}
	})

	t.Run("run messenger with timeout context", func(t *testing.T) {
		mockManager := transport.NewManager(nil, nil, nil)
		mockRouter := mocks.NewMockRouter(ctrl)
		usedTransports := []string{"inmemory"}

		mockRouter.EXPECT().GetUsedTransports().Return(usedTransports)

		m := messenger.NewMessenger("default", mockManager, nil, mockRouter)

		ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
		defer cancel()

		err := m.Run(ctx)

		assert.Equal(t, context.DeadlineExceeded, err)
	})

	t.Run("run messenger with empty transport list", func(t *testing.T) {
		mockManager := transport.NewManager(nil, nil, nil)
		mockRouter := mocks.NewMockRouter(ctrl)
		usedTransports := []string{}

		mockRouter.EXPECT().GetUsedTransports().Return(usedTransports)

		m := messenger.NewMessenger("default", mockManager, nil, mockRouter)

		ctx, cancel := context.WithCancel(t.Context())

		errChan := make(chan error, 1)
		go func() {
			errChan <- m.Run(ctx)
		}()

		time.Sleep(10 * time.Millisecond)
		cancel()

		select {
		case runErr := <-errChan:
			assert.Equal(t, context.Canceled, runErr)
		case <-time.After(1 * time.Second):
			t.Fatal("Run did not complete within timeout")
		}
	})

	t.Run("run messenger with nil router", func(t *testing.T) {
		mockManager := transport.NewManager(nil, nil, nil)

		m := messenger.NewMessenger("default", mockManager, nil, nil)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		assert.Panics(t, func() {
			_ = m.Run(ctx)
		})
	})
}

func TestMessenger_Integration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("full workflow with real components", func(t *testing.T) {
		busLocator := bus.NewLocator()
		defaultBus := bus.NewBus()
		asyncBus := bus.NewBus(&helpers.TestMiddleware{})

		require.NoError(t, busLocator.Register("default", defaultBus))
		require.NoError(t, busLocator.Register("async", asyncBus))

		mockManager := transport.NewManager(nil, nil, nil)
		mockRouter := mocks.NewMockRouter(ctrl)

		usedTransports := []string{"inmemory"}
		mockRouter.EXPECT().GetUsedTransports().Return(usedTransports)

		m := messenger.NewMessenger("default", mockManager, busLocator, mockRouter)

		defaultBusResult, err := m.GetDefaultBus()
		require.NoError(t, err)
		assert.Equal(t, defaultBus, defaultBusResult)

		asyncBusResult, err := m.GetBusWith("async")
		require.NoError(t, err)
		assert.Equal(t, asyncBus, asyncBusResult)

		_, err = m.GetBusWith("non-existing")
		require.Error(t, err)

		ctx, cancel := context.WithCancel(t.Context())

		errChan := make(chan error, 1)
		go func() {
			errChan <- m.Run(ctx)
		}()

		time.Sleep(10 * time.Millisecond)
		cancel()

		select {
		case runErr := <-errChan:
			assert.Equal(t, context.Canceled, runErr)
		case <-time.After(1 * time.Second):
			t.Fatal("Run did not complete within timeout")
		}
	})
}
