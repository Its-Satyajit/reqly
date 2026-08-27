# T2 — RequestFile + CLI overrides

> **Spec:** `docs/spec/m47-proxy-tls-per-request.md`
> **Blocks:** T1

- `internal/requestfile` round-trip `proxy` + `tls` (atomic save, `0644`)
- CLI `reqly run/test --proxy --insecure --ca-file` overrides file (precedence like `--env`)

**Done when:** `go test ./internal/requestfile` + manual `reqly run --help` shows flags
