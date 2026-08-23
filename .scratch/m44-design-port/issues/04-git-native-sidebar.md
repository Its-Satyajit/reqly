# M44 T4 — Git-native collections sidebar

Blocked by: T2.
Blocks: T7.

## Goal

Source-control state lives inside the collections sidebar (the demo's defining trait): file status markers on requests/folders, branch chip, workspace pill, staged/unstaged grouping of pending collection changes.

## Requirements

- Status glyphs (modified/added/deleted/conflicted) inline per tree node, driven by existing git bindings.
- Pending-changes strip inside the sidebar: stage/unstage, commit message, commit action.
- Workspace pill + environment pill in the shell header (T2 slots) wired to real stores.
- Loading, empty-repo, error and conflict states as simulated in the demo.

## Acceptance criteria

- [x] Editing a request file updates its sidebar status without manual refresh. — 4s poll + window-focus refresh; fs-event push deferred to T7 polish. 2026-08-25
- [x] Commit flow works end-to-end against a real local repo. — internal/git + bridge integration tests (real temp repos); UI strip with stage/unstage toggle + commit box. 2026-08-25
- [x] All states token-driven; both themes verified. 2026-08-25
- [x] typecheck + lint + React Doctor clean vs baseline; vitest 32/32; go test ./... all pkgs ok. Conflict glyphs render verbatim (UU) with error tone — full resolver is T7. — 2026-08-25

## Reference

- `shared/data.js` → `gitState`, workspace fixtures
- `shared/engine.js` → sidebar mount, `commitStripHtml`, branch-menu actions
