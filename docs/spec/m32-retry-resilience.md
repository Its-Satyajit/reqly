# Spec: Retry & Resilience (Milestone 32)

> **Status:** Shipped 2026-08-23 — grill settled 2026-08-23 (Q1–Q12, engine-level MVP + desktop UI pass)
> **Scope:** P1 `ROADMAP.md:347` — count/delay/backoff, status + network-error + 429 handling, `Retry-After` respect, engine-level so CLI + desktop + runners inherit; desktop request-editor retry section via `/frontend-design` + `/design-principles`
> **Stack:** `internal/request` (Retry type + loop in `Client.Execute`) + `internal/response` (`Attempts` field) + `apps/cli/cmd/run.go` flags + desktop request editor — no new deps, stdlib `time`/`net/http`
> **Predecessor:** M31 Bulk `docs/adr/0023` (deferred retry to M32), auth one-shot retries in `client.go:123–177`

## Problem Statement

`ROADMAP.md:347` requires retry with count/delay/backoff and status/network-error/429 handling — none exists today. Transient failures (429 rate limits, 502/503/504 gateway blips, connection resets) fail immediately; users must wrap `reqly run` in shell loops. The only re-send logic is the bounded 401 auth path (Digest challenge, OAuth2 token refresh) in `internal/request/client.go`, and `Retry-After` is never read anywhere.

## Solution

Retry lives **inside `request.Client.Execute`**, config declarative on the Request:

* **`request.Retry` block** — `Retry *Retry` on `request.Request` (`json:"retry,omitempty" yaml:"retry,omitempty"`), mirroring `Pagination *Pagination`:

  ```yaml
  request:
    url: https://api.example.com/flaky
    retry:
      count: 3            # retries after first attempt; 0/absent = off
      delayMs: 1000       # base delay
      strategy: exponential   # fixed | exponential (default exponential when retry set)
      maxDelayMs: 30000   # cap
      retryOn: [429, 502, 503, 504]   # optional override
  ```

* **Loop in `Client.Execute`** — attempt = full send including in-attempt auth refresh/digest (orthogonal, never consumes budget). After a failed attempt (network error, or status in `retryOn` default `429,502,503,504`), wait then resend until `count` exhausted or success. Delay: `fixed` → `delayMs`; `exponential` → `delayMs * 2^(attempt-1)` capped at `maxDelayMs`. `Retry-After` (seconds or HTTP-date) on 429/503 overrides computed delay, clamped to `maxDelayMs`. Backoff sleep selects on `ctx.Done()` — cancellation aborts mid-wait. All methods retried regardless of idempotency (opt-in per file, documented caveat).

* **Observability** — `response.Response.Attempts int` (new field, 1 = no retries). CLI prints `retrying in 2s (attempt 2/3)` per retry (masked output); silent otherwise. History records final attempt only via existing `HistoryService.Record`, carrying `Attempts`.

* **CLI flags** — `reqly run --retries N --retry-delay <duration>` gated by `flags.Changed()`, overriding file values; URL-mode requests accept the same flags.

* **Desktop UI** — compact Retry collapsible section next to the timeout control in the request editor (count, delayMs, strategy dropdown, maxDelayMs), collapsed by default showing "Retries: off" when unset (progressive disclosure). Designed per `/frontend-design` + `/design-principles`; saves through the existing format-preserving save seam. No new bindings needed — `Execute` behavior flows through the shared engine.

## User Stories

1. As a user, I want `retry: {count: 3}` in my request file so flaky endpoints are automatically retried without shell loops.
2. As a user, I want exponential backoff capped at `maxDelayMs` so hammering a struggling server doesn't escalate.
3. As a user, I want `Retry-After` on a 429 respected so I don't re-hit the rate limit early.
4. As a user, I want network errors (timeouts, refused connections) retried by default so transient blips don't fail my run.
5. As a user, I want `retryOn: [500, 503]` to override the default status set so bare 500s retry when I choose.
6. As a user, I want Ctrl-C / cancel to abort instantly mid-backoff so waits never block shutdown.
7. As a user, I want `--retries 5` on the CLI to override the file without editing it.
8. As a desktop user, I want a small Retry section beside timeout so I can enable retries visually without YAML.
9. As a developer, I want deterministic delays (no jitter) so tests assert exact timing with stub clocks/servers.
10. As a developer, I want the auth-refresh re-send to stay inside one attempt so retry counts stay meaningful.

## Implementation Decisions

