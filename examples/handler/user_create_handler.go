package handler

import (
	"context"
	"fmt"

	"github.com/gerfey/messenger/examples/messages"
)

type UserCreateHandler struct {
	_ struct{} `messenger:"bus=message.bus"`
}

func (u *UserCreateHandler) Handle(ctx context.Context, msg *messages.UserCreatedMessage) error {
	fmt.Printf("Handled user: ID=%d, Name=%s\n", msg.ID, msg.Name)

	return nil
}
