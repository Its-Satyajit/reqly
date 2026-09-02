# Spec: OpenAPI Editor MVP (Milestone 45 — §33 slice)

> **Status:** Draft — backfilled 2026-08-27 from grill Q1–Q4 (recommended a/b/a/a)
> **Scope:** `ROADMAP.md:513` §33 — in-app spec authoring (M39b deferred explorer table is M39, CLI explore/generate shipped `internal/openapi/explore.go:64` + `generate.go:385`)
> **Grill:** Q1 Editor MVP (tree+YAML+validation+generate) / Q2 arbitrary file + workspace `openapi.yaml` convention / Q3 split tree+editor+preview, live `openapi.Load` debounced 400ms, gutter + Problems panel / Q4 tree checkboxes → `Generate` → `./requests` with `-2` dedup
> **Stack:** `frontend/src/features/spec-editor/SpecEditorView.tsx:1`, `frontend/src/lib/specTree.ts:1`, `frontend/src/stores/useSpecEditorStore.ts:1`, `apps/desktop/backend/spec.go:1` + `openapiexplorer.go:57`, `internal/openapi`

## Problem
`reqly import openapi` is whole-spec import; `reqly openapi explore/generate` is CLI-only. No in-app authoring with live validation and selective request generation. Existing `SpecEditorView` is static (`SPEC_SECTIONS` hardcoded, no file I/O, no diagnostics).

## Solution
- **Editor:** CodeMirror YAML (`yaml` lang) bound to `useSpecEditorStore.content`; `setContent` → `diagnosticsForSpec` (missing `openapi`/`info.title`/`paths`, yaml parse error) + `nodesForContent` derives `paths:*` children via indentation-aware parse (no new dep, reuses `internal/openapi.Load` via debounced Wails `SpecValidate` in follow-up).
- **Tree:** `nodesForContent(content)` — `Info`, `Servers`, `Paths` (dynamic children from `paths:`), `Components` (`schemas`/`security` when present). Selection `selectedId`, per-path checkbox `selectedOps:Set<string>` for generate.
- **File:** `filePath` (default `openapi.yaml`, workspace-relative via `resolveTestPath` `apps/desktop/backend/test.go:214`, `0600/0644` atomic `SpecRead`/`SpecSave` `apps/desktop/backend/spec.go:1`). Browser dev fallback: `<input type=file>` → `file.text()` → `loadContent`.
- **Generate:** Header **Generate (n)** disabled when 0 → `OpenapiGenerateRequests(specPath, selections, "generated")` `openapiexplorer.go:140` (reuses `Generate` `operationFilename` dedup, `{{baseUrl}}`, param/body resolution, bearer/basic/apikey-header → native `Auth` blocks). Warnings → `generateWarnings` in Problems panel.
- **Problems:** bottom panel merges `diagnostics` (`error`/`warning`) + `generate: …` lines; issue badge count.

## Data Model
`SpecNode {id,label,children?}` `SpecDiagnostic {message,severity}` store fields: `content,filePath,selectedId,dirty,diagnostics,selectedOps,generateWarnings`.

## API Surface
Frontend store `setContent/setSelected/setFilePath/loadContent/toggleOp/setGenerateWarnings/markSaved`; backend `SpecRead(path)→string`, `SpecSave(path,content)`, `OpenapiGenerateRequests`.

## Edge Cases
Empty spec → `spec is empty` error; invalid yaml → `yaml: …` error, tree falls back to static sections; missing `openapi` → error, missing `paths` → warning; outside-workspace path → `outside workspace` error; bridge unavailable (browser dev) → `bridge unavailable` warning, no throw.

## Testing Strategy
Pure lib `specTree.test.ts` (nodesForContent, diagnosticsForSpec, flatten), store `useSpecEditorStore.test.ts`, go `internal/openapi` existing explore/generate tests, manual Wails `SpecRead/Save` + `OpenapiGenerateRequests` round-trip.
