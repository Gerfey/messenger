package transport

import (
	"context"
	"fmt"
	"sync"

	"github.com/gerfey/messenger/api"
)

type Manager struct {
	transports []api.Transport
	handler    func(context.Context, api.Envelope) error
	wg         sync.WaitGroup
	mu         sync.Mutex
	running    bool
}

func NewManager(handler func(context.Context, api.Envelope) error) *Manager {
	return &Manager{
		handler: handler,
	}
}

func (m *Manager) AddTransport(t api.Transport) {
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

func (m *Manager) receiveTransport(ctx context.Context, t api.Transport) {
	m.wg.Add(1)
	go func(t api.Transport) {
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
