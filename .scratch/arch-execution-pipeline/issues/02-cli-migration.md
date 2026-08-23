# 02 — Migrate the five CLI commands; delete the preambles

**What to build:** `run`, `test`, `pagination run`, `bulk run`, and `collection run/test` stop rebuilding env/client/masking wiring and call `core.RequestService.Run` (or pass a Run-backed sendFn). The duplicated preambles, `openTokenStore`, and `activeEnvironment` shrink to flag-parsing plus rendering; CLI output stays byte-identical for existing tests.

**Blocked by:** 01 — Run pipeline + workspace-wiring seams

**Status:** ready-for-agent

- [x] run.go: flags → RunOptions; rendering consumes masked RunResult; retry notices via OnRetry
- [x] test.go: same seam; assertion output unchanged
- [x] pagination.go / bulk.go: sendFn delegates to Run with per-row RuntimeVars; step printers unchanged; drop private contains/indexOf helpers if trivially replaceable
- [x] collection.go: both call sites migrated
- [x] Contract step: unused preamble helpers deleted (`mergeEnvScope`, per-command client wiring); auth.go keeps openTokenStore only via OpenForWorkspace
- [x] Existing CLI httptest suite passes unmodified (except where it asserted internals); new parity case: `reqly run` records a history entry when a workspace exists, none without one
- [x] `go vet` + `gofmt -l` + `go test -race ./apps/cli/...` green; `reqly run --help` smoke
