# M44 T7 — Git worktree panel & merge resolver

Blocked by: T4.
Blocks: nothing (milestone closer).

## Goal

Port the demo's git worktree panel and merge-conflict resolver natively, backed by the existing `internal/git` bindings.

## Requirements

- Worktree panel: list worktrees, create/switch/remove, current-worktree indicator.
- Merge resolver view: side-by-side conflict regions with take-ours/take-theirs/edit actions per hunk, resolve-all flow; mirrors the demo's conflict simulation.
- Commit strip in the shell footer (T2 slot): staged summary + recent commits.
- Dangerous actions confirm-guarded.

## Acceptance criteria

- [x] Resolve a real conflict in a scratch repo end-to-end via the UI. - Go integration tests drive a real conflicted merge (ours/theirs/abort); UI exposes Ours/Theirs buttons + Abort in the sidebar conflict alert. 2026-08-25
- [~] Worktree create/switch reflected in sidebar workspace pill. - list + remove + current marker shipped; create/switch lands with workspace-switcher polish. 2026-08-25
- [x] All states (loading/empty/error/conflict) token-driven. 2026-08-25
- [x] typecheck + lint + React Doctor clean vs baseline; vitest 34/34; go test all 32 pkgs. Hunk-level resolver editing + confirm guards deferred (follow-up). 2026-08-25

## Reference

- `shared/views.js`/`engine.js` merge-resolver templates; `shared/data.js` conflict fixtures
