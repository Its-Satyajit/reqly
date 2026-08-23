# 03 — Docs + quality gate + ROADMAP tick

**What to build:** Close M29 with docs parity and green quality gates so `ROADMAP.md:344` JWT decode is shippable and M29b (verify/sign + desktop) has a clean base.

**Blocked by:** 02 — CLI reqly jwt decode

**Status:** ready-for-agent

- [ ] Verify `docs/adr/0021-jwt-tooling-decode.md` + `CONTEXT.md:263` `JWT Tooling`/`JWT Decode`/`Verify`/`Sign` landed in grill (already committed) — fixup if drift from `internal/jwt` API
- [ ] Tick `ROADMAP.md:344` M29 `reqly jwt decode (header/claims viewer, expiry detection)` + `docs/features.md:18` JWT Tooling docs update (decode shipped, verify/sign deferred)
- [ ] Full verification: `go test ./...` + `go test -race ./...` + `go vet ./...` + `gofmt -l` (must be clean) + `go build -o reqly ./apps/cli` + `reqly jwt decode --help` + real token smoke (`--json` valid, expired, malformed `exit 1`)
- [ ] `npm run lint` exit 0 (frontend untouched, anti-slop 0/0 enforced) — no new frontend code for M29
- [ ] Update Progress Tracker Phase 2 % after M29 ship (if P1 % moves)
