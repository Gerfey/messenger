package helpers

import (
	"context"
	"errors"

	"github.com/gerfey/messenger/api"
)

type TestMessage struct {
	ID      string
	Content string
}

type SimpleMessage string

type ComplexMessage struct {
	ID       string
	Type     string
	Metadata map[string]string
	Payload  any
}

type TestStamp struct {
	Value string
}

type AnotherStamp struct {
	Number int
}

func NewTestMessage(id, content string) *TestMessage {
	return &TestMessage{
		ID:      id,
		Content: content,
	}
}

func NewComplexMessage(id, msgType string) *ComplexMessage {
	return &ComplexMessage{
		ID:       id,
		Type:     msgType,
		Metadata: make(map[string]string),
		Payload:  nil,
	}
}

type TestEvent struct {
	ID      string
	Message string
}

type AnotherTestEvent struct {
	Value int
	Data  string
}

type ErrorEvent struct {
	ShouldFail bool
}

func SimpleEventListener(_ *TestEvent) error {
	return nil
}

func TestEventListenerWithContext(_ context.Context, _ *TestEvent) error {
	return nil
}

func ErrorEventListener(event *ErrorEvent) error {
	if event.ShouldFail {
		return errors.New("listener error")
	}

	return nil
}

type TestEventHandler struct {
	CallCount int
}

func (h *TestEventHandler) Handle(_ *TestEvent) error {
	h.CallCount++

	return nil
}

type TestEventHandlerWithContext struct {
	CallCount int
}

func (h *TestEventHandlerWithContext) Handle(_ context.Context, _ *TestEvent) error {
	h.CallCount++

	return nil
}

type ErrorEventHandler struct {
	ShouldFail bool
}

func (h *ErrorEventHandler) Handle(_ *ErrorEvent) error {
	if h.ShouldFail {
		return errors.New("handler error")
	}

	return nil
}

type InvalidEventHandler struct{}

type InvalidEventHandlerWrongSignature struct{}

func (h *InvalidEventHandlerWrongSignature) Handle() error {
	return nil
}

type InvalidEventHandlerTooManyParams struct{}

func (h *InvalidEventHandlerTooManyParams) Handle(_ context.Context, _ *TestEvent, _ string) error {
	return nil
}

type TestMiddleware struct {
	CallCount   int
	LastMessage any
	Called      bool
}

func (m *TestMiddleware) Handle(ctx context.Context, env api.Envelope, next api.NextFunc) (api.Envelope, error) {
	m.CallCount++
	m.LastMessage = env.Message()
	m.Called = true

	return next(ctx, env)
}

type ErrorMiddleware struct {
	Error error
}

func (m *ErrorMiddleware) Handle(_ context.Context, _ api.Envelope, _ api.NextFunc) (api.Envelope, error) {
	return nil, m.Error
}

type ContextMiddleware struct{}

func (m *ContextMiddleware) Handle(ctx context.Context, env api.Envelope, next api.NextFunc) (api.Envelope, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return next(ctx, env)
}

type OrderedMiddleware struct {
	Name           string
	ExecutionOrder *[]string
}

func (m *OrderedMiddleware) Handle(ctx context.Context, env api.Envelope, next api.NextFunc) (api.Envelope, error) {
	*m.ExecutionOrder = append(*m.ExecutionOrder, m.Name)

	return next(ctx, env)
}

type TestTransport struct {
	TransportName string
	Messages      []api.Envelope
	IsStarted     bool
	IsStopped     bool
	SendError     error
}

func (t *TestTransport) Name() string {
	if t.TransportName != "" {
		return t.TransportName
	}

	return "test-transport"
}

func (t *TestTransport) Send(_ context.Context, env api.Envelope) error {
	if t.SendError != nil {
		return t.SendError
	}

	t.Messages = append(t.Messages, env)

	return nil
}

func (t *TestTransport) Receive(_ context.Context, _ func(context.Context, api.Envelope) error) error {
	return nil
}

func (t *TestTransport) Start(_ context.Context) error {
	t.IsStarted = true

	return nil
}

func (t *TestTransport) Stop() error {
	t.IsStopped = true

	return nil
}

type TestTransportFactory struct {
	TransportName string
	Transport     api.Transport
	CreateError   error
}

func (f *TestTransportFactory) Name() string {
	return f.TransportName
}

func (f *TestTransportFactory) Supports(_ string) bool {
	return true
}

func (f *TestTransportFactory) Create(_ string, _ string, _ []byte, _ api.Serializer) (api.Transport, error) {
	if f.CreateError != nil {
		return nil, f.CreateError
	}

	return f.Transport, nil
}
