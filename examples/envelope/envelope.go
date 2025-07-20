package main

import (
	"fmt"

	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/stamps"
)

const (
	defaultDelayMs = 5000
)

type exampleHelloMessage struct {
	Text string
}

func main() {
	msg := exampleHelloMessage{Text: "Hello World"}

	env := envelope.NewEnvelope(msg).
		WithStamp(stamps.BusNameStamp{Name: "default"}).
		WithStamp(stamps.DelayStamp{Milliseconds: defaultDelayMs})

	if message, ok := env.Message().(exampleHelloMessage); ok {
		fmt.Println("Message:", message.Text)
	}

	busName, _ := envelope.LastStampOf[stamps.BusNameStamp](env)
	fmt.Println("BusName:", busName.Name)

	delay, _ := envelope.LastStampOf[stamps.DelayStamp](env)
	fmt.Println("Delay:", delay.Milliseconds)
}
