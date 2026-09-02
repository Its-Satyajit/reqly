# Spec: Binary + GraphQL Body Editors (Milestone 21)

> **Status:** Draft — grill settled 2026-08-21
> **Scope:** Finish P0 body gap `ROADMAP.md:131` (binary + GraphQL) + `ROADMAP.md:74` file upload, `docs/features.md:8`
> **Stack:** `internal/request` + `internal/requestfile` + `frontend/src/lib/body.ts` + `frontend/src/editors/CodeMirrorEditor.tsx` + `frontend/src/features/request-editor/RequestEditor.tsx` + `apps/desktop/frontend/src/bridge.ts`
> **Predecessor:** M20 `docs/adr/0012-aws-edgegrid-auth.md`, M14 `Body Type` + `serializeBody`

## Problem Statement

`ROADMAP.md:131` lists body editors as `~` with JSON/XML/raw/form-data/urlencoded done, **binary + GraphQL pending**; `ROADMAP.md:74` file upload/multipart is unchecked. `docs/features.md:8` lists Binary/File upload/GraphQL/Multipart as supported types, but desktop can’t send a file or a GraphQL query without hand-editing the request file. `frontend/src/lib/body.ts:6` `BodyType` is `none|json|xml|form-data|urlencoded|raw`; `CodeMirrorEditor.tsx:1` supports `json/js/xml/yaml/markdown/text` but not `graphql`. `form-data` already does `multipart/form-data` with `boundaryFor` (`body.ts:32`) but its rows can’t carry a file.

## Solution

Extend `BodyType` to `...|binary|graphql` and make `form-data` file-aware:

* **`binary`** — single file picker (Wails file dialog + drag-drop, preview name/size/type), stores `request.body: { file: "./relative/path" }` Git-native, `serializeBody` reads file bytes at send, `contentTypeFor` → file mime or `application/octet-stream`.
* **`graphql`** — two editors: query (CodeMirror `graphql` lang) + variables JSON (CodeMirror `json`, optional `{}`), stored as structured `body: { query: "query { ... }", variables: { "id": 1 } }`, `serializeBody` → `JSON.stringify({ query, variables })` with `application/json`.
* **`form-data` file rows** — each key-value row gains `file?: "./path"` + `filename?` (file picker per row), `serializeBody` builds `multipart/form-data` with `boundaryFor`, file parts read at send.

All file paths are relative to the request file’s collection dir, Git-native, resolved at `RequestService.Send` time. Desktop shows file inputs; CLI resolves same `requestfile` shape.

## User Stories

1. As a desktop user, I want a **Binary** body type with a file picker, so I can send an image/PDF without editing the file.
2. As a desktop user, I want **GraphQL** body with query + variables editors, so I can send `{"query","variables"}` without hand-JSON.
3. As a desktop user, I want `form-data` rows to attach files, so I can send `multipart/form-data` with mixed fields + files.
4. As a CLI user, I want `requestfile` `body.file` / `body.query` to work via `reqly run`, with missing file or invalid JSON variables failing fast.

## Body Type & Requestfile

**BodyType** (`frontend/src/lib/body.ts:6`):
```ts
export type BodyType = 'none'|'json'|'xml'|'form-data'|'urlencoded'|'raw'|'binary'|'graphql'
```

**requestfile** (`internal/requestfile`):
```yaml
# binary
request:
  body: { file: "./data/image.png" }           # BodyType binary
# graphql
request:
  body: { query: "query GetUser($id:ID!){user(id:$id){name}}", variables: { id: 1 } }
# form-data with file
request:
  body:
    - { key: "avatar", file: "./avatar.png", filename: "profile.png" }
    - { key: "name", value: "Alice" }
```

**contentTypeFor** (`body.ts:26`): `binary` → mime from file ext or `application/octet-stream`, `graphql` → `application/json`, `form-data` → `multipart/form-data; boundary=...` (existing).

## Desktop UX

* **Body tab picker** — adds `Binary` and `GraphQL` to `BodyType` select (`RequestEditor.tsx`), `form-data` rows get file toggle (text ↔ file picker, `filename` override).
* **Binary** — file input (Wails `OpenFileDialog` via `bridge.ts` seam or `<input type=file>` + drag-drop), shows file name/size/type, stores relative path in draft, `authWarnings`-style save warnings if file missing.
* **GraphQL** — `CodeMirrorEditor` with `graphql` lang (add `@codemirror/lang-graphql` or `lang-graphql` stub), query editor + variables JSON editor below, variables validated as JSON on save (non-blocking warning if invalid).
* **CodeMirror** — extend `frontend/src/editors/CodeMirrorEditor.tsx:20` `languageExtensions` with `graphql: graphql()` (`satisfies Record<EditorLanguage,...>`).

## Validation

* **Save**: non-blocking warnings — `binary` missing file → “File is required for Binary body”, `graphql` variables invalid JSON → “Variables must be valid JSON”, `form-data` file row missing file → warning. Reuses `RequestEditor` save-warnings banner (M19 `authWarnings` pattern).
* **Send**: `internal/request` `Apply`/`Send` validates — file not found → error surfaced in `ResponseViewer`/toast, `graphql` variables not JSON → error. File read at send time, not at save.

## Out of Scope

* File upload progress, chunked transfer, streaming, drag-drop for non-binary types, GraphQL introspection/autocomplete/schema browser (`ROADMAP.md:150` P0 GraphQL protocol is separate), binary preview, `multipart` as distinct type (keep `form-data`).

## Verification

* `internal/request` + `requestfile` tests: `BodyType` binary/graphql enums, `serializeBody` binary file read + mime, graphql JSON wrap, form-data file multipart with `boundaryFor`, file-not-found and invalid JSON errors.
* `frontend/src/lib/body.test.ts` (new) — `contentTypeFor` binary/graphql, `serializeBody` table-driven.
* `go test ./...` + `go test -race`, `npm run lint` 0, `npm run typecheck` both frontends, `wails3 task build` from `apps/desktop/backend`, manual drag-drop + file dialog smoke.

## Docs

* New `docs/adr/0013-binary-graphql-body.md` (why extend `form-data` not new `multipart`, why structured `body.query` not stringified, why file-path not base64).
* Update `CONTEXT.md` glossary: `Binary Body`, `GraphQL Body`, `File Upload`.
* Tick `ROADMAP.md:74` + `131` + `docs/features.md:8` + Progress Tracker Phase 1 %.

## Ticket Split (5)

* T1 Core `internal/request` + `requestfile` — BodyType, file handling, graphql wrap, tests
* T2 Bridge `apps/desktop/frontend/src/bridge.ts` + `frontend/src/lib/body.ts` + `BodyType` union, file picker seam
* T3 Body tab `RequestEditor.tsx` + `CodeMirrorEditor.tsx` (graphql lang) — binary file input, graphql editors, form-data file toggle
* T4 Validation + warnings `RequestEditor.tsx` + `internal/request` errors
* T5 Docs ADR + CONTEXT + ROADMAP
