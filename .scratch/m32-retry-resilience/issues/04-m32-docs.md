# 04 — M32 docs + quality gates

**What to build:** Milestone closure: roadmap truthfulness, spec finalized, full-suite race green, shipped binary smokes retry end-to-end.

**Blocked by:** 01, 02, 03

**Status:** ready-for-agent

- [x] ROADMAP: tick Phase 2 retry bullet + Next milestone #32 marked shipped with ADR 0024 link; Phase 2 progress ~40%
- [x] Spec Status header → Shipped (settled Q1–Q12); ADR 0024 stays Accepted
- [x] Full gates: `go test -race ./...`, `go vet`, `gofmt -l` clean, `go build -o reqly ./apps/cli`, frontend typecheck
- [x] Smoke: `reqly run <flaky-file> --retries 2` against a fail-once httptest fixture succeeds with retry lines
- [x] Commit docs artifacts on docs/internal branch only
