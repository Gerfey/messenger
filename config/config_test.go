package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/config"
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
		path := "../tests/fixtures/configs/valid_config.yaml"
		cfg, err := config.LoadConfig(path)

		require.NoError(t, err)
		assert.Equal(t, "default", cfg.DefaultBus)
		assert.Equal(t, "failure", cfg.FailureTransport)
		assert.Len(t, cfg.Buses, 2)
		assert.Len(t, cfg.Transports, 2)
		assert.Len(t, cfg.Routing, 2)
	})

	t.Run("load config with env variables", func(t *testing.T) {
		path := "../tests/fixtures/configs/config_with_env.yaml"

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
		assert.Equal(t, 15, cfg.Transports["default"].Options.ConsumerPoolSize)
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
		path := "../tests/fixtures/configs/invalid_config.yaml"
		cfg, err := config.LoadConfig(path)

		require.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("load with custom processor", func(t *testing.T) {
		path := "../tests/fixtures/configs/valid_config.yaml"

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
		assert.True(t, cfg.Transports["default"].Options.AutoSetup)
		assert.Equal(t, 10, cfg.Transports["default"].Options.ConsumerPoolSize)
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
		assert.Equal(t, "messages", cfg.Transports["amqp"].Options.Exchange.Name)
		assert.Equal(t, "topic", cfg.Transports["amqp"].Options.Exchange.Type)
		assert.True(t, cfg.Transports["amqp"].Options.Exchange.Durable)
		assert.False(t, cfg.Transports["amqp"].Options.Exchange.AutoDelete)
		assert.False(t, cfg.Transports["amqp"].Options.Exchange.Internal)
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
		assert.Contains(t, cfg.Transports["amqp"].Options.Queues["default"].BindingKeys, "#")
		assert.True(t, cfg.Transports["amqp"].Options.Queues["default"].Durable)
		assert.False(t, cfg.Transports["amqp"].Options.Queues["default"].Exclusive)
		assert.False(t, cfg.Transports["amqp"].Options.Queues["default"].AutoDelete)
	})
}
