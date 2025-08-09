# AMQP Transport Benchmark Report

* Transport: `amqp://` (Messenger `v0.8.0`)
* Publishing mode: sync

## Overall Performance

| Benchmark                 | Time (ns/op) | Throughput (msg/sec) | Memory (B/op) | Allocs/op |
| ------------------------- | -----------: | -------------------: | ------------: | --------: |
| `BenchmarkSend`           |      378,376 |              \~2,643 |         6,792 |       130 |
| `BenchmarkConcurrentSend` |      245,631 |              \~4,071 |         6,790 |       129 |

*Parallel sending gives ~1.5x increase in bandwidth with similar memory consumption.*

---

## Message Size Impact

| Message Size | Time (ns/op) | Throughput (msg/sec) | Memory (B/op) | Allocs/op |
| ------------ | -----------: | -------------------: | ------------: | --------: |
| 100 B        |      384,571 |              \~2,600 |         6,797 |       130 |
| 1 KB         |      397,139 |              \~2,518 |         9,058 |       130 |
| 10 KB        |      445,552 |              \~2,244 |        31,667 |       130 |
| 100 KB       |    1,397,238 |                \~716 |       320,669 |       136 |

*Increasing the payload size is expected to reduce RPS and increase memory consumption.*
