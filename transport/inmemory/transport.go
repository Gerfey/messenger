package inmemory

import (
	"context"
	"sync"

	"github.com/gerfey/messenger/api"
)

type Transport struct {
	cfg   TransportConfig
	queue []api.Envelope
	lock  sync.Mutex
}

func NewTransport(cfg TransportConfig) api.Transport {
	return &Transport{
		cfg:   cfg,
		queue: []api.Envelope{},
	}
}

func (t *Transport) Name() string {
	return t.cfg.Name
}

func (t *Transport) Send(_ context.Context, env api.Envelope) error {
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
