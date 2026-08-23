# 02 — Workspace writer with folders (internal/importer)

**What to build:** `PostmanResult.Write(dir)` writes the parsed tree as a Git-native workspace: `<dir>/reqly.yaml`, `collections/<name>/reqly.yaml` (+ collection variables), nested folder dirs each with a `reqly.yaml` name descriptor, request files as YAML via `yaml.Marshal` (0600/0644 modes matching the other importers).

**Blocked by:** 01

**Status:** done

- [x] Folder recursion preserves nesting; sanitized + deduplicated dir names per folder scope
- [x] Request filenames sanitized/deduped within their containing folder
- [x] Collection variables land in the collection descriptor's `variables:` map
- [x] Test: write → reload with `collections.LoadWorkspace` → tree shape, vars, and requests round-trip
- [x] `go test ./internal/importer ./internal/collections` green
