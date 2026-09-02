# ADR 0014: History, Cookie Jar, Table View & Binary Preview (M22)

## Status
Accepted

## Context
`ROADMAP.md:75` (SQLite history + replay), `ROADMAP.md:131-138` (response views — Table pending, Cookies display-only, binary/CSV preview pending) and `docs/features.md:11-12` are the last P0 desktop polish gaps after M21. `internal/history/doc.go:19` is a stub, `frontend/src/features/response-viewer/ResponseViewer.tsx` has Raw/Pretty/Tree/Headers/Cookies(display-only) + JSONPath but no Table, and `Response Cookie` is parsed but not persisted. The design question is where history/cookies live, how they stay local-first/Git-native vs Git-ignored, and what Table/Binary ships without bloating the response viewer.

## Decision
1. **Per-workspace SQLite in `<workspace>/.reqly/history.db` (pure-Go `modernc.org/sqlite`, WAL, FTS5).** One DB, two tables: `history(id, request_path, method, url, env, status, duration_ms, size, req_headers_json, req_body_path, resp_headers_json, resp_body_path, created_at)` with FTS5 on `url, request_path`, and `cookies(name, value, domain, path, expires_at, secure, http_only, same_site, env)` partitioned by `env`. Keeps history project-scoped like `tokens.json` (mirrors `Workspace.Root` lookup), 0600, `.reqly/` is `.gitignore`d — not Git-native by design (local metadata). File spills: bodies >1MB → `<workspace>/.reqly/history/blobs/<id>.bin`. Prune keeps last 500, oldest deleted.

2. **Cookie jar piggybacks the same DB.** Populate via `net/http` `ReadSetCookies` on every response (RFC 6265), matching via domain/path/secure before `request.Client.Send` auto-attaches `Cookie:`. Partition `env` column supports "Clear per environment" vs "Clear per workspace". UI is view+delete+clear; edit is delete+re-add. Opt-out per request via `Request.Settings.CookieJar: false` (default on).

3. **Table = JSON array-of-objects + CSV only; Binary = image/PDF/hex.** Table tab always visible, disabled with hint when not tabular (probe via `Content-Type` + `JSON.parse`). JSON array union of keys → columns, first 1000 rows virtualized, search filters rows; CSV via `encoding/csv`. Binary preview: `image/*` inline `data:`, `application/pdf` download banner, else hex 4KB + download. Keeps deps minimal.

4. **History = exact replay, masked display.** Store the fully-resolved request (after inheritance/interpolation/cookie attach) for faithful replay: `history replay <id>` and History view Replay button do `Client.Send(storedReq)` with no re-interpolation. Display masks via `environments.MaskValues` (Authorization, Cookie, secret keys); DB stores plaintext (file is 0600). Search via FTS5, pagination 50/page, status filter 2xx/4xx/5xx. Retention and masking mirror `internal/environments` and `internal/auth`.

5. **`internal/history` behind `*sql.DB` + `internal/core` HistoryService; frontend via HistoryAdapter.** Shared by Desktop (Wails bindings `List/Show/Search/Replay/Clear`) and CLI (`reqly history list|show|replay|clear|search`). Same adapter pattern as request/auth/environment — host injects Wails adapter, browser fallback is read-only.

## Considered Options
- **Global `~/.config/reqly/history.db`** — rejected: leaks cross-workspace, complicates `workspace.Root` discovery; per-workspace matches `tokens.json` pattern.
- **`mattn/go-sqlite3` (CGO)** — rejected: breaks cross-platform Wails builds (§1.12); `modernc.org/sqlite` is pure-Go and swappable behind `*sql.DB`.
- **Mask on write** — rejected: would break faithful replay and hex preview; mask on display keeps DB exact while UI/CLI never leaks secrets.
- **Separate `cookies.db`** — rejected: one DB means one WAL/retention/permission model and atomic transactions with history.

## Consequences
- **Positive:** Closes last P0 desktop polish (history/replay + persistent cookies + Table + binary preview) with one SQLite seam, minimal new UI (one History view + one Table tab + cookie delete/clear), and consistent local-first privacy (0600, `.gitignore`, masked display).
- **Trade-off:** Exact replay ignores current `{{variables}}` drift (future `history replay --env <name>` needed for env-targeted replay); 1MB spill threshold means very large bodies live as blobs.

## Amendment (2026-08-24)

The query surface described above is now generated: `internal/history` compiles its SQL via **sqlc** into typed Go (ADR 0027). Storage engine, per-workspace location, WAL, FTS5 index, spill files, and retention are unchanged. GUI History search moved to Fuse.js client-side; CLI search keeps FTS5 behind a sanitizer.
