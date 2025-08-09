package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/core/config"
)

type testProcessor struct {
	processFunc func(content []byte) ([]byte, error)
}

func (p *testProcessor) Process(content []byte) ([]byte, error) {
	if p.processFunc != nil {
		return p.processFunc(content)
	}

	return content, nil
}

func TestLoadConfig(t *testing.T) {
	t.Run("load valid config", func(t *testing.T) {
		path := "../../tests/fixtures/configs/valid_config.yaml"
		cfg, err := config.LoadConfig(path)

		require.NoError(t, err)
		assert.Equal(t, "default", cfg.DefaultBus)
		assert.Equal(t, "failure", cfg.FailureTransport)
		assert.Len(t, cfg.Buses, 2)
		assert.Len(t, cfg.Transports, 2)
		assert.Len(t, cfg.Routing, 2)
	})

	t.Run("load config with env variables", func(t *testing.T) {
		path := "../../tests/fixtures/configs/config_with_env.yaml"

		t.Setenv("MEMORY_HOST", "test-host")
		t.Setenv("CONSUMER_POOL_SIZE", "15")
		t.Setenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
		t.Setenv("MAX_RETRIES", "5")
		defer func() {
			_ = os.Unsetenv("MEMORY_HOST")
			_ = os.Unsetenv("CONSUMER_POOL_SIZE")
			_ = os.Unsetenv("RABBITMQ_URL")
			_ = os.Unsetenv("MAX_RETRIES")
		}()

		cfg, err := config.LoadConfig(path)

		require.NoError(t, err)
		assert.Equal(t, "memory://test-host", cfg.Transports["default"].DSN)
		assert.Equal(t, 15, cfg.Transports["default"].Options["consumer_pool_size"])
		assert.Equal(t, "amqp://guest:guest@localhost:5672/", cfg.Transports["amqp"].DSN)
		assert.Equal(t, uint(5), cfg.Transports["amqp"].RetryStrategy.MaxRetries)
	})

	t.Run("load non-existent config", func(t *testing.T) {
		path := "non_existent_config.yaml"
		cfg, err := config.LoadConfig(path)

		require.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("load invalid config", func(t *testing.T) {
		path := "../../tests/fixtures/configs/invalid_config.yaml"
		cfg, err := config.LoadConfig(path)

		require.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("load with custom processor", func(t *testing.T) {
		path := "../../tests/fixtures/configs/valid_config.yaml"

		customProcessor := &testProcessor{
			processFunc: func(content []byte) ([]byte, error) {
				modified := string(content)
				modified = strings.Replace(modified, "handle_message", "custom_middleware", 1)

				return []byte(modified), nil
			},
		}

		cfg, err := config.LoadConfig(path, customProcessor)

		require.NoError(t, err)
		assert.Contains(t, cfg.Buses["default"].Middleware, "custom_middleware")
	})
}

func TestMessengerConfig(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		var cfg config.MessengerConfig

		parser := &config.YAMLParser{}
		err := parser.Parse([]byte(`{}`), &cfg)

		require.NoError(t, err)
		assert.Equal(t, "default", cfg.DefaultBus)
		assert.Empty(t, cfg.FailureTransport)
		assert.Empty(t, cfg.Buses)
		assert.Empty(t, cfg.Transports)
		assert.Empty(t, cfg.Routing)
	})

	t.Run("transport options default values", func(t *testing.T) {
		var cfg config.MessengerConfig

		parser := &config.YAMLParser{}
		err := parser.Parse([]byte(`
transports:
  default:
    dsn: memory://default
`), &cfg)

		require.NoError(t, err)
		assert.Equal(t, "memory://default", cfg.Transports["default"].DSN)
		autoSetup, ok := cfg.Transports["default"].Options["auto_setup"]
		if ok && autoSetup != nil {
			assert.True(t, autoSetup.(bool))
		}
		consumerPoolSize, ok := cfg.Transports["default"].Options["consumer_pool_size"]
		if ok && consumerPoolSize != nil {
			assert.Equal(t, 10, consumerPoolSize)
		}
	})

	t.Run("exchange options default values", func(t *testing.T) {
		var cfg config.MessengerConfig

		parser := &config.YAMLParser{}
		err := parser.Parse([]byte(`
transports:
  amqp:
    dsn: amqp://localhost
    options:
      exchange:
        name: messages
`), &cfg)

		require.NoError(t, err)
		exchangeOptions, ok := cfg.Transports["amqp"].Options["exchange"].(map[string]any)
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

		internal, ok := exchangeOptions["internal"]
		if ok && internal != nil {
			assert.False(t, internal.(bool))
		}
	})

	t.Run("queue options default values", func(t *testing.T) {
		var cfg config.MessengerConfig

		parser := &config.YAMLParser{}
		err := parser.Parse([]byte(`
transports:
  amqp:
    dsn: amqp://localhost
    options:
      queues:
        default:
          binding_keys:
            - "#"
`), &cfg)

		require.NoError(t, err)

		queuesOptions, ok := cfg.Transports["amqp"].Options["queues"].(map[string]any)
		require.True(t, ok)

		defaultQueue, ok := queuesOptions["default"].(map[string]any)
		require.True(t, ok)

		bindingKeys, ok := defaultQueue["binding_keys"]
		if ok && bindingKeys != nil {
			assert.Contains(t, bindingKeys, "#")
		}

		durable, ok := defaultQueue["durable"]
		if ok && durable != nil {
			assert.True(t, durable.(bool))
		}

		exclusive, ok := defaultQueue["exclusive"]
		if ok && exclusive != nil {
			assert.False(t, exclusive.(bool))
		}

		autoDelete, ok := defaultQueue["auto_delete"]
		if ok && autoDelete != nil {
			assert.False(t, autoDelete.(bool))
		}
	})
}
