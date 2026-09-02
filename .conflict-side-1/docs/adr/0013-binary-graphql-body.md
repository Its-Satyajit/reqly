# ADR 0013: Binary and GraphQL Body Editors

## Status
Accepted

## Context
`ROADMAP.md:74` (file upload) and `ROADMAP.md:131` (body editors) show `form-data`/`json`/`xml`/`urlencoded`/`raw` done, `binary` + `GraphQL` pending. `frontend/src/lib/body.ts:6` `BodyType` and `frontend/src/features/request-editor/RequestEditor.tsx:34` `bodyLanguage` map only those 5, and `frontend/src/editors/CodeMirrorEditor.tsx:20` supports `json`/`xml`/`yaml` but not `graphql`. `form-data` rows (`frontend/src/lib/request.ts:25` `KeyValueRow`) can’t carry a file, so `multipart/form-data` with files and single-file `binary` bodies are impossible without hand-editing the request file. The question is whether to add new `BodyType`s, how file paths should be stored Git-natively, and whether `multipart` should be a distinct type.

## Decision
1. **Extend `BodyType` with `binary` and `graphql`.** `frontend/src/lib/body.ts:6` → `none|json|xml|form-data|urlencoded|raw|binary|graphql`; `bodyTypes` adds `Binary` and `GraphQL` labels. No new `multipart` type — `form-data` stays as `multipart/form-data` with `boundaryFor` (`body.ts:32`), extended to file-aware rows.
2. **File paths are Git-native relative strings.** `binary` stores `request.body: { file: "./relative/path" }` (relative to collection dir) in `requestfile` YAML/JSON, resolved at `RequestService.Send` time. `form-data` rows gain `file?: "./path"` + `filename?` (file picker per row). Desktop shows file input + drag-drop + preview (name/size/type, `filename` override); CLI resolves same `requestfile` shape via `os.ReadFile`. No base64 inline — file bytes are read at send, not at save.
3. **GraphQL is structured, not stringified.** `graphql` stores `request.body: { query: "...", variables: { ... } }` in the file, edited as two editors (query `graphql` lang + variables `json`), `serializeBody` (`frontend/src/lib/body.ts:71`) does `JSON.stringify({ query, variables })` with `Content-Type: application/json` (`contentTypeFor` `body.ts:26` → `graphql: application/json`, `binary: application/octet-stream`). `RequestInput` gains `graphqlQuery`/`graphqlVariables` for the draft → `serializeBody` bridge.
4. **CodeMirror adds `graphql`.** `frontend/src/editors/CodeMirrorEditor.tsx:20` `languageExtensions` gains `graphql: graphql()` (`satisfies Record<EditorLanguage,...>`), `bodyLanguage` maps `binary|graphql` to `text`/`graphql`.
5. **Non-blocking warnings, hard send errors.** Like ADR 0011/0012: missing file or invalid JSON variables → `saveWarnings` (`RequestEditor.tsx:43`) banner (yellow, non-blocking); file-not-found or invalid JSON at send → `internal/request` error surfaced in `ResponseViewer`/toast.

## Considered Options
- **New `multipart` BodyType distinct from `form-data`** — rejected: `form-data` already is `multipart/form-data` with `boundaryFor`; adding `multipart` would duplicate and confuse the picker.
- **Base64 inline file bytes in `request.body`** — rejected: Git-native file path is diff-friendly and keeps files out of the YAML; base64 would bloat the file and hide the file reference.
- **Single raw `graphql` string (query + variables concatenated)** — rejected: structured `body.query`/`body.variables` keeps query and variables separately editable and matches the `requestfile` JSON shape that `serializeBody` can JSON-stringify.

## Consequences
- **Positive:** P0 body completes to 100% (binary/file upload + GraphQL + file-aware multipart) with minimal new UI (2 picker entries + file inputs), Git-native file paths, and `multipart/form-data` reuse.
- **Trade-off:** `binary`/`graphql` bodies that reference a file outside the workspace or with absolute paths will fail at send (file-not-found); variables JSON errors are only warnings at save, hard errors at send.
