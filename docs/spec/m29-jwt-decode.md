# Spec: JWT Tooling — Decode MVP (Milestone 29)

> **Status:** Draft — grill settled 2026-08-23 (Q1–Q5, decode-only MVP)
> **Scope:** P1 `ROADMAP.md:214`/`344` — `reqly jwt decode` header/claims viewer + expiry detection (sign/verify deferred to M29b), `docs/features.md:18` JWT Tooling
> **Stack:** `internal/jwt` (new) + `apps/cli/cmd/jwt.go` + `internal/version` (creator) — no new deps, stdlib `base64`+`json`+`time`, reuses `internal/auth/jwt.go:79` hashes only in M29b
> **Predecessor:** M28 `docs/adr/0020-har-import-export.md` (T1–T3 MVP in progress), M29 `docs/adr/0021-jwt-tooling-decode.md` (Q1–Q5)

## Problem Statement

`ROADMAP.md:97` `JWT — HS256/384/512 per-request signing shipped` but `reqly jwt` decode/claims-viewer deferred (`docs/adr/0005:6`). `docs/features.md:18` requires JWT decoder + header inspection + claims viewer + expiration detection + signature info + signing/verification — none is usable offline today. Users pasting `Authorization: Bearer eyJ...` from Reqly or an external IdP have no local way to inspect `header.alg`/`payload.exp`/`payload.sub` or see `expired`/`not yet valid` without calling jwt.io or adding a library. `internal/auth/jwt.go:88` `signJWT` creates tokens but never decodes them; no `internal/jwt` seam exists.

## Solution

One pure decode seam plus one Cobra command group:

* **`internal/jwt.Decode(token string) (*Token, error)`** — base64url-decodes header + payload (no verification), JSON-unmarshals each into `map[string]any` (preserves `alg`/`kid`/`typ` + arbitrary claims), captures raw segments, computes `ExpiryStatus` from numeric `exp`/`nbf`/`iat` (RFC 7519 §2: int or float, seconds since epoch). Works for any `alg` (`HS*`, `RS*`, `ES*`, `none`) and for tokens with missing `exp` (`no_expiry`). `Bearer ` prefix stripped before decode; whitespace trimmed. Errors are per-segment (`invalid token: expected 3 segments`, `invalid header: base64url`, `invalid payload: not JSON`) — never silent.

* **`reqly jwt decode <token> [--json] [-]`** — Cobra `jwt` root + `decode` subcommand (`apps/cli/cmd/jwt.go`, registered in `apps/cli/cmd/root.go:57`, one-file-per-group). Human default: pretty-printed `Header:` JSON + `Payload:` JSON + `Expiry:` line (`expired 2h ago`, `valid for 3h`, `not yet valid (nbf in 5m)`, `no expiry`). Machine mode `--json`: single JSON `{"header":{...},"payload":{...},"signature":"...","alg":"HS256","expiry":{"status":"expired|valid|not_yet_valid|no_expiry","remaining": -7200, "exp":123, "nbf":...}}`. Stdin `-` reads token from pipe (`echo $JWT | reqly jwt decode -`). No secret, no network, no masking (input is the token itself). `exit 1` on malformed token with stderr error.

Desktop JWT inspector (`AppService.DecodeJWT` binding) and `verify`/`sign` (HS HMAC via `internal/auth/jwt.go:79` `jwtHashes`) are **M29b** — M29 ships the highest reusable seam (`internal/jwt`) the desktop can bind without new Go work.

## User Stories

