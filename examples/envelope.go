package main

import (
	"fmt"
	"reflect"

	"github.com/gerfey/messenger/envelope"
	"github.com/gerfey/messenger/examples/messages"
	"github.com/gerfey/messenger/stamps"
)

func main() {
	msg := messages.HelloMessage{Text: "Hello World"}

	env := envelope.NewEnvelope(msg).
		WithStamp(stamps.BusNameStamp{Name: "default"}).
		WithStamp(stamps.DelayStamp{Milliseconds: 5000})

	fmt.Println("Message:", env.Message().(messages.HelloMessage).Text)

	busName := env.LastStampOfType(reflect.TypeOf(stamps.BusNameStamp{}))
	if busName != nil {
		fmt.Println("BusName:", busName.(stamps.BusNameStamp).Name)
	}

	delay := env.LastStampOfType(reflect.TypeOf(stamps.DelayStamp{}))
	if delay != nil {
		fmt.Println("Delay:", delay.(stamps.DelayStamp).Milliseconds)
	}
}
