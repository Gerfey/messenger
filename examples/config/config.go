package main

import (
	"fmt"
	"log"

	"github.com/gerfey/messenger/config"
)

func main() {
	cfg, err := config.LoadConfig("examples/config/messenger.yaml")
	if err != nil {
		log.Fatal(err)
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

		fmt.Printf("    Options:\n")
		fmt.Printf("      AutoSetup: %v\n", transport.Options.AutoSetup)
		fmt.Printf("      Exchange:\n")
		fmt.Printf("        Name: %s\n", transport.Options.Exchange.Name)
		fmt.Printf("        Type: %s\n", transport.Options.Exchange.Type)
		fmt.Printf("        Durable: %v\n", transport.Options.Exchange.Durable)
		fmt.Printf("        AutoDelete: %v\n", transport.Options.Exchange.AutoDelete)
		fmt.Printf("        Internal: %v\n", transport.Options.Exchange.Internal)
	}
}
