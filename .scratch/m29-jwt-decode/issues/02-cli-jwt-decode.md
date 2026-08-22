# 02 — CLI reqly jwt decode

**What to build:** `reqly jwt decode <token> [--json]` (plus `echo $JWT | reqly jwt decode -`) prints header + payload + expiry for copy-paste or CI piping, with per-segment errors.

**Blocked by:** 01 — internal/jwt decode seam + expiry

**Status:** ready-for-agent

- [ ] Cobra `jwt` root + `decode` subcommand in `apps/cli/cmd/jwt.go` (one file per group), registered in `apps/cli/cmd/root.go`, flag `--json` bool, positional `<token>` may be `"-"` for stdin
- [ ] Human default: `Header:\n<pretty>\n\nPayload:\n<pretty>\n\nExpiry: <line>` (`expired 2h ago`/`valid for 3h`/`not yet valid (nbf in 5m)`/`no expiry` + `issued` when `iat` present) via `json.MarshalIndent`
- [ ] Machine mode `--json`: single indented JSON `{"header":{...},"payload":{...},"signature":"...","alg":"HS256","expiry":{...}}` to stdout
- [ ] Reads stdin when `<token>` is `"-"` (pipe `echo $JWT | reqly jwt decode -`), trims whitespace, strips `Bearer ` prefix via `internal/jwt.Decode`
- [ ] Malformed token → stderr per-segment error + `exit 1`; `reqly jwt --help` / `decode --help` documented
- [ ] Integration tests `apps/cli/cmd/jwt_test.go` — pretty vs `--json`, stdin `-`, Bearer prefix, malformed `exit 1`, `--help` contains `decode` (prior art `auth_test.go`)
- [ ] `go test ./...` + `go build -o reqly ./apps/cli` + `reqly jwt decode --help` smoke
