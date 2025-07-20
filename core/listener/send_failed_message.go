package listener

import (
	"context"
	"log/slog"
	"time"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/event"
	"github.com/gerfey/messenger/core/retry"
	"github.com/gerfey/messenger/core/stamps"
)

type SendFailedMessageForRetryListener struct {
	transport        api.RetryableTransport
	failureTransport api.Transport
	retryStrategy    retry.Strategy
	logger           *slog.Logger
}

func NewSendFailedMessageForRetryListener(
	transport api.RetryableTransport,
	failureTransport api.Transport,
	strategy retry.Strategy,
	logger *slog.Logger,
) *SendFailedMessageForRetryListener {
	return &SendFailedMessageForRetryListener{
		transport:        transport,
		failureTransport: failureTransport,
		retryStrategy:    strategy,
		logger:           logger,
	}
}

func (l *SendFailedMessageForRetryListener) Handle(ctx context.Context, evt event.SendFailedMessageEvent) {
	env := evt.Envelope

	receivedStamp, ok := envelope.LastStampOf[stamps.ReceivedStamp](evt.Envelope)
	if !ok {
		return
	}

	if receivedStamp.Transport != evt.TransportName {
		return
	}

	var nextRetry uint
	retryStamp, ok := envelope.LastStampOf[stamps.RedeliveryStamp](env)
	if ok {
		nextRetry = retryStamp.RetryCount + 1
	}

	errorStamp := stamps.ErrorDetailsStamp{
		ErrorMessage: evt.Error.Error(),
		FailedAt:     time.Now(),
		RetryCount:   nextRetry,
	}
	env = env.WithStamp(errorStamp)

	delay, shouldRetry := l.retryStrategy.ShouldRetry(nextRetry)
	if !shouldRetry {
		if l.failureTransport != nil {
			err := l.failureTransport.Send(ctx, env)
			if err != nil {
				l.logger.ErrorContext(ctx, "failed to send message to failure transport", "error", err)
			}
		}

		return
	}

	newEnv := env.WithStamp(stamps.RedeliveryStamp{RetryCount: nextRetry})

	time.AfterFunc(delay, func() {
		err := l.transport.Retry(ctx, newEnv)
		if err != nil {
			l.logger.ErrorContext(ctx, "retry dispatch failed", "error", err)
		}
	})
}
