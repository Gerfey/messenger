# Redis Transport Benchmark Report

* Transport: `redis://` (Messenger `v0.8.0`)
* Publishing mode: sync

## Overall Performance

| Benchmark                 | Time (ns/op) | Throughput (msg/s) | Memory (B/op) | Allocs/op |
| ------------------------- | ------------ | ------------------ | ------------- | --------- |
| `BenchmarkSend`           | 133,748      | \~7,477            | 7,361         | 106       |
| `BenchmarkConcurrentSend` | 32,616       | \~30,675           | 7,397         | 106       |

---

## Message Size Impact

| Message Size | Time (ns/op) | Throughput (msg/s) | Memory (B/op) | Allocs/op |
| ------------ | ------------ | ------------------ | ------------- | --------- |
| 100 B        | 127,118      | \~7,866            | 7,362         | 106       |
| 1 KB         | 139,020      | \~7,193            | 9,619         | 106       |
| 10 KB        | 219,779      | \~4,551            | 32,072        | 106       |
| 100 KB       | 1,009,267    | \~991              | 274,182       | 107       |

*Redis Streams shows good rps in parallel, but as the payload increases, the speed decreases and memory increases.*
