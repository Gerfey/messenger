package helpers

import (
	"context"
	"errors"
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
