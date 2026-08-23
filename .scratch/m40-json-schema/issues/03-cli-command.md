# 03 — CLI commands (`reqly schema validate|inspect|generate`)

**Blocked by:** 01, 02

**Status:** done

- [x] apps/cli/cmd/schema.go: three subcommands with flags per spec
- [x] validate exit codes (1 on violations), stdin support, --json violations
- [x] e2e tests: table/JSON outputs, stdin pipe simulation, exit codes
- [x] go vet/gofmt/go test green; go build -o reqly ./apps/cli
