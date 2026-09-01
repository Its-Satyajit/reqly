# Reqly — Agent & Developer Guidelines

Local-first, Git-native API development environment. Go core (`internal/` + `apps/cli`), Wails v3 desktop (`apps/desktop`).

## 0. Non-Negotiable Boundaries

- **Local-first & Git-native:** Collections, environments, request files, and tests are plain-text (JSON/YAML) versioned with Git. No cloud backends or accounts.
- **Zero telemetry:** Payload, header, credential, secret, and traffic data stays local, always.
- **Dual parity:** Core features live in Go (`internal/`) and are exposed via CLI (`apps/cli`) and Desktop (`apps/desktop`).
- **No CGO:** `modernc.org/sqlite` only (WAL + FTS5 history/cookie jar); no `mattn/go-sqlite3`. History queries are sqlc-generated (`internal/history/db`; edit `db/*.sql`, run `sqlc generate`) — see ADR 0027.

## 1. Skill Pipeline — _grill → spec → tickets → implement → review_

For features or architecture changes, run the 5-stage pipeline in order (`~/.agents/skills/`):

1. `/grill-with-docs` — stress-test requirements against `CONTEXT.md` (one question per turn); update `CONTEXT.md` and `docs/adr/` inline. **Done when:** every requirement has a clear acceptance criterion and no open questions remain.
2. `/to-spec` — produce the technical spec/design doc. **Done when:** the spec covers data model, API surface, edge cases, and testing strategy.
3. `/to-tickets` — split the spec into actionable tickets. **Done when:** each ticket is independently shippable with clear blocking edges.
4. `/implement` — execute tickets with strict TDD (tests first). For frontend work (`apps/desktop/backend/frontend`), always follow `/frontend-design`. **Done when:** all tests pass, no dummy fallbacks, no skipped assertions.
5. `/code-review` — audit against conventions; verify clean test output. **Done when:** all checks pass and the reviewer raises no blocking findings.

## 2. Verification Gates — _all green_

A change is done only when every gate below passes. Run each gate; fix failures before proceeding.

| Gate | Command | Pass condition |
|------|---------|----------------|
| Go tests | `go test ./...` | Exit 0, no skips |
| Go race tests | `go test -race ./...` | Exit 0 (run before shipping) |
| Go vet | `go vet ./...` | Exit 0 |
| Go format | `gofmt -l` | Empty output |
| CLI build | `go build -o reqly ./apps/cli` | Binary produced |
| Frontend typecheck | `nub run typecheck` | Exit 0 |
| Lint | `nub run lint` from repo root | Exit 0 (warnings from four relaxed JSON-boundary rules are expected) |
| React doctor | `nubx react-doctor@latest` (or `npm run react:doctor`) | Passes with no new warnings vs the baseline: **0 errors / 0 warnings / 0 bugs** (score 100) |

**Style:** follow [`docs/reference/go.md`](./docs/reference/go.md) + `~/.agents/skills/cc-skills-golang/` (`golang-code-style`, `golang-naming`, `golang-error-handling`, etc.)

**Release:** `Taskfile.yml` OS matrix + `release.yml` checksums verified on `v*.*.*` tags (ADR 0019). File modes: `0600` for secrets/tokens/history.db, `0644` for request/workspace files.

## 3. Milestone Tracking — _dual-roadmap_

When completing a milestone or shipping a feature, update the relevant roadmap(s) **before committing**.

### What to update

| Scope changed | File to update | Section |
|---------------|---------------|---------|
| Core Go logic (`internal/`) | [`ROADMAP.md`](ROADMAP.md) | Corresponding Phase/section |
| CLI command (`apps/cli/`) | [`ROADMAP.md`](ROADMAP.md) | Corresponding Phase/section |
| Desktop GUI (`frontend/` or `apps/desktop/`) | [`docs/internal/gui-roadmap.md`](docs/internal/gui-roadmap.md) | GUI-0 through GUI-16 |
| Domain terms | [`CONTEXT.md`](CONTEXT.md) | Relevant term |
| Design review finding | [`docs/internal/gui-roadmap.md`](docs/internal/gui-roadmap.md) | GUI-4 (Design Quality) |

