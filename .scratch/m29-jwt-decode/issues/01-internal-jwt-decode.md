# 01 — internal/jwt decode seam + expiry

**What to build:** Offline JWT decoding without verification — `internal/jwt.Decode(token)` returns header + payload + signature + expiry status for any `alg` (HS*/RS*/ES*/none), with `Bearer ` prefix strip and RFC 7519 numeric date handling. Enables `reqly jwt decode` and future desktop inspector via one pure seam.

**Blocked by:** None — can start immediately

**Status:** ready-for-agent

- [ ] `internal/jwt.Decode(token string) (*Token, error)` — base64url decode header/payload, JSON unmarshal to `map[string]any`, capture raw segments, report `Alg`, work with `time.Now().UTC()` injectable `Now` for tests
- [ ] `ExpiryStatus` from `exp`/`nbf`/`iat` numeric (int/float) — `expired`/`not_yet_valid`/`valid`/`no_expiry` + `Remaining` seconds + non-numeric `exp` surfaces as `invalid exp: not numeric` but still returns decoded JSON
- [ ] `Bearer ` prefix stripped case-insensitively, whitespace trimmed, token must have 3 dot-segments (signature may be empty for `none`)
- [ ] Per-segment explicit errors (`expected 3 segments`, `invalid header: base64url`, `invalid payload: not JSON`)
- [ ] Table-driven unit tests `internal/jwt/decode_test.go` — valid HS256/RS256/none, expired/not-yet-valid/no-expiry, Bearer prefix, whitespace, float exp, non-numeric exp, 2 segments, bad base64url, non-JSON payload (inline tokens, injectable Now)
- [ ] Stdlib only (`encoding/base64`, `encoding/json`, `strings`, `time`) — no new go.mod deps; `go test ./internal/jwt -race` + `go vet` + `gofmt -l` green
