package inmemory

import (
	"context"
	"sync"

	"github.com/gerfey/messenger/api"
)

type Transport struct {
	queue []api.Envelope
	lock  sync.Mutex
}

func NewTransport() api.Transport {
	return &Transport{
		queue: []api.Envelope{},
	}
}

func (t *Transport) Send(ctx context.Context, env api.Envelope) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.queue = append(t.queue, env)

	return nil
}

func (t *Transport) Receive(ctx context.Context, handler func(context.Context, api.Envelope) error) error {
	for {
		t.lock.Lock()

		if len(t.queue) == 0 {
			t.lock.Unlock()

			return nil
		}

		env := t.queue[0]
		t.queue = t.queue[1:]
		t.lock.Unlock()

		if err := handler(ctx, env); err != nil {
			return err
		}
	}
}
