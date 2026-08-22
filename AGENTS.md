# Reqly — Agent & Developer Guidelines

Local-first, Git-native API development environment. Go core (`internal/` + `apps/cli`), Wails v3 desktop (`apps/desktop`).

## 1. Non-Negotiable Boundaries

- **Local-first & Git-native:** Collections, environments, request files, and tests are plain-text (JSON/YAML) versioned with Git. No cloud backends or accounts.
- **Zero telemetry:** Payload, header, credential, secret, and traffic data stays local, always.
- **Dual parity:** Core features live in Go (`internal/`) and are exposed via CLI (`apps/cli`) and Desktop (`apps/desktop`).
- **No CGO:** `modernc.org/sqlite` only (WAL + FTS5 history/cookie jar); no `mattn/go-sqlite3`.

## 2. Session Memory (context-mode)

Use the `context-mode` plugin for session continuity: before asking the user what was decided or being worked on, search prior context via `context-mode_ctx_search` (timeline/relevance). Store decisions and constraints with `context-mode_ctx_index` as they're made. Session history persists across `/clear`/`/compact`; only `ctx purge` wipes it.

## 3. Layout

- `internal/` — shared domain logic. Current packages: `auth` (incl. aws/edgegrid/oauth2), `collections`, `core`, `diffing`, `docs`, `environments`, `exporter`, `git`, `graphql`, `grpc`, `history` (SQLite + FTS5 + cookie jar), `importer`, `mcp` (stub), `mocking`, `openapi`, `request`, `requestfile`, `response`, `runner`, `scripting` (Goja), `secrets` (FileStore + KeychainStore via `go-keyring`), `sse`, `testing` (assertions/JSONPath), `validation`, `variables` (6 scopes + `{{$tag}}` + `.env`), `version`, `websocket`.
- `apps/cli` — Cobra CLI. `apps/cli/cmd/*.go` is one command per file; `root.go` holds the root command. Shipped commands: `run`, `test`, `collection` (run/list/test), `import` (curl/openapi), `export` (postman/code/workspace), `ws`, `sse`, `mock`, `validate`, `diff`, `docs generate`, `env` (list/show/use/validate/diff), `auth` (login/status/logout), `history` (list/show/search/clear/replay).
- `apps/desktop` — Wails v3. Go backend in `backend/` (`AppService` via `NewAppService()`, `application.NewService`); React frontend in `backend/frontend` (Vite + TS, **no test script** — use `npm run typecheck`). Bindings: `wails3 generate bindings`.
- `docs/` — specs, ADRs (`docs/adr/0019` latest: cross-platform desktop), and references (`docs/reference/`). `CONTEXT.md` is the canonical glossary; `ROADMAP.md` tracks milestone status (P0 100% through M27, P1 ~15%).
- Root release assets: `Taskfile.yml` (OS matrix), `.goreleaser.yaml`, `install.sh` (Linux/macOS multi-distro), `install.ps1` (Windows), `.github/workflows/release.yml`.

## 4. Skill Pipeline

For features or architecture changes, run the 5-stage pipeline in order (`~/.agents/skills/`):

1. `/grill-with-docs` — stress-test requirements against `CONTEXT.md` (one question per turn); update `CONTEXT.md` and `docs/adr/` inline.
2. `/to-spec` — produce the technical spec/design doc.
3. `/to-tickets` — split the spec into actionable tickets.
4. `/implement` — execute tickets with strict TDD (tests first). For frontend work (`apps/desktop/backend/frontend`), always follow `/frontend-design`.
5. `/code-review` — audit against conventions; verify clean test output.

## 5. Verification & Quality Gates

A change is done only when all unit and integration tests pass with no dummy fallbacks or skipped assertions.

- **Go core & CLI:** `go test ./...` (use `go test -race ./...` before shipping; `go vet` + `gofmt -l` must be clean)
- **CLI binary:** `go build -o reqly ./apps/cli`
- **Frontend:** `cd apps/desktop/backend/frontend && npm run typecheck` (no test suite exists — Vitest TBD)
- **Lint:** `npm run lint` from the repo root (oxlint + vendored anti-slop rules; must exit 0 — warnings from the four relaxed JSON-boundary rules are expected)
- **Style:** gofmt; follow [`docs/reference/go.md`](./docs/reference/go.md) + `~/.agents/skills/cc-skills-golang/` (`golang-code-style`, `golang-naming`, `golang-error-handling`, etc.)
- **Release:** `Taskfile.yml` OS matrix + `release.yml` checksums verified on `v*.*.*` tags (ADR 0019). File modes: `0600` for secrets/tokens/history.db, `0644` for request/workspace files.

## 6. Domain Glossary

Consult [`CONTEXT.md`](./CONTEXT.md) before naming entities or creating abstractions. Key recent entries: `Template Tag`/`TagGenerator`/`Dynamic Tag Picker` (ADR 0015), `Code Generation`/`Exporter` (ADR 0016), `Workspace Save`/`SaveWorkspace` (ADR 0017), `Docs Generation` (ADR 0018), `Release Pipeline`/`Unsigned Release Fallback`/`Multi-Distro Install Script` (ADR 0019).
