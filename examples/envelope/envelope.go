package main

import (
	"fmt"

	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/stamps"
)

type exampleHelloMessage struct {
	Text string
}

func main() {
	msg := exampleHelloMessage{Text: "Hello World"}

	env := envelope.NewEnvelope(msg).
		WithStamp(stamps.BusNameStamp{Name: "default"}).
		WithStamp(stamps.DelayStamp{Milliseconds: 5000})

	fmt.Println("Message:", env.Message().(exampleHelloMessage).Text)

	busName := envelope.LastStampOf[stamps.BusNameStamp](env)
	fmt.Println("BusName:", busName.Name)

	delay := envelope.LastStampOf[stamps.DelayStamp](env)
	fmt.Println("Delay:", delay.Milliseconds)
}
