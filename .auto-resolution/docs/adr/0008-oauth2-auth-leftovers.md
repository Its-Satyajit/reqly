# ADR 0008: OAuth 2.0 auth leftovers — device flow, keychain store, custom redirects, desktop auth

## Status

Accepted

## Context

Milestones 11–12 shipped OAuth 2.0 **Client Credentials** and **Authorization Code + PKCE** on the `TokenSource` + `secrets.Store` seams ([ADR 0006](0006-oauth2-client-credentials.md), [ADR 0007](0007-oauth2-authorization-code-pkce.md)). Four gaps block real-world use ([Spec #60](https://github.com/Its-Satyajit/reqly/issues/60)):

1. **Device flow** — providers with no browser callback (GitHub CLI-style, Cognito, Slack sign-in) require a `user_code` + `verification_uri` shown to the user and a **poll** of the token endpoint.
2. **Token storage** — cached tokens live in a plain-text 0600 JSON file. ADR 0006 promised an OS-keychain drop-in behind `secrets.Store`; none existed.
3. **Custom redirect schemes** — `loopbackRedirect` (ADR 0007) hard-rejected any non-loopback `redirect_uri`. Desktop apps conventionally register a custom scheme (e.g. `reqly://callback`) for deep-link callbacks.
4. **Desktop auth** — auth was CLI-only. The desktop bridge created a cache-less request client (noted as a gap in milestone 11) and had no login/status/logout surface.

## Decision

Four increments on the existing seams, each keeping the request engine and cache untouched:

1. **Device flow (RFC 8628)** — `DeviceCodeSource` (`internal/auth/oauth2_device.go`), dispatched by `grant_type: device_code`. `StartDeviceAuthorization` POSTs the device-authorization request (form `client_id` + `scope`/`audience`, Basic client auth); `PollDeviceToken` polls the token endpoint with `grant_type=urn:ietf:params:oauth:grant-type:device_code`, honoring the provider `interval` (default 5s), retrying on `authorization_pending`, adding 5s on `slow_down` (RFC 8628 §3.5), and terminating with distinct errors on `expired_token`/`access_denied`/non-2xx. The flow reports progress through an injectable `Status` callback; `reqly auth login --flow device` prints the verification URI (preferring `verification_uri_complete`) + `user_code` and waits. The granted token lands in the same per-workspace cache, so Bearer attach, expiry refresh, and 401 refresh+retry work unchanged. Automatic acquisition reports via `SetOAuth2DeviceStatus` (CLI → stderr). No browser is ever opened.
2. **OS-keychain store** — `secrets.KeychainStore` (`internal/secrets/keychain.go`) implementing `Store` via `zalando/go-keyring` (Secret Service / Keychain / WinCred; already an indirect dependency). Same `fs.ErrNotExist` semantics as `FileStore`. Secret values live in the keychain; the **set of keys** is tracked in a 0600 index file (`<workspace>/.reqly/keychain.index`) because keychain APIs have no portable list operation — account names are hash-derived cache keys, not credentials. `Keys()` self-heals entries deleted out-of-band. The real keyring call is isolated behind a `keychainOps` adapter so unit tests use an in-memory fake and CI (no Secret Service) never touches the OS; a `REQLY_TEST_KEYCHAIN=1` smoke test exercises the real driver. **Selection:** `--store file|keychain` (auth commands) > `REQLY_TOKEN_STORE` > default `file`; a keychain that cannot be opened falls back to the file store with a warning. `auth status` reports the active backend.
3. **Custom redirect schemes** — the auth-code flow's callback becomes a transport seam (`internal/auth/oauth2_authcode.go`): loopback URIs keep the one-shot HTTP listener unchanged; a non-loopback scheme (e.g. `reqly://callback`) is accepted **only when a receiver is registered** (`RegisterCustomSchemeReceiver`, called by the desktop app at startup) and completes via `DeliverCustomSchemeCallback(uri)`, which verifies `state` and extracts the code through the same `parseCallbackQuery` path as the HTTP handler, one-shot (the flow is removed from the registry at delivery). The CLI registers nothing, so a non-loopback `redirect_uri` fails fast with an actionable error instead of hanging.
4. **Desktop auth** — `internal/core.AuthService` (Login/Status/Logout, UI-agnostic, masked output) mirrors `reqly auth` on the same seams; `NewCachedRequestService` wires store-backed token caching into the desktop client (milestone-11 gap closed). The Wails `AppService` resolves the workspace root, opens the token store (**keychain default**, file fallback with warning), and exposes `AuthLogin`/`AuthStatus`/`AuthLogout`/`DeliverCustomSchemeCallback`. The React sidebar gains an **AuthPanel**: masked token list, flow picker (browser/device), JSON config editor, login/logout. The shared frontend defines an `AuthAdapter` + `useAuthStore`; the host injects the Wails-backed adapter (browser dev falls back), mirroring the request-sender pattern.

## Consequences

- **Positive:** headless auth (device flow) with no browser dependency; tokens at rest in the OS credential store by choice, with a safe default and graceful fallback; desktop auth-code logins can complete via `reqly://` deep links; desktop requests authenticate identically to the CLI; every increment was testable with `httptest`/fakes (race-clean) plus an env-guarded real-keychain smoke test.
- **Trade-off:** the keychain **key index** (account names only, 0600) is needed for `Keys()`; a keychain that is locked/unavailable degrades to the file store rather than failing; device-flow auto-acquisition prints progress to stderr rather than a GUI; the desktop's OS-level `reqly://` scheme registration (Wails URL-scheme support, `.desktop`/Info.plist/registry entries) is the host's job and lands with desktop packaging, not in this ADR.
- **Deferred:** Password/ROPC (deprecated in OAuth 2.1); OAuth 1.0, AWS SigV4, Akamai EdgeGrid; refresh-token rotation policy beyond "rotate when the server returns a new one"; device-flow UX in the desktop panel (the bridge can trigger it later).

## References

- [Spec #60](https://github.com/Its-Satyajit/reqly/issues/60), tickets [#61–#65](https://github.com/Its-Satyajit/reqly/issues/61), PRs #66–#69.
- RFC 8628 (Device Authorization Grant), RFC 6749 §6 (refresh), RFC 7636 (PKCE).
- ADR 0006 (Client Credentials), ADR 0007 (Authorization Code + PKCE) — the seams this builds on.
