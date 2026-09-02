# Spec: Perf Runner — CLI + GUI (Milestone 48 — §37)

> **Status:** Draft — grill Q1–Q4 done 2026-08-27 (CLI+GUI TanStack Charts, constant RPS, snapshot, separate store), ADR 0032
> **Scope:** `ROADMAP.md:517` §37 — lightweight RPS/latency P95/P99/error-rate/status-distribution; TanStack Charts `pnpm add @tanstack/charts` `defineChart/lineY/barY/mountChart`
> **Stack:** `internal/perf` (ticker + worker pool), `apps/cli/cmd/perf.go`, `frontend/src/features/perf-view/PerfView.tsx` (DOM host `mountChart`)

## Data Model
`PerfResult { RPS, Duration, Total, Latencies []int64 (ms), P50,P95,P99 int64, StatusCounts map[int]int, ErrorRate float64, StartedAt, DurationMs }` — sorted slice for percentiles (no hdrhistogram).

## API Surface
- **Go core:** `internal/perf.Run(ctx, req, vars, opts{RPS, Duration, Concurrency, SendFn}) (PerfResult, error)` — ticker `time.NewTicker(1/RPS)`, semaphore `concurrency`, `OnTick` streaming optional; context cancel aborts.
- **CLI:** `reqly perf run <request-file> [--rps 10] [--duration 30s] [--concurrency 5] [--json]` — loads `requestfile`, interpolates vars, reuses `internal/request.Client` via `core.RunService`; `--json` machine-readable `PerfResult`.
- **GUI:** Perf view — form RPS/duration/concurrency + Run → `AppService.PerfRun(specPath)` → snapshot charts: `lineY(latencyOverTime)` (x: second via `scalePoint`, y: `scaleLinear` ms, `tooltip`) + `barY(statusCounts)` histogram; `host.update` on new result, `host.destroy` on unmount.

## Edge Cases
RPS 0 → error; duration 0 → single burst; concurrency > RPS → cap; non-2xx counted in `StatusCounts` + `ErrorRate`; context cancel → partial `PerfResult` with `error: canceled`; empty latencies → P* = 0.

## Testing Strategy
`internal/perf` unit (ticker cap, sorted P95, status counts), `apps/cli/cmd/perf_test.go` (flag parsing, `--json`), frontend `PerfView` mount/update/destroy via `jsdom` + `ResizeObserver` mock; `go test -race` + `nub run typecheck`.
