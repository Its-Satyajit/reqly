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

### Environment Manager
The desktop surface for managing a workspace's environments (list, create, edit, delete, set active). Edits are held in memory as an **Environment Draft** until explicitly saved, which writes the `environments/<name>.yaml` file back to disk — the filesystem remains the source of truth.

### Environment Draft
The in-memory, unsaved edit state of an **Environment** in the desktop Environment Manager. Editing a secret into the draft marks it changed but never reads an existing secret's plaintext back into the UI; a changed secret is written on save, an unchanged one is left as-is.

### Environment Adapter
The frontend seam through which the desktop UI reads and writes environments. The host injects a Wails-backed adapter into the shared frontend; browser dev mode uses a read-only fallback, mirroring the request/auth adapter pattern.

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

### AWS Signature V4
A per-request signing scheme (`internal/auth/aws.go`, `auth.type: aws`) for AWS APIs: HMAC-SHA256 over the canonical request (method, URI, query, headers, payload hash) with `accessKey`/`secretKey`/`region`/`service` (plus optional `sessionToken` for STS), producing `Authorization: AWS4-HMAC-SHA256 ...` + `X-Amz-Date`/`X-Amz-Content-Sha256`.

### Akamai EdgeGrid
A per-request signing scheme (`internal/auth/edgegrid.go`, `auth.type: edgegrid`) for Akamai OPEN APIs: HMAC-SHA256 over timestamp + method + host + path with `clientToken`/`clientSecret`/`accessToken`/`host`, producing `Authorization: EG1-HMAC-SHA256 ...`.

### Request Builder Tabs
The desktop request editor's tab bar (**Params / Headers / Auth / Body / Variables**) from milestones 14 and 19. Params and Headers are key-value row editors (add/remove/toggle-enabled); the Auth tab edits the request's own auth (see **Auth Editor**); the Body tab holds a body-type picker. The send path maps the tabs onto `request.Request{Query, Headers, Auth, Body}`; the engine's existing merge semantics apply (query params override the URL's, a manual `Content-Type` header beats the body type's default).

### Auth Editor
The desktop request-editor surface (the **Auth** tab) for editing a request's own authentication. It shows a scheme picker (**Inherit / No Auth / Basic / Bearer / API Key / JWT / Digest / OAuth 2.0 / AWS SigV4 / Akamai EdgeGrid**) with per-scheme field forms; sensitive config values are plaintext inputs flagged as sensitive, mirroring the Git-native request file. Distinct from the sidebar **Auth Panel**, which manages the workspace's OAuth token lifecycle (login/status/logout); the editor edits grant *config* only.

