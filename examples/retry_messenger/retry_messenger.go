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

func main() {
	ctx := context.Background()

	log := slog.Default()

	cfg, err := config.LoadConfig("./examples/retry_messenger/messenger.yaml")
	if err != nil {
		log.Error("ERROR load config", "error", err)
		return
	}

	b := builder.NewBuilder(cfg)

	_ = b.RegisterHandler(&handler.ExampleHelloErrorHandler{})

	b.RegisterListener(event.SendFailedMessageEvent{}, &listener.TestListener{})

	messenger, err := b.Build()
	if err != nil {
		log.Error("builder messenger", "error", err)
	}

	go func() {
		if err := messenger.Run(ctx); err != nil {
			log.Error("consumer error", "error", err)
		}
	}()

	messengerBus, err := messenger.GetDefaultBus()
	if err != nil {
		log.Error("messenger bus", "error", err)
	}

	_, _ = messengerBus.Dispatch(ctx, &message.ExampleHelloMessage{
		Text: "Hello World",
	})

	time.Sleep(20 * time.Second)

	//<-ctx.Done()
}
