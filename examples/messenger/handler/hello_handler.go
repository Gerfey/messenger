package handler

import (
	"context"
	"fmt"

	"github.com/gerfey/messenger/examples/messenger/message"
)

type ExampleHelloHandler struct{}

func (u *ExampleHelloHandler) Handle(_ context.Context, msg *message.ExampleHelloMessage) error {
	fmt.Printf("Handled: Text=%v\n", msg.Text)

	return nil
}

func (u *ExampleHelloHandler) GetBusName() string {
	return "default"
}