### Auth Draft
The in-memory, editable auth state of a **Request Tab** — one of **Inherit** (the request declares no auth and receives its containers'), **No Auth** (`auth.type: none`, which explicitly disables inherited auth), or a typed scheme with config values. It is dirty-tracked like any builder field and written to the file on save: choosing **Inherit** *removes* any existing auth block, so the file truly has none.

### Inherited Auth
The effective auth a request receives from its container chain (workspace → collection → folder) under **Auth Inheritance**, shown read-only in the **Auth Editor** when the request has no own auth — mirroring the inherited-headers group in the Headers tab.

### Body Type
A request-body kind selected in the builder's Body tab: **none / JSON / XML / form-data / urlencoded / raw text / binary / GraphQL**. JSON/XML/raw/GraphQL use a CodeMirror editor with the matching language; form-data and urlencoded use key-value rows (form-data rows can carry a file). Binary is a single file picker. Selecting a type sets the appropriate `Content-Type` (or `multipart/form-data` with a generated boundary) unless the user set one manually.

### Binary Body
A body kind for single-file uploads (`BodyType: binary`): a file picker (and drag-drop) stores a Git-native relative file path (`request.body: { file: "./path" }`), `serializeBody` reads the file at send and sends its bytes with `Content-Type: application/octet-stream` (or the file’s mime).

### GraphQL Body
A body kind for GraphQL (`BodyType: graphql`): two editors — query (`graphql` lang) and variables JSON — stored as structured `request.body: { query, variables }`, `serializeBody` JSON-stringifies them to `{"query","variables"}` with `Content-Type: application/json`.

### File Upload
Attaching a file to a request: `binary` as a single file body, or `form-data` rows with `file` + optional `filename` (each row toggles between text value and file path), producing `multipart/form-data` via `boundaryFor` (`frontend/src/lib/body.ts:32`). File paths are Git-native relative and resolved at send time.

### Response View
One of the desktop response viewer's tabs: **Raw** (as-received), **Pretty** (JSON pretty-printed, XML re-indented when parseable), **Headers** (key/value table), **Tree** (recursive expand/collapse JSON tree), or **Cookies** (parsed `Set-Cookie` response headers). A search box filters the active view; the JSONPath query bar replaces the body area with query matches while active.

### JSONPath Query
A dependency-free evaluator (`frontend/src/lib/jsonpath.ts`) over the response's JSON body: `$` root, dot/bracket segments (`$.user.name`, `$['users'][0]`), wildcard `*`, and array indexes. Returns a match list with canonical paths, or a specific per-segment error for invalid paths; zero matches render as an empty state, not an error.

### Response Cookie
A cookie parsed from a `Set-Cookie` response header (RFC 6265 §5.2) for display: name, value, domain, path, expiry (normalized from `Expires` or `Max-Age`), and `Secure`/`HttpOnly`/`SameSite` flags. Milestone 14 was display-only; Milestone 22 adds the **Cookie Jar** (see below).

### Cookie Jar
A per-workspace, `env`-partitioned persistent store for `Response Cookie`s in the same SQLite database as history (`<workspace>/.reqly/history.db` `cookies` table: `name, value, domain, path, expires_at, secure, http_only, same_site, env`). Populated automatically from `Set-Cookie` on each response (via `net/http` `ReadSetCookies`, RFC 6265), queried by `request.Client` before send for domain/path/secure matching (auto-attached as `Cookie:`), and shown in the desktop response Cookies tab with delete/clear controls. Partitioned by `env` so **Clear per environment** (`DELETE WHERE env=?`) and **Clear per workspace** (`DELETE WHERE workspace=?`) are distinct. Edit is delete+re-add; a per-request opt-out (`CookieJar: false`) is available.

### Table View
A response-viewer tab for tabular bodies: a **JSON array of objects** (union of keys → columns, first 1000 rows virtualized, client-side search/sort) or **CSV** (`encoding/csv`). The Table tab is always visible but disabled with a hint when the body is not tabular (needs JSON array or CSV, detected via `Content-Type` + `JSON.parse` probe). CSV pretty is the Table itself; no separate CSV pretty.

### Binary Preview
How non-JSON bodies are shown: `image/*` renders inline via `data:` URL, `application/pdf` shows a download banner, other binary shows a hex dump (first 4KB) + download. This covers the remaining `ROADMAP.md:74` response body parsing (XML/HTML/CSV/binary) — JSON stays Pretty/Tree, XML stays Pretty-indented, CSV uses Table, binary uses this preview.

### History Entry
One row per executed request in the per-workspace SQLite history (`<workspace>/.reqly/history.db` `history` table): `id, request_path, method, url, env, status, duration_ms, size, req_headers_json, req_body_path, resp_headers_json, resp_body_path, created_at` plus FTS5 on `url, request_path`. Bodies >1MB spill to `<workspace>/.reqly/history/blobs/<id>.bin`; otherwise inline. The file is 0600 and `.reqly/` is `.gitignore`d. Stored bytes are exact (for faithful replay) but masked on display via `environments.MaskValues`.

### History Store
The `internal/history` persistence layer over `modernc.org/sqlite` (pure-Go, no CGO, WAL mode): `List`, `Show`, `Search` (FTS5), `Clear`, and `EnforceRetention` (keep last 500 per workspace, prune oldest). Shared by Desktop and CLI via `internal/core`.

### Replay
Re-sending a **History Entry**'s fully-resolved request verbatim (`Client.Send(storedReq)`) via the History view's **Replay** button or `reqly history replay <id>`. Exact replay only for M22 — no re-interpolation of `{{variables}}`; a future `history replay --env <name>` can target another env.

### History View
The desktop surface for local history: sidebar **History** entry → table (`time ↓, method, url/path, status, duration, env`, paginated 50/page, status filter 2xx/4xx/5xx, FTS search, Clear button). Clicking a row shows masked request/response with a Replay button; replay uses the History Entry's stored request.

### History Adapter
The frontend seam through which the desktop UI reads history and triggers replay/clear. Host injects a Wails-backed adapter; browser dev mode uses a read-only fallback, mirroring the request/auth/environment adapter pattern.

### Workspace
The top-level Git-native unit of Reqly: a directory containing a `reqly.yaml` descriptor plus optional `collections/` and `environments/` directories. Discovered by walking up from the current directory to the nearest descriptor. It owns the workspace-level configuration (base URL, headers, auth, variables, active environment) that its collections inherit.

### Collection
A named group of requests under a workspace's `collections/` directory. A collection is a directory with its own `reqly.yaml` descriptor and may nest **Folders**. It contributes configuration (base URL, headers, auth, variables) that its requests and folders inherit, with child values overriding parent values.

### Folder
A nested container inside a collection (recursively nestable), also a directory carrying a `reqly.yaml` descriptor. Only subdirectories with a descriptor are part of the workspace; a descriptor-less subdirectory is ignored.

### Request Entry
A request file (`.json`/`.yaml`/`.yml` in `requestfile` format) located directly inside a collection or folder. It is the leaf of the workspace tree; there are no inline requests — the file *is* the request, and its position in the tree defines its identity.

### File Request
The raw, unmerged request a **Request Entry** declares (url/method/headers/query/body — the *builder fields*) plus the file's own auth, timeout, and request-level variables. It is the editor seed: a **Request Tab** opens from the File Request, and only its builder fields are editable. Everything else is preserved verbatim when the tab is saved.

### Request Path
A workspace-relative identifier locating a **Request Entry**, e.g. `users/auth/login` (the file name minus its extension). This is the stable identity used to find a request, open it, and deduplicate request tabs.

### Inherited Configuration
The effective base URL, headers, and auth a request receives from the **Workspace → Collection → Folder** chain before its own fields apply: headers merge key-wise (child wins), base URLs join (an absolute child replaces the parent), auth is replaced when the child defines one and cleared by `auth.type: none`. Inherited values are resolved lazily so `{{variable}}` placeholders survive until interpolation.

### Resolved Request
A **Request Entry** combined with its **Inherited Configuration** and the full variable chain (workspace → collection → folder → request scopes), ready for execution. Opening a request in the desktop **Collections Browser** loads its Resolved Request alongside the raw **File Request**: the resolved form drives display (Effective URL line, inherited-headers group, Variables tab) while the editor edits the File Request's builder fields and **re-resolves** the live draft through the inheritance chain at send time.

### Collections Browser
The desktop sidebar surface that loads the workspace and renders its collections, folders, and request entries as an expandable/collapsible tree. Clicking a request opens its **File Request** into an editable **Request Tab**; a manual refresh reloads the tree from disk. File-backed tabs can save their builder fields back to the file (format-preserving, atomic) with dirty tracking and changed-on-disk conflict handling.

### Request Tab
The desktop's per-request working area, one per opened request (deduplicated by **Request Path**, plus a persistent *New Request* scratchpad for ad-hoc sends). File-backed tabs edit a draft of the **File Request**'s builder fields and can save them to disk; each tab shows its response, a read-only effective-variables view, a live **Effective URL** line, an inherited-headers group, and an environment pill showing which environment the tab will send with — the request file's `environment:` if set, else the header-selected one. Sends re-resolve the draft through the inheritance chain at send time; the scratchpad sends raw.

### Collection Run
Executing every **Request Entry** in a **Collection** (or a **Folder**) in deterministic, name-sorted order through the shared runner engine (`internal/runner`). Steps share one variable store across the run, so a post-request script can extract a token that a later request interpolates; pre/post scripts and `reqly.test()` assertions are evaluated per step. A run reads **fresh from disk** — it is a statement about the workspace, not about the editor session, so unsaved **Request Tab** drafts are never part of a run. Runs are sequential and single-flight in the desktop: only one run at a time.

### Run Step
One **Request Entry** as executed within a **Collection Run**: its request name and **Request Path**, whether it passed (request succeeded *and* every test passed), the transport/pre-script error if any, the response received, the `reqly.test()` results, and the script console logs. Credential values are resolved per step for masking and never serialized.

### Run Report
The aggregate result of a **Collection Run**: the ordered **Run Steps**, start/finish times, totals (total/passed/failed), and duration. The desktop receives it streamed — one event per completed **Run Step** for live progress, then a final event carrying the complete report.

### Run View
The desktop surface that presents a **Collection Run**: a dedicated tab (distinct from **Request Tab**) showing each **Run Step** with live status, status code, duration, expandable tests/logs/response, a fail-fast toggle, and a cancel control for the in-flight run. Clicking a **Run Step** opens its request into a normal **Request Tab** for inspection without stopping the run.