### How to update

1. Find the matching milestone in the roadmap file.
2. Change `[ ]` → `[x]` (shipped) or `[~]` → `[x]` (partial → complete).
3. Add a brief note: `— <short description> (<date>)`.
4. If new domain terms were introduced, add/update entries in `CONTEXT.md`.
5. If a design review finding (GUI-4) was resolved, mark the `G-4.*` item as `[x]` in `docs/internal/gui-roadmap.md`.

**Dual-roadmap rule:** Features with both core logic and a GUI surface must be marked complete in **both** roadmaps. A feature is not "done" until both `ROADMAP.md` and `docs/internal/gui-roadmap.md` reflect its completion.

```markdown
# Before
- [ ] **G-5.1.1** Import dialog — modal with format auto-detection

# After
- [x] **G-5.1.1** Import dialog — modal with format auto-detection (cURL, OpenAPI, HAR, Postman) — 2026-08-25
```

## 4. Layout

- `internal/` — shared domain logic: `auth` (aws/edgegrid/oauth2), `collections`, `core`, `diffing`, `docs`, `environments`, `exporter`, `git`, `graphql`, `grpc`, `history` (SQLite + FTS5 + cookie jar), `importer`, `mcp` (stub), `mocking`, `openapi`, `request`, `requestfile`, `response`, `runner`, `scripting` (Goja), `secrets` (FileStore + KeychainStore), `sse`, `testing` (assertions/JSONPath), `validation`, `variables` (6 scopes + `{{$tag}}` + `.env`), `version`, `websocket`.
- `apps/cli` — Cobra CLI. One command per file in `apps/cli/cmd/*.go`; `root.go` holds root.
- `apps/desktop` — Wails v3. Go backend in `backend/` (`AppService`); React frontend in `backend/frontend` (Vite + TS). Bindings: `wails3 generate bindings`.
- `docs/` — ADRs, specs, references. `CONTEXT.md` is the glossary; `ROADMAP.md` tracks core/CLI; `docs/internal/gui-roadmap.md` tracks desktop GUI.

## 5. Session Memory — _context-mode_

Use the `context-mode` plugin for session continuity across `/clear`/`/compact` (only `ctx purge` wipes it). there is also skills there for `context-mode` tha patten is `ctx-*`.

| Need | Command |
|------|---------|
| Resume prior work | `context-mode_ctx_search(queries: [...], sort: "timeline")` |
| Store a decision | `context-mode_ctx_index(content, source)` |
| Check what was decided | `context-mode_ctx_search(queries: ["decision"], source: "decision", sort: "timeline")` |

Before asking the user what was decided or being worked on, search first.

### Tool priority — _strict order_

1. **context-mode** (`ctx_*`) — highest priority. Use for any operation a `ctx_*` tool can do: analysis, filtering, counting, searching, multi-command gathers (`ctx_batch_execute`), web fetches (`ctx_fetch_and_index`), file analysis (`ctx_execute_file`).
2. **Harness tools** (Read/Edit/Write/Grep/Glob) — when context-mode has no suitable operation. Read/Edit stay correct for file *mutation*; `ctx_execute_file` for read-only *analysis*.
3. **Bash** — last resort only. Allowed always: git/but mutations, `mkdir`/`rm`/`mv`, `go build/test/vet`, `npm`. Never bypass context-mode just because Bash is more convenient, especially for outputs that may be large or need processing.

## 6. Context

- Consult [`CONTEXT.md`](./CONTEXT.md) before naming entities or creating abstractions.

## 8. Agent skills

### Issue tracker

GitHub Issues via the `gh` CLI (historical `.scratch/<feature>/` markdown on `docs/internal` is an archive). See `docs/agents/issue-tracker.md`.

### Triage labels

Default vocabulary: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context — `CONTEXT.md` + `docs/adr/` on the quarantined `docs/internal` branch. See `docs/agents/domain.md`.

### Version control

GitButler (`but`) for all VCS writes; never raw git. See `docs/agents/gitbutler.md` and the `but` skill.

### Session memory

context-mode (`ctx_*`) tools first — context protection + persistent session memory across `/clear`/`/compact`. See `docs/agents/context-mode.md`.
