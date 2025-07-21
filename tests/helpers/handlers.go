package helpers

import "context"

func TestEventListener(ctx context.Context, msg *TestMessage) error {
	return nil
}

func AnotherTestEventListener(ctx context.Context, msg *ComplexMessage) error {
	return nil
}

type ValidHandler struct{}

func (h *ValidHandler) Handle(ctx context.Context, msg *TestMessage) error {
	return nil
}

type ValidHandlerWithResult struct{}

func (h *ValidHandlerWithResult) Handle(ctx context.Context, msg *TestMessage) (*TestMessage, error) {
	return msg, nil
}

type ValidHandlerWithBusName struct{}

func (h *ValidHandlerWithBusName) Handle(ctx context.Context, msg *TestMessage) error {
	return nil
}

func (h *ValidHandlerWithBusName) GetBusName() string {
	return "test-bus"
}

type AnotherValidHandler struct{}

func (h *AnotherValidHandler) Handle(ctx context.Context, msg *ComplexMessage) error {
	return nil
}

type InvalidHandlerNoMethod struct{}

type InvalidHandlerWrongParams struct{}

func (h *InvalidHandlerWrongParams) Handle(msg *TestMessage) error {
	return nil
}

type InvalidHandlerWrongFirstParam struct{}

func (h *InvalidHandlerWrongFirstParam) Handle(wrongType string, msg *TestMessage) error {
	return nil
}

type InvalidHandlerTooManyParams struct{}

func (h *InvalidHandlerTooManyParams) Handle(ctx context.Context, msg *TestMessage, extra string) error {
	return nil
}

type InvalidHandlerNoReturn struct{}

func (h *InvalidHandlerNoReturn) Handle(ctx context.Context, msg *TestMessage) {
}

type InvalidHandlerWrongReturn struct{}

func (h *InvalidHandlerWrongReturn) Handle(ctx context.Context, msg *TestMessage) string {
	return ""
}

type InvalidHandlerTooManyReturns struct{}

func (h *InvalidHandlerTooManyReturns) Handle(ctx context.Context, msg *TestMessage) (string, int, error) {
	return "", 0, nil
}

type TestMessageHandler struct {
	CallCount int
}

func (h *TestMessageHandler) Handle(ctx context.Context, msg *TestMessage) error {
	h.CallCount++
	return nil
}

type AnotherTestMessageHandler struct {
	CallCount int
}

func (h *AnotherTestMessageHandler) Handle(ctx context.Context, msg *TestMessage) error {
	h.CallCount++
	return nil
}

type ErrorTestMessageHandler struct {
	CallCount int
	Error     error
}

func (h *ErrorTestMessageHandler) Handle(ctx context.Context, msg *TestMessage) error {
	h.CallCount++
	return h.Error
}

type ResultTestMessageHandler struct {
	CallCount int
	Result    any
}

func (h *ResultTestMessageHandler) Handle(ctx context.Context, msg *TestMessage) (any, error) {
	h.CallCount++
	return h.Result, nil
}
