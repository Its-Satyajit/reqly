# Domain Glossary

### Release Pipeline Architecture
Dual-orchestration pipeline publishing releases to GitHub (`Its-Satyajit/reqly`). Separates headless Go CLI compilation (GoReleaser) from native GUI desktop app bundling (Wails v3 OS matrix across macOS, Linux, Windows).

### Linux Artifact Matrix
Multi-distro Linux distribution strategy for Reqly binaries. Covers Debian `.deb`, universal `.AppImage`, compressed `.tar.gz`, Arch Linux AUR (`reqly-bin`), and multi-distro detection in `install.sh` supporting `pacman`, `apt`, `dnf`, and binary fallback.

### Unsigned Release Fallback
Zero-cost release strategy for open-source distributions. macOS binaries use ad-hoc code signing (`codesign -s -`) with quarantine removal (`xattr -d com.apple.quarantine`), Windows binaries use unsigned `.exe`/`.zip`, and Linux binaries use native `.AppImage`/`.tar.gz`.

### Automated Tag Release Trigger
Event-driven GitHub Actions workflow triggered by pushing semver tags (`git push origin v*.*.*`). Executes GoReleaser and Wails v3 OS matrix, generates release notes from Conventional Commits, uploads binaries, and updates checksums.

### Multi-Distro Install Script
POSIX-compliant installation script (`curl -fsSL https://raw.githubusercontent.com/Its-Satyajit/reqly/main/install.sh | sh`). Detects OS and Linux package manager (`pacman`, `apt`, `dnf`, `zypper`), installing binaries or system packages to `/usr/local/bin/reqly`.

### AI Agent Protocol & Workspace Rules
Standardized instruction configuration (`AGENTS.md`, `GEMINI.md`, `.cursorrules`, `.github/copilot-instructions.md`) governing AI coding assistant contributions to Reqly. Enforces TDD verification (`go test ./...`), local-first/Git-native constraints, and the 5-stage skill pipeline (`/grill-with-docs` → `/to-spec` → `/to-tickets` → `/implement` → `/code-review`) in `~/.agents/skills/`.

### Specification & Project Validation Engine
Static verification subsystem (`internal/validation`) backing `reqly validate`. Audits OpenAPI 2.x/3.x specifications (`reqly validate openapi <path>`) and local Git-native project descriptors (`reqly validate project [path]`) for broken schema references, missing required attributes, and invalid descriptor syntax.

### Structural Diff Engine
AST and JSON/YAML key-aware diffing engine (`internal/diffing`) backing `reqly diff`. Computes semantic differences between API specifications (`reqly diff spec <spec1> <spec2>`), request descriptors (`reqly diff request <file1> <file2>`), and JSON/YAML response payloads (`reqly diff response <file1> <file2>`).

### Environment
A named set of variables (plus optional secrets) selectable per workspace and applied to requests through the variable precedence chain between the global and collection scopes.

### Active Environment
The environment currently selected for a workspace; resolved with the following precedence (highest wins): `REQLY_ENV` process variable, then the `--env` CLI flag, then a request/test file's `environment:` field, then the workspace descriptor's `environment:` field.

### Environment File
A YAML file under `environments/<name>.yaml` holding an environment's `variables:` and `secrets:` maps. Resolved from the nearest `environments/` directory — CWD (or the request file's directory) for standalone `run`/`test`, the workspace root for collection commands. The filesystem is the source of truth.

### Secret
An environment variable marked as sensitive; its value is masked in CLI and test output and never printed by `reqly env show`.

### Process Environment Scope
The lowest-precedence variable scope fed by the OS environment and the nearest-directory `.env` file, mirroring Node's `process.env`. When both define a key, the OS environment wins, so CI can override local `.env` values.

### Environment Validation
Workspace-aware static checks (`reqly env validate <name>`) over an environment: file syntax, secret-exposure warnings (name-heuristic for `key/token/secret/password/credential` plus duplicate keys in both `variables:` and `secrets:`), and undefined-variable detection across the workspace's requests under the active env.

