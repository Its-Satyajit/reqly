# T1 — Dynamic tree + diagnostics (lib + store)

> **Spec:** `docs/spec/m45-openapi-editor-mvp.md`
> **Blocks:** none — first slice
> **Shippable:** `nodesForContent` + `diagnosticsForSpec` + Problems badge, no file I/O

- Pure lib `frontend/src/lib/specTree.ts:1` — `nodesForContent(content)`, `diagnosticsForSpec`, `SpecDiagnostic`
- Store `frontend/src/stores/useSpecEditorStore.ts:1` — `diagnostics`, `setContent→diagnostics`
- Tests `frontend/src/lib/specTree.test.ts:1` (9) — paths children, missing fields, flatten

**Done when:** `nub run typecheck` + `vitest specTree` green, `SpecEditorView` shows dynamic tree + Problems.

**Implemented:** 2026-08-27 (patch debt: migrate to `typeGuards` + `js-yaml` vs manual parse)
