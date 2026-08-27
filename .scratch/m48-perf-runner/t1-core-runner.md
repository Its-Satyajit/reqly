# T1 — Core perf runner

> **Spec:** `docs/spec/m48-perf-runner.md`
> **Blocks:** T2, T3

- `internal/perf.Run` — `RPS` ticker `time.NewTicker`, concurrency semaphore, `SendFn` (like `pagination.Run`/`bulk.Run`), sorted latencies → P50/P95/P99, `StatusCounts`, `ErrorRate`
- `PerfResult` JSON

**Done when:** `go test -race ./internal/perf` green (P95 calc, status histogram, cancel)
