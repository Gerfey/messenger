# Messenger

[![CI](https://github.com/Gerfey/messenger/actions/workflows/ci.yml/badge.svg)](https://github.com/Gerfey/messenger/actions/workflows/ci.yml)
[![Security](https://github.com/Gerfey/messenger/actions/workflows/security.yml/badge.svg)](https://github.com/Gerfey/messenger/actions/workflows/security.yml)
[![codecov](https://codecov.io/gh/Gerfey/messenger/branch/main/graph/badge.svg)](https://codecov.io/gh/Gerfey/messenger)
[![Go Report Card](https://goreportcard.com/badge/github.com/Gerfey/messenger)](https://goreportcard.com/report/github.com/Gerfey/messenger)
[![Go Reference](https://pkg.go.dev/badge/github.com/Gerfey/messenger.svg)](https://pkg.go.dev/github.com/Gerfey/messenger)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> `v0.7.0` is a pre-release version — feel free to test and report issues!

> Full documentation available in the [GitHub Wiki](https://github.com/Gerfey/messenger/wiki/Documentation)

🇷🇺 [Русская версия](README.ru.md)

## Features
- **Multiple Transports**: AMQP (RabbitMQ), In-Memory (sync)
- **Middleware Chain**: Extensible middleware system for message processing
- **Event-Driven**: Built-in event dispatcher for lifecycle hooks
- **Retry Mechanism**: Configurable retry strategies with exponential backoff
- **Message Routing**: Flexible routing system for message distribution
- **Stamps System**: Metadata attachment for message tracking
- **YAML Configuration**: Easy configuration management with `%env(...)%` support

## Installation
> Requires Go 1.24+
```bash
go get github.com/gerfey/messenger@v0.7.0
```

## Quick Start

### Define Your Message

```go
package main

type HelloMessage struct {
    Text string
}
```

### Create a Handler

```go
package main

type HelloHandler struct{}

func (h *HelloHandler) Handle(ctx context.Context, msg *HelloMessage) error {
    fmt.Println("Hello:", msg.Text)
    return nil
}
```

### Create config file `messenger.yaml`:

```yaml
default_bus: default

buses:
  default: ~
```
> 💡 If no transport is configured for a message, it will be executed synchronously by default (inline handler execution).

### Initialize messenger:

```go
cfg, errConfig := config.LoadConfig("messenger.yaml")
if errConfig != nil {
    fmt.Println("ERROR load config", "error", errConfig)
    return
}

b := builder.NewBuilder(cfg, slog.Default())
b.RegisterHandler(&HelloHandler{})
m, _ := b.Build()

ctx := context.Background()
go m.Run(ctx)

bus, _ := m.GetDefaultBus()
_, _ = bus.Dispatch(ctx, &HelloMessage{Text: "World"})
```

## More Examples

* Commands with void return
* Queries with return value access
* Retry and Dead Letter Queue
* Custom Middleware and Transports
* Event Listeners and Lifecycle Hooks

> See [Usage Scenarios](https://github.com/Gerfey/messenger/wiki/Usage-Scenarios) for commands, queries, return values and advanced use-cases.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## ⚖License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

If you find this project useful, please consider starring ⭐️ it and sharing with others!

## Acknowledgments

- Inspired by [Symfony Messenger](https://symfony.com/doc/current/messenger.html)
- Built with ❤️ for the Go community
