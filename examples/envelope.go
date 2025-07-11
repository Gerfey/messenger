package main

import (
	"fmt"
	"reflect"

	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/stamps"
	"github.com/gerfey/messenger/examples/messages"
)

func main() {
	msg := messages.ExampleHelloMessage{Text: "Hello World"}

	env := envelope.NewEnvelope(msg).
		WithStamp(stamps.BusNameStamp{Name: "default"}).
		WithStamp(stamps.DelayStamp{Milliseconds: 5000})

	fmt.Println("Message:", env.Message().(messages.ExampleHelloMessage).Text)

	busName := env.LastStampOfType(reflect.TypeOf(stamps.BusNameStamp{}))
	if busName != nil {
		fmt.Println("BusName:", busName.(stamps.BusNameStamp).Name)
	}

	delay := env.LastStampOfType(reflect.TypeOf(stamps.DelayStamp{}))
	if delay != nil {
		fmt.Println("Delay:", delay.(stamps.DelayStamp).Milliseconds)
	}
}
