package transport

import (
	"context"
	"fmt"
	"sync"

	"github.com/gerfey/messenger/envelope"
)

type Manager struct {
	transports []Transport
	handler    func(context.Context, *envelope.Envelope) error
	wg         sync.WaitGroup
	mu         sync.Mutex
	running    bool
}

func NewManager(handler func(context.Context, *envelope.Envelope) error) *Manager {
	return &Manager{
		handler: handler,
	}
}

func (m *Manager) AddTransport(t Transport) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.transports = append(m.transports, t)

	if m.running {
		m.receiveTransport(context.Background(), t)
	}
}

func (m *Manager) Start(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return
	}

	m.running = true

	for _, t := range m.transports {
		m.receiveTransport(ctx, t)
	}
}

func (m *Manager) receiveTransport(ctx context.Context, t Transport) {
	m.wg.Add(1)
	go func(t Transport) {
		defer m.wg.Done()

		err := t.Receive(ctx, m.handler)
		if err != nil {
			_ = fmt.Errorf("receive: %w", err)
		}
	}(t)
}

func (m *Manager) Stop() {
	m.mu.Lock()
	m.running = false
	m.mu.Unlock()

	m.wg.Wait()
}
