package config_test

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/config"
)

type MockProcessor struct {
	ProcessFunc func(content []byte) ([]byte, error)
}

func (p *MockProcessor) Process(content []byte) ([]byte, error) {
	if p.ProcessFunc != nil {
		return p.ProcessFunc(content)
	}

	return content, nil
}

func TestFileReader_Read(t *testing.T) {
	reader := &config.FileReader{}

	t.Run("read existing file", func(t *testing.T) {
		path := "../tests/fixtures/configs/valid_config.yaml"
		content, err := reader.Read(path)

		require.NoError(t, err)
		assert.NotEmpty(t, content)
		assert.Contains(t, string(content), "default_bus: default")
	})

	t.Run("read non-existing file", func(t *testing.T) {
		path := "non_existing_file.yaml"
		content, err := reader.Read(path)

		require.Error(t, err)
		assert.True(t, os.IsNotExist(err))
		assert.Empty(t, content)
	})

	t.Run("read with single processor", func(t *testing.T) {
		path := "../tests/fixtures/configs/valid_config.yaml"
		processor := &MockProcessor{
			ProcessFunc: func(content []byte) ([]byte, error) {
				return append(content, []byte("\n# Processed")...), nil
			},
		}

		content, err := reader.Read(path, processor)

		require.NoError(t, err)
		assert.Contains(t, string(content), "# Processed")
	})

	t.Run("read with multiple processors", func(t *testing.T) {
		path := "../tests/fixtures/configs/valid_config.yaml"
		processor1 := &MockProcessor{
			ProcessFunc: func(content []byte) ([]byte, error) {
				return append(content, []byte("\n# Processor1")...), nil
			},
		}
		processor2 := &MockProcessor{
			ProcessFunc: func(content []byte) ([]byte, error) {
				return append(content, []byte("\n# Processor2")...), nil
			},
		}

		content, err := reader.Read(path, processor1, processor2)

		require.NoError(t, err)
		assert.Contains(t, string(content), "# Processor1")
		assert.Contains(t, string(content), "# Processor2")
	})

	t.Run("read with processor error", func(t *testing.T) {
		path := "../tests/fixtures/configs/valid_config.yaml"
		expectedErr := errors.New("processor error")
		processor := &MockProcessor{
			ProcessFunc: func(_ []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}

		content, err := reader.Read(path, processor)

		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Empty(t, content)
	})

	t.Run("read with multiple processors, second fails", func(t *testing.T) {
		path := "../tests/fixtures/configs/valid_config.yaml"
		expectedErr := errors.New("processor 2 error")
		processor1 := &MockProcessor{
			ProcessFunc: func(content []byte) ([]byte, error) {
				return append(content, []byte("\n# Processor1")...), nil
			},
		}
		processor2 := &MockProcessor{
			ProcessFunc: func(_ []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}

		content, err := reader.Read(path, processor1, processor2)

		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Empty(t, content)
	})

	t.Run("read empty file", func(t *testing.T) {
		tempDir := t.TempDir()
		tempFile, err := os.CreateTemp(tempDir, "empty_config*.yaml")
		require.NoError(t, err)
		defer func(name string) {
			_ = os.Remove(name)
		}(tempFile.Name())
		_ = tempFile.Close()

		content, err := reader.Read(tempFile.Name())

		require.NoError(t, err)
		assert.Empty(t, content)
	})
}

func TestFileReader_Integration(t *testing.T) {
	reader := &config.FileReader{}

	t.Run("read config with env processor", func(t *testing.T) {
		path := "../tests/fixtures/configs/config_with_env.yaml"

		t.Setenv("MEMORY_HOST", "localhost")
		t.Setenv("CONSUMER_POOL_SIZE", "20")

		processor := &config.EnvVarProcessor{}

		content, err := reader.Read(path, processor)

		require.NoError(t, err)
		assert.Contains(t, string(content), "memory://localhost")
		assert.Contains(t, string(content), "consumer_pool_size: 20")
	})

	t.Run("read config with multiple processors", func(t *testing.T) {
		path := "../tests/fixtures/configs/config_with_env.yaml"

		t.Setenv("MEMORY_HOST", "localhost")

		envProcessor := &config.EnvVarProcessor{}
		customProcessor := &MockProcessor{
			ProcessFunc: func(content []byte) ([]byte, error) {
				modified := []byte(string(content))

				return modified, nil
			},
		}

		content, err := reader.Read(path, envProcessor, customProcessor)

		require.NoError(t, err)
		assert.Contains(t, string(content), "memory://localhost")
	})
}
