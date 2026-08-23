# 02 — Request-file round-trip + CLI flags

**What to build:** Users declare retry policy in plain-text request files that save/reload losslessly, or override per-invocation with `reqly run --retries N --retry-delay <duration>` without editing the file; history shows one record per logical execution carrying its attempt count.

**Blocked by:** 01 — Engine retry loop

**Status:** ready-for-agent

- [x] `retry:` block parses from YAML+JSON request files and round-trips format-preserving through Save (atomic, comment/indentation preserving)
- [x] `--retries int` and `--retry-delay duration` flags on `run` only; gated by Changed() so unset flags preserve file values; work in URL-mode too
- [x] History records final attempt only, with Attempts visible in `history show` output
- [x] CLI prints masked `retrying in <delay>s (attempt i/j)` line per retry; silent otherwise
- [x] CLI integration tests: temp file + httptest fail-once server — file-only, flag-override, URL-mode flag paths
- [x] `go vet` + `gofmt -l` + `go test -race ./internal/requestfile ./apps/cli/...` green; `go build -o reqly ./apps/cli`