### Environment Diff
Secret-aware comparison of two environments (`reqly env diff <nameA> <nameB>`): added/removed/changed keys via the structural diff engine, with changed secret values rendered as `[SECRET]` so the diff shows *which* secret changed without leaking it.

### Authentication Scheme
A named credential-application strategy (`internal/auth`) that mutates an outgoing request. Dispatched by a request's `auth.type` string through a registered `Scheme` interface; each scheme owns its `config` keys and applies them to the request before send.

### Auth Config
The flat string map on a request/collection descriptor (`auth.config`) holding a scheme's fields (e.g. `token`, `username`, `password`, `key`, `in`). Values are interpolated like variables and fed to the environment masker so secrets never leak in output.

### Auth Inheritance
A request resolves its auth from the nearest enclosing collection/folder that defines one; a request's own non-empty `auth.type` overrides it, and `auth.type: none` explicitly disables inherited auth for that request.

### No Auth
The `none` scheme: clears any inherited auth on a request so it is sent unauthenticated, even under an auth-bearing collection/folder.

### JWT Auth
A scheme that signs a JSON Web Token per request from `config` (secret, algorithm, claims) and sends it as `Authorization: Bearer <token>`. Distinct from JWT *tooling* (decode/claims viewer), which is a separate CLI feature.

### OAuth 2.0 Auth Scheme
An authentication scheme (`internal/auth`) implementing OAuth 2.0 grants dispatched on `grant_type`: **Client Credentials** (RFC 6749 §4.4), **Authorization Code + PKCE** (RFC 6749 §4.1, RFC 7636 — see below), and **Device flow** (RFC 8628 — see below). Applied as `Authorization: Bearer <access_token>`. Configured via flat `auth.config` keys (`grant_type`, `token_url`, `authorization_url`, `device_authorization_url`, `client_id`, `client_secret`, `redirect_uri`, `scope`, `audience`, `token_name`). Client Credentials acquisition is automatic on first use; Authorization Code and Device flow are interactive (`reqly auth login`) or auto-acquired on first request.

### Token Source
An optional capability (`auth.TokenSource`) a scheme may implement to acquire a token ahead of request application. The request engine invokes it before `Apply` and injects the resolved token into a copy of the auth config under `token`; the request's own config is never mutated. The raw token is returned on the response as `AuthToken` for post-request masking.

### Token Store
A key/value secret backend (`internal/secrets`) persisting acquired OAuth tokens per workspace. Two implementations behind the `secrets.Store` interface (Get/Set/Delete/Keys): **FileStore** at `<workspace>/.reqly/tokens.json` (0600, atomic temp-file writes, the default) and **KeychainStore** in the OS credential store (Secret Service / Keychain / WinCred via `go-keyring`; 0600 key index for enumeration). Backend selection: `--store file|keychain` (auth commands) > `REQLY_TOKEN_STORE` > default `file`; a keychain that cannot be opened falls back to the file store with a warning. `reqly auth status` reports the active backend. Cache keys are a hash of the workspace root and the canonicalized auth config.

### Token Cache
A `CachedTokenSource` decorating a `TokenSource` with store-backed reuse: a persisted token is reused while fresh (30s expiry skew absorbs clock drift), otherwise re-acquired and persisted. The engine serializes acquisition per config so concurrent requests do not double-acquire.

### Token Refresh
Self-healing token lifetime: expiry-driven re-acquisition before send, plus a reactive path where a 401 forces the cached token out, renews, and retries exactly once (a second 401 is returned as-is). When the cached token carries a refresh token, renewal uses the refresh-token grant (RFC 6749 §6) via the optional `RefreshingTokenSource` capability — the browser flow is never re-run while a refresh token exists.

