package main

import (
	"fmt"
	"reflect"

	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/stamps"
)

type exampleMessage struct {
	Text string
}

func main() {
	msg := &exampleMessage{Text: "Test message"}
	env := envelope.NewEnvelope(msg)

	env = env.WithStamp(stamps.HandledStamp{
		Handler:    "Handler1",
		Result:     "Result1",
		ResultType: reflect.TypeOf(""),
	})

	env = env.WithStamp(stamps.HandledStamp{
		Handler:    "Handler2",
		Result:     "Result2",
		ResultType: reflect.TypeOf(""),
	})

	handledStamps := envelope.StampsOf[stamps.HandledStamp](env)
	fmt.Println("All HandledStamp:")
	for i, stamp := range handledStamps {
		fmt.Printf("  %d: Handler=%s, Result=%v\n", i, stamp.Handler, stamp.Result)
	}

	firstHandled := envelope.FirstStampOf[stamps.HandledStamp](env)
	lastHandled := envelope.LastStampOf[stamps.HandledStamp](env)
	fmt.Printf("First HandledStamp: Handler=%s, Result=%v\n", firstHandled.Handler, firstHandled.Result)
	fmt.Printf("Last HandledStamp: Handler=%s, Result=%v\n", lastHandled.Handler, lastHandled.Result)

	hasHandledStamp := envelope.HasStampOf[stamps.HandledStamp](env)
	hasSentStamp := envelope.HasStampOf[stamps.SentStamp](env)
	fmt.Printf("HandledStamp: %v\n", hasHandledStamp)
	fmt.Printf("SentStamp: %v\n", hasSentStamp)
}
