package builder_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/builder"
	"github.com/gerfey/messenger/config"
	"github.com/gerfey/messenger/core/stamps"
	"github.com/gerfey/messenger/tests/helpers"
)

func TestNewBuilder(t *testing.T) {
	t.Run("create builder with valid config", func(t *testing.T) {
		cfg := &config.MessengerConfig{
			DefaultBus: "default",
			Buses: map[string]config.BusConfig{
				"default": {
					Middleware: []string{},
				},
			},
			Transports: map[string]config.TransportConfig{
				"inmemory": {
					DSN: "",
					Options: config.OptionsConfig{
						AutoSetup:        true,
						ConsumerPoolSize: 10,
					},
				},
			},
			Routing: map[string]string{},
		}
		logger := slog.Default()

		builderInstance := builder.NewBuilder(cfg, logger)

		assert.NotNil(t, builderInstance)
	})

	t.Run("create builder with nil config", func(t *testing.T) {
		logger := slog.Default()

		require.NotPanics(t, func() {
			builderInstance := builder.NewBuilder(nil, logger)
			assert.NotNil(t, builderInstance)
		})
	})

	t.Run("create builder with nil logger", func(t *testing.T) {
		cfg := &config.MessengerConfig{
			DefaultBus: "default",
		}

		require.NotPanics(t, func() {
			builderInstance := builder.NewBuilder(cfg, nil)
			assert.NotNil(t, builderInstance)
		})
	})
}

func TestBuilder_RegisterMessage(t *testing.T) {
	cfg := &config.MessengerConfig{DefaultBus: "default"}
	logger := slog.Default()
	builderInstance := builder.NewBuilder(cfg, logger)

	t.Run("register message successfully", func(t *testing.T) {
		msg := &helpers.TestMessage{ID: "test", Content: "content"}

		require.NotPanics(t, func() {
			builderInstance.RegisterMessage(msg)
		})
	})

	t.Run("register multiple different messages", func(t *testing.T) {
		msg1 := &helpers.TestMessage{}
		msg2 := &helpers.ComplexMessage{}
		msg3 := helpers.SimpleMessage("test")

		require.NotPanics(t, func() {
			builderInstance.RegisterMessage(msg1)
			builderInstance.RegisterMessage(msg2)
			builderInstance.RegisterMessage(msg3)
		})
	})
}

func TestBuilder_RegisterHandler(t *testing.T) {
	cfg := &config.MessengerConfig{DefaultBus: "default"}
	logger := slog.Default()
	builderInstance := builder.NewBuilder(cfg, logger)

	t.Run("register valid handler with context", func(t *testing.T) {
		handler := &helpers.TestEventHandlerWithContext{}

		err := builderInstance.RegisterHandler(handler)
		require.NoError(t, err)
	})

	t.Run("register invalid handler", func(t *testing.T) {
		handler := &helpers.InvalidEventHandler{}

		err := builderInstance.RegisterHandler(handler)
		require.Error(t, err)
	})

	t.Run("register nil handler panics", func(t *testing.T) {
		require.Panics(t, func() {
			_ = builderInstance.RegisterHandler(nil)
		})
	})
}

func TestBuilder_RegisterMiddleware(t *testing.T) {
	cfg := &config.MessengerConfig{DefaultBus: "default"}
	logger := slog.Default()
	builderInstance := builder.NewBuilder(cfg, logger)

	t.Run("register middleware successfully", func(t *testing.T) {
		middleware := &helpers.TestMiddleware{}

		require.NotPanics(t, func() {
			builderInstance.RegisterMiddleware("test_middleware", middleware)
		})
	})

	t.Run("register multiple middleware", func(t *testing.T) {
		mw1 := &helpers.TestMiddleware{}
		mw2 := &helpers.ErrorMiddleware{}

		require.NotPanics(t, func() {
			builderInstance.RegisterMiddleware("test_mw1", mw1)
			builderInstance.RegisterMiddleware("test_mw2", mw2)
		})
	})

	t.Run("register middleware with empty name", func(t *testing.T) {
		middleware := &helpers.TestMiddleware{}

		require.NotPanics(t, func() {
			builderInstance.RegisterMiddleware("", middleware)
		})
	})
}

