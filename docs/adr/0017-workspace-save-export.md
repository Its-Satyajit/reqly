# ADR 0017: Workspace Save & Export (M25)

## Status
Accepted

## Context
`ROADMAP.md:119` (save/export a workspace) is the last P0 file round-trip gap: `internal/collections` is read-only (load + inheritance, no write) and `internal/requestfile` is read via `LoadFile` + per-file `Save` (used by desktop `WorkspaceSaveRequest` with format-preserving `isJSONPath`, temp-file + rename, `changed-on-disk` version). The desktop already does per-file save, but there is no bulk `SaveWorkspace` for CLI or for copying a workspace to a new directory. The design questions are whether save is in-place vs export copy, whether deleted collections/requests are pruned, whether format is preserved, and whether bulk needs `changed-on-disk` version checks.

## Decision
1. **Both Save in-place + Export copy, same seam `internal/collections.SaveWorkspace`.** `SaveWorkspace(root string, ws *Workspace) error` writes `reqly.yaml` + `collections/<coll>/reqly.yaml` + `collections/<coll>/requests/*.yaml` (or `.json` per `isJSONPath`) via `requestfile.Save` (format-preserving, atomic, `0644`/`0600`) and writes `collections`/`environments` dirs as needed. It **prunes** `collections/<coll>` dirs/files that no longer exist in `ws` (walk `root/collections` and `os.RemoveAll` for deleted). No `changed-on-disk` version check for bulk (bulk is “write what’s in memory”, like `Export`); per-file `WorkspaceSaveRequest` keeps version check. `Export` is thin: `reqly export workspace [src] --out <dir>` → `LoadWorkspace(src)` → `SaveWorkspace(out, ws)` (copy, no `tar` in M25).
2. **`internal/collections` is the highest seam (beside `LoadWorkspace`).** No new `internal/exporter` code; `SaveWorkspace` reuses `requestfile.Save` for request files and `yaml.Marshal` for descriptors. Keeps blast radius to one file `collections/save.go` (plus `workspace_test.go` round-trip). Desktop bulk UI is M25b — desktop keeps per-file `WorkspaceSaveRequest` for M25.
3. **CLI `reqly export workspace [src] --out <dir>` (like `reqly run`).** `src` defaults to `.`, `--out` required (no `--out` → error; in-place bulk is `SaveWorkspace` directly, not via CLI). `--env` not needed for save (save writes raw descriptors, not resolved). No `tar.gz`/`zip`/`--dry-run`/`--diff` in M25 (M25b adds `archive/tar`).

## Considered Options
- **Only Export copy, no Save in-place** — rejected: desktop bulk “Save All” needs in-place save, and `SaveWorkspace` is the primitive `Export` reuses.
- **No prune (only upsert)** — rejected: deleted collections would linger on disk as orphaned dirs, diverging from Git-native `ws` truth.
- **New `internal/workspace` package** — rejected: `internal/collections` already owns `Workspace` + `LoadWorkspace`; adding a new package fragments the seam.
- **Version check for bulk** — rejected: bulk is “write what’s in memory” for export/copy; per-file version check already guards `WorkspaceSaveRequest` for interactive edits.

## Consequences
- **Positive:** One `SaveWorkspace` seam closes the P0 file round-trip for CLI and for future desktop bulk, with pruning and format-preserving atomic writes, and `export workspace` is a one-liner copy.
- **Trade-off:** M25 `tar.gz`/`Save All` UI/`--dry-run` are deferred to M25b — export is directory copy only, bulk has no `changed-on-disk` guard.

