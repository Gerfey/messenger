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

	cfg, err := config.LoadConfig("examples/messenger/messenger.yaml")
	if err != nil {
		log.Error("load config: %v", err)
	}

	b := builder.NewBuilder(cfg)

	_ = b.RegisterHandler(&handler.ExampleHelloHandler{})

	b.RegisterMiddleware("logger", middleware.NewExampleLoggerMiddleware(log))

	messenger, err := b.Build()
	if err != nil {
		log.Error("builder messenger: %v", err)
	}

	go func() {
		if err := messenger.Run(ctx); err != nil {
			log.Error("consumer error: %v", err)
		}
	}()

	messengerBus, err := messenger.GetDefaultBus()
	if err != nil {
		log.Error("messenger bus: %v", err)
	}

	_, _ = messengerBus.Dispatch(ctx, &message.ExampleHelloMessage{
		Text: "Hello World",
	})

	time.Sleep(3 * time.Second)

	//<-ctx.Done()
}
