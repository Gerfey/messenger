package handler

import (
	"context"
	"errors"

	"github.com/gerfey/messenger/examples/retry_messenger/message"
)

type ExampleHelloErrorHandler struct{}

func (u *ExampleHelloErrorHandler) Handle(_ context.Context, _ *message.ExampleHelloMessage) error {
	return errors.New("simulated failure at attempt")
}

func (u *ExampleHelloErrorHandler) GetBusName() string {
	return "default"
}
