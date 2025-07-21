package handlers

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/gerfey/messenger/tests/helpers"
)

type E2ETestHandler struct {
	callCount   int64
	lastMessage *helpers.TestMessage
	mu          sync.RWMutex
}

func NewE2ETestHandler() *E2ETestHandler {
	return &E2ETestHandler{}
}

func (h *E2ETestHandler) Handle(ctx context.Context, msg *helpers.TestMessage) error {
	atomic.AddInt64(&h.callCount, 1)

	h.mu.Lock()
	h.lastMessage = msg
	h.mu.Unlock()

	return nil
}

func (h *E2ETestHandler) GetCallCount() int64 {
	return atomic.LoadInt64(&h.callCount)
}

func (h *E2ETestHandler) GetLastMessage() *helpers.TestMessage {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastMessage
}
