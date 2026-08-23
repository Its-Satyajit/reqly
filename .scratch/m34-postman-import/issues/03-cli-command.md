# 03 — CLI `reqly import postman`

**What to build:** Subcommand on the existing import command: `reqly import postman <file> [--output <dir>]`; default output = sanitized collection name in CWD; warnings printed to stderr; success prints the workspace path + request count.

**Blocked by:** 02

**Status:** done

- [x] Cobra subcommand wired in `apps/cli/cmd/import.go` following the har subcommand shape
- [x] Warnings surfaced (`--quiet` not needed for MVP)
- [x] Test: end-to-end import of a fixture collection into t.TempDir, assert workspace loads via LoadWorkspace and first request matches source
- [x] `go build -o reqly ./apps/cli` green
