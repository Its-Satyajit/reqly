# 01 — Core ctx seam (internal/core)

**What to build:** `RequestService.Send` accepts a `context.Context` as its first parameter and threads it to `client.Execute`, removing the last core send path that hard-codes `context.Background()`. All internal callers updated; behavior unchanged when callers pass `context.Background()`.

**Blocked by:** None — can start immediately

**Status:** done

- [x] Signature becomes `Send(ctx context.Context, r request.Request, vars ...*variables.Set) (*SendResponse, error)` in `internal/core/request.go`
- [x] `ctx` passed to `s.client.Execute` (replacing the internal `context.Background()`)
- [x] All internal callers updated (desktop bridge, tests, any CLI seams) — grep for `.requests.Send(` / `RequestService).Send(`
- [x] Test: a pre-cancelled ctx makes `Send` return promptly with a cancellation error and record nothing in history
- [x] `go vet` + `gofmt -l` + `go test ./internal/core` green