### Authorization Code + PKCE Flow
The browser-based grant (`AuthorizationCodeSource`, RFC 6749 §4.1 + RFC 7636): validate config, generate a `code_verifier` (32 random bytes, base64url) and its S256 `code_challenge` plus a per-flow `state`, build the authorization URL, start a one-shot callback transport, open the browser (via an injectable `Open` hook — the CLI's platform launcher), wait for the redirect, verify `state`, and exchange the code (`code`, `redirect_uri`, `client_id`, `code_verifier`, Basic client auth) for a token with an optional `refresh_token`. The callback transport is either the **loopback HTTP listener** on an ephemeral `127.0.0.1` port (default, and the only option the CLI accepts) or a **custom scheme** (`reqly://callback`): the desktop app registers its scheme via `RegisterCustomSchemeReceiver` and feeds deep links through `DeliverCustomSchemeCallback`, which verifies `state` one-shot. A config-provided `redirect_uri` must be loopback unless a receiver is registered.

### Device Flow
The headless grant (`DeviceCodeSource`, RFC 8628): POST the device-authorization request, surface `verification_uri` (or `verification_uri_complete`) + `user_code`, then poll the token endpoint honoring the provider `interval`, retrying on `authorization_pending`, adding 5s on `slow_down`, and terminating on `expired_token`/`access_denied` with clear errors. `reqly auth login --flow device` prints the URI + code and waits; the token is cached like any other grant. Automatic acquisition reports progress via `SetOAuth2DeviceStatus` (CLI → stderr).

### Loopback Callback
The one-shot local HTTP listener that receives the provider's authorization redirect: a single GET is accepted, `state` is verified against the flow, the `code` (or `error`/`error_description`) is extracted, a small page is rendered, and the listener shuts down. `WaitCode` blocks with a 10-minute hard cap when the caller's context carries no deadline.

### Refresh Token Grant
RFC 6749 §6 renewal: a form POST to `token_url` with `grant_type=refresh_token` and the stored `refresh_token`, using the cached credentials. The response yields a new access token and may carry a new refresh token (rotation — persisted when present; the previous one is kept otherwise). `reqly auth login` and the engine's refresh paths use it so expired tokens recover without reopening the browser.

### Request Builder Tabs
The desktop request editor's tab bar (**Params / Headers / Body**) from milestone 14. Params and Headers are key-value row editors (add/remove/toggle-enabled); the Body tab holds a body-type picker. The send path maps the tabs onto `request.Request{Query, Headers, Body}`; the engine's existing merge semantics apply (query params override the URL's, a manual `Content-Type` header beats the body type's default).

### Body Type
A request-body kind selected in the builder's Body tab: **none / JSON / XML / form-data / urlencoded / raw text**. JSON/XML/raw use a CodeMirror editor with the matching language; form-data and urlencoded use key-value rows. Selecting a type sets the appropriate `Content-Type` (or `multipart/form-data` with a generated boundary) unless the user set one manually.

### Response View
One of the desktop response viewer's tabs: **Raw** (as-received), **Pretty** (JSON pretty-printed, XML re-indented when parseable), **Headers** (key/value table), **Tree** (recursive expand/collapse JSON tree), or **Cookies** (parsed `Set-Cookie` response headers). A search box filters the active view; the JSONPath query bar replaces the body area with query matches while active.

### JSONPath Query
A dependency-free evaluator (`frontend/src/lib/jsonpath.ts`) over the response's JSON body: `$` root, dot/bracket segments (`$.user.name`, `$['users'][0]`), wildcard `*`, and array indexes. Returns a match list with canonical paths, or a specific per-segment error for invalid paths; zero matches render as an empty state, not an error.

### Response Cookie
A cookie parsed from a `Set-Cookie` response header (RFC 6265 §5.2) for display: name, value, domain, path, expiry (normalized from `Expires` or `Max-Age`), and `Secure`/`HttpOnly`/`SameSite` flags. Milestone 14 is display-only — there is no cookie jar, so cookies are not persisted or replayed on later requests (separate roadmap item).

