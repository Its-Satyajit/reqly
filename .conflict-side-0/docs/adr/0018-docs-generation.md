# ADR 0018: Docs Generation (M26)

## Status
Accepted

## Context
`ROADMAP.md:186` (`reqly docs`) is the last P0 docs gap: `internal/collections` + `internal/requestfile` + `internal/exporter` already exist (Postman export, code generation, history), but there is no `docs generate` that turns a `Workspace` into Markdown documentation (collections list + per-request method/URL/headers/query/body/auth + `cURL` example). The design questions are which format ships in M26 (Markdown vs HTML), which content (collections only vs openapi/history), which template, and where the feature is exposed (CLI vs desktop).

## Decision
1. **Markdown per collection + `index.md` via `text/template`, `internal/docs` package.** `internal/docs.Generate(outDir string, ws *Workspace, env string) error` (new `internal/docs` package, beside `exporter`, pure function) writes `<out>/index.md` (collections list) + `<out>/<coll>.md` per collection via `text/template` (method/URL/headers/query/body/auth + `cURL` block via `exporter.Generate` `curl` with `[SECRET]` masked via `environments.MaskValues` + `auth.MaskValues`). Shows raw `{{var}}` as in file plus resolved `cURL` example (when `--env` set via `environments.ResolveSet`, otherwise raw). `openapi`/`history`/`GraphQL` not in M26 (M26b adds `goldmark` HTML + `openapi` endpoints).
2. **`internal/docs` is the highest seam (beside `exporter`), not `internal/exporter/docs.go`.** Keeps `exporter` focused on code/postman, `docs` is single purpose: `LoadWorkspace` + `flattenWorkspace` + `template` + `exporter.Generate` for `curl`, `os.MkdirAll` + atomic `WriteFile` (`0644`). CLI `docs generate` is thin wrapper (`LoadWorkspace(src)` → `Generate(out, ws, env)`); desktop UI is M26b (no `AppService.DocsGenerate` for M26).
3. **CLI `reqly docs generate [src] --out <dir> [--env <name>]` (like `export workspace`).** `src` defaults to `.`, `--out` required, `env` optional (like `run`, when set resolves `{{var}}` for the `cURL` block via `environments.ResolveSet`). No `--format` in M26 (Markdown only, HTML is M26b), no `--serve`/`--watch`.

## Considered Options
- **HTML via `goldmark`/`templ` in M26** — rejected: Markdown `text/template` is `0` deps and `0` `goldmark` for M26; HTML is M26b.
- **`internal/exporter/docs.go` (add to exporter)** — rejected: `exporter` already owns code/postman (sharing request shapes); `docs` is a distinct product (Markdown generation), keeping blast radius to one package `docs`.
- **Content `openapi` + `history` examples in M26** — rejected: `openapi` spec endpoints and `history` last response examples are P1 follow-ups; M26 is `collections` + `requestfile` only (like `export postman` flat list).
- **Desktop `DocsView` + `AppService.DocsGenerate` in M26** — rejected: CLI `generate --out` covers CI/docs generation; desktop UI is M26b (same `Generate` can be exposed later via `HistoryAdapter` pattern).

## Consequences
- **Positive:** One `internal/docs.Generate` closes P0 docs with one file + `text/template` + `exporter` for `curl`, `Workspace` + `requestfile` only, `CLI` `generate --out` mirrors `export workspace` (`--env`), `0600` not needed, `0644`, and `M26b` path for HTML/`openapi`/`history` without new seams.
- **Trade-off:** M26 docs are Markdown only, `collections` only, no `openapi`/`history`/`GraphQL`, no desktop UI, no `--format` — all are M26b follow-ups.

