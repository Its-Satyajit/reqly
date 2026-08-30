# Phase 0: Foundation

## Phase 0 — Foundation (100% complete)

Project skeleton, build system, and the first core primitives.

### 0.1 Repository & build infra

- [x] Go module `github.com/Its-Satyajit/reqly` (Go 1.25)
- [x] npm workspaces + nub package manager (`pnpm-lock.yaml` committed)
- [x] Wails v3 desktop project (`apps/desktop/backend`) with Taskfile + build assets
- [x] CI workflow (frontend typecheck/build job; Go vet/gofmt/race/coverage job)
- [x] Makefile task aliases
- [x] Apache-2.0 license + SPDX headers on all Go sources
- [x] GoReleaser + Wails OS-matrix release pipeline (`release.yml`, `Taskfile.yml`, `install.sh`/`install.ps1`, ADR 0019)

### 0.2 Desktop shell (Wails v3)

- [x] `main.go` — Wails v3 `application.New`, window (1280×800), dark background (`NewAppService()` constructor)
- [x] `AppService` binding registered + `Greet` bridge proof → replaced by real `SendRequest` binding (see §1.5)
- [x] Go ↔ TypeScript bindings generated (`wails3 generate bindings`)
- [x] Host app (`apps/desktop/backend/frontend`) — Vite + React + Tailwind, wails vite plugin, port 9245
- [x] `wails3 build` produces `bin/reqly`
- [x] Backend warning/error log mirror — slog handler emits `reqly.golog` events so desktop crash reports include Go-side diagnostics
- [x] sqlc-generated typed query layer over `modernc.org/sqlite` for the history store (`internal/history/db`; schema/query SQL in-repo, zero reflection, no CGO)

### 0.3 Shared UI shell (`frontend/`)

- [x] App shell (header, sidebar, split request/response panes)
- [x] Light/dark theming with Reqly brand colors + theme store + toggle
- [x] Dark/light logo in header; logo as app icon
- [x] Base UI via shadcn CLI (`button` component, `#`-alias imports)
- [x] CodeMirror 6 editor wrapper (json/js/xml/yaml/markdown/text)

### 0.4 Core primitives (shipped, TDD)

- [x] `internal/variables` — 6-scope resolution + `{{key}}`/`{{$tag}}` interpolation + `.env` process-env scope + env-file validation/diff
- [x] `internal/scripting` — Goja runtime with `reqly` sandbox (request/response access, variable get/set, `reqly.test()`, console) + pre/post wiring + dynamic values
- [x] `internal/request` + `internal/response` — request engine + response model (see §1.1)
- [x] `internal/testing` — assertion engine + JSONPath + suite runner + test-file loader (see §1.11)
- [x] `internal/history` + `internal/secrets` — SQLite history/cookie jar + token store (FileStore + KeychainStore)

### 0.5 CLI skeleton

- [x] Cobra command tree: `run`, `test`, `collection run`, `mock`, `validate`, `diff`, `docs` (+ `collection list`/`test`, `import`, `export`, `env`, `auth`, `history`, `ws`, `sse`)
- [x] 15 CLI commands wired to the Go core — `run`, `test`, `collection run`/`list`/`test`, `import curl`/`openapi`, `export postman`/`code`/`workspace`, `ws`, `sse`, `mock`, `validate`, `diff`, `docs generate`, `env` (list/show/use/validate/diff), `auth` (login/status/logout), `history` (list/show/search/clear/replay)

---

## Code Review Gate (`/code-review` — two-axis)

- [ ] Standards: `oxlint` + `gofmt`/`go vet` + `anti-slop` + Fowler smell baseline — `git diff main...HEAD` (three-dot, merge-base) — no `as` without `// SAFETY:`, no hard violations
- [ ] Spec: this milestone (`Milestones/01`) vs implementation (`ROADMAP.md` Phase 0 + DoD: core+UI/CLI+tests) — `git log main..HEAD` + `git diff main...HEAD` — both axes must be green before ticking `[x]` above; fix `main...HEAD` diff until green — run `/code-review`