func TestBuilder_RegisterStamp(t *testing.T) {
	cfg := &config.MessengerConfig{DefaultBus: "default"}
	logger := slog.Default()
	builderInstance := builder.NewBuilder(cfg, logger)

	t.Run("register stamp successfully", func(t *testing.T) {
		stamp := &stamps.BusNameStamp{Name: "test"}

		require.NotPanics(t, func() {
			builderInstance.RegisterStamp(stamp)
		})
	})

	t.Run("register multiple stamps", func(t *testing.T) {
		stamp1 := &stamps.BusNameStamp{}
		stamp2 := &stamps.SentStamp{}
		stamp3 := &helpers.TestStamp{}

		require.NotPanics(t, func() {
			builderInstance.RegisterStamp(stamp1)
			builderInstance.RegisterStamp(stamp2)
			builderInstance.RegisterStamp(stamp3)
		})
	})
}

func TestBuilder_RegisterListener(t *testing.T) {
	cfg := &config.MessengerConfig{DefaultBus: "default"}
	logger := slog.Default()
	builderInstance := builder.NewBuilder(cfg, logger)

	t.Run("register listener successfully", func(t *testing.T) {
		event := &helpers.TestEvent{}
		listener := helpers.SimpleEventListener

		require.NotPanics(t, func() {
			builderInstance.RegisterListener(event, listener)
		})
	})

	t.Run("register multiple listeners", func(t *testing.T) {
		event1 := &helpers.TestEvent{}
		event2 := &helpers.AnotherTestEvent{}
		listener1 := helpers.SimpleEventListener
		listener2 := helpers.TestEventListenerWithContext

		require.NotPanics(t, func() {
			builderInstance.RegisterListener(event1, listener1)
			builderInstance.RegisterListener(event2, listener2)
		})
	})
}

func TestBuilder_RegisterTransportFactory(t *testing.T) {
	t.Run("register transport factory", func(t *testing.T) {
		cfg := &config.MessengerConfig{
			DefaultBus: "default",
			Buses: map[string]config.BusConfig{
				"default": {
					Middleware: []string{},
				},
			},
			Transports: map[string]config.TransportConfig{
				"inmemory": {
					DSN: "in-memory://test",
					Options: config.OptionsConfig{
						AutoSetup:        true,
						ConsumerPoolSize: 10,
					},
				},
			},
			Routing: map[string]string{},
		}
		logger, _ := helpers.NewFakeLogger()
		builderInstance := builder.NewBuilder(cfg, logger)

		testTransport := &helpers.TestTransport{TransportName: "test_transport"}
		testFactory := &helpers.TestTransportFactory{
			TransportName: "test_transport",
			Transport:     testTransport,
		}

		require.NotPanics(t, func() {
			builderInstance.RegisterTransportFactory(testFactory)
		})

		msg := &helpers.TestMessage{ID: "123", Content: "test"}
		builderInstance.RegisterMessage(msg)

		cfg.Transports = map[string]config.TransportConfig{
			"test_transport": {
				DSN: "test://localhost",
			},
		}
		cfg.Routing = map[string]string{
			"*helpers.TestMessage": "test_transport",
		}

		msgHandler := &helpers.TestMessageHandler{}
		err := builderInstance.RegisterHandler(msgHandler)
		require.NoError(t, err)

		messenger, err := builderInstance.Build()
		require.NoError(t, err)
		require.NotNil(t, messenger)
	})
}

