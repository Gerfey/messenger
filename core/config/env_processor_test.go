package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/core/config"
)

func TestEnvVarProcessor_Process(t *testing.T) {
	processor := &config.EnvVarProcessor{}

	t.Run("replace single env variable", func(t *testing.T) {
		t.Setenv("TEST_VAR", "test_value")
		defer func() {
			_ = os.Unsetenv("TEST_VAR")
		}()

		content := []byte("test_url: %env(TEST_VAR)%")
		result, err := processor.Process(content)

		require.NoError(t, err)
		assert.Equal(t, "test_url: test_value", string(result))
	})

	t.Run("replace multiple env variables", func(t *testing.T) {
		t.Setenv("HOST", "localhost")
		t.Setenv("PORT", "1234")
		defer func() {
			_ = os.Unsetenv("HOST")
			_ = os.Unsetenv("PORT")
		}()

		content := []byte("dsn: memory://user:pass@%env(HOST)%:%env(PORT)%")
		result, err := processor.Process(content)

		require.NoError(t, err)
		assert.Equal(t, "dsn: memory://user:pass@localhost:1234", string(result))
	})

	t.Run("replace same env variable multiple times", func(t *testing.T) {
		t.Setenv("COMMON_VALUE", "shared")
		defer func() {
			_ = os.Unsetenv("COMMON_VALUE")
		}()

		content := []byte("key1: %env(COMMON_VALUE)%\nkey2: %env(COMMON_VALUE)%")
		result, err := processor.Process(content)

		require.NoError(t, err)
		assert.Equal(t, "key1: shared\nkey2: shared", string(result))
	})

	t.Run("replace with empty env variable", func(t *testing.T) {
		t.Setenv("EMPTY_VAR", "")
		defer func() {
			_ = os.Unsetenv("EMPTY_VAR")
		}()

		content := []byte("value: %env(EMPTY_VAR)%")
		result, err := processor.Process(content)

		require.NoError(t, err)
		assert.Equal(t, "value: ", string(result))
	})

	t.Run("replace with non-existing env variable", func(t *testing.T) {
		_ = os.Unsetenv("NON_EXISTING_VAR")

		content := []byte("value: %env(NON_EXISTING_VAR)%")
		result, err := processor.Process(content)

		require.NoError(t, err)
		assert.Equal(t, "value: ", string(result))
	})

	t.Run("no env variables to replace", func(t *testing.T) {
		content := []byte("simple: value\nother: data")
		result, err := processor.Process(content)

		require.NoError(t, err)
		assert.Equal(t, "simple: value\nother: data", string(result))
	})

	t.Run("invalid env variable format ignored", func(t *testing.T) {
		content := []byte("invalid1: %env%\ninvalid2: %env()%\ninvalid3: %env(lowercase)%")
		result, err := processor.Process(content)

		require.NoError(t, err)
		assert.Equal(t, "invalid1: %env%\ninvalid2: %env()%\ninvalid3: %env(lowercase)%", string(result))
	})

	t.Run("mixed valid and invalid env variables", func(t *testing.T) {
		t.Setenv("VALID_VAR", "valid_value")
		defer func() {
			_ = os.Unsetenv("VALID_VAR")
		}()

		content := []byte("valid: %env(VALID_VAR)%\ninvalid: %env(lowercase)%")
		result, err := processor.Process(content)

		require.NoError(t, err)
		assert.Equal(t, "valid: valid_value\ninvalid: %env(lowercase)%", string(result))
	})

	t.Run("numeric env variable name replaced with empty", func(t *testing.T) {
		_ = os.Unsetenv("123")

		content := []byte("numeric: %env(123)%")
		result, err := processor.Process(content)

		require.NoError(t, err)
		assert.Equal(t, "numeric: ", string(result))
	})

	t.Run("env variable with underscores and numbers", func(t *testing.T) {
		t.Setenv("TEST_VAR_123", "complex_value")
		defer func() {
			_ = os.Unsetenv("TEST_VAR_123")
		}()

		content := []byte("complex: %env(TEST_VAR_123)%")
		result, err := processor.Process(content)

		require.NoError(t, err)
		assert.Equal(t, "complex: complex_value", string(result))
	})

	t.Run("process empty content", func(t *testing.T) {
		content := []byte("")
		result, err := processor.Process(content)

		require.NoError(t, err)
		assert.Empty(t, string(result))
	})
}

func TestEnvVarProcessor_Integration(t *testing.T) {
	processor := &config.EnvVarProcessor{}

	t.Run("process realistic config with env variables", func(t *testing.T) {
		t.Setenv("DB_HOST", "localhost")
		t.Setenv("DB_PORT", "5432")
		t.Setenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
		defer func() {
			_ = os.Unsetenv("DB_HOST")
			_ = os.Unsetenv("DB_PORT")
			_ = os.Unsetenv("RABBITMQ_URL")
		}()

		content := []byte(`
default_bus: default
transports:
  database:
    dsn: postgres://user:pass@%env(DB_HOST)%:%env(DB_PORT)%/messenger
  amqp:
    dsn: %env(RABBITMQ_URL)%
routing:
  "MyMessage": database
`)

		result, err := processor.Process(content)

		require.NoError(t, err)
		expected := `
default_bus: default
transports:
  database:
    dsn: postgres://user:pass@localhost:5432/messenger
  amqp:
    dsn: amqp://guest:guest@localhost:5672/
routing:
  "MyMessage": database
`
		assert.Equal(t, expected, string(result))
	})
}
