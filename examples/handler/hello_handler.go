package handler

import (
	"context"
	"fmt"

	"github.com/gerfey/messenger/examples/messages"
)

type ExampleHelloHandler struct {
	_ struct{} `messenger:"bus=default"`
}

func (u *ExampleHelloHandler) Handle(ctx context.Context, msg *messages.ExampleHelloMessage) error {
	fmt.Printf("Handled: Text=%v\n", msg.Text)

	return nil
}
