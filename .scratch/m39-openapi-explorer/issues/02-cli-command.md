# 02 — CLI commands (`reqly openapi explore|generate`)

**Blocked by:** 01

**Status:** done

- [x] `apps/cli/cmd/openapi.go`: `explore <spec> [--tag]... [--json]` table + JSON output
- [x] `generate <spec> [--operation]... [--method --path] [--tag]... [--all] [--output dir]`; no-selector error lists operations
- [x] Warnings to stderr, exit 0; generate exits 1 only when zero files written
- [x] e2e tests against import-suite/openapi fixtures; help text with examples
- [x] go vet/gofmt/go test green; `go build -o reqly ./apps/cli`
