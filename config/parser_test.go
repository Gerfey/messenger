package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/config"
)

func TestYAMLParser_Parse(t *testing.T) {
	parser := &config.YAMLParser{}

	t.Run("parse valid yaml", func(t *testing.T) {
		content := []byte(`
default_bus: default
buses:
  default:
    middleware:
      - handle_message
`)
		var cfg config.MessengerConfig
		err := parser.Parse(content, &cfg)

		require.NoError(t, err)
		assert.Equal(t, "default", cfg.DefaultBus)
		assert.Contains(t, cfg.Buses, "default")
		assert.Contains(t, cfg.Buses["default"].Middleware, "handle_message")
	})

	t.Run("parse yaml with default values", func(t *testing.T) {
		content := []byte(`
default_bus: custom
transports:
  default:
    dsn: memory://default
    options: {}
`)
		var cfg config.MessengerConfig
		err := parser.Parse(content, &cfg)

		require.NoError(t, err)
		assert.Equal(t, "custom", cfg.DefaultBus)

		autoSetup, ok := cfg.Transports["default"].Options["auto_setup"]
		if ok && autoSetup != nil {
			assert.True(t, autoSetup.(bool))
		}

		consumerPoolSize, ok := cfg.Transports["default"].Options["consumer_pool_size"]
		if ok && consumerPoolSize != nil {
			assert.Equal(t, 10, consumerPoolSize)
		}
	})

	t.Run("parse invalid yaml", func(t *testing.T) {
		content := []byte(`
default_bus: default
buses:
  default:
    middleware:
  - handle_message
`)
		var cfg config.MessengerConfig
		err := parser.Parse(content, &cfg)

		require.Error(t, err)
	})

	t.Run("parse empty yaml", func(t *testing.T) {
		content := []byte(`{}`)
		var cfg config.MessengerConfig
		err := parser.Parse(content, &cfg)

		require.NoError(t, err)
		assert.Equal(t, "default", cfg.DefaultBus)
		assert.Empty(t, cfg.Buses)
		assert.Empty(t, cfg.Transports)
		assert.Empty(t, cfg.Routing)
	})

	t.Run("parse yaml with exchange options", func(t *testing.T) {
		content := []byte(`
transports:
  amqp:
    dsn: amqp://localhost
    options:
      exchange:
        name: messages
        type: topic
`)
		var cfg config.MessengerConfig
		err := parser.Parse(content, &cfg)

		require.NoError(t, err)

		exchangeOptions, ok := cfg.Transports["amqp"].Options["exchange"].(map[string]interface{})
		require.True(t, ok)

		name, ok := exchangeOptions["name"]
		if ok && name != nil {
			assert.Equal(t, "messages", name)
		}

		exchangeType, ok := exchangeOptions["type"]
		if ok && exchangeType != nil {
			assert.Equal(t, "topic", exchangeType)
		}

		durable, ok := exchangeOptions["durable"]
		if ok && durable != nil {
			assert.True(t, durable.(bool))
		}

		autoDelete, ok := exchangeOptions["auto_delete"]
		if ok && autoDelete != nil {
			assert.False(t, autoDelete.(bool))
		}
	})

	t.Run("parse yaml with queue options", func(t *testing.T) {
		content := []byte(`
transports:
  amqp:
    dsn: amqp://localhost
    options:
      queues:
        default:
          binding_keys:
            - "#"
`)
		var cfg config.MessengerConfig
		err := parser.Parse(content, &cfg)

		require.NoError(t, err)

		queuesOptions, ok := cfg.Transports["amqp"].Options["queues"].(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, queuesOptions, "default")

		defaultQueue, ok := queuesOptions["default"].(map[string]interface{})
		require.True(t, ok)

		bindingKeys, ok := defaultQueue["binding_keys"].([]interface{})
		if ok && bindingKeys != nil {
			assert.Contains(t, bindingKeys, "#")
		}

		durable, ok := defaultQueue["durable"]
		if ok && durable != nil {
			assert.True(t, durable.(bool))
		}
	})

	t.Run("parse yaml with retry strategy", func(t *testing.T) {
		content := []byte(`
transports:
  amqp:
    dsn: amqp://localhost
    retry_strategy:
      max_retries: 3
      delay: 1s
      multiplier: 2
      max_delay: 60s
`)
		var cfg config.MessengerConfig
		err := parser.Parse(content, &cfg)

		require.NoError(t, err)
		assert.NotNil(t, cfg.Transports["amqp"].RetryStrategy)
		assert.Equal(t, uint(3), cfg.Transports["amqp"].RetryStrategy.MaxRetries)
		assert.InDelta(t, float64(2), cfg.Transports["amqp"].RetryStrategy.Multiplier, 0.001)
	})
}

func TestYAMLParser_Integration(t *testing.T) {
	parser := &config.YAMLParser{}
	reader := &config.FileReader{}

	t.Run("parse real config file", func(t *testing.T) {
		content, err := reader.Read("../tests/fixtures/configs/valid_config.yaml")
		require.NoError(t, err)

		var cfg config.MessengerConfig
		err = parser.Parse(content, &cfg)

		require.NoError(t, err)
		assert.Equal(t, "default", cfg.DefaultBus)
		assert.Equal(t, "failure", cfg.FailureTransport)
		assert.Len(t, cfg.Buses, 2)
		assert.Len(t, cfg.Transports, 2)
		assert.Len(t, cfg.Routing, 2)
	})
}
