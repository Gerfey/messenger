# Messenger

[![CI](https://github.com/Gerfey/messenger/actions/workflows/ci.yml/badge.svg)](https://github.com/Gerfey/messenger/actions/workflows/ci.yml)
[![Security](https://github.com/Gerfey/messenger/actions/workflows/security.yml/badge.svg)](https://github.com/Gerfey/messenger/actions/workflows/security.yml)
[![codecov](https://codecov.io/gh/Gerfey/messenger/branch/main/graph/badge.svg)](https://codecov.io/gh/Gerfey/messenger)
[![Go Report Card](https://goreportcard.com/badge/github.com/Gerfey/messenger)](https://goreportcard.com/report/github.com/Gerfey/messenger)
[![Go Reference](https://pkg.go.dev/badge/github.com/Gerfey/messenger.svg)](https://pkg.go.dev/github.com/Gerfey/messenger)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> Версия `v0.8.0` — это пре-релиз. Тестируйте и сообщайте о багах!

> Полная документация доступна на [GitHub Wiki](https://github.com/Gerfey/messenger/wiki/Documentation)

🇬🇧 [English README](README.md)

---

## Возможности
- **Множественные транспорты**: AMQP (RabbitMQ), Kafka, Redis (Stream), In-Memory (sync)
- **Цепочка middleware**: Расширяемая система промежуточной обработки
- **Событийный движок**: Встроенный dispatcher событий жизненного цикла
- **Механизм повторов**: Настраиваемые стратегии ретраев с поддержкой DLQ
- **Маршрутизация сообщений**: Гибкое сопоставление сообщений и транспортов
- **Система метаданных (Stamps)**: Для трассировки и поведения сообщений
- **YAML-конфигурация**: С поддержкой переменных окружения `%env(...)%`

## Установка
> Требуется Go версии **1.24+**
```bash
go get github.com/gerfey/messenger@v0.8.0
```

## Быстрый старт

### Определите сообщение

```go
package main

type HelloMessage struct {
    Text string
}
```

### Создайте обработчик

```go
package main

type HelloHandler struct{}

func (h *HelloHandler) Handle(ctx context.Context, msg *HelloMessage) error {
    fmt.Println("Hello:", msg.Text)
    return nil
}
```

### Создайте конфигурационный файл `messenger.yaml`:

```yaml
default_bus: default

buses:
  default: ~
```
> 💡 Если транспорт для сообщения не указан — оно будет выполнено синхронно (inline).

### Инициализация и запуск:

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

## Больше примеров

* Команды без возврата значения
* Запросы с возвратом результата
* Повторные попытки и Dead Letter Queue
* Пользовательские middleware и транспорты
* Слушатели событий и хуки жизненного цикла

> Смотри [Сценарии использования](https://github.com/Gerfey/messenger/wiki/Сценарии-использования).

## Показатели

- AMQP (RabbitMQ): [AMQP Transport Benchmark Report](docs/benchmark/AMQP-Benchmark.md)
- Redis (Stream): [Redis Transport Benchmark Report](docs/benchmark/Redis-Benchmark.md)
- Sync: [Sync Transport Benchmark Report](docs/benchmark/Sync-Benchmark.md)
- Kafka (Async): [Kafka Transport Async Benchmark Report](docs/benchmark/Kafka-async-Benchmark.md)

## Как внести вклад

1. Форкните репозиторий
2. Создайте новую ветку (`git checkout -b feature/amazing-feature`)
3. Сделайте коммит (`git commit -m 'Add some amazing feature'`)
4. Запушьте изменения (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

## Лицензия

Проект лицензирован под [MIT](LICENSE).

## Поддержка

Если вам полезен этот проект — поставьте ⭐️ и расскажите другим!

## Благодарности

- Вдохновлено [Symfony Messenger](https://symfony.com/doc/current/messenger.html)
- Сделано с ❤️ для сообщества Go
