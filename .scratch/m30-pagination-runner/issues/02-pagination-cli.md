# 02 — Pagination CLI

**What to build:** `reqly pagination run <request-file> [--max-pages <n>]` reads `requestfile.File` `pagination:` block, loops via `internal/pagination.Run` with real `core.RequestService.Send`, streams per-step `step/status/url/duration`.

**Blocked by:** 01 — Pagination loop (internal/pagination)

**Status:** ready-for-agent

- [ ] Cobra `pagination` root + `run` subcommand in `apps/cli/cmd/pagination.go` (one file per group), registered in `apps/cli/cmd/root.go`, flag `--max-pages` int overrides file `maxPages`
- [ ] Load `<request-file>` via `requestfile.LoadFile`, resolve `pagination` block, call `pagination.Run` with `sendFn` wired to `request.Client`/`core.RequestService`, print per-step line via `OnStep` to stdout
- [ ] Errors: missing pagination block → explicit error, malformed `nextPath` → stop with warning, non-2xx → fail step
- [ ] Integration tests `apps/cli/cmd/pagination_test.go` with `httptest.NewServer` paging 3× for page/offset/cursor/link-header, `--max-pages` override, `--help` contains `run`
- [ ] `go test ./...` + `go build -o reqly ./apps/cli` + `reqly pagination run --help` smoke
