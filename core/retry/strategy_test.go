package retry_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/core/retry"
)

func TestNewMultiplierRetryStrategy(t *testing.T) {
	t.Run("creates strategy with valid parameters", func(t *testing.T) {
		maxRetries := uint(3)
		delay := 100 * time.Millisecond
		multiplier := 2.0
		maxDelay := 1 * time.Second

		strategy := retry.NewMultiplierRetryStrategy(maxRetries, delay, multiplier, maxDelay)

		require.NotNil(t, strategy)
		assert.IsType(t, &retry.MultiplierRetryStrategy{}, strategy)
	})

	t.Run("creates strategy with zero values", func(t *testing.T) {
		strategy := retry.NewMultiplierRetryStrategy(0, 0, 0, 0)

		require.NotNil(t, strategy)
	})
}

func TestMultiplierRetryStrategy_ShouldRetry(t *testing.T) {
	t.Run("should retry within max retries", func(t *testing.T) {
		strategy := retry.NewMultiplierRetryStrategy(3, 100*time.Millisecond, 2.0, 1*time.Second)

		delay, shouldRetry := strategy.ShouldRetry(0)
		assert.True(t, shouldRetry)
		assert.Equal(t, 100*time.Millisecond, delay)

		delay, shouldRetry = strategy.ShouldRetry(1)
		assert.True(t, shouldRetry)
		assert.Equal(t, 200*time.Millisecond, delay)

		delay, shouldRetry = strategy.ShouldRetry(2)
		assert.True(t, shouldRetry)
		assert.Equal(t, 400*time.Millisecond, delay)
	})

	t.Run("should not retry when max retries exceeded", func(t *testing.T) {
		strategy := retry.NewMultiplierRetryStrategy(3, 100*time.Millisecond, 2.0, 1*time.Second)

		delay, shouldRetry := strategy.ShouldRetry(3)
		assert.False(t, shouldRetry)
		assert.Equal(t, time.Duration(0), delay)

		delay, shouldRetry = strategy.ShouldRetry(4)
		assert.False(t, shouldRetry)
		assert.Equal(t, time.Duration(0), delay)
	})

	t.Run("should respect max delay", func(t *testing.T) {
		strategy := retry.NewMultiplierRetryStrategy(10, 100*time.Millisecond, 2.0, 500*time.Millisecond)

		delay, shouldRetry := strategy.ShouldRetry(0)
		assert.True(t, shouldRetry)
		assert.Equal(t, 100*time.Millisecond, delay)

		delay, shouldRetry = strategy.ShouldRetry(1)
		assert.True(t, shouldRetry)
		assert.Equal(t, 200*time.Millisecond, delay)

		delay, shouldRetry = strategy.ShouldRetry(2)
		assert.True(t, shouldRetry)
		assert.Equal(t, 400*time.Millisecond, delay)

		delay, shouldRetry = strategy.ShouldRetry(3)
		assert.True(t, shouldRetry)
		assert.Equal(t, 500*time.Millisecond, delay)

		delay, shouldRetry = strategy.ShouldRetry(4)
		assert.True(t, shouldRetry)
		assert.Equal(t, 500*time.Millisecond, delay)
	})

	t.Run("should handle zero max retries", func(t *testing.T) {
		strategy := retry.NewMultiplierRetryStrategy(0, 100*time.Millisecond, 2.0, 1*time.Second)

		delay, shouldRetry := strategy.ShouldRetry(0)
		assert.False(t, shouldRetry)
		assert.Equal(t, time.Duration(0), delay)
	})

	t.Run("should handle zero delay", func(t *testing.T) {
		strategy := retry.NewMultiplierRetryStrategy(3, 0, 2.0, 1*time.Second)

		delay, shouldRetry := strategy.ShouldRetry(0)
		assert.True(t, shouldRetry)
		assert.Equal(t, time.Duration(0), delay)

		delay, shouldRetry = strategy.ShouldRetry(1)
		assert.True(t, shouldRetry)
		assert.Equal(t, time.Duration(0), delay)
	})

	t.Run("should handle zero multiplier", func(t *testing.T) {
		strategy := retry.NewMultiplierRetryStrategy(3, 100*time.Millisecond, 0, 1*time.Second)

		delay, shouldRetry := strategy.ShouldRetry(0)
		assert.True(t, shouldRetry)
		assert.Equal(t, 100*time.Millisecond, delay)

		delay, shouldRetry = strategy.ShouldRetry(1)
		assert.True(t, shouldRetry)
		assert.Equal(t, time.Duration(0), delay)
	})

	t.Run("should handle multiplier less than 1", func(t *testing.T) {
		strategy := retry.NewMultiplierRetryStrategy(3, 100*time.Millisecond, 0.5, 1*time.Second)

		delay, shouldRetry := strategy.ShouldRetry(0)
		assert.True(t, shouldRetry)
		assert.Equal(t, 100*time.Millisecond, delay)

		delay, shouldRetry = strategy.ShouldRetry(1)
		assert.True(t, shouldRetry)
		assert.Equal(t, 50*time.Millisecond, delay)

		delay, shouldRetry = strategy.ShouldRetry(2)
		assert.True(t, shouldRetry)
		assert.Equal(t, 25*time.Millisecond, delay)
	})

	t.Run("should handle large retry counts", func(t *testing.T) {
		strategy := retry.NewMultiplierRetryStrategy(1000, 1*time.Millisecond, 1.1, 10*time.Second)

		delay, shouldRetry := strategy.ShouldRetry(100)
		assert.True(t, shouldRetry)
		assert.Equal(t, 10*time.Second, delay)
	})
}

func TestMultiplierRetryStrategy_EdgeCases(t *testing.T) {
	t.Run("should handle very large multiplier", func(t *testing.T) {
		strategy := retry.NewMultiplierRetryStrategy(5, 1*time.Millisecond, 1000.0, 1*time.Second)

		delay, shouldRetry := strategy.ShouldRetry(0)
		assert.True(t, shouldRetry)
		assert.Equal(t, 1*time.Millisecond, delay)

		delay, shouldRetry = strategy.ShouldRetry(1)
		assert.True(t, shouldRetry)
		assert.Equal(t, 1*time.Second, delay)
	})

	t.Run("should handle zero max delay", func(t *testing.T) {
		strategy := retry.NewMultiplierRetryStrategy(3, 100*time.Millisecond, 2.0, 0)

		delay, shouldRetry := strategy.ShouldRetry(0)
		assert.True(t, shouldRetry)
		assert.Equal(t, time.Duration(0), delay)

		delay, shouldRetry = strategy.ShouldRetry(1)
		assert.True(t, shouldRetry)
		assert.Equal(t, time.Duration(0), delay)
	})
}
