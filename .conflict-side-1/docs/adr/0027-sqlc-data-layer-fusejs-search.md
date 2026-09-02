# ADR 0027: sqlc for the Data Layer, Fuse.js for GUI Search

## Status
Accepted (2026-08-24)

## Context

ADR 0014 put history and the cookie jar in per-workspace SQLite (`modernc.org/sqlite`, WAL, FTS5) behind `*sql.DB`, with hand-written inline SQL in `internal/history`. Two frictions surfaced:

1. **Query/typo drift.** The big SELECT column list is duplicated across four call sites; a new column means editing every instance by hand. Nothing ties the Go mapping to the schema.
2. **FTS5 syntax leaks.** User search text flowed straight into `MATCH ?`. Innocent input like `todo/` produced "SQL logic error: fts5: syntax error" in the desktop History view — query syntax is a programmer interface, not a user interface.

Separately, the GUI needed search semantics FTS5 doesn't naturally provide (punctuation tolerance, typo tolerance, weighted fields).

## Decision

**Data layer — sqlc over `database/sql` + `modernc.org/sqlite` (no CGO, unchanged).**

- SQL lives in reviewed files (`internal/history/db/schema.sql`, `db/query.sql`) and compiles via `sqlc generate` into typed Go (`internal/history/db`). Queries are checked against the schema at generation time.
- The generated package is aliased `sqlcdb` at import sites so the handle variable `db *sql.DB` never shadows it.
- Datetime columns are declared `TEXT NOT NULL` in the generation schema and keep the RFC3339Nano wire format — byte-compatible with databases written before this change.
- Runtime startup migrations (`CREATE TABLE IF NOT EXISTS` + legacy column adds) remain the source of truth for physical schemas; `schema.sql` is the generation snapshot kept in sync.
- FTS5 MATCH uses per-column predicates (`url MATCH ? OR request_path MATCH ?`) because sqlc cannot parse bare table-name match syntax; combined with the `ftsQuery` sanitizer (quoted prefix phrases per token, punctuation-only tokens dropped), CLI search accepts arbitrary user text safely.
- DDL, PRAGMA inspection, and the legacy-column bridge stay hand-written — sqlc covers queries, not migrations.

**GUI search — Fuse.js (frontend).**

- Desktop History search runs Fuse.js over a recent-entry pool (500) client-side: punctuation-safe, typo-tolerant, weighted (url 0.5 / requestPath 0.4 / method 0.1). The GUI no longer issues FTS5 queries at all.
- The core `Search` API remains for the CLI, guarded by the sanitizer.

## Alternatives Considered

- **Keep hand-written inline SQL** — rejected: it caused both frictions above; nothing validates column lists or query syntax until runtime.
- **ORM (GORM/ent)** — rejected: model magic and reflection contradict the minimalist `database/sql` stack; SQL-first keeps the FTS5 specifics visible.
- **Fuse.js-equivalent fuzzy search in Go core** — rejected: adds a dependency to core for a purely presentation-layer concern, and the GUI already owns result presentation.

## Consequences

- **Positive:** compile-time-checked queries, single source per statement, safe user-facing search on both surfaces, zero CGO regression, byte-compatible storage.
- **Negative:** one more codegen step (`sqlc generate`; tool pinned at v1.27.0), and two sources of schema truth (runtime migrations vs `schema.sql`) that must be kept in sync manually.
- **Neutral:** generated code is committed; regeneration is deterministic from the SQL files.
