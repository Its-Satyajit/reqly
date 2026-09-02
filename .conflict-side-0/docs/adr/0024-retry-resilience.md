# ADR 0024: Retry & Resilience (M32)

## Status
Accepted

## Context
`ROADMAP.md:347` (Retry & resilience — count, delay, backoff, status/network error, 429 handling) is unchecked. The only re-send logic in the engine is the bounded 401 path (Digest challenge, OAuth2 refresh) in `internal/request/client.go`; `Retry-After` is never read. Recent features (pagination M30, bulk M31) established a pure-loop + injected-sendFn pattern with CLI-only wiring, and M31 explicitly deferred retry to M32. Design questions: where the retry loop lives (engine vs orchestration package), whether non-idempotent methods are retried, how auth re-sends compose with retry budget, jitter vs determinism, and which surfaces get config (CLI flags, request file, desktop UI).

## Decision
1. **Retry lives inside `request.Client.Execute`, configured by a declarative `request.retry` block** (`count`, `delayMs`, `strategy: fixed|exponential`, `maxDelayMs`, `retryOn`) — not as a separate pure-loop package like pagination/bulk. Retry is transport-level: every consumer of the engine (CLI run/test/collection, bulk/pagination sendFns, desktop) inherits it without per-surface wiring.
2. **Default retryable set is network errors + `429/502/503/504`**, overridable via `retryOn`; bare `500` excluded by default (usually a deterministic server bug). All HTTP methods are retried regardless of idempotency — this is an opt-in, per-request-file dev tool and the caveat is documented; forcing idempotency checks would surprise users more than a duplicated POST.
3. **Auth re-sends stay orthogonal:** digest-challenge and token-refresh re-sends resolve within a single attempt and never consume retry budget.
4. **Exponential backoff default (factor 2, capped at `maxDelayMs`), fixed available, no jitter for M32** — deterministic delays keep tests exact; jitter can be added later without breaking the schema. `Retry-After` on 429/503 overrides the computed delay whenever present, clamped to `maxDelayMs` so a hostile server cannot stall the client.
5. **Observability via `response.Attempts`**: one history record per logical execution (final attempt only), CLI prints a masked `retrying in Ns (attempt i/j)` line.

## Considered Options
- **Pure-loop package (`internal/retry`) wrapping sendFn, CLI-only command** — rejected: unlike pagination/bulk (orchestration over many sends), retry is a property of one send; a wrapper would leave desktop, collection runner, and bulk un-retried or duplicated.
- **Idempotency guard (skip retry for POST/PUT)** — rejected: Postman-style tools retry everything; users opt in per request file where they know the semantics.
- **Jittered backoff in MVP** — rejected: nondeterminism complicates tests for marginal benefit at dev-tool scale.
- **Per-attempt history rows** — rejected: pollutes history with failures the user never acted on; final attempt + `Attempts` count suffices.

## Consequences
- **Positive:** One implementation covers every surface; request files stay Git-native declarative config; M31's deferral is closed cleanly; desktop gets the behavior for free plus a small editor UI pass (`/frontend-design` + `/design-principles`).
- **Trade-off:** Non-idempotent requests can double-execute on transient failures when retries are enabled; no jitter means synchronized retry storms are possible against rate-limiting servers (mitigated by `Retry-After` respect and per-user scale).
