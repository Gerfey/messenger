package transport

import (
	"context"
	"log/slog"
	"sync"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/event"
)

type Manager struct {
	transports      []api.Transport
	handler         func(context.Context, api.Envelope) error
	eventDispatcher api.EventDispatcher
	logger          *slog.Logger
	wg              sync.WaitGroup
	mu              sync.Mutex
	running         bool
}

func NewManager(
	handler func(context.Context, api.Envelope) error,
	eventDispatcher api.EventDispatcher,
	logger *slog.Logger,
) *Manager {
	return &Manager{
		handler:         handler,
		eventDispatcher: eventDispatcher,
		logger:          logger,
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

func (m *Manager) Start(ctx context.Context, consumeOnly []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return
	}

	m.running = true

	for _, t := range m.transports {
		if !m.stringInSlice(t.Name(), consumeOnly) {
			continue
		}
		m.receiveTransport(ctx, t)
	}
}

func (m *Manager) Stop() {
	m.mu.Lock()
	m.running = false
	m.mu.Unlock()

	m.wg.Wait()
}

func (m *Manager) HasTransport(name string) bool {
	for _, transport := range m.transports {
		if transport.Name() == name {
			return true
		}
	}

	return false
}

func (m *Manager) receiveTransport(ctx context.Context, t api.Transport) {
	m.wg.Add(1)
	go func(t api.Transport) {
		defer m.wg.Done()

		err := t.Receive(ctx, func(ctx context.Context, env api.Envelope) error {
			errMessageReceived := m.eventDispatcher.Dispatch(ctx, event.WorkerMessageReceivedEvent{
				Ctx:           ctx,
				Envelope:      env,
				TransportName: t.Name(),
			})
			if errMessageReceived != nil {
				return errMessageReceived
			}

			err := m.handler(ctx, env)

			if err != nil {
				errMessageFailed := m.eventDispatcher.Dispatch(ctx, event.WorkerMessageFailedEvent{
					Ctx:           ctx,
					Envelope:      env,
					TransportName: t.Name(),
					Error:         err,
				})
				if errMessageFailed != nil {
					return errMessageFailed
				}

				errSendFailed := m.eventDispatcher.Dispatch(ctx, event.SendFailedMessageEvent{
					Envelope:      env,
					Error:         err,
					TransportName: t.Name(),
				})
				if errSendFailed != nil {
					return errSendFailed
				}
			} else {
				errMessageHandled := m.eventDispatcher.Dispatch(ctx, event.WorkerMessageHandledEvent{
					Ctx:           ctx,
					Envelope:      env,
					TransportName: t.Name(),
				})
				if errMessageHandled != nil {
					return errMessageHandled
				}
			}

			return err
		})

		if err != nil {
			m.logger.Error("receive error", "error", err)
		}
	}(t)
}

func (m *Manager) stringInSlice(s string, list []string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}

	return false
}
