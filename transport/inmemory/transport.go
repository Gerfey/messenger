package inmemory

import (
	"context"
	"sync"
	"time"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/stamps"
)

type Transport struct {
	cfg   TransportConfig
	queue []api.Envelope
	lock  sync.Mutex
}

func NewTransport(cfg TransportConfig) api.Transport {
	return &Transport{
		cfg:   cfg,
		queue: make([]api.Envelope, 0),
	}
}

func (t *Transport) Name() string {
	return t.cfg.Name
}

func (t *Transport) Send(ctx context.Context, env api.Envelope) error {
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
				
				time.Sleep(10 * time.Millisecond)
				
				continue
			}

			env := t.queue[0]
			t.queue = t.queue[1:]
			t.lock.Unlock()

			envWithReceivedStamp := env.WithStamp(stamps.ReceivedStamp{Transport: t.cfg.Name})

			if err := handler(ctx, envWithReceivedStamp); err != nil {
				return err
			}
		}
	}
}
