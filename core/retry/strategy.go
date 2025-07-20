package retry

import (
	"math"
	"time"
)

type Strategy interface {
	ShouldRetry(attempt uint) (time.Duration, bool)
}
type MultiplierRetryStrategy struct {
	MaxRetries uint
	Delay      time.Duration
	Multiplier float64
	MaxDelay   time.Duration
}

func NewMultiplierRetryStrategy(
	maxRetries uint,
	delay time.Duration,
	multiplier float64,
	maxDelay time.Duration,
) Strategy {
	return &MultiplierRetryStrategy{
		MaxRetries: maxRetries,
		Delay:      delay,
		Multiplier: multiplier,
		MaxDelay:   maxDelay,
	}
}

func (s *MultiplierRetryStrategy) ShouldRetry(retryCount uint) (time.Duration, bool) {
	if retryCount >= s.MaxRetries {
		return 0, false
	}

	delay := float64(s.Delay) * math.Pow(s.Multiplier, float64(retryCount))
	if time.Duration(delay) > s.MaxDelay {
		return s.MaxDelay, true
	}

	return time.Duration(delay), true
}
