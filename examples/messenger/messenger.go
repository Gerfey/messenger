package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/gerfey/messenger/builder"
	"github.com/gerfey/messenger/config"
	"github.com/gerfey/messenger/examples/messenger/handler"
	"github.com/gerfey/messenger/examples/messenger/message"
	"github.com/gerfey/messenger/examples/messenger/middleware"
)

func main() {
	ctx := context.Background()

	log := slog.Default()

	cfg, err := config.LoadConfig("./examples/messenger/messenger.yaml")
	if err != nil {
		log.Error("ERROR load config", "error", err)
		return
	}

	b := builder.NewBuilder(cfg)

	_ = b.RegisterHandler(&handler.ExampleHelloHandler{})

	b.RegisterMiddleware("logger", middleware.NewExampleLoggerMiddleware(log))

	messenger, err := b.Build()
	if err != nil {
		log.Error("failed to build messenger", "error", err)
		return
	}

	go func() {
		if err := messenger.Run(ctx); err != nil {
			log.Error("messenger run failed", "error", err)
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

	time.Sleep(20 * time.Second)

	<-ctx.Done()
}
