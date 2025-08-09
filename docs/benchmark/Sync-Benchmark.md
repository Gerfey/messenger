# Sync Transport Benchmark Report

* Transport: `sync://` (Messenger `v0.8.0`)
* Publishing mode: sync

## Overall Performance

| Benchmark                 | Time (ns/op) | Throughput (msg/sec) | Memory (B/op) | Allocs/op |
| ------------------------- | -----------: | -------------------: | ------------: | --------: |
| `BenchmarkSend`           |        1,110 |            \~900,901 |         1,368 |        38 |
| `BenchmarkConcurrentSend` |        932.9 |          \~1,071,926 |         1,377 |        38 |

---

## Message Size Impact

| Message Size | Time (ns/op) | Throughput (msg/sec) | Memory (B/op) | Allocs/op |
| ------------ | -----------: | -------------------: | ------------: | --------: |
| 100 B        |        1,152 |            \~868,056 |         1,368 |        38 |
| 1 KB         |        1,250 |            \~800,000 |         2,281 |        38 |
| 10 KB        |        2,091 |            \~478,240 |        11,505 |        38 |
| 100 KB       |       11,583 |             \~86,333 |       107,832 |        38 |

*ultra‑low latency and allocations — ideal for unit tests, CQRS commands/queries and inline middleware.*
