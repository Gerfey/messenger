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
	transportName string
	transport     api.RetryableTransport
	retryStrategy retry.RetryStrategy
}

func NewSendFailedMessageForRetryListener(transportName string, transport api.RetryableTransport, strategy retry.RetryStrategy) *SendFailedMessageForRetryListener {
	return &SendFailedMessageForRetryListener{
		transportName: transportName,
		transport:     transport,
		retryStrategy: strategy,
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

	var next uint = 1

	retryStamp, ok := envelope.LastStampOf[stamps.RedeliveryStamp](env)
	if ok {
		next = retryStamp.RetryCount + 1
	}

	delay, shouldRetry := l.retryStrategy.ShouldRetry(next, evt.Error)
	if !shouldRetry {
		return
	}

	newEnv := env.WithStamp(stamps.RedeliveryStamp{RetryCount: next})

	time.AfterFunc(delay, func() {
		err := l.transport.Retry(ctx, newEnv)
		if err != nil {
			fmt.Printf("retry dispatch failed: %v\n", err)
		}
	})
}
