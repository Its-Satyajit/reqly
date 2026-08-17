# ADR 0005: Git-Native Authentication Schemes

## Status
Accepted

## Context
Reqly needs first-class authentication. Today `request.Auth{Type, Config}` exists as a serializable shape (used by request files and collection descriptors) with a hard-coded `applyAuth` switch in `request/client.go` covering `bearer`, `basic`, and `apikey`. The roadmap (§1.3) requires Basic, Bearer, API key, JWT, Digest, NTLM, OAuth 1.0/2.0, AWS Signature, Akamai EdgeGrid. The way schemes are dispatched, configured, and inherited is hard to reverse once users have Git-tracked projects, so we record it.

## Decision
1. A dedicated `internal/auth` package with a `Scheme` interface: each scheme applies itself to an outgoing request given a `request.Auth` and the variable set. A registry dispatches by `auth.type` string; unknown types are an error at apply time, not a silent no-op.
2. `request.Auth{Type string, Config map[string]string}` stays the single Git-native serializable shape — flat string config, no per-scheme structs. Schemes validate their own required `config` keys.
3. Ship in this milestone: `basic`, `bearer`, `apikey` (formalized from the current switch), plus `digest` (challenge/response) and `jwt` (per-request signing). Defer `ntlm`, `oauth1`/`oauth2`, `aws` (SigV4), and `akamai` to later milestones; OAuth 2.0 is its own feature area (§17).
4. Auth values are secrets: scheme config values are interpolated and fed through the environment `Masker` so nothing sensitive leaks in output, errors, or logs.
5. `auth.type: none` explicitly clears inherited auth on a request (public endpoints under an auth-bearing collection). Inheritance rules stay as-is: request-level auth replaces the inherited one when non-empty.
6. No new CLI flags this milestone — auth flows through request files and collection descriptors. A JWT decode/claims-viewer CLI (`reqly jwt`) is a separate follow-up.

## Consequences
- **Positive:** matches the minimal-dependency posture; flat string config stays diffable and Git-native; a registry makes future schemes (OAuth, AWS SigV4) additive; masking keeps secrets out of output everywhere.
- **Trade-off:** `map[string]string` config can't express nested/typed values (OAuth token responses, refresh flows) — those land with the OAuth milestone; Digest is request-body aware and needs care with streaming bodies.
- **Deferred:** NTLM, OAuth 1.0/2.0, AWS SigV4, Akamai, auth plugins, and the JWT tooling CLI.