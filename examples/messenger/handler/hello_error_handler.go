package handler

import (
	"context"
	"fmt"

	"github.com/gerfey/messenger/examples/messenger/message"
)

type ExampleHelloErrorHandler struct{}

func (u *ExampleHelloErrorHandler) Handle(_ context.Context, msg *message.ExampleHelloMessage) error {

	fmt.Printf("Handled: Text=%v\n", msg.Text)

	return fmt.Errorf("simulated failure at attempt")
}

func (u *ExampleHelloErrorHandler) GetBusName() string {
	return "default"
}
