# Spec: Pagination Runner (Milestone 30)

> **Status:** Draft — grill settled 2026-08-23 (Q1–Q4, 4 strategies MVP)
> **Scope:** P1 `ROADMAP.md:345` — `page`/`offset`/`cursor`/`link-header` traversal, stop conditions, `maxPages` (aggregation/export deferred to M30b), `docs/features.md:27`
> **Stack:** `internal/pagination` (new pure loop) + `apps/cli/cmd/pagination.go` + `internal/collections` request file `pagination:` block + `core.RequestService.Send` seam reuse — no new deps, stdlib only
> **Predecessor:** M28 HAR `docs/adr/0020` shipped, M29 JWT decode `docs/adr/0021` shipped (decode MVP), M30 `docs/adr/0022-pagination-runner.md` (Q1–Q4)

## Problem Statement

`ROADMAP.md:345` and `docs/features.md:27` require automatic paginated traversal (page/pageSize, offset/limit, cursor, Link headers, stop conditions, aggregation, maxPages) — none exists today. Users must manually `reqly run` each page, hand-edit `{{page}}`/`{{cursor}}`, inspect `Link` headers, and stop by eye. `internal/runner` is sequential single-flight but only for multi-request collections, not for looping one endpoint; no `pagination:` config exists on request files.

## Solution

One declarative pagination block on the request file plus one pure loop seam plus one CLI command:

* **`request.pagination: {strategy, param, nextPath, maxPages}`** — flat YAML on the request file (like `auth`/`body`), e.g. `pagination: {strategy: cursor, cursorParam: cursor, nextPath: $.nextCursor, maxPages: 50}`. Four strategies: `page` (`?page=1` + `pageSize` via `pageParam`/`pageSizeParam`), `offset` (`?offset=0`+`limit` via `offsetParam`/`limitParam`), `cursor` (`?cursor=<next>` via `cursorParam` + `nextPath` JSONPath into prior body), `link-header` (`rel="next"` URL from `Link` header, no `nextPath`). Loop via `internal/pagination.Run(ctx, req, opts, sendFn, onStep)` pure over `sendFn func(Request)(Response,error)` (prod injects `core.RequestService.Send`, tests inject stub) streaming `onStep func(step int, req Request, resp Response, err error)` like `collection_run.go:OnStep`; mutated `Request.Query` copy per step, not runtime vars.

* **`reqly pagination run <request-file> [--max-pages <n>]`** — Cobra `pagination` root + `run` subcommand (`apps/cli/cmd/pagination.go`, registered in `root.go:70`, one-file-per-group). Reads request file (or workspace `requestfile`), resolves `pagination` block, loops until structural stop (empty `[]` or missing `next`, non-2xx, or `maxPages` default 100, CLI flag overrides file). Prints per-step `step/status/url/duration` to stdout; full response bodies available via verbose `--json` (defer). No aggregation concat or `--out` for M30 — per-step streaming suffices, M30b adds `--out results.json`.

## User Stories

1. As a user, I want `reqly pagination run ./collections/items/list.yaml` (page strategy) to fetch `?page=1`, `?page=2` ... until empty or `maxPages`, so I can crawl a `page/pageSize` API without editing `{{page}}`.
2. As a user, I want offset strategy to repeat `?offset=0&limit=20`, `?offset=20&limit=20` until `[]` or missing next, so `offset/limit` APIs are covered.
3. As a user, I want cursor strategy `cursorParam: cursor, nextPath: $.nextCursor` to take `nextCursor` from prior JSON body (`{"items":[...], "nextCursor":"abc"}`) and set `?cursor=abc`, so cursor APIs work via JSONPath.
4. As a user, I want link-header strategy to follow `Link: <https://api.example.com/items?page=2>; rel="next"` (RFC 8288) without `nextPath`, so GitHub-style pagination works.
5. As a user, I want loop to stop on `[]` body, missing `next`/`rel="next"`, non-2xx, or `maxPages` (default 100), so infinite loops are bounded.
6. As a user, I want `--max-pages 5` to override file `maxPages: 50`, so ad-hoc limits need no file edit.
7. As a user, I want pagination via a plain request file (`pagination:` block) Git-native, `git diff` shows the strategy, so review is Git-native like `auth`.
8. As a user, I want per-step progress (`step 1: 200 124ms https://...?page=1`) streamed, so long crawls are observable like `collection run` OnStep.
9. As a developer, I want `internal/pagination.Run` unit-tested with stub `sendFn` (no network), so tests are deterministic and fast.

## Implementation Decisions

- **Seam: `internal/pagination` pure package, highest seam.** `func Run(ctx context.Context, req Request, opts Options, sendFn SendFunc, onStep func(Step)) error` where `Options{Strategy string, PageParam, PageSizeParam, OffsetParam, LimitParam, CursorParam, NextPath, MaxPages int, InitialValues map[string]string}` and `Step{Index int, Request Request, Response Response, Err error, Next string}`. Imports only `encoding/json`, `net/http`, `strconv`, `strings`, `net/url` — stdlib only, same posture as `internal/jwt`. No new SQLite/history table; history/cookies flow through `sendFn`'s existing `core.RequestService`.

