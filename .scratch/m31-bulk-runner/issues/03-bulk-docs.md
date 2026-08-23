# 03 — Bulk docs + quality gate

**What to build:** Close M31 with docs parity and green gates so `ROADMAP.md:346` bulk CSV/JSON + modes is shippable and M31b (variable/generated + desktop) has clean base.

**Blocked by:** 02 — Bulk CLI

**Status:** ready-for-agent

- [ ] Finalize `docs/adr/0023-bulk-runner.md` (promote Draft → Accepted after CSV+JSON lands) — fixup if `internal/bulk` API drift
- [ ] Verify `CONTEXT.md:275` `Bulk Runner`/`Bulk Input Row`/`Bulk Concurrency` grilling landed, fixup if drift
- [ ] Tick `ROADMAP.md:346` M31 shipped (CSV+JSON + sequential/parallel ordered + `--continue-on-error`, `internal/bulk` + `reqly bulk run --data`) — variable/generated + desktop deferred to M31b
- [ ] Full verification: `go test ./...` + `go test -race ./...` + `go vet ./...` + `gofmt -l` (clean) + `go build -o reqly ./apps/cli` + `reqly bulk run --help` + smoke CSV 2 rows via httptest
- [ ] `npm run lint` exit 0 (frontend untouched) + `docs/spec/m31-bulk-runner.md` already drafted
