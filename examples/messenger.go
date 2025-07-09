package main

import (
	"context"
	"log"
	"time"

	"github.com/gerfey/messenger/config"
	"github.com/gerfey/messenger/examples/handler"
	"github.com/gerfey/messenger/examples/messages"
	"github.com/gerfey/messenger/internal/messenger/bootstrap"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig("examples/messenger.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	builder := bootstrap.NewBuilder(cfg)

	_ = builder.RegisterHandler(&handler.UserCreateHandler{})

	messenger, err := builder.Build()
	if err != nil {
		log.Fatalf("builder messenger: %v", err)
	}

	go func() {
		if err := messenger.Run(ctx); err != nil {
			log.Fatalf("consumer error: %v", err)
		}
	}()

	messengerBus, err := messenger.GetMessageBus()
	if err != nil {
		log.Fatalf("messenger bus: %v", err)
	}

	_, _ = messengerBus.Dispatch(ctx, &messages.UserCreatedMessage{
		ID:   1,
		Name: "Alice",
	})

	time.Sleep(3 * time.Second)

	//<-ctx.Done()
}
