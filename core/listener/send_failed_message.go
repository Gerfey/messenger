package listener

import (
	"context"
	"fmt"
	"time"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/event"
	"github.com/gerfey/messenger/core/retry"
	"github.com/gerfey/messenger/core/stamps"
)

type SendFailedMessageForRetryListener struct {
	transportName    string
	transport        api.RetryableTransport
	failureTransport api.Transport
	retryStrategy    retry.RetryStrategy
}

func NewSendFailedMessageForRetryListener(
	transportName string,
	transport api.RetryableTransport,
	failureTransport api.Transport,
	strategy retry.RetryStrategy,
) *SendFailedMessageForRetryListener {
	return &SendFailedMessageForRetryListener{
		transportName:    transportName,
		transport:        transport,
		failureTransport: failureTransport,
		retryStrategy:    strategy,
	}
}

func (l *SendFailedMessageForRetryListener) Handle(ctx context.Context, evt event.SendFailedMessageEvent) {
	env := evt.Envelope

	receivedStamp, ok := envelope.LastStampOf[stamps.ReceivedStamp](evt.Envelope)
	if !ok {
		return
	}

	if receivedStamp.Transport != l.transportName {
		return
	}

	var nextRetry uint = 0
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

	delay, shouldRetry := l.retryStrategy.ShouldRetry(nextRetry, evt.Error)
	if !shouldRetry {
		if l.failureTransport != nil {
			err := l.failureTransport.Send(ctx, env)
			if err != nil {
				fmt.Printf("failed to send message to failure transport: %v\n", err)
			}
		}

		return
	}

	newEnv := env.WithStamp(stamps.RedeliveryStamp{RetryCount: nextRetry})

	time.AfterFunc(delay, func() {
		err := l.transport.Retry(ctx, newEnv)
		if err != nil {
			fmt.Printf("retry dispatch failed: %v\n", err)
		}
	})
}
