# 03 — Pagination docs + quality gate

**What to build:** Close M30 with docs parity and green gates so `ROADMAP.md:345` pagination strategies + stop + maxPages is shippable and M30b (aggregation + desktop) has clean base.

**Blocked by:** 02 — Pagination CLI

**Status:** ready-for-agent

- [ ] Finalize `docs/adr/0022-pagination-runner.md` (promote Draft → Accepted after 4 strategies land) — fixup if `internal/pagination` API drift
- [ ] Verify `CONTEXT.md:275` `Pagination Runner`/`Strategy`/`Stop`/`Aggregation` grilling landed, fixup if drift
- [ ] Tick `ROADMAP.md:345` M30 shipped (4 strategies + stop + maxPages, `internal/pagination` + `reqly pagination run`) — `verify`/`aggregate`/`desktop` deferred to M30b
- [ ] Full verification: `go test ./...` + `go test -race ./...` + `go vet ./...` + `gofmt -l` (clean) + `go build -o reqly ./apps/cli` + `reqly pagination run --help` + smoke with httptest paging 3×
- [ ] `npm run lint` exit 0 (frontend untouched) + `docs/spec/m30-pagination-runner.md` already drafted
