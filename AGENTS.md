# Reqly — Agent & Developer Guidelines

Local-first, Git-native API development environment. Go core (`internal/` + `apps/cli`), Wails v3 desktop (`apps/desktop`).

## 1. Non-Negotiable Boundaries

- **Local-first & Git-native:** Collections, environments, request files, and tests are plain-text (JSON/YAML) versioned with Git. No cloud backends or accounts.
- **Zero telemetry:** Payload, header, credential, secret, and traffic data stays local, always.
- **Dual parity:** Core features live in Go (`internal/`) and are exposed via CLI (`apps/cli`) and Desktop (`apps/desktop`).

## 2. Session Memory (context-mode)

Use the `context-mode` plugin for session continuity: before asking the user what was decided or being worked on, search prior context via `context-mode_ctx_search` (timeline/relevance). Store decisions and constraints with `context-mode_ctx_index` as they're made. Session history persists across `/clear`/`/compact`; only `ctx purge` wipes it.

## 3. Layout

- `internal/` — shared domain logic (request engine, runner, auth, mocking, openapi, importer/exporter, websocket, sse, validation, diffing, mcp, requestfile parser).
- `apps/cli` — Cobra CLI. `apps/cli/cmd/*.go` is one command per file; `root.go` holds the root command.
- `apps/desktop` — Wails v3. Go backend in `backend/`; React host frontend in `frontend/` (Vite + TS, **no test script** — use `npm run typecheck`), shared UI in the root `frontend/` workspace.
- `docs/` — specs, ADRs (`docs/adr/`), and references (`docs/reference/`). `CONTEXT.md` is the canonical glossary; `ROADMAP.md` tracks milestone status.

## 4. Skill Pipeline

For features or architecture changes, run the 5-stage pipeline in order (`~/.agents/skills/`):

1. `/grill-with-docs` — stress-test requirements against `CONTEXT.md` (one question per turn); update `CONTEXT.md` and `docs/adr/` inline.
2. `/to-spec` — produce the technical spec/design doc.
3. `/to-tickets` — split the spec into actionable tickets.
4. `/implement` — execute tickets with strict TDD (tests first). For frontend work (`apps/desktop/frontend` + shared `frontend/`), always follow `/frontend-design`.
5. `/code-review` — audit against conventions; verify clean test output.

## 5. Verification & Quality Gates

A change is done only when all unit and integration tests pass with no dummy fallbacks or skipped assertions.

- **Go core & CLI:** `go test ./...` (use `go test -race ./...` before shipping)
- **CLI binary:** `go build -o reqly ./apps/cli`
- **Frontend:** `cd apps/desktop/frontend && npm run typecheck` (no test suite exists)
- **Style:** gofmt; follow [`docs/reference/go.md`](./docs/reference/go.md)

## 6. Domain Glossary

Consult [`CONTEXT.md`](./CONTEXT.md) before naming entities or creating abstractions.
