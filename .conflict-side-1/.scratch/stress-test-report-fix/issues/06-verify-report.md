# 06: Verify repro pack and update STRESS_TEST_REPORT.md

**What to build:** All repros from §6 are green, verification gates all green, and `STRESS_TEST_REPORT.md` gets a Fix Verification appendix linking to commits and showing before/after.

**Blocked by:** 01, 02, 03, 04, 05 (needs all fixes landed)

**Status:** done — verified 2026-09-01 (commit `umz` on `fix/stress-report-verification`)

- [x] Build `/tmp/reqly` from current HEAD (`go build -o /tmp/reqly ./apps/cli`) — built 2026-09-01, `version: 1.2.0 commit: unknown`
- [x] Re-run §6 repro pack: B1 changelog YAML, B2 generic YAML, B3 typed bodies, B4 help dedup, B5 openapi positional, B6 test exit 1, B7 bruno, A1 mask, A3 empty list, A7 graphql, A8 bulk empty — capture outputs — all green (see `STRESS_TEST_REPORT.md:310` §9)
- [x] Run verification gates: `go test ./...` 0, `go test -race ./internal/diffing ./internal/request ./internal/requestfile ./internal/importer` 0, `go vet ./...` 0, `gofmt -l` empty — 44 ok, 4 race ok, vet 0, fmt 0
- [x] Update `STRESS_TEST_REPORT.md` with Fix Verification appendix (date 2026-09-01, commits `2a132025`, `52b2deef`, `ee81c9cd`, per-issue before/after + reproduction evidence) — `STRESS_TEST_REPORT.md:310`
- [x] `STRESS_TEST_REPORT.md` status header changed to `FIXED` or `VERIFIED 2026-09-01` — `STRESS_TEST_REPORT.md:3` `> **Status: FIXED — Verified 2026-09-01**`
- [x] No new failures introduced (positive controls §5 still green: env/history/collection/import/export/network/schema/jwt/governance) — `go test ./...` ok
