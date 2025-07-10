package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/gerfey/messenger/config"
	"github.com/gerfey/messenger/examples/handler"
	"github.com/gerfey/messenger/examples/messages"
	"github.com/gerfey/messenger/examples/middleware"
	"github.com/gerfey/messenger/internal/messenger/bootstrap"
)

func main() {
	ctx := context.Background()

	log := slog.Default()

	cfg, err := config.LoadConfig("examples/messenger.yaml")
	if err != nil {
		log.Error("load config: %v", err)
	}

	builder := bootstrap.NewBuilder(cfg)

	_ = builder.RegisterHandler(&handler.UserCreateHandler{})

	builder.RegisterMiddleware("logger", middleware.NewLoggerMiddleware(log))

	messenger, err := builder.Build()
	if err != nil {
		log.Error("builder messenger: %v", err)
	}

	go func() {
		if err := messenger.Run(ctx); err != nil {
			log.Error("consumer error: %v", err)
		}
	}()

	messengerBus, err := messenger.GetBus()
	if err != nil {
		log.Error("messenger bus: %v", err)
	}

	_, _ = messengerBus.Dispatch(ctx, &messages.UserCreatedMessage{
		ID:   1,
		Name: "Alice",
	})

	time.Sleep(3 * time.Second)

	//<-ctx.Done()
}
