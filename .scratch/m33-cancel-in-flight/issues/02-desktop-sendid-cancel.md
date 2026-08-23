# 02 — Desktop sendId registry + CancelSend binding (apps/desktop/backend)

**What to build:** `SendOptions` gains an optional `SendID string`. When non-empty, `SendRequest` derives a cancellable ctx around the pipeline call and registers the cancel func in a mutex-guarded map, mirroring `runCancels`; new binding `CancelSend(sendID)` cancels it.

**Blocked by:** 01

**Status:** done

- [x] `sendMu sync.Mutex` + `sendCancels map[string]context.CancelFunc` on AppService (mirrors `runMu`/`runCancels`)
- [x] Non-empty SendID → `context.WithCancel`, register under lock, delete-on-finish via defer
- [x] Empty SendID → no registration, current behavior unchanged
- [x] New binding `CancelSend(sendID string) error` — missing/expired id = no-op returning nil
- [x] Table-driven tests in `app_test.go`: cancel mid-flight aborts send with cancellation error; no-op cancel after finish; cancelled send records no history entry
- [x] `wails3 generate bindings` run; `go vet` + `go test ./apps/desktop/backend` green
