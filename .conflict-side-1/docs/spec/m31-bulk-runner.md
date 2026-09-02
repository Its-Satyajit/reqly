# Spec: Bulk Runner (Milestone 31)

> **Status:** Draft — grill settled 2026-08-23 (Q1–Q4, CSV+JSON MVP)
> **Scope:** P1 `ROADMAP.md:346` — CSV/JSON inputs, sequential default, parallel with `--parallel` + `--concurrency N` ordered, `--continue-on-error` (variable/generated datasets deferred to M31b), `docs/features.md:26`
> **Stack:** `internal/bulk` (new pure loop) + `apps/cli/cmd/bulk.go` + `internal/requestfile` CSV/JSON row parsing + `core.RequestService`/`request.Client` send seam — no new deps, stdlib `encoding/csv`+`encoding/json`
> **Predecessor:** M30 Pagination `docs/adr/0022` shipped, M31 `docs/adr/0023-bulk-runner.md` (Q1–Q4)

## Problem Statement

`ROADMAP.md:346` and `docs/features.md:26` require executing one request against many inputs (CSV, JSON, variables, generated datasets) sequential or parallel with configurable concurrency — none exists today. Users must manually loop `reqly run` per row, hand-edit `{{var}}`, and track failures. `internal/pagination` loops via `sendFn` for next-link, `internal/runner` is collection-level, but no row-driven repetition seam exists.

## Solution

One row-driven pure loop plus one CLI command, no `request.bulk` block:

* **`internal/bulk.Run(ctx, req, rows, opts, sendFn, onStep)`** — `rows []map[string]string` from CSV header→values or JSON array of objects (values stringified), each row interpolated as `{{key}}` via `variables.ScopeRuntime` per send (copy `variables.Set`), `opts{Parallel bool, Concurrency int, ContinueOnError bool}`. Sequential default (1 at a time), parallel via `--parallel` + semaphore `concurrency` (default 5 parallel, 1 sequential); ordered output by input index but concurrent send (errgroup + semaphore); stop on first non-2xx unless `ContinueOnError`. Pure `sendFn func(context.Context, request.Request) (*response.Response, error)` like pagination (prod `core.RequestService.Send`, tests stub).

* **`reqly bulk run <request-file> --data <csv|json> [--parallel] [--concurrency <n>] [--continue-on-error]`** — Cobra `bulk` root + `run` subcommand (`apps/cli/cmd/bulk.go`, registered in `root.go:70`, one-file-per-group). Parses `--data` CSV (header first row) or JSON array, loads `<request-file>` via `requestfile.LoadFile`, loops via `bulk.Run` with real `sendFn`, streams `onStep` `step: status duration url` to stdout (masked via `masker`). No `request.bulk` block for M31 — request files stay plain Git-native; variable/generated rows deferred to M31b.

## User Stories

1. As a user, I want `reqly bulk run ./collections/users/create.yaml --data users.csv` (CSV `id,name` header) to send the request per row interpolating `{{id}}`/`{{name}}`, so batch creation works without manual `reqly run` loop.
2. As a user, I want JSON `--data users.json` (`[{"id":1},{"id":2}]`) to work like CSV, so JSON exports are usable.
3. As a user, I want sequential default (row1, row2, row3 in order) with per-step `step 1: 200 10ms` streamed, so progress is observable like `collection run`.
4. As a user, I want `--parallel --concurrency 10` to run rows concurrent via semaphore (default 5) but report in input order, so large batches are fast yet deterministic.
5. As a user, I want non-2xx to stop on first failure, but `--continue-on-error` to keep going, so bulk can be fail-fast or exhaustive.
6. As a user, I want `{{key}}` from each row to be interpolated via runtime variables per send, so existing request `url: https://api.example.com/users/{{id}}` works without file edits.
7. As a developer, I want `internal/bulk.Run` unit-tested with stub `sendFn` (no network), so bulk logic is deterministic and fast.

## Implementation Decisions

