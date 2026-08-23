# 01 — Engine retry loop (internal/request)

**What to build:** A request whose file (or flags, ticket 02) declares a retry policy is automatically re-sent on transient failure: network errors always, plus statuses 429/502/503/504 by default; exponential backoff doubling from the base delay and capped, fixed delays selectable; `Retry-After` honored and clamped; cancellation aborts mid-wait; every consumer of the engine (CLI, desktop, runners) inherits the behavior with zero per-surface wiring.

**Blocked by:** None — can start immediately

**Status:** ready-for-agent

- [x] `request.Retry` struct (`count`, `delayMs`, `strategy`, `maxDelayMs`, `retryOn`) as optional pointer field on Request, json+yaml tags mirroring pagination block
- [x] Defaults in-package when retry present: exponential strategy, delayMs 1000, maxDelayMs 30000, retryOn {429,502,503,504}; count <= 0 → off
- [x] Attempt loop around the send path: attempt = full send including existing auth refresh/digest re-send (auth stays inside one attempt, never consumes budget)
- [x] Delay computation: fixed constant; exponential `delayMs * 2^(attempt-1)` capped at maxDelayMs; Retry-After (seconds or HTTP-date) overrides when larger, clamped to cap
- [x] Backoff wait selects on ctx.Done() — cancellation returns immediately mid-wait
- [x] Network errors always retryable; ctx.Canceled never retried; final failure returns last response/error unchanged
- [x] `response.Response.Attempts int` set on every Execute return (1 = no retries); Wails bindings regenerated
- [x] Table-driven tests with stub round-tripper/httptest fail-N-then-succeed: attempts counting, both strategies + cap, Retry-After seconds/date/clamp, retryOn override (500 excluded default, retried when listed), ctx-cancel during backoff, auth-401 orthogonality, no-retry-config → unchanged behavior
- [x] `go vet` + `gofmt -l` + `go test -race ./internal/request ./internal/response` green
