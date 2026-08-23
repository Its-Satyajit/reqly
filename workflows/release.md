# Workflow: Release

> **Trigger:** Event — `git push origin v*.*.*` (semver tag, e.g. `v1.3.0`) or `workflow_dispatch` for dry-run.
> **Checkpoint:** Human-in-the-loop after build — agent presents a **brief** (binaries + checksums + notes, not raw logs) and waits for approve before publishing.
> **Push-right:** Agent does maximal work before checkpoint — builds, generates notes, verifies checksums — so the human is asked once, late, with everything prepared.

## Goal

Turn a tag push into a published GitHub release with GoReleaser CLI binaries + Wails v3 desktop matrix, without asking the human to run builds or write notes.

## Steps

1. **Detect** tag push via `git rev-list --tags --max-count=1` or `gh release view <tag> --json tagName` / `gh api /repos/Its-Satyajit/reqly/releases/tags/<tag>`. If `NOTES.md` is thin, first interview about versioning (Conventional Commits → `release-please` #142) and record.

2. **Build** — `GoReleaser` headless CLI: `go build -o reqly ./apps/cli` (mirrors `release.yml` `goreleaser` job) + Wails v3 OS matrix: `cd apps/desktop/backend && wails3 build` for `linux` (WebKit), `darwin` (WebKit, ad-hoc `codesign -s -`), `windows` (WebView2) via `Taskfile.yml` `linux:build` / `darwin:build` / `windows:build` (or `wails3 task` per `docs/adr/0001...`). Collect artifacts in `dist/` + `bin/`.

3. **Generate notes** — `GoReleaser` changelog from Conventional Commits (`docs/adr/0001...` `Automated Tag Release Trigger`) + `checksums.txt` (`sha256sum dist/*`).

4. **Verify** — `go test -race ./...`, `npm run lint` (expect 41 warns from 4 relaxed `oxlint.config.ts:56-59`), `npx tsc --noEmit` both frontends, `wails3 task build` smoke.

5. **Draft brief** — per-release one-liner: `Release v1.3.0 — CLI binaries (linux/amd64, darwin/arm64, windows) + desktop (AppImage, .deb, .dmg, .exe) — checksums ok — notes: <link>` with link to `dist/` + `CHANGELOG.md`.

6. **Present brief** — at the checkpoint, show the brief (tight, decision-ready, with link to the drafted `gh release create --draft` assets). Do **not** publish yet.

7. **Await approve** — human reads the brief, not the raw build logs. On **approve**, `gh release create <tag> dist/* --notes-file CHANGELOG.md --latest` (or `GoReleaser` publish) + push `checksums.txt`. On **reject**, record reason in `NOTES.md` and delete draft: `gh release delete <tag> --yes`.

## Brief Format

```
Release v1.3.0 — CLI (3) + desktop (4) — checksums ok — notes: <link>
Link: https://github.com/Its-Satyajit/reqly/releases/tag/v1.3.0 (draft)
```

Speed of review is imperative — <3 lines.

## Tools & Channels

- **In-workspace** via `gh` + `but` + `go` + `wails3` + `GoReleaser`; reads `docs/adr/0001...`, `NOTES.md` for version language.
- **Output:** GitHub release (draft → published) + `NOTES.md` update.

## Definition of Done

Implementer could build without asking a question: trigger is tag push, checkpoint is post-build brief, brief format is fixed, tools are `gh`/`GoReleaser`/`wails3`, and `NOTES.md` holds version language. Done when `workflows/release.md` exists and `NOTES.md` has been sharpened.
