# 02 — Bulk CLI

**What to build:** `reqly bulk run <request-file> --data <csv|json> [--parallel] [--concurrency <n>] [--continue-on-error]` loads request file, parses `--data` rows, loops via `internal/bulk.Run` with real `core.RequestService`, streams per-step `step: status duration url`.

**Blocked by:** 01 — Bulk loop (internal/bulk)

**Status:** ready-for-agent

- [ ] Cobra `bulk` root + `run` subcommand in `apps/cli/cmd/bulk.go` (one file per group), registered in `apps/cli/cmd/root.go`, flags `--data` required string, `--parallel` bool, `--concurrency` int (default 5, ignored when not parallel with warning), `--continue-on-error` bool
- [ ] Load `<request-file>` via `requestfile.LoadFile`, parse `--data` CSV/JSON into rows, validate `--data` exists, call `bulk.Run` with `sendFn` wired to `request.Client`/`core.RequestService` + `variables.ScopeRuntime` per row, `onStep` prints to stdout (masked)
- [ ] Errors: missing `--data`, non-existent file, malformed CSV/JSON → explicit error, non-2xx per spec
- [ ] Integration tests `apps/cli/cmd/bulk_test.go` with `httptest` 3 rows sequential/parallel/continue-on-error, CSV header `id,name` → `{{id}}`, JSON `[{"id":1}]`, `--help` contains `run`, missing `--data` error
- [ ] `go test ./...` + `go build -o reqly ./apps/cli` + `reqly bulk run --help` smoke
