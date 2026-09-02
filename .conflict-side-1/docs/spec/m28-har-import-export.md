# Spec: HAR Import / Export + Replay (Milestone 28)

> **Status:** Draft — grill settled 2026-08-23 (Q1–Q4)
> **Scope:** P1 `ROADMAP.md:212` + `ROADMAP.md:343` — HAR 1.2 import → workspace + history → HAR export + replay via `history` (5 tickets)
> **Stack:** `internal/importer/har.go` + `internal/exporter/har.go` + `internal/history` + `apps/cli/cmd/import.go`/`export.go` + `internal/collections.SaveWorkspace`
> **Predecessor:** M27 `docs/adr/0019-cross-platform-desktop.md` (P0 100%), M28 `docs/adr/0020-har-import-export.md` (grill Q1–Q4)

## Problem Statement

`ROADMAP.md:212` (`HAR import/export + replay`) and `docs/features.md:37` (HAR Support) remain unchecked alongside `docs/features.md:38` (Browser Integrations). Users cannot import `HAR` captured from Chrome DevTools `Network → Export HAR` into a Git-native workspace, nor export a workspace's captured traffic (history) as HAR for sharing/debugging. Existing importers are `curl` + `openapi` (`internal/importer/curl.go`, `openapi.go`) and exporters are `postman` + `code` (`internal/exporter/postman.go`, `code.go`); history is SQLite `modernc.org/sqlite` WAL (`internal/history`). No HAR seam exists.

## Solution

Two seams, one per direction, reusing `SaveWorkspace` + `requestfile.Save` + `history.Store`:

* **Import HAR → workspace (`internal/importer/har.go`)**: `ParseHAR([]byte) (*Workspace, []string, error)` parses HAR JSON (`log.entries[]`), maps each entry's `request` to a `RequestEntry` file in a single flat collection `har-import` (default, `validName` sanitized, `--collection` override), deduped names `get-users`, `get-users-2`. Mapping: `headers[]+cookies[]→Headers` (`Cookie:` merged, duplicates kept), `queryString[]→Query`, `postData.text→Body` (base64 decoded when `encoding=="base64"`, `mimeType→Content-Type` only when no explicit header wins, like `frontend/src/lib/body.ts:32`), bodies >1MB spill to `blobs/<id>.bin` via `request.body: {file:"./blobs/..."}` (reuses `history` 1MB spill seam). Drops `pageref`/`timings`/`cache`/`_resourceType` with `unsupported-feature` warning (parity `importer/curl.go`).

* **Export history → HAR (`internal/exporter/har.go`)**: `Export([]history.Entry, HarOptions) ([]byte,error)` pure function beside `postman.go`/`code.go`, serializes `history.Store` entries (filtered by `env` partition + `--limit 500` `EnforceRetention`) into HAR `log.creator={name:"reqly", version, comment}` + `log.entries[]` (`request` from exact `ReqHeaders/ReqBody`, `response` from `RespHeaders/RespBody`, `response.content.text` base64 when binary, `timings` synthesized from `DurationMS`/`send`/`wait`/`receive`, `startedDateTime` from `created_at`). Secrets masked to `[SECRET]` via `environments.MaskValues`. Atomic `WriteFile` `0644`.

Replay = `import har` materializes traffic as `har-import` collection, then `reqly collection run har-import`; or `reqly history replay <id>` verbatim `Client.Send` (`CONTEXT.md:154`). No new `history replay --env` for M28.

## User Stories

1. As a user, I want `reqly import har capture.har` to create `./collections/har-import/<method>-<host>-<path>.yaml` per entry, so DevTools HAR becomes a Git-native workspace I can `git diff`/`collection run`.
2. As a user, I want `reqly import har capture.har --out ./ws --collection chrome` to control output dir/collection name with deduped sanitized filenames.
3. As a user, I want `reqly export har --env staging --limit 100 --out traffic.har` to share history (with responses) as HAR for bug reports, with secrets masked.
4. As a user, I want `reqly export har` (no `--out`) to print HAR JSON to stdout for `| jq` pipelines.
5. As a user, I want `postData.text` base64 + `response.content.encoding` + `Cookie:` merge + `queryString` map handled, with warnings for dropped HAR fields, so fidelity is lossy-with-warnings not silent.

## CLI Contract

```bash
reqly import har <har-file> [--out <dir>] [--collection <name>]
  # har-file: HAR JSON 1.2 (log.entries[]). --out default "." (cwd, creates collections/<name>/ + reqly.yaml if missing). --collection default "har-import" (validName, /\_\-. disallow "../").
  # Writes 0644 request files, no .reqly/history.db creation on import. Dup names: get-users, get-users-2. Exit 0 + stderr warnings for unsupported fields.

reqly export har [--out <file.har>] [--env <name>] [--limit 500]
  # history → HAR. --out absent → stdout (0644 when file). --env filters history partition (like HistoryList env), --limit caps entries (default 500, EnforceRetention). Secrets masked. Timings synthesized.
```

