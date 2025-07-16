package handler

import (
	"context"
	"fmt"

	"github.com/gerfey/messenger/examples/messenger/message"
)

type ExampleHelloHandler struct{}

func (u *ExampleHelloHandler) Handle(_ context.Context, msg *message.ExampleHelloMessage) error {

	fmt.Printf("Handled: Text=%v\n", msg.Text)

	return fmt.Errorf("simulated failure at attempt")
}

func (u *ExampleHelloHandler) GetBusName() string {
	return "default"
}