1. As a developer, I want `reqly jwt decode eyJhbG...` to print header claims (alg, typ, kid) pretty-printed, so I can see which algorithm signed the token.
2. As a developer, I want the payload claims (sub, exp, iat, iss, aud, custom) pretty-printed, so I can inspect who the token is for and what it carries.
3. As a developer, I want `Expiry: expired 2h ago` or `valid for 15m` derived from `exp`/`nbf` at decode time (`time.Now().UTC()`, no leeway), so I know if I need to refresh without sending a request.
4. As a developer, I want a token with no `exp` to report `no expiry` (not an error), so non-expiring service tokens are distinguishable from broken ones.
5. As a developer, I want `reqly jwt decode --json eyJ...` to emit machine JSON (`header`+`payload`+`expiry`), so I can pipe to `jq '.payload.sub'` in CI.
6. As a developer, I want `echo $JWT | reqly jwt decode -` to read from stdin, so shell variables and `pbpaste` work without quoting long tokens.
7. As a developer, I want `Authorization: Bearer eyJ...` pasted raw to be accepted (prefix stripped), so copy-paste from Reqly history or DevTools needs no editing.
8. As a developer, I want `alg: none` and `RS256`/`ES256` tokens to decode (no verification), so I can inspect IdP tokens even though Reqly only signs HS*.
9. As a developer, I want malformed tokens (2 segments, bad base64url, non-JSON payload) to fail with a per-segment error and `exit 1`, so scripting can detect bad input.
10. As a developer, I want `exp`/`nbf` that are non-numeric strings to surface `invalid exp: not numeric` but still show decoded JSON, so I see what the issuer sent.
11. As a developer, I want `iat` shown as age info (`issued 3h ago`) when present, without affecting validity, so I can reason about token freshness.
12. As a CLI user, I want `reqly jwt --help` and `reqly jwt decode --help` to document usage, so Cobra discoverability matches `reqly docs`/`diff`/`history`.
13. As a future desktop user, I want the same `internal/jwt.Decode` seam to power a desktop JWT inspector (M29b) without changing Go, so CLI and desktop stay parity via Go core.

## Implementation Decisions

- **Seam: `internal/jwt` pure package, highest seam.** `func Decode(token string) (*Token, error)` + `type Token struct { Header, Payload map[string]any; RawHeader, RawPayload, Signature string; Alg string; Expiry ExpiryStatus }` + `type ExpiryStatus struct { Status string; Remaining int64; Exp, Nbf, Iat *int64 }` + `func isExpired(expiry int64) bool`. No `DecodeHeader`/`DecodePayload` split for M29 — single `Decode` keeps the seam minimal; helpers can be added in M29b without breaking. Package imports only `encoding/base64`, `encoding/json`, `strings`, `time` — zero new `go.mod` deps, same stdlib-only posture as `internal/auth/jwt.go:21`.

- **No verification in M29.** `Alg` is reported from decoded header, `Signature` is captured as raw third segment (may be empty for `none`), but no `hmac.New` call. HS verify (`--secret`) reuses `internal/auth/jwt.go:79` `jwtHashes` in M29b — decode stays algorithm-agnostic so `RS256`/`ES256` from external IdPs decode.

- **Numeric date handling (RFC 7519 §2).** `exp`/`nbf`/`iat` accepted as `json.Number` → `float64` → `int64` seconds (truncated); fractional seconds preserved via float but expiry computed on int seconds. Non-numeric type → `FieldError` but decode still succeeds (payload JSON is shown). Missing `exp` → `Status=no_expiry`, `Remaining=0`.

- **Time source: `time.Now().UTC()` at decode, no leeway/skew.** Unlike `internal/secrets` 30s skew for `CachedTokenSource`, tooling reports raw delta; `--leeway` deferred — keeps output deterministic for tests via injectable `Now func() time.Time` (default `time.Now`, tests override).

- **CLI shape: `jwt` root + `decode` subcommand.** `apps/cli/cmd/jwt.go` defines `jwtCmd = &cobra.Command{Use:"jwt", Short:"JWT tooling"}` + `jwtDecodeCmd = &cobra.Command{Use:"decode <token> [--json]", Args: cobra.ExactArgs(1)}`; `root.go:57` registers `jwtCmd`. Flags: `--json` bool (default false). Positional `<token>` may be `"-"` for stdin. `Bearer ` prefix stripped case-insensitively; surrounding whitespace trimmed; token with 3 dot-segments required (header.payload.signature, signature may be empty for `alg:none` → allow `header.payload.` as 3 segments via `strings.SplitN` with trailing empty). Errors write to `cmd.ErrOrStderr()` + `exit 1`.

- **Output contracts.** Default: `fmt.Fprintf(out, "Header:\n%s\n\nPayload:\n%s\n\nExpiry: %s\n", pretty(header), pretty(payload), expiryLine)` where `pretty` is `json.MarshalIndent(m, "", "  ")` + expiry line from `ExpiryStatus` (`expired 2h ago` via `time.Duration.String()` rounded, `valid for 3h`, `not yet valid (nbf in 5m)`, `no expiry`, `issued 1h ago` appended when `iat` present). `--json`: `json.MarshalIndent(map[string]any{"header":tok.Header,"payload":tok.Payload,"signature":tok.Signature,"alg":tok.Alg,"expiry":tok.Expiry}, "", "  ")` to stdout. No `--header-only`/`--payload-only` for M29 (defer).

