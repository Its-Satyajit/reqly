# Spec: Desktop Import Dialog (GUI-5.1)

> **Status:** Shipped 2026-08-24 — grill settled 2026-08-24 (Q1–Q9 confirmed)
> **Scope:** `docs/internal/gui-roadmap.md` §5.1 items G-5.1.1–G-5.1.8 — the highest-impact missing GUI feature
> **Stack:** `internal/importer` (new `Detect`) + `apps/desktop/backend/app.go` (new `Import` bridge) + root `frontend/` (`features/import-dialog/`)
> **Builds on:** M42 import preservation — every file-format parser returns a structured `*ImportReport`

## Problem Statement

A desktop user with a Postman collection, HAR capture, cURL command, OpenAPI spec, Insomnia export, or Bruno collection must fall back to the CLI (`reqly import …`) to get their data into Reqly. The GUI — the primary surface for most developers — has no import path at all. When imports do degrade features (M42), the structured report is invisible in the GUI.

## Solution

An Import dialog, opened from the workspace UI: drop a file or paste a command, Reqly auto-detects the format (six supported: cURL, OpenAPI 3.x, HAR 1.2, Postman v2.1, Insomnia v4/v5, Bruno) with a manual override, shows a preview of what will be created together with the M42 degradation report grouped by category and severity, then writes it into the workspace on commit. cURL is the exception: it opens as an unsaved request tab instead of writing files. Conflicts fail fast; malformed input errors inline.

## User Stories

1. As a developer migrating from Postman, I drop my collection JSON onto the import dialog and see it detected automatically.
2. As a developer with mixed-format files, I override the detected format from a dropdown when detection guesses wrong.
3. As a developer debugging an API via Chrome DevTools, I import a captured HAR and preview every request before anything is written.
4. As a developer sharing a repro, I paste a cURL command and get an editable request tab without any file being written.
5. As an API consumer with an OpenAPI spec, I preview operations grouped by tag before importing hundreds of endpoints.
6. As a cautious importer, I see the target folder name before commit and can rename it if it collides.
7. As a careful maintainer, I read the degradation report — entries grouped by category with severity tallies — before committing an imperfect import.
8. As a user who pasted garbage, I get an inline error in the dialog and can fix my input without losing it.
9. As a user importing into a dirty workspace, a commit that would overwrite existing files fails with a clear message instead of silently merging.
10. As an Insomnia/Bruno user, my collections import through the same dialog with no format-specific UI to learn.

## Implementation Decisions

**Bridge (Q1–Q3):**
- One generic bridge method `AppService.Import(req)` where req carries content, filename, format hint, dry-run flag, and target dir; returns `ImportResult`. No per-format methods. Shipped amendment: content crosses as a string (imports are text-only), avoiding Wails []byte encoding ambiguity; an additional `Detect(content)` bridge method backs the dialog's live format badge (T5 acceptance criterion).
- New pure function `importer.Detect(data []byte) (Format, bool)` in core — content sniffing (JSON structure keys, YAML shape, cURL verb prefix); `FormatHint` overrides detection when valid.
- `ImportReport` crosses the Wails boundary as the core type (plain strings; bindings generate TS models automatically). No DTO layer.
- `ImportResult` carries: `Kind` ("workspace" | "request"), the `*ImportReport`, a title, item counts, target dir, parsed `Request` for cURL, and `[]openapi.Endpoint` for OpenAPI dry-runs.

**Write semantics (Q2, Q6, Q7):**
- `DryRun=true` parses and returns preview + report without touching the filesystem; the same code path serves preview and commit.
- File formats write a workspace folder named `SanitizeDirName(result.Title)` at workspace root — CLI parity. `TargetDir` from the editable preview field names the folder at commit.
- Commit fails fast when the target directory already exists — never silent merge/overwrite (CLI's current `MkdirAll` behavior is deliberately not replicated).
- cURL bypasses the write path entirely: `Kind="request"`, dialog hands the parsed request to the request store as an unsaved tab; DryRun/commit do not apply.

**Dialog UX (Q4, Q8, Q9):**
- Single staged modal in `frontend/src/features/import-dialog/`: input stage (drop zone + paste textarea shared by files and commands) → preview stage → results stage, all in-modal.
- On input change, auto-detect runs and renders a detected-format badge; a select offers all six formats plus "auto" as override.
- Preview shows the item tree grouped by primary tag for OpenAPI (fallback "untagged"), capped at ~50 visible entries with "+N more"; other formats show counts + top-level structure. Report renders beneath, grouped by category with severity tallies (G-5.1.8 component, reused at preview and results).
- Parse/detection failures render as an inline error banner inside the modal; input stays editable; badge flips to "unknown". Parse failures never surface as toasts.
- Dialog shell and form controls reuse the existing ui primitives; no new component library.

**Bindings:** generated locally via `wails3 generate bindings -ts -i` (the bindings directory is gitignored by policy) after any AppService change; generated output lands under the desktop shell's bindings directory consumed by the frontend bridge module.

## Testing Decisions

- **Seam 1 — `importer.Detect`:** table-driven tests over real fixture bytes for each of the six formats plus ambiguous/unknown inputs (existing fixtures under `internal/importer/testdata/` are reused).
- **Seam 2 — `AppService.Import`:** Go tests against a temp workspace covering per-format dry-run shapes, report passthrough, hint override, unknown-format error, conflict fail-fast on commit, actual workspace mutation on commit, and cURL kind dispatch. This seam carries all behavioral verification.
- **Frontend:** verified by `npm run typecheck` + lint gates (no frontend test runner exists; none introduced). Component correctness rests on Seam 2 semantics plus manual/preview checks during `/code-review`.
- Prior art: existing AppService tests and importer table-driven suites.

## Out of Scope

- WSDL import in the dialog (core support exists; roadmap lists six formats — adding it later is one dispatch line once Detect knows it).
- Parent-folder picker for import destination (fixed workspace-root target in v1).
- Export dialogs (§5.2, separate milestone).
- Drag-and-drop reordering, partial selection of previewed items, environment-only import toggles.
- Frontend component test infrastructure.

## Further Notes

- `apps/desktop/backend/frontend/` is a stale stub left over from the b19d625 frontend move; AGENTS.md §5 still points there. Cleanup candidate — not part of this feature.
- ADR judgment: the generic-Import bridge decision sets precedent for future AppService surface (export dialogs follow the same shape). Recorded here rather than as an ADR — reversible, unsurprising given the existing thin-service convention.