- **Request file `pagination:` block schema (flat, Git-native):** `strategy: page|offset|cursor|link-header` (required), `pageParam: page` (default `page`), `pageSizeParam: pageSize` (default `pageSize`), `offsetParam: offset` (default `offset`), `limitParam: limit` (default `limit`), `cursorParam: cursor` (default `cursor`), `nextPath: $.nextCursor` (required for `cursor`, ignored for `link-header`/`page`/`offset`), `maxPages: 100` (default), `pageSize/limit: 20` (defaults when omitted). Stored alongside `request.url`/`method`/`headers` in `requestfile.File`.

- **4 strategies mutating Query copy per step:** `page` increments `page` param starting 1; `offset` adds `limit` to offset starting 0; `cursor` replaces `cursorParam` with `next` extracted from prior body via `jsonpath` `$.nextCursor` (reusing `internal/testing` JSONPath or simple `encoding/json` map walk for `$.field` for M30); `link-header` parses `Link` header `rel="next"` URL and uses it as next request URL (overrides `Query` for that step). All via `request.Request{Query []Parameter}` copy, not `variables.ScopeRuntime`.

- **Stop is structural, no scripting:** after each `sendFn`, if `status != 2xx` → stop with error Step; else if `len(body)==0` or body is `[]` or `{"items":[]}` and `next==""` → stop; else if `next==""` (cursor/link missing) → stop; else if `step >= MaxPages` → stop. No `while` expression for M30.

- **CLI shape: `pagination` root + `run` subcommand.** `apps/cli/cmd/pagination.go` defines `paginationCmd = &cobra.Command{Use:"pagination"}` + `paginationRunCmd = &cobra.Command{Use:"run <request-file> [--max-pages <n>]"}`; `root.go` registers `paginationCmd`. Reads `<request-file>` via `requestfile.LoadFile`, resolves `pagination` block, loops via `pagination.Run` with real `sendFn` (`request.Client` or `core.RequestService`). Flag `--max-pages` int overrides file `maxPages`.

- **Output contracts.** Default per-step line `fmt.Fprintf(out, "step %d: %d %dms %s\n", step, status, duration, url)` to stdout; `--json` verbose (maybe defer) — for M30 single line per step suffices like `collection run`. No aggregation file for M30.

## Testing Decisions

- **What makes a good test:** assert loop behavior via stub `sendFn` (per-step request URL/query + stop condition), not base64 internals. Table-driven `testing` like `internal/auth/jwt_test.go`.

- **Seams to test (highest first):**
  - `internal/pagination` unit: stub `sendFn` returning paginated bodies (`{items:[1], nextCursor:"a"}`, `Link: <...>; rel="next"`, `page` increment, `offset` add, `cursor` replace), assert `onStep` calls + `maxPages` cap + empty/missing-next/non-2xx stops, `MaxPages` default 100, CLI override, default param names.
  - CLI integration `apps/cli/cmd/pagination_test.go`: temp request file + `httptest.NewServer` paging 3× (page/offset/cursor/link), `--max-pages` override, `--help` contains `run`, malformed pagination block error.
  - No golden files — payloads short inline `httptest` ATS, deterministic `Now` not needed.

- **Prior art:** `internal/runner` collection runner `OnStep` streaming, `internal/variables/tag_test.go` stub generation, `apps/cli/cmd/collection_test.go` Cobra `Execute` harness.

- **Quality gates:** `go test ./...` + `go test -race ./...` + `go vet` + `gofmt -l` clean; `go build -o reqly ./apps/cli` + `reqly pagination run --help` smoke.

## Out of Scope

- Aggregation concat (`--out results.json`, merging `items[]` across steps) — M30b.
- `while` expression / scripting stop predicate — M30 structural stop only.
- Collection-level `pagination:` on `collections/<name>/reqly.yaml` (all requests paginated) — request file only for M30.
- Desktop Pagination tab + `AppService.Paginate` binding — M30b, reuses same `internal/pagination.Run` seam.
- XPath `nextPath`, non-JSON `next` extraction, `Cursor` from headers other than `Link` — JSONPath `$.field` only for M30.
- Retry/backoff during pagination (M32 `Retry & resilience`) — non-2xx stops, no retry for M30.

## Further Notes

- **ADR:** `docs/adr/0022-pagination-runner.md` (draft Q1–Q3, Q4 pending) — 4 strategies MVP, file-only, structural stop, `internal/pagination` pure, M30b aggregation.
- **Glossary:** `CONTEXT.md:275` `Pagination Runner`/`Strategy`/`Stop`/`Aggregation` (this spec's grilling).
- **ROADMAP:** `ROADMAP.md:345` M30 tick after ship; M28+ M29 already shipped (`adr/0020`+`0021`).
- **Ticket split (see `to-tickets`):** T1 `internal/pagination` strategies + stop + tests, T2 `apps/cli/cmd/pagination.go` + `root.go` wiring + CLI httptest, T3 docs (ADR 0022 + CONTEXT + ROADMAP + this spec) + `go test -race` green.
