# T2 — CLI perf run

> **Spec:** `docs/spec/m48-perf-runner.md`
> **Blocks:** T1

- `apps/cli/cmd/perf.go` — `reqly perf run <file> [--rps 10] [--duration 30s] [--concurrency 5] [--json]` (file load via `requestfile`, vars via `variables`, `core.RunService`)
- `--json` machine-readable `PerfResult`

**Done when:** `go test ./apps/cli/cmd -run Perf` + `go build -o /tmp/reqly` shows flags
