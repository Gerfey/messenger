package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/gerfey/messenger/core/config"

	"github.com/gerfey/messenger/transport/amqp"
)

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg, err := config.LoadConfig("examples/config/messenger.yaml")
	if err != nil {
		logger.ErrorContext(ctx, "Failed to load config", "error", err)
		os.Exit(1)
	}

	fmt.Println("Default bus:", cfg.DefaultBus)

	fmt.Println("\nBuses:")
	for name, bus := range cfg.Buses {
		fmt.Printf("  %s:\n", name)
		fmt.Printf("    Middleware: %v\n", bus.Middleware)
	}

	fmt.Println("\nTransports:")
	for name, transport := range cfg.Transports {
		fmt.Printf("  %s:\n", name)
		fmt.Printf("    DSN: %s\n", transport.DSN)

		if transport.RetryStrategy != nil {
			fmt.Printf("    Retry_strategy:\n")
			fmt.Printf("      MaxRetries: %v\n", transport.RetryStrategy.MaxRetries)
			fmt.Printf("      Delay: %v\n", transport.RetryStrategy.Delay)
			fmt.Printf("      Multiplier: %v\n", transport.RetryStrategy.Multiplier)
			fmt.Printf("      MaxDelay: %v\n", transport.RetryStrategy.MaxDelay)
		}

		rawYAML, _ := yaml.Marshal(transport.Options)

		var opts amqp.OptionsConfig
		_ = yaml.Unmarshal(rawYAML, &opts)

		fmt.Printf("    Options:\n")
		fmt.Printf("      AutoSetup: %v\n", opts.AutoSetup)
		fmt.Printf("      Exchange:\n")
		fmt.Printf("        Name: %s\n", opts.Exchange.Name)
		fmt.Printf("        Type: %s\n", opts.Exchange.Type)
		fmt.Printf("        Durable: %v\n", opts.Exchange.Durable)
		fmt.Printf("        AutoDelete: %v\n", opts.Exchange.AutoDelete)
		fmt.Printf("        Internal: %v\n", opts.Exchange.Internal)
	}
}
