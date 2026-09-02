# Notes — api-client loops & language

## World

- **Tools:** `gh` (issues/PRs, `gh pr view/list/checks`, `gh issue view/edit/close`, `gh api`), `but` (GitButler `status/diff/commit/branch/push/pull/land`, stacked branches `m19→m20→m21`), `go` (`go test -race`, `go vet`, `gofmt`), `wails3` (`wails3 generate bindings -d ../frontend/bindings`, `wails3 dev/build` from `apps/desktop/backend`), `nub` (package manager `0.7.5`, `nub ci`, `nub run typecheck`), `oxlint` (`oxlint.config.ts` with vendored `anti-slop` at `tools/oxlint/anti-slop`, 11 error + 4 warn), `vite` + `tsc`.
- **Channels:** GitHub issues (`spec` label, `ready-for-agent`), PRs (stacked `m19→m20`, `oxc`, `m21`), `Dependency Dashboard` #102, `release-please` #142, `ROADMAP.md` Phase 1 ~65-67%, `CONTEXT.md` glossary, `docs/adr/` (0011 auth editing, 0012 AWS/EdgeGrid, 0013 binary/GraphQL), `workflows/` (triage, renovate, release).
- **Workspace:** `/home/satyajit/Documents/GitHub/OSS/api-client` with `frontend/` + `apps/desktop/frontend` (Vite + React + Tailwind, `CodeMirrorEditor`), `apps/desktop/backend` (Wails v3 Go, `AppService`, `build/config.yml`), `internal/` (Go core).

## Terminology (canonical)

- **Triage loop:** Event(new issue/PR) → categorise → verify repro → grill → draft brief → checkpoint(brief) → approve → `ready-for-agent` (push-right).
- **Renovate 5:** The 5 dependency PRs #157 `golang-1.26-trixie`, #158 `alpine`, #159 `golang-1.x`, #177 `wails/v3 beta.11`, #178 `goja digest` (scope of renovate workflow, not `release-please`).
- **Release loop:** Event(`git push origin v*.*.*`) → GoReleaser + Wails matrix → checksums/notes → brief → approve → `gh release create`.
- **Stacked PRs:** GitButler stacks (`m19` base, `m20`/`oxc`/`m21` on top) merged via `but land --yes` (direct to `origin/main`, orphans PRs) vs `gh pr merge --auto` (not supported for stacks).
- **Brief:** Tight, decision-ready summary (<5 lines) with link, not raw output — speed of review is imperative.
- **Push-right:** Defer checkpoint as far as it will go — do maximal work before involving the human, so they are asked once, late, with everything prepared.

## Decisions

- Triage and Renovate workflows run **in-workspace** via `gh`+`but`, brief as GitHub comment + label, spec as `workflows/*.md`.
- Release workflow trigger is tag push, checkpoint is post-build brief.
