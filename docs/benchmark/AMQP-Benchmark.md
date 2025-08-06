# AMQP Benchmark Report

Current performance results of the AMQP transport in Messenger (`v0.8.0`), tested with `RabbitMQ` using the `amqp091-go` client.

## Overall Performance

| Benchmark                     | Time (ns/op) | Throughput (msg/sec) | Memory (B/op) | Allocs/op |
|------------------------------|---------------|---------------------|----------------|-----------|
| `BenchmarkAMQPSend`          | 465,508       | ~2,148              | 13,822         | 267       |
| `BenchmarkAMQPConcurrentSend`| 296,922       | ~3,368              | 13,829         | 266       |

*Concurrent sending is ~36% faster with 100B messages.*

---

## Message Size Impact

| Message Size | Time (ns/op) | Throughput (msg/sec) | Memory (B/op) | Allocs/op |
|------------------|---------------|---------------------|----------------|-----------|
| 100 B            | 472,831       | ~2,115              | 13,826         | 267       |
| 1 KB             | 557,170       | ~1,794              | 19,874         | 267       |
| 10 KB            | 704,831       | ~1,418              | 82,185         | 269       |
| 100 KB           | 1,788,004     | ~559                | 726,393        | 276       |

*As the payload size increases, throughput decreases and memory pressure on GC grows, as expected.*

---

## Allocations: `pprof` Analysis
> Collected using `go test -bench=BenchmarkAMQP -benchmem -memprofile mem.out`, analyzed via `pprof`.

(pprof) top

- encoding/json: ~20%
- amqp091-go (sendOpen, Ack, readLongstr): ~25%
- Envelope.WithStamp: 5.75%
- Middleware chain: ~5%

---

## Summary

- AMQP transport in Messenger demonstrates stable and predictable performance
- Memory and allocation optimization opportunities are being addressed in upcoming versions

