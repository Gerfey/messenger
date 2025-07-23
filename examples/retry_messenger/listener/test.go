package listener

import (
	"context"
	"fmt"

	"github.com/gerfey/messenger/core/event"
)

type TestListener struct{}

func (l *TestListener) Handle(_ context.Context, evt event.SendFailedMessageEvent) {
	fmt.Printf("Failed message: %v\n", evt.Error)
}