No `reqly har` root, no `history export har` alias for M28. `apps/cli/cmd/import.go` + `export.go` one-file-per-command (`AGENTS.md:16`).

## Mapping Details (Import)

| HAR `request` | Reqly `request.Request` |
|---|---|
| `method` | `Method` |
| `url` (full) | `URL` |
| `headers[] {name,value}` + `cookies[] {name,value}` | `Headers[]` (`Cookie: a=b; c=d` merged, duplicates kept) |
| `queryString[] {name,value}` | `Query[]` |
| `postData {mimeType,text,encoding,params}` | `Body` string (`text` base64-decoded when `encoding=="base64"`; `params` ignored with warning; `mimeType` → `Content-Type` header only if no explicit header) |
| `bodySize`/`headersSize` | ignored |

Responses discarded (no `requestfile.File` response field). `response` captured only via later `history` on run.

## Mapping Details (Export)

| `history.Entry` | HAR `entry` |
|---|---|
| `method`/`url` | `request.method`/`url` |
| `req_headers_json` | `request.headers[]` + `cookies[]` split from `Cookie:` |
| `req_body_path` (or inline) | `request.postData {mimeType,text,encoding}` (text base64 when binary detected via `Content-Type` non-`text/*`/`json`) |
| `status` | `response.status` |
| `resp_headers_json` | `response.headers[]` |
| `resp_body_path` | `response.content {text,encoding,mimeType,size}` |
| `duration_ms` | `timings {send,wait,receive}` synthesized (`wait=duration*0.8`, `send=receive=duration*0.1`, clamped) |
| `created_at` | `startedDateTime` ISO8601 |
| `env` | omitted (HAR has no env; filter via `--env` pre-serialisation) |

`log.creator = {name:"reqly", version: internal/version.Version, comment:"exported from history"}`.

## Validation

* **Import**: HAR JSON `encoding/json` unmarshal + `log.version` must be `"1.2"` (warn if other, still parse), `log.entries` must be `[]`; empty HAR → empty collection + warning (no error). Missing `request.url` → entry skipped + warning. Invalid base64 → entry error + skip. Secrets masked on warnings via `MaskValues`.
* **Export**: `history.Store` may be empty → `log.entries:[]` + empty HAR (valid). `--limit` must be `>0` (else error). `--env` missing env → no entries (not error).

## Out of Scope (M28b)

* `pageref` → `Folder` grouping or `host` grouping (M28 flat `har-import` only)
* Workspace → HAR export (request files have no response; M28 history→HAR only)
* Desktop drag-drop HAR onto Collections Browser + File Upload HAR (M28 CLI only)
* `_resourceType` filtering (e.g. only `xhr`/`fetch`), `cookies` separate `har` jar, `cache`/`serverIPAddress`/`connection`
* `history replay --env <name>` re-interpolation against HAR

## Verification

* `internal/importer/har_test.go`: table-driven `ParseHAR` (GET+headers+cookies+query, POST json base64, POST form params warning, empty HAR, invalid HAR, duplicate name dedupe, large body spill to `file:` ref, pageref dropped warning, secrets masked in warnings). Fixture `testdata/har/*.har` + `.golden` yaml for `RequestEntry`.
* `internal/exporter/har_test.go`: table-driven `Export` (single GET history entry → HAR json, binary response base64, filtered by env, limit, empty history, secrets masked to `[SECRET]`, timings synthesized, `log.creator`). Golden `testdata/har/*.har.golden`.
* `go test ./...` + `go test -race ./...`, `gofmt -l` clean, `nub run typecheck` both frontends, `nub run lint` exit 0.
* `go vet` clean, `modernc.org/sqlite` no CGO.

## Docs

* New `docs/adr/0020-har-import-export.md` (already accepted, Q1–Q4)
* Update `CONTEXT.md` `HAR`/`HAR Import`/`HAR Export`/`HAR Replay` (done)
* Tick `ROADMAP.md:212` (`HAR import/export + replay`) + `docs/features.md:37` + Progress Tracker Phase 2 % after ship.

## Ticket Split (5)

* T1 `internal/importer/har.go` + `har_test.go` + `testdata` (ParseHAR + mapping + warnings + spill)
* T2 `internal/exporter/har.go` + `har_test.go` + `testdata` (Export history→HAR + base64 + timings + mask)
* T3 CLI `apps/cli/cmd/import.go` (`import har`) + `apps/cli/cmd/export.go` (`export har`) + `validateName`/`dedupe`
* T4 Integration `go test ./internal/...` + `ROADMAP/CONTEXT` docs + `EnforceRetention` limit + `0644` atomic
* T5 Docs ADR 0020 + CONTEXT `HAR*` + ROADMAP docs + spec this file (grill closure)
