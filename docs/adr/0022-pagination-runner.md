# ADR 0022: Pagination Runner (M30)

## Status
Draft — grill Q1 settled (4 strategies MVP)

## Context
`ROADMAP.md:345` (`Pagination runner — page/offset/cursor/link-header traversal, stop conditions, aggregation`) and `docs/features.md:27` (page/pageSize, offset/limit, cursor, link headers, automatic traversal, stop, aggregation, maxPages) remain unchecked. `internal/runner` is sequential single-flight, `internal/collections` descriptors hold plain request files, and `internal/variables` already interpolates `{{var}}` per request. Design questions: which strategies ship in M30, where pagination config lives (`request.pagination` vs collection-level), how `next` is extracted (JSONPath vs header), stop semantics, and CLI shape (`reqly pagination run` vs `reqly collection run --paginate`).

## Decision
1. **M30 is 4 strategies MVP without aggregation export.** `page` (`?page=1&pageSize=20` via `pageParam`/`pageSizeParam`), `offset` (`?offset=0&limit=20` via `offsetParam`/`limitParam`), `cursor` (`?cursor=<next>` via `cursorParam` + `nextPath` JSONPath `$.nextCursor`), and `link-header` (`rel="next"` URL) — declarative `pagination: {strategy, param, nextPath, maxPages}` on the request file (mirroring `request.auth`/`request.body` flat config). Loop up to `maxPages` (default 100) via `internal/pagination` pure loop reusing `RequestService.Send` per step; M30b adds aggregate concat + `--out`.
2. **Loop reuses runner seams, not a new engine.** `internal/pagination.Run(ctx, req, opts, sendFn)` is pure function over `sendFn func(Request) (Response,error)` (like `runner.Run` `OnStep` streaming) — no new SQLite, no new `history` spill; history/cookies flow through existing `core.RequestService`.
3. **Stop is structural.** Empty array body, missing `next` (cursor/link empty), non-2xx status, or `maxPages` — no `while` expression or scripting for M30; `M30b` adds predicate stop.

## Considered Options
- **Page+offset only for M30, cursor/link deferred** — rejected: cursor + link-header are the dominant long-tail APIs and share the same `nextPath` extraction seam as `page`/`offset`.
- **Aggregation concat in M30** — rejected: requires JSON merge semantics (array vs object) and `--out` file modes; streaming per-step via `OnStep` already satisfies debugging, aggregation is additive in M30b.
- **`reqly collection run --paginate` flag** — rejected: pagination is a request-level concern (one endpoint), not a collection-level flag; collection pagination is a loop over one request, not over many requests.

## Consequences
- **Positive:** Closes `ROADMAP.md:345` strategies + stop + maxPages with one pure package, one Cobra seam, no new storage, and a clear M30b path for aggregation.
- **Trade-off:** M30 has no `--out` aggregated file; `nextPath` is JSONPath only (XPath deferred).
