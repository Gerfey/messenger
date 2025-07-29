package inmemory

import (
	"context"
	"sync"
	"time"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/stamps"
)

const sleepDuration = 10 * time.Millisecond

type Transport struct {
	name  string
	queue []api.Envelope
	lock  sync.Mutex
}

func NewTransport(name string) api.Transport {
	return &Transport{
		name:  name,
		queue: make([]api.Envelope, 0),
	}
}

func (t *Transport) Name() string {
	return t.name
}

func (t *Transport) Send(_ context.Context, env api.Envelope) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.queue = append(t.queue, env)

	return nil
}

func (t *Transport) Receive(ctx context.Context, handler func(context.Context, api.Envelope) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			t.lock.Lock()
			if len(t.queue) == 0 {
				t.lock.Unlock()

				time.Sleep(sleepDuration)

				continue
			}

			env := t.queue[0]
			t.queue = t.queue[1:]
			t.lock.Unlock()

			envWithReceivedStamp := env.WithStamp(stamps.ReceivedStamp{Transport: t.name})

			if err := handler(ctx, envWithReceivedStamp); err != nil {
				return err
			}
		}
	}
}