func TestBuilder_Build_Errors(t *testing.T) {
	t.Run("build fails with unknown message type in routing", func(t *testing.T) {
		cfg := &config.MessengerConfig{
			DefaultBus: "default",
			Buses: map[string]config.BusConfig{
				"default": {
					Middleware: []string{},
				},
			},
			Transports: map[string]config.TransportConfig{
				"inmemory": {
					DSN: "in-memory://test",
					Options: config.OptionsConfig{
						AutoSetup:        true,
						ConsumerPoolSize: 10,
					},
				},
			},
			Routing: map[string]string{
				"unknown.MessageType": "inmemory",
			},
		}
		logger := slog.Default()
		builderInstance := builder.NewBuilder(cfg, logger)

		messenger, err := builderInstance.Build()

		require.Error(t, err)
		assert.Nil(t, messenger)
		assert.Contains(t, err.Error(), "failed to resolve message type")
	})

	t.Run("build fails with unknown middleware", func(t *testing.T) {
		cfg := &config.MessengerConfig{
			DefaultBus: "default",
			Buses: map[string]config.BusConfig{
				"default": {
					Middleware: []string{"unknown_middleware"},
				},
			},
			Transports: map[string]config.TransportConfig{
				"inmemory": {
					DSN: "in-memory://test",
					Options: config.OptionsConfig{
						AutoSetup:        true,
						ConsumerPoolSize: 10,
					},
				},
			},
			Routing: map[string]string{},
		}
		logger := slog.Default()
		builderInstance := builder.NewBuilder(cfg, logger)

		messenger, err := builderInstance.Build()

		require.Error(t, err)
		assert.Nil(t, messenger)
		assert.Contains(t, err.Error(), "middleware")
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestBuilder_Build_Success(t *testing.T) {
	t.Run("build messenger with complete configuration", func(t *testing.T) {
		cfg := &config.MessengerConfig{
			DefaultBus: "default",
			Buses: map[string]config.BusConfig{
				"default": {
					Middleware: []string{"test_middleware"},
				},
				"async": {
					Middleware: []string{"test_middleware"},
				},
			},
			Transports: map[string]config.TransportConfig{
				"inmemory": {
					DSN: "in-memory://test",
					Options: config.OptionsConfig{
						AutoSetup:        true,
						ConsumerPoolSize: 10,
					},
				},
			},
			Routing: map[string]string{
				"*helpers.TestMessage": "inmemory",
			},
		}
		logger, _ := helpers.NewFakeLogger()
		builderInstance := builder.NewBuilder(cfg, logger)

		msg := &helpers.TestMessage{ID: "123", Content: "test"}
		builderInstance.RegisterMessage(msg)

		msgHandler := &helpers.TestMessageHandler{}
		err := builderInstance.RegisterHandler(msgHandler)
		require.NoError(t, err)

		middleware := &helpers.TestMiddleware{}
		builderInstance.RegisterMiddleware("test_middleware", middleware)

		stamp := &helpers.TestStamp{Value: "test"}
		builderInstance.RegisterStamp(stamp)

		event := &helpers.TestEvent{ID: "123", Message: "test"}
		listener := helpers.SimpleEventListener
		builderInstance.RegisterListener(event, listener)

		messenger, err := builderInstance.Build()
		require.NoError(t, err)
		require.NotNil(t, messenger)
	})

	t.Run("build messenger with retry configuration", func(t *testing.T) {
		cfg := &config.MessengerConfig{
			DefaultBus: "default",
			Buses: map[string]config.BusConfig{
				"default": {
					Middleware: []string{},
				},
			},
			Transports: map[string]config.TransportConfig{
				"inmemory": {
					DSN: "in-memory://test",
					Options: config.OptionsConfig{
						AutoSetup:        true,
						ConsumerPoolSize: 10,
					},
				},
			},
			Routing: map[string]string{
				"*helpers.TestMessage": "inmemory",
			},
		}
		logger, _ := helpers.NewFakeLogger()
		builderInstance := builder.NewBuilder(cfg, logger)

		msg := &helpers.TestMessage{ID: "123", Content: "test"}
		builderInstance.RegisterMessage(msg)

		msgHandler := &helpers.TestMessageHandler{}
		err := builderInstance.RegisterHandler(msgHandler)
		require.NoError(t, err)

		messenger, err := builderInstance.Build()
		require.NoError(t, err)
		require.NotNil(t, messenger)
	})

	t.Run("build messenger with multiple buses and transports", func(t *testing.T) {
		cfg := &config.MessengerConfig{
			DefaultBus: "default",
			Buses: map[string]config.BusConfig{
				"default": {
					Middleware: []string{},
				},
				"async": {
					Middleware: []string{},
				},
				"sync": {
					Middleware: []string{},
				},
			},
			Transports: map[string]config.TransportConfig{
				"inmemory1": {
					DSN: "in-memory://test1",
				},
				"inmemory2": {
					DSN: "in-memory://test2",
				},
			},
			Routing: map[string]string{
				"*helpers.TestMessage":    "inmemory1",
				"*helpers.ComplexMessage": "inmemory2",
			},
		}
		logger, _ := helpers.NewFakeLogger()
		builderInstance := builder.NewBuilder(cfg, logger)

		msg1 := &helpers.TestMessage{ID: "123", Content: "test"}
		msg2 := &helpers.ComplexMessage{ID: "456", Type: "complex"}
		builderInstance.RegisterMessage(msg1)
		builderInstance.RegisterMessage(msg2)

		handler1 := &helpers.TestMessageHandler{}
		handler2 := &helpers.AnotherValidHandler{}
		err := builderInstance.RegisterHandler(handler1)
		require.NoError(t, err)
		err = builderInstance.RegisterHandler(handler2)
		require.NoError(t, err)

		messenger, err := builderInstance.Build()
		require.NoError(t, err)
		require.NotNil(t, messenger)
	})
}
