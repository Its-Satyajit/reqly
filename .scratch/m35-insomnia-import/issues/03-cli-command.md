# 03 — CLI `reqly import insomnia`

**What to build:** `reqly import insomnia <file> [--output <dir>]`; auto-detect v4/v5 by content; default output = sanitized workspace/collection name; warnings on stderr.

**Blocked by:** 02

**Status:** done

- [x] Cobra subcommand wired following postman subcommand shape
- [x] End-to-end test: import insomnia-v5.yaml fixture into t.TempDir, LoadWorkspace asserts
- [x] go build -o reqly ./apps/cli green
