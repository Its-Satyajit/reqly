# 02 — CLI subcommand (`reqly import wsdl`)

**Blocked by:** 01

**Status:** done

- [x] `import wsdl <file> [--output dir]` wired into apps/cli/cmd/import.go with directory guidance output
- [x] Warnings to stderr, exit 0; e2e test against vendored fixture
- [x] go vet/gofmt/go test green; go build -o reqly ./apps/cli