- **Seam: `internal/bulk` pure package, highest seam.** `func Run(ctx context.Context, req request.Request, rows []map[string]string, opts Options, sendFn SendFunc, onStep func(Step)) error` + `Options{Parallel bool, Concurrency int, ContinueOnError bool}` + `Step{Index int, Row map[string]string, Request request.Request, Response *response.Response, Err error}` + `SendFunc func(context.Context, request.Request) (*response.Response, error)`. Stdlib only (`encoding/csv`, `encoding/json`, `golang.org/x/sync/errgroup` or hand-rolled semaphore); no new SQLite/history; `variables.ScopeRuntime` per row.
- **Row parsing:** `--data` file ext `.csv` → `encoding/csv.NewReader` header first row → `map[col]value` per record (trim spaces, string values); `.json` → `json.Unmarshal` into `[]map[string]any` → stringified `map[string]string` (numbers via `strconv`, objects via `json.Marshal`). Empty file → 0 rows → no steps (not error).
- **Interpolation:** for each row, `vars := baseVars.Clone(); vars.Set(variables.ScopeRuntime, k, v)` for `k,v` in row, then `client.Execute(ctx, &reqCopy, vars)` where `reqCopy` is `req` plus `variables.Interpolate` inside `client.Execute` picks `ScopeRuntime`. No mutation of `req` struct itself; pagination mutates `Query` but bulk mutates vars.
- **Concurrency:** `Parallel==false` → sequential loop `for i, row := range rows` send one by one, `onStep` ordered. `Parallel==true` → `sem := make(chan struct{}, concurrency)` + `errgroup` (or `sync.WaitGroup` + mutex for ordered results); send concurrent but `onStep` called in input order via results slice indexed by `i` (or stream as completed but report ordered at end). `Concurrency` default 5 when parallel, 1 sequential; flag `--concurrency` ignored when not parallel (warn).
- **Stop:** after each send, if `resp.StatusCode != 0 && (resp.StatusCode <200 || >=300)` and `!ContinueOnError` → return early (or break); else continue. `ctx` cancellation stops all.
- **CLI shape:** `bulk` root + `run` subcommand `Use:"run <request-file> --data <csv|json> [--parallel] [--concurrency <n>] [--continue-on-error]"` `Args: cobra.ExactArgs(1)` plus required `--data` flag; `root.go` registers `bulkCmd`. Errors: missing `--data` → `required flag`, non-existent file → error, malformed CSV/JSON → per-spec error.
- **Output:** default per-step `fmt.Fprintf(out, "step %d: %d %s (%s) %s\n", step, status, statusText, duration, urlWithQuery)` via `onStep` to stdout (masked); no `history` bulk spillage for M31.

## Testing Decisions

- **What makes a good test:** assert bulk behavior via stub `sendFn` (row interpolation + sequential/parallel ordered + stop vs continue), not CSV parsing internals. Table-driven `testing` like `internal/pagination`.
- **Seams to test (highest first):**
  - `internal/bulk` unit: stub `sendFn` captures `Request` interpolated URL/query per row (`{{id}}`→`1`), sequential order index, parallel concurrency 5 ordered (via `httptest` or stub sleep), `--continue-on-error` vs stop on 500, CSV header→map, JSON array→map, empty rows.
  - CLI integration `apps/cli/cmd/bulk_test.go`: temp CSV/JSON `--data` + `httptest` 3 rows sequential/parallel/continue, `--help` contains `run`, missing `--data` error, malformed CSV error.
  - No golden files — rows short inline `httptest` ATS, deterministic via `httptest` server per row.
- **Prior art:** `internal/pagination` stub `sendFn`, `apps/cli/cmd/pagination_test.go` temp file + `httptest`, `internal/auth/jwt_test.go` table-driven.
- **Quality gates:** `go test ./...` + `go test -race ./...` + `go vet` + `gofmt -l` clean; `go build -o reqly ./apps/cli` + `reqly bulk run --help` smoke.

## Out of Scope

- `--var key=value` rows without file + `{{$randomInt}}`/`{{$uuid}}` generated datasets (M31b, reuses `TagGenerator`).
- `--limit N` rows cap — M31 runs all rows (CSV/JSON already bounded).
- Aggregation `--out` file or `history` bulk persistence — per-step streaming only for M31.
- Desktop `Bulk` tab + `AppService.BulkRun` binding — M31b, reuses same `internal/bulk.Run` seam.
- Variable precedence beyond `ScopeRuntime` (row overrides env but not file vars) — scoped as above.
- Retry/backoff during bulk (M32 `Retry & resilience`) — non-2xx stops, no retry for M31.

## Further Notes

- **ADR:** `docs/adr/0023-bulk-runner.md` (draft Q1–Q3, Q4 pending) — CSV+JSON MVP, CLI `--data` only, sequential/parallel semaphore ordered, `--continue-on-error`.
- **Glossary:** `CONTEXT.md:275` `Bulk Runner`/`Bulk Input Row`/`Bulk Concurrency` (this spec's grilling).
- **ROADMAP:** `ROADMAP.md:346` M31 tick after ship; M30 pagination already shipped (`adr/0022`).
- **Ticket split (see `to-tickets`):** T1 `internal/bulk` rows + loop + tests, T2 `apps/cli/cmd/bulk.go` + `root.go` wiring + CLI httptest, T3 docs (ADR 0023 + CONTEXT + ROADMAP + this spec) + `go test -race` green.
