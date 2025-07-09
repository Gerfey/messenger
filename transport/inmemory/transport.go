package inmemory

import (
	"context"
	"sync"

	"github.com/gerfey/messenger/envelope"
)

type InMemoryTransport struct {
	queue []*envelope.Envelope
	lock  sync.Mutex
}

func New() *InMemoryTransport {
	return &InMemoryTransport{
		queue: []*envelope.Envelope{},
	}
}

func (t *InMemoryTransport) Send(ctx context.Context, env *envelope.Envelope) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.queue = append(t.queue, env)

	return nil
}

func (t *InMemoryTransport) Receive(ctx context.Context, handler func(context.Context, *envelope.Envelope) error) error {
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
