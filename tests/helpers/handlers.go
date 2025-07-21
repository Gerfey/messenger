package helpers

import "context"

func TestEventListener(_ context.Context, _ *TestMessage) error {
	return nil
}

func AnotherTestEventListener(_ context.Context, _ *ComplexMessage) error {
	return nil
}

type ValidHandler struct{}

func (h *ValidHandler) Handle(_ context.Context, _ *TestMessage) error {
	return nil
}

type ValidHandlerWithResult struct{}

func (h *ValidHandlerWithResult) Handle(_ context.Context, msg *TestMessage) (*TestMessage, error) {
	return msg, nil
}

type ValidHandlerWithBusName struct{}

func (h *ValidHandlerWithBusName) Handle(_ context.Context, _ *TestMessage) error {
	return nil
}

func (h *ValidHandlerWithBusName) GetBusName() string {
	return "test-bus"
}

type AnotherValidHandler struct{}

func (h *AnotherValidHandler) Handle(_ context.Context, _ *ComplexMessage) error {
	return nil
}

type InvalidHandlerNoMethod struct{}

type InvalidHandlerWrongParams struct{}

func (h *InvalidHandlerWrongParams) Handle(_ *TestMessage) error {
	return nil
}

type InvalidHandlerWrongFirstParam struct{}

func (h *InvalidHandlerWrongFirstParam) Handle(_ string, _ *TestMessage) error {
	return nil
}

type InvalidHandlerTooManyParams struct{}

func (h *InvalidHandlerTooManyParams) Handle(_ context.Context, _ *TestMessage, _ string) error {
	return nil
}

type InvalidHandlerNoReturn struct{}

func (h *InvalidHandlerNoReturn) Handle(_ context.Context, _ *TestMessage) {
}

type InvalidHandlerWrongReturn struct{}

func (h *InvalidHandlerWrongReturn) Handle(_ context.Context, _ *TestMessage) string {
	return ""
}

type InvalidHandlerTooManyReturns struct{}

func (h *InvalidHandlerTooManyReturns) Handle(_ context.Context, _ *TestMessage) (string, int, error) {
	return "", 0, nil
}

type TestMessageHandler struct {
	CallCount int
}

func (h *TestMessageHandler) Handle(_ context.Context, _ *TestMessage) error {
	h.CallCount++

	return nil
}

type AnotherTestMessageHandler struct {
	CallCount int
}

func (h *AnotherTestMessageHandler) Handle(_ context.Context, _ *TestMessage) error {
	h.CallCount++

	return nil
}

type ErrorTestMessageHandler struct {
	CallCount int
	Error     error
}

func (h *ErrorTestMessageHandler) Handle(_ context.Context, _ *TestMessage) error {
	h.CallCount++

	return h.Error
}

type ResultTestMessageHandler struct {
	CallCount int
	Result    any
}

func (h *ResultTestMessageHandler) Handle(_ context.Context, _ *TestMessage) (any, error) {
	h.CallCount++

	return h.Result, nil
}
