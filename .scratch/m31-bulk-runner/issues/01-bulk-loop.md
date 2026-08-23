# 01 — Bulk loop (internal/bulk)

**What to build:** Pure bulk loop `internal/bulk.Run` reading CSV/JSON rows into `[]map[string]string`, interpolating per row via `ScopeRuntime`, sequential default, parallel `--parallel` + `--concurrency` ordered, stop on non-2xx unless `--continue-on-error`.

**Blocked by:** None — can start immediately

**Status:** ready-for-agent

- [ ] `internal/bulk` pure stdlib (`encoding/csv`, `encoding/json`, `strconv`) — `Run(ctx, req, rows, opts, sendFn, onStep)` + `Options{Parallel, Concurrency, ContinueOnError}` + `Step{Index, Row, Request, Response, Err}` + `SendFunc`
- [ ] Row parsing: CSV header→map, JSON array of objects → stringified map, empty file → 0 steps
- [ ] Per-row `variables.ScopeRuntime` interpolation (clone base `Set`, set row `k=v`, pass to `sendFn` via `client.Execute`)
- [ ] Sequential loop ordered, parallel via semaphore `concurrency` (default 5 parallel, 1 sequential) ordered output, errgroup, ctx cancellation
- [ ] Stop on first non-2xx unless `ContinueOnError` (log failed step and continue)
- [ ] Table-driven unit tests `internal/bulk/bulk_test.go` with stub `sendFn` (row interpolation, sequential vs parallel ordered, stop vs continue, CSV/JSON parsing, empty rows, concurrency)
- [ ] `go vet` + `gofmt -l` + `go test -race ./internal/bulk` green
