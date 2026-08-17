# ADR 0006: OAuth 2.0 Client Credentials with Store-Backed Token Caching

## Status
Accepted

## Context
The roadmap (§1.3) requires OAuth 2.0. ADR 0005 deferred it, noting that the flat `map[string]string` auth config can't express nested token responses or refresh flows. Reqly is local-first and Git-native: auth configs live in request/collection descriptors, and credential material must never be committed. We need a token lifecycle (acquire, cache, refresh, logout) that is invisible to the user and survives across CLI invocations without an interactive login.

## Decision
1. Ship **Client Credentials** (RFC 6749 §4.4) first: a form-encoded POST to `token_url` with HTTP Basic client auth (`client_id`/`client_secret`), applied as `Authorization: Bearer <access_token>`. Config keys: `grant_type` (default `client_credentials`), `token_url`, `client_id`, `client_secret`, `scope`, `audience`, `token_name` (JSON response field holding the token, default `access_token`).
2. **Split acquisition from application** via a new `auth.TokenSource` interface. Schemes that need a pre-request token implement `Token(ctx, cfg, vars) (Token, error)`; the request engine calls it before `Apply`, injects the resolved token into a copy of the config under the `"token"` key, and returns it as `response.AuthToken` (`json:"-"`, never serialized) for post-request masking. The request's own config is never mutated.
3. **Tokens live outside the descriptor** in a `secrets.Store` at `<workspace>/.reqly/tokens.json` (0600, atomic temp-file writes), git-ignored via `.reqly/`. The `secrets.Store` interface (Get/Set/Delete/Keys) abstracts the backend so an OS-keychain implementation is a drop-in replacement (§1.4). Cache keys are `sha256(workspace root + canonicalized auth config)`.
4. **Caching and refresh** sit in a `CachedTokenSource` decorating the scheme's `TokenSource`: reuse a persisted token while fresh, re-acquire within a 30-second expiry skew, persist atomically. The engine serializes acquisition per config with a per-client mutex so concurrent requests never double-acquire. A reactive path forces the cached token out on a 401 and retries exactly once; a second 401 returns as-is (no retry loop).
5. **No interactive login.** Acquisition is automatic on the first request. `reqly auth status` lists per-workspace cached tokens (endpoint, expiry, masked token, state) and `reqly auth logout` clears them; both resolve the workspace root by walking up to the descriptor, matching the environments discovery.
6. **Defer Authorization Code + PKCE** (system-browser flow, state/verifier) to the next milestone, and **defer the Password/ROPC grant** (deprecated in OAuth 2.1).

## Consequences
- **Positive:** no secrets in Git-tracked descriptors; token reuse across CLI invocations and collection steps; bounded, self-healing refresh with no retry loops; additive — Auth Code + PKCE plugs into the same `TokenSource` + store seams; masking covers both `client_secret` and the acquired token.
- **Trade-off:** plain-text (0600) token files at rest until the keychain backend ships; `token_name` indirection adds a config knob; the flat-config shape persists (no nested OAuth payloads in descriptors).
- **Deferred:** Password/ROPC, OS keychain, custom redirect schemes, device flow. Authorization Code + PKCE and refresh-token reuse shipped in [ADR 0007](0007-oauth2-authorization-code-pkce.md).