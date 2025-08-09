package redis_test

import (
	"fmt"
	"log/slog"
	"sync"
	"testing"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/builder"
	"github.com/gerfey/messenger/core/config"
)

const (
	benchmarkDSN = "redis://localhost:6379/0"
)

type BenchmarkMessage struct {
	ID      string
	Content string
	Data    []byte
}

func setupMessenger(b *testing.B) api.MessageBus {
	b.Helper()

	logger := slog.New(slog.DiscardHandler)

	cfg := &config.MessengerConfig{
		DefaultBus: "default",
		Buses: map[string]config.BusConfig{
			"default": {},
		},
		Transports: map[string]config.TransportConfig{
			"redis": {
				DSN:        benchmarkDSN,
				Serializer: "default.transport.serializer",
				Options: map[string]any{
					"auto_setup": true,
					"stream":     "benchmark_stream",
					"group":      "benchmark_group",
					"consumer":   "benchmark_consumer",
				},
			},
		},
		Routing: map[string]string{
			"*redis_test.BenchmarkMessage": "redis",
		},
	}

	builderInstance := builder.NewBuilder(cfg, logger)

	builderInstance.RegisterMessage(&BenchmarkMessage{})

	messenger, err := builderInstance.Build()
	if err != nil {
		b.Fatalf("Build messenger failed: %v", err)
	}

	bus, err := messenger.GetDefaultBus()
	if err != nil {
		b.Fatalf("Get default bus failed: %v", err)
	}

	return bus
}

func dispatchMessages(b *testing.B, bus api.MessageBus, size int, parallel bool) {
	ctx := b.Context()
	b.ResetTimer()
	b.ReportAllocs()

	if parallel {
		concurrency := 10
		var wg sync.WaitGroup
		messagesPerWorker := b.N / concurrency
		for w := range concurrency {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for i := range messagesPerWorker {
					bus.Dispatch(ctx, &BenchmarkMessage{
						ID:      fmt.Sprintf("worker-%d-msg-%d", id, i),
						Content: "benchmark content",
						Data:    make([]byte, size),
					})
				}
			}(w)
		}
		wg.Wait()
	} else {
		for i := range b.N {
			bus.Dispatch(ctx, &BenchmarkMessage{
				ID:      fmt.Sprintf("msg-%d", i),
				Content: "benchmark content",
				Data:    make([]byte, size),
			})
		}
	}
}

func BenchmarkSend(b *testing.B) {
	bus := setupMessenger(b)
	dispatchMessages(b, bus, 100, false)
}

func BenchmarkConcurrentSend(b *testing.B) {
	bus := setupMessenger(b)
	dispatchMessages(b, bus, 100, true)
}

func BenchmarkMessageSizes(b *testing.B) {
	sizes := []int{100, 1024, 10240, 102400}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size_%dB", size), func(b *testing.B) {
			bus := setupMessenger(b)
			dispatchMessages(b, bus, size, false)
		})
	}
}
