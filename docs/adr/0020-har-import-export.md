# ADR 0020: HAR Import / Export + Replay (M28)

## Status
Accepted

## Context
`ROADMAP.md:164` (`Import: Postman, Insomnia, OpenAPI, Swagger, cURL, WSDL, HAR`) and `ROADMAP.md:212` (`HAR import/export + replay`) are the last P1 import gap alongside browser integrations: `docs/features.md:37` (HAR Support — capture/debug/replay) and `docs/features.md:38` (Browser Integration — DevTools `Copy as HAR`). `internal/importer` has `curl.go`/`openapi.go` (flat `har-import` vs per-spec collection) and `internal/exporter` has `postman.go`/`code.go` (`Generate` pure function + `postman` workspace export) and `internal/history` (`Store` WAL + `Entry` + `blobs/` spill). The design questions are: flat `har-import` collection vs `pageref` folders, history→HAR vs workspace→HAR export source, fidelity (base64 `postData`/`content.encoding`, binary spill, timings), and CLI shape (`import har`/`export har` vs `history export`).

## Decision
1. **Import HAR → flat `har-import` collection, file-per-entry, `pageref` deferred.** `reqly import har <har-file> [--out <dir>] [--collection <name>]` (`--collection har-import` default, `validName` sanitized) writes `collections/<name>/<method>-<host>-<path>.yaml` (`RequestEntry` per `log.entries[]`), deduped `get-users`, `get-users-2`. Maps `request.headers[]` + `cookies[]` → `request.Headers` (`Cookie:` merged, duplicates kept), `queryString[]` → `request.Query`, `postData.text` → `request.Body` (base64 `encoding` decoded first), `postData.mimeType` → `Content-Type` header only when no explicit header wins (like `frontend/src/lib/body.ts:32`). Drops `pageref`/`timings`/`cache`/`_resourceType` with `unsupported-feature` warning (like `importer/curl.go`). Bodies >1MB spill to `blobs/<id>.bin` via `request.body: {file:"./blobs/..."}` reuse of `history` spill seam, not inlined. `M28` no `_pageref` folders; `M28b` adds grouping without breaking `RequestPath` identity.

2. **Export history→HAR via `internal/exporter/har.go` (`Export([]history.Entry) ([]byte,error)`), workspace→HAR deferred.** `reqly export har [--out <file.har>] [--env <name>] [--limit 500]` serializes `history.Store` entries (filtered by `env` partition, `EnforceRetention` 500) into HAR `log.entries[]` (`request` from `ReqHeaders/ReqBody` exact bytes, `response` from `RespHeaders/RespBody`, `content.text` base64 when binary, `timings` synthesized from `DurationMS`/`send`/`wait`/`receive`). Secrets masked to `[SECRET]` via `environments.MaskValues` (never write raw tokens). `M28` no `workspace→HAR` (request files have no `response` → incomplete HAR); `M28b` adds `workspace export har`. Replay = `import har` then `collection run` plus existing `history replay <id>` verbatim `Client.Send` (`CONTEXT.md:154`).

3. **Fidelity lossy-with-warnings, `0400`+`base64`, spill, synthesized timings, masked.** Handles `postData.text`+`encoding`, `content.encoding`, duplicates, `Cookie:` merge; warns on dropped fields; atomic `WriteFile` `0644` (like `requestfile.Save`), `0600` for `tokens.json`/`history.db` unchanged.

4. **CLI shape `import har` + `export har`, stdout default, one-file-per-command, `GetUsers` dedupe, `ValidName` sanitisation; no `har` root command and no `history export` alias for `M28`.** Reuses `apps/cli/cmd/import.go`/`export.go` pattern (`import curl/openapi`, `export postman/code/workspace`) and `SaveWorkspace` naming; `apps/desktop` desktop drag-drop deferred to `M28b`.

## Considered Options
- **`pageref` → `Folder` grouping for `M28`** — rejected: HAR `pageref` semantics are browser-page, not Git-native container; `workspace.go:214` descriptor-less dirs ignored, so invented folders break discovery. Deferred to `M28b` with explicit opt-in.
- **`workspace→HAR` as `M28` primary, history→HAR deferred** — rejected: `requestfile.File` has no `response`/`timings` → HAR `log.entries[]` would be half-empty and overlaps `export postman`. History is the only source with full `request`+`response`.
- **Strict verbatim HAR (preserve `timings`, `cache`, `serverIPAddress`) via `requestfile` extensions** — rejected: pollutes `RequestEntry` with HAR-only keys and breaks Git-diffability. Kept `M28` plain-text with base64-for-binary only.
- **`har` root command (`reqly har import/export`) or `history export har`** — rejected: fragments `cmd/*.go` one-file rule and hides export under `history` (hurts `docs/reference/go.md` discoverability). Kept `import`/`export` roots.

## Consequences
- **Positive:** Closes `ROADMAP.md:212` with one exporter (`har.go` beside `postman.go`/`code.go`), one importer (`har.go` beside `curl.go`/`openapi.go`), reusable `SaveWorkspace`+`requestfile.Save` seams, `history` spill reuse, and `unsupported-feature` parity with `curl`/`openapi`.
- **Trade-off:** `pageref` folders, workspace→HAR, desktop drag-drop, and `_resourceType` filtering are `M28b` — `M28` is flat `har-import` collection + history→HAR only.
