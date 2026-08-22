# 01 — Pagination loop (internal/pagination)

**What to build:** Pure pagination loop `internal/pagination.Run` handling `page`/`offset`/`cursor`/`link-header` strategies, advancing `Request.Query` per step, extracting `next` via JSONPath `$.nextCursor` or `Link: rel="next"`, stopping on empty/missing next/non-2xx/`maxPages` (100 default), streaming `OnStep`. No aggregation for M30.

**Blocked by:** None — can start immediately

**Status:** ready-for-agent

- [ ] `internal/pagination` package pure stdlib (`encoding/json`, `net/url`, `strings`, `strconv`) — `Run(ctx, req, opts, sendFn, onStep) error` + `Options{Strategy, PageParam, PageSizeParam, OffsetParam, LimitParam, CursorParam, NextPath, MaxPages}` + `Step{Index, Request, Response, Err, Next}`
- [ ] 4 strategies: `page` increments `page` from 1, `offset` adds `limit` from 0, `cursor` replaces `cursorParam` with `next` at `NextPath`, `link-header` follows `Link: <url>; rel="next"` (parse RFC 8288, no nextPath)
- [ ] Stop: empty `[]` / missing `next` / non-2xx / `step >= MaxPages` (file `maxPages` default 100)
- [ ] Mutate `Request.Query` copy per step, not `variables.ScopeRuntime`; inject `sendFn func(Request)(Response,error)` for determinism
- [ ] Table-driven unit tests `internal/pagination/pagination_test.go` with stub `sendFn` (paginated `{items:[1], nextCursor:"a"}`, Link header, page increment, offset add, cursor replace, empty/missing-next/non-2xx/maxPages stops)
- [ ] `go vet` + `gofmt -l` + `go test -race ./internal/pagination` green
