# Messenger Documentation

[![CI](https://github.com/Gerfey/messenger/actions/workflows/ci.yml/badge.svg)](https://github.com/Gerfey/messenger/actions/workflows/ci.yml)
[![Security](https://github.com/Gerfey/messenger/actions/workflows/security.yml/badge.svg)](https://github.com/Gerfey/messenger/actions/workflows/security.yml)
[![codecov](https://codecov.io/gh/Gerfey/messenger/branch/main/graph/badge.svg)](https://codecov.io/gh/Gerfey/messenger)
[![Go Report Card](https://goreportcard.com/badge/github.com/Gerfey/messenger)](https://goreportcard.com/report/github.com/Gerfey/messenger)
[![Go Reference](https://pkg.go.dev/badge/github.com/Gerfey/messenger.svg)](https://pkg.go.dev/github.com/Gerfey/messenger)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> Pre-release MVP version. Requires Go 1.24+

## Installation

```bash
go get github.com/gerfey/messenger@v0.6.0
```

## Overview
- **Multiple Transports**: AMQP (RabbitMQ), In-Memory
- **Middleware Chain**: Extensible middleware system for message processing
- **Event-Driven**: Built-in event dispatcher for lifecycle hooks
- **Retry Mechanism**: Configurable retry strategies with exponential backoff
- **Message Routing**: Flexible routing system for message distribution
- **Stamps System**: Metadata attachment for message tracking
- **YAML Configuration**: Easy configuration management

## Quick Start

### 1. Define Your Message

```go
type HelloMessage struct {
    Text string
}
```

### 2. Create a Handler

```go
type HelloHandler struct{}

func (h *HelloHandler) Handle(ctx context.Context, msg *HelloMessage) error {
    fmt.Printf("Received: %s\n", msg.Text)
    return nil
}
```

### 3. Configure and Run

```yaml
# messenger.yaml
default_bus: default

buses:
  default: ~

transports:
  amqp:
    dsn: "amqp://guest:guest@localhost:5672/"
    options:
      auto_setup: true
      exchange:
        name: messages
        type: topic

routing:
  package.HelloMessage: amqp
```

```go
package main

import (
    "context"
    "log/slog"

    "github.com/Gerfey/messenger/builder"
    "github.com/Gerfey/messenger/config"
)

func main() {
    ctx := context.Background()
    logger := slog.Default()

    // Load configuration
    cfg, err := config.LoadConfig("messenger.yaml")
    if err != nil {
        logger.Error("failed to load config", "error", err)
        return
    }

    // Build messenger
    b := builder.NewBuilder(cfg, logger)
    b.RegisterHandler(&HelloHandler{})

    messenger, err := b.Build()
    if err != nil {
        logger.Error("failed to build messenger", "error", err)
        return
    }

    // Start consuming
    go func() {
        if err := messenger.Run(ctx); err != nil {
            logger.Error("messenger failed", "error", err)
        }
    }()

    // Send message
    bus, _ := messenger.GetDefaultBus()
    bus.Dispatch(ctx, &HelloMessage{Text: "Hello, World!"})
}
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by [Symfony Messenger](https://symfony.com/doc/current/messenger.html)
- Built with ❤️ for the Go community