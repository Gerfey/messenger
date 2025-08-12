package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gerfey/messenger/core/builder"
	"github.com/gerfey/messenger/core/config"
	"github.com/gerfey/messenger/examples/messenger/handler"
	"github.com/gerfey/messenger/examples/messenger/message"
	"github.com/gerfey/messenger/examples/messenger/middleware"
)

const (
	waitDurationSeconds = 5
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := slog.Default()

	cfg, err := config.LoadConfig("./examples/messenger/messenger.yaml")
	if err != nil {
		log.Error("ERROR load config", "error", err)

		return
	}

	b := builder.NewBuilder(cfg, log)

	_ = b.RegisterHandler(&handler.ExampleHelloHandler{})

	b.RegisterMiddleware("logger", middleware.NewExampleLoggerMiddleware(log))

	messenger, err := b.Build()
	if err != nil {
		log.Error("failed to build messenger", "error", err)

		return
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

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

	select {
	case sig := <-sigChan:
		log.Info("Received signal, shutting down gracefully", "signal", sig)
		cancel()
	case <-time.After(waitDurationSeconds * time.Second):
		log.Info("Timeout reached, shutting down")
		cancel()
	}

	time.Sleep(waitDurationSeconds * time.Second)
}
