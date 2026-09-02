# ADR 0021: JWT Tooling — Decode MVP (M29)

## Status
Accepted

## Context
`ROADMAP.md:344` (`JWT tooling — reqly jwt decode + signing/verification`) and `docs/features.md:18` (JWT decoder, header/claims viewer, expiry, signature info) remain the last P1 auth gap. `internal/auth/jwt.go:39` already signs HS256/384/512 per request (`secret`/`algorithm`/`claims`/`expiresIn` → `Bearer`, `jwtHashes`, `signJWT`), but has no inspection CLI. `docs/adr/0005:6` deferred `reqly jwt` decode/claims-viewer. Design questions: decode-only vs decode+verify/sign scope, CLI shape (`reqly jwt` vs `reqly auth jwt`), algorithm coverage for decode, output shape (pretty vs JSON), expiry semantics, desktop parity, and where the new code lives (`internal/jwt` vs `internal/auth` extension).

## Decision
1. **M29 is decode-only; verify/sign are M29b.** `reqly jwt decode <token> [--json]` ships in M29. `reqly jwt verify --secret` (HS verify) and `reqly jwt sign --secret --alg --claims` reuse the same `signJWT`/`jwtHashes` seam and defer to M29b — the CLI shape is additive (`jwt` root with `decode`/`verify`/`sign` subcommands) so M29 does not break.
2. **New `internal/jwt` package beside `internal/auth`, pure functions, no network, no CGO.** `Decode(token string) (*Token, error)` + `DecodeHeader`/`DecodePayload` + `ExpiryStatus` (expired / not-yet-valid / ttl) — hand-rolled base64url + JSON, no `golang-jwt` dependency, mirroring `jwt.go` stdlib-only posture. `internal/auth/jwt.go` keeps signing; `internal/jwt` owns inspection. CLI `apps/cli/cmd/jwt.go` is thin Cobra wrapper (one file per command group, `root.go` registers `jwtCmd`).
3. **Decode is algorithm-agnostic and never verifies.** Header `alg` is reported (including `none`, `RS*`, `ES*`) but no crypto is attempted — the existing HS HMAC stays in `internal/auth`. Malformed segments, invalid base64url, or non-JSON claims surface as explicit errors; `exp`/`nbf`/`iat` are parsed as numeric dates (int/float, per RFC 7519 §2) for expiry detection.
4. **CLI shape `reqly jwt decode`, stdout default, `--json` machine mode.** `reqly jwt decode <token>` prints header JSON + payload JSON pretty + `Expiry:` line (`expired 2h ago`, `valid for 3h`, `nbf in future`). `--json` emits single JSON `{header,payload,expiry}` for scripting/CI. Token may have optional `Bearer ` prefix stripped. One positional arg; piping via `echo $JWT | reqly jwt decode -` (stdin `-`) supported.
5. **Desktop inspector deferred to M29b.** M29 ships CLI + `internal/jwt` only; the desktop JWT inspector (auto-decode of `Authorization: Bearer` responses + dedicated tool) reuses `internal/jwt.Decode` via Wails binding in M29b without changing the core seam.

## Considered Options
- **Decode+verify+sign in one milestone** — rejected: doubles scope (secret handling, masking, `SecretKeys()` parity) and blocks the simpler, already-deferred `decode` value.
- **`internal/auth` extension (no new package)** — rejected: tooling is orthogonal to request auth dispatch and should work for tokens not tied to `auth.config`; a separate `internal/jwt` keeps `auth` focused on `Apply` and avoids importing `secrets`/`mask` for tooling.
- **`reqly auth jwt decode` or `reqly jwt` flat args** — rejected: `auth` holds OAuth login/status/logout lifecycle; JWT inspection is a stateless utility, like `reqly docs`/`reqly diff`, so a top-level `reqly jwt` root matches `docs/reference/go.md` Cobra discoverability and `ROADMAP.md:344`.
- **`golang-jwt/jwt` library** — rejected: stdlib `encoding/base64` + `encoding/json` already suffices for decode and keeps the binary minimal (same rationale as hand-rolled `SHA256/512` in `jwt.go:21`).

## Consequences
- **Positive:** Closes the `ROADMAP.md:344` decode gap with one pure package, one Cobra file, no new dependencies, algorithm-agnostic inspection, and a clear M29b path for HS sign/verify reusing existing seams.
- **Trade-off:** M29 desktop has no JWT inspector; HS verify/sign UX (secret flag, `--alg` default, expiry override) stays open until M29b.
