package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/gerfey/messenger/builder"
	"github.com/gerfey/messenger/config"
	"github.com/gerfey/messenger/core/event"
	"github.com/gerfey/messenger/examples/retry_messenger/handler"
	"github.com/gerfey/messenger/examples/retry_messenger/listener"
	"github.com/gerfey/messenger/examples/retry_messenger/message"
)

const (
	waitDurationSeconds = 20
)

func main() {
	ctx := context.Background()

	log := slog.Default()

	cfg, err := config.LoadConfig("./examples/retry_messenger/messenger.yaml")
	if err != nil {
		log.Error("ERROR load config", "error", err)

		return
	}

	b := builder.NewBuilder(cfg, log)

	_ = b.RegisterHandler(&handler.ExampleHelloErrorHandler{})

	b.RegisterListener(event.SendFailedMessageEvent{}, &listener.TestListener{})

	messenger, err := b.Build()
	if err != nil {
		log.Error("failed to build messenger", "error", err)

		return
	}

	go func() {
		if runErr := messenger.Run(ctx); runErr != nil {
			log.Error("messenger run failed", "error", runErr)

			return
		}
	}()

	messengerBus, err := messenger.GetDefaultBus()
	if err != nil {
		log.Error("failed to get default bus", "error", err)

		return
	}

	_, err = messengerBus.Dispatch(ctx, &message.ExampleHelloMessage{
		Text: "Hello World",
	})
	if err != nil {
		log.Error("failed to dispatch message", "error", err)

		return
	}

	time.Sleep(waitDurationSeconds * time.Second)
}
