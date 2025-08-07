package redis_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/builder"
	"github.com/gerfey/messenger/config"
)

const (
	benchmarkDSN     = "redis://localhost:6379/0"
	benchmarkTimeout = 60 * time.Second
)

type BenchmarkMessage struct {
	ID      string
	Content string
	Data    []byte
}

func (m *BenchmarkMessage) RoutingKey() string {
	return "benchmark_routing_key"
}

type BenchmarkHandler struct {
	processedCount *int64
	wg             *sync.WaitGroup
}

func (h *BenchmarkHandler) Handle(_ context.Context, _ *BenchmarkMessage) error {
	atomic.AddInt64(h.processedCount, 1)
	if h.wg != nil {
		h.wg.Done()
	}

	return nil
}

func (h *BenchmarkHandler) GetBusName() string {
	return "default"
}

func setupMessenger(b *testing.B, withWaitGroup bool) api.MessageBus {
	b.Helper()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

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

	processedCount := int64(0)
	var wg *sync.WaitGroup
	if withWaitGroup {
		wg = &sync.WaitGroup{}
	}

	handler := &BenchmarkHandler{
		processedCount: &processedCount,
		wg:             wg,
	}

	builderInstance := builder.NewBuilder(cfg, logger)
	if err := builderInstance.RegisterHandler(handler); err != nil {
		b.Fatalf("Register handler failed: %v", err)
	}

	messenger, err := builderInstance.Build()
	if err != nil {
		b.Fatalf("Build messenger failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(b.Context(), benchmarkTimeout)
	go func() {
		defer cancel()
		if runErr := messenger.Run(ctx); runErr != nil && !errors.Is(runErr, context.Canceled) {
			b.Logf("Messenger run error: %v", runErr)
		}
	}()

	time.Sleep(2 * time.Second)

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
	bus := setupMessenger(b, false)
	dispatchMessages(b, bus, 100, false)
}

func BenchmarkConcurrentSend(b *testing.B) {
	bus := setupMessenger(b, false)
	dispatchMessages(b, bus, 100, true)
}

func BenchmarkMessageSizes(b *testing.B) {
	sizes := []int{100, 1024, 10240, 102400}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size_%dB", size), func(b *testing.B) {
			bus := setupMessenger(b, false)
			dispatchMessages(b, bus, size, false)
		})
	}
}