- **Error messages preserve segment context.** `invalid token: expected 3 segments, got 2`; `invalid header: base64url: illegal base64 data`; `invalid payload: not JSON: ...`; `invalid exp: not numeric (got string)` — tests assert substring per `golang-error-handling` `%w` wrapping.

- **File modes unchanged.** No new files on disk; `internal/jwt` is pure; CLI reads nothing from workspace.

## Testing Decisions

- **What makes a good test:** assert external behavior (`Decode` returns header/payload/expiry + CLI prints correct stdout/stderr + exit code), not base64 internals. Table-driven `testing` (no testify required, but testify allowed where prior art uses it) — like `internal/auth/jwt_test.go:44`.

- **Seams to test (highest first):**
  - `internal/jwt` unit: `Decode` valid HS256 (header alg/payload sub), RS256/ES256/none (no verify), `Bearer ` prefix, whitespace, stdin-equivalent string, expired (`exp` in past → `expired`), not-yet-valid (`nbf` future → `not_yet_valid`), `no_expiry`, `iat` age, numeric float `exp`, non-numeric `exp` error, 2 segments error, bad base64url, non-JSON payload.
  - CLI integration: `apps/cli/cmd/jwt_test.go` — `jwt decode <token>` pretty vs `--json`, `jwt decode -` from stdin, `Bearer ` prefix, malformed token `exit 1` + stderr, `--help` contains `decode`.
  - No `testdata/*.golden` — JWT strings are short (inline fixtures, deterministic via injected `Now`), unlike `internal/exporter/code_test.go` golden for longer snippets.

- **Prior art:** `internal/auth/jwt_test.go` (table-driven HS*, base64 `RawURLEncoding`, HMAC verify), `apps/cli/cmd/auth_test.go` (Cobra `Execute` harness), `internal/variables/tag_test.go` (fixed time injection via `Now func`).

- **Quality gates:** `go test ./...` + `go test -race ./...` + `go vet` + `gofmt -l` clean; `go build -o reqly ./apps/cli` + `reqly jwt decode --help` smoke; `npm run lint` exit 0 (frontend untouched).

## Out of Scope

- `reqly jwt verify <token> --secret <s> [--alg HS256]` — HS256/384/512 HMAC verification (M29b, reuses `jwtHashes` + `signJWT` seam, adds masking for `--secret` in output).
- `reqly jwt sign --secret --alg --claims --expiresIn` — compact JWS creation (M29b, same `signJWT` seam as `auth.type: jwt`, secrets never printed).
- `reqly jwt decode --header-only` / `--payload-only` / `--raw` / `--leeway` flags — single unified view for M29.
- Desktop JWT inspector (ResponseViewer `Authorization: Bearer` auto-decode + dedicated JWT tool panel + `wails3 generate bindings`) — M29b, via `AppService.DecodeJWT` binding to `internal/jwt.Decode`.
- Private claims validation (e.g. `iss`/`aud` against a spec) and `kid`/`jwk` fetching — decode only.
- Library `golang-jwt/jwt` or `lestrrat-go/jwx` — stdlib suffices for decode.

## Further Notes

- **ADR:** `docs/adr/0021-jwt-tooling-decode.md` (accepted, Q1–Q5) — decode-only MVP; M29b adds `verify`/`sign` + desktop.
- **Glossary:** `CONTEXT.md:263` `JWT Tooling`/`JWT Decode`/`JWT Verify`/`JWT Sign` (this spec's grilling).
- **ROADMAP:** `ROADMAP.md:344` M29 `JWT tooling` — tick `reqly jwt decode` (header/claims viewer, expiry detection) after ship; `signing/verification` stays unchecked for M29b.
- **Ticket split (see `to-tickets`):** T1 `internal/jwt` decode + expiry + tests, T2 `apps/cli/cmd/jwt.go` + `root.go` wiring + CLI tests, T3 docs (ADR 0021 + CONTEXT + ROADMAP + this spec) + `go test -race` green.

