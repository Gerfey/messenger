package handler

import (
	"context"
	"fmt"

	"github.com/gerfey/messenger/examples/messenger/message"
)

type ExampleHelloHandler struct {
	_ struct{} `messenger:"bus=default"`
}

func (u *ExampleHelloHandler) Handle(ctx context.Context, msg *message.ExampleHelloMessage) error {
	fmt.Printf("Handled: Text=%v\n", msg.Text)

	return nil
}