- **Seam: engine-level, not a pure-loop package.** Unlike pagination/bulk (orchestration concerns over a sendFn), retry is transport-level: implement in `internal/request/client.go` around the send path of `Execute`. Config type `type Retry struct { Count int; DelayMs int64; Strategy string; MaxDelayMs int64; RetryOn []int }` in `internal/request/request.go` with json+yaml tags; pointer on `Request`.
- **Defaults filled in-package** when `Retry != nil`: `Strategy "" → "exponential"`, `DelayMs <= 0 → 1000`, `MaxDelayMs <= 0 → 30000`, `RetryOn empty → {429, 502, 503, 504}`. `Count <= 0` → treat as off (single attempt).
- **Attempt loop:** for attempt := 1..count+1: send (existing build/do/auth-refresh path untouched); classify result — nil error + status not in retryOn → return; else if attempts remain, compute delay (Retry-After parse `strconv.Atoi` seconds or `http.ParseTime` date; clamp to MaxDelayMs; else backoff formula), select on ctx.Done() vs `time.After`, continue. Final failure returns last response/error unchanged plus `Attempts` set.
- **Network errors always retryable** (any non-nil send error except ctx cancellation — cancelled ctx returns immediately).
- **Auth orthogonality:** digest challenge and token-refresh re-sends happen inside one attempt exactly as today; the retry loop wraps them and does not interact.
- **Response:** add `Attempts int` to `response.Response` (json tag for bindings); set on every Execute return (1 when no retry configured). Regenerate Wails bindings.
- **CLI:** `--retries int`, `--retry-delay duration` on `run` only (not test/collection — those inherit via files); applyRunOverrides-style gating with `flags.Changed()`.
- **History:** single record per logical execution (final response), `attempts` visible in `reqly history show`; no per-attempt rows.
- **Desktop:** frontend-only addition to the request editor (Settings area near timeout); binding regeneration picks up `Attempts`/`Retry` models; UI pass follows `/frontend-design` + `/design-principles` skills.

## Testing Decisions

- **What makes a good test:** stub the HTTP round-tripper (`http.Client` injected via `WithHTTPClient`) or use `httptest` servers that fail N times then succeed; assert attempt counts, exact delays (short values), Retry-After precedence, ctx-cancel mid-backoff, and `Attempts` reporting. Table-driven stdlib `testing`, matching `internal/pagination`.
- **Seams to test (highest first):**
  - `internal/request/client_test.go`: success-first-attempt (`Attempts=1`), fails-twice-then-success (`Attempts=3`), exhausts count returns last error/status, fixed vs exponential delays, `maxDelayMs` cap, `retryOn` override (500 not retried by default, retried when listed), network error retried, ctx cancellation aborts during backoff, Retry-After seconds + HTTP-date honored + clamped, auth 401 refresh doesn't consume budget, retry absent → zero behavior change.
  - `internal/requestfile`: YAML+JSON round-trip of `retry:` block via `Save` (format-preserving).
  - CLI `apps/cli/cmd/run_test.go`: `--retries`/`--retry-delay` override file values; httptest fail-then-succeed integration.
  - Frontend: `npm run typecheck` (no unit suite yet).
- **Quality gates:** `go test ./...` + `go test -race ./...` + `go vet` + `gofmt -l` clean; `go build -o reqly ./apps/cli`; frontend `npm run typecheck`.

## Out of Scope

- Jittered backoff (deterministic MVP; additive later).
- Circuit breakers, health-check hooks, global retry budgets — P2/P3 territory.
- Per-attempt history rows / replay of individual attempts.
- Desktop Run View retry visualization (attempt badges) — follow-up after editor section ships.
- Retry inside `test`/`collection` CLI commands beyond what request files declare.

## Further Notes

- **ADR:** `docs/adr/0024-retry-resilience.md` — engine-level placement, auth-orthogonality, idempotency stance, no-jitter trade-off.
- **Glossary:** `CONTEXT.md` `Retry Policy`/`Retry Attempt`/`Backoff`/`Retryable Response` (this spec's grilling Q8–Q12).
- **ROADMAP:** tick `ROADMAP.md:347` (§ Phase 2 bullet + Next milestones #32) after ship; Phase 2 % bump.
- **Ticket split (see `to-tickets`):** T1 `internal/request` Retry type + Execute loop + tests, T2 response.Attempts + history carry + CLI flags + tests, T3 desktop editor retry section (+bindings regen) via `/frontend-design`+`/design-principles`, T4 docs (ADR 0024 + CONTEXT + ROADMAP + this spec) + `-race` green.
