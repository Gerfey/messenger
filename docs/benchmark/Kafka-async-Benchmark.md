# Kafka Transport Async Benchmark Report

* Transport: `kafka://` (Messenger `v0.8.0`)
* Publishing mode: async (`async=true`)

## Overall Performance

| Benchmark                 | Time (ns/op) | Throughput (msg/sec) | Memory (B/op) | Allocs/op |
| ------------------------- | -----------: | -------------------: | ------------: | --------: |
| `BenchmarkSend`           |        3,018 |            \~331,345 |         3,072 |        47 |
| `BenchmarkConcurrentSend` |        1,859 |            \~537,924 |         3,067 |        47 |


*Async publishing provides high throughput with a stable allocation profile.*

---

## Message Size Impact

| Message Size | Time (ns/op) | Throughput (msg/sec) | Memory (B/op) | Allocs/op |
| ------------ | -----------: | -------------------: | ------------: | --------: |
| 100 B        |        2,913 |            \~343,289 |         3,062 |        47 |
| 1 KB         |        4,147 |            \~241,138 |         5,304 |        47 |
| 10 KB        |        9,800 |            \~102,041 |        27,723 |        47 |
| 100 KB       |       87,745 |             \~11,397 |       254,397 |        48 |


*Increasing the message size is expected to reduce throughput and increase memory consumption, but the curve looks smooth and predictable.*
