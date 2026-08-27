# Domain Glossary

> **Related:** [`ROADMAP.md`](ROADMAP.md) (core + CLI milestones), [`docs/internal/gui-roadmap.md`](docs/internal/gui-roadmap.md) (desktop GUI milestones), [`docs/internal/frontend-design-review-2026-08-23.md`](docs/internal/frontend-design-review-2026-08-23.md) (design review)

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
One of the desktop response viewer's tabs: **Raw** (as-received), **Pretty** (JSON pretty-printed, XML re-indented when parseable), **Headers** (key/value table), **Tree** (recursive expand/collapse JSON tree), **Cookies** (parsed `Set-Cookie` response headers), or **Table** (tabular rendering of JSON arrays / CSV; disabled unless the body is tabular). A search box filters the active view; the JSONPath query bar replaces the body area with query matches while active.

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

### Dynamic Value
A generated, non-persistent value produced at request-send time via a template tag (e.g. `{{$uuid}}`), distinct from `variables` scopes. Generated per occurrence (each `{{$tag}}` placeholder in one request yields a fresh value) and interpolated through the same `variables.Interpolate` pass that resolves URLs, headers, query params, bodies, auth config, and scripting scopes. Stored in history as the resolved bytes, not the raw tag; the request file on disk retains the raw `{{$tag}}` template.

### Template Tag
The `{{$name}}` syntax for dynamic values (Postman-compatible `{{$` prefix, zero-arg for M23; space-separated args reserved for future parametric tags e.g. `{{$randomInt 1 100}}`). Five built-ins for M23: `{{$uuid}}` (v4), `{{$timestamp}}` (unix sec), `{{$isoTimestamp}}` (ISO8601), `{{$randomInt}}` (0-1000), `{{$randomString}}` (8-char alphanumeric). Unknown tags are left literal with a non-blocking `saveWarnings` ("Unknown dynamic tag `{{$unknown}}` will be sent as-is"); args are ignored in M23 and generate the default range; custom tags are out of scope for M23.

### Tag Generator
The internal `variables` seam that produces dynamic values: `Generate(tag string, args []string) (string, bool)` where `bool` indicates a known tag. The default generator uses `uuid.New`, `time.Now`, `math/rand`; tests inject a `fixedGenerator` for deterministic `FixedUUID` etc. Custom `RegisterTag` is deferred — the interface is in place for a future plugin registry while M23 ships only the 5 built-ins.

### Dynamic Tag Picker
The desktop affordance for inserting template tags: a `{{$}}` pill button beside URL/Body/Params/Headers editors plus `{{$` autocomplete (filtering the 5 tags) — typing inserts the tag at the cursor. History stores resolved values, so the picker never persists beyond insertion; the tag remains as `{{$name}}` in the request file.

### Parametric Tag (deferred)
Space-separated args for a tag (`{{$randomInt min max}}`, `{{$randomString length}}`) — parsed but ignored in M23 (generates default). Tracks as M23b follow-up; the regex `\{\{\$(.*?)\}\}` already captures args for future use without changing `Interpolate` seams.

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

### Code Generation
Turning a resolved `Request` (method, URL, headers, body, auth) into a snippet for another client: `cURL` (`curl --request --header --data-raw/--form/--data-binary`), `JavaScript` (`fetch`), `Python` (`requests`), `Go` (`net/http`). Generated via `internal/exporter.Generate(req Request, lang string, mask func(string) string)` (beside `postman.go`) — pure function, no network, reuses `request.Request` directly; secrets render as `[SECRET]` with a comment, never plaintext. Single-request only for M24 (history entry or file-backed draft or scratchpad); collection bulk and signing (`aws/edgegrid/oauth2/digest`) are deferred.

### Code Export (CLI)
`reqly export code <request-file|collection-path> --lang cURL|js|python|go [--out <file>]` (like `reqly run`) — resolves the request through the workspace/env chain, then `exporter.Generate` to stdout (or `--out` file). `--env` respected, secrets masked. History entry can also be exported via `reqly history show <id> | reqly export code` (resolved bytes).

### Copy as (Desktop)
The desktop “Copy as ▾” affordance (RequestEditor header bar + HistoryView row) that copies the generated snippet to the clipboard via `copyText` (no file download for M24). Shares `internal/exporter` via the `HistoryAdapter`/`RequestAdapter` pattern; the picker offers `cURL` (default), `JavaScript`, `Python`, `Go`. The ResponseViewer has plain Copy/Copied for the body only — it is not part of the Copy-as picker.

### Exporter
The `internal/exporter` package seam for sharing request shapes: `postman.go` (Postman v2.1) and `code.go` (`Generate` for code generation). The highest seam for code generation — `exporter.Generate` is the single pure function both CLI and desktop call; `request`/`variables` stay untouched.

### Golden File
A `exporter/testdata/<lang>.golden` fixture for table-driven `TestGenerate_<Lang>` — input `request.Request` fixture → expected snippet literal. Prior art `postman_test.go`; deterministic, no `rand`, no network.

### Workspace Save
Bulk in-place persistence of a `Workspace` to its root: writes `reqly.yaml` + `collections/<coll>/reqly.yaml` + `collections/<coll>/requests/*.yaml` (or `.json` per `isJSONPath`) via `requestfile.Save` (format-preserving, atomic temp-file + rename, `0644`, `0600` for secrets) and `collections.SaveWorkspace` (creates `collections/` + `environments/` dirs as needed). Prunes `collections/<coll>` dirs/files that no longer exist in `ws` (delete on disk). No `changed-on-disk` version check for bulk (bulk is “write what’s in memory”); per-file `WorkspaceSaveRequest` keeps version check.

### Workspace Export
Copying a `Workspace` to a new directory via `SaveWorkspace(out, ws)` after `LoadWorkspace(src)` — `reqly export workspace [src] --out <dir>` (src `.` when omitted, `--out` required; no `--out` → error, in-place bulk is `SaveWorkspace` directly). Reuses `SaveWorkspace` — no `tar.gz` in M25 (M25b adds `archive/tar`). Git-native, plain-text, like `Save`.

### SaveWorkspace
The `internal/collections.SaveWorkspace(root string, ws *Workspace) error` seam beside `LoadWorkspace` — writes descriptors + request files, creates dirs, `0600`/`0644`, atomic, prune deleted. Highest seam for save/export — `export workspace` is thin Cobra wrapper (`LoadWorkspace(src)` → `SaveWorkspace(out, ws)`), no new `internal/exporter` code.

### Docs Generation
Generating Markdown documentation from a `Workspace` (`collections` + `requestfile` only for M26): `reqly docs generate [src] --out <dir> [--env <name>]` writes `<out>/index.md` (collections list) + `<out>/<coll>.md` per collection via `text/template` (method/URL/headers/query/body/auth + `cURL` example via `exporter.Generate` `curl` with `[SECRET]` masked). Shows raw `{{var}}` as in file plus resolved `cURL` block (when `--env` set via `environments.ResolveSet`). No `openapi`/`history`/`GraphQL` in M26 (M26b adds HTML/`goldmark`/`openapi`).

### Docs Generator
The `internal/docs.Generate(outDir string, ws *Workspace, env string) error` seam in new `internal/docs` package (beside `exporter`, pure function, `LoadWorkspace` + `flattenWorkspace` + `template` + `exporter.Generate` for `curl`, `os.MkdirAll` + atomic `WriteFile`, `0644`). Highest seam for docs — CLI `docs generate` is thin wrapper, desktop UI is M26b.

### Docs Golden File
A `docs/testdata/<collection>.golden` fixture for table-driven `TestGenerate` — `Workspace` fixture (2 colls, `{{var}}` + `auth` + `body`) → expected Markdown `index.md` + `<coll>.md` literals. Prior art `exporter/postman_test.go` + `collections/save_test.go`; deterministic, no `rand`, no network.

### HAR
HTTP Archive 1.2 (`log.entries[]` with `request`/`response`/`timings`/`startedDateTime`): `request: {method,url,headers,cookies,queryString,postData:{mimeType,text,encoding,params}}` + `response: {status,headers,content:{text,encoding,mimeType}}`. Browser DevTools `Copy as HAR` / `Export HAR` produces it; Reqly consumes it for import and produces it from history for sharing. Stored on disk as plain JSON (`.har`), never as workspace YAML.

### HAR Import
`reqly import har <har-file> [--out <dir>] [--collection <name>]` (default `--collection har-import`): parses HAR JSON, maps each `log.entries[i].request` to a `RequestEntry` file (`collections/<name>/<method>-<host>-<path>.yaml`, deduped `get-users-2`), `headers+cookies→Headers` (`Cookie:` merged), `queryString→Query`, `postData.text→Body` (base64 decoded when `encoding=="base64"`, `mimeType→Content-Type` only when no explicit header). Bodies >1MB spill to `blobs/<id>.bin` via `request.body: {file:"./blobs/..."}`. Unknown `pageref`/`timings`/`cache`/`_resourceType` dropped with `unsupported-feature` warning (like `curl`/`openapi`).

### HAR Export
`reqly export har [--out <file.har>] [--env <name>] [--limit 500]` (default stdout when `--out` absent): serializes `history.Store` entries (filtered by `env` partition) into HAR `log.entries[]` via `internal/exporter/har.go` `Export([]history.Entry) ([]byte,error)` beside `postman.go`/`code.go` (pure function, `0644` atomic). `request` from `history.Entry.ReqHeaders/ReqBody` exact bytes, `response.content.text` base64 when binary, `timings` synthesized from `DurationMS` (`send/wait/receive`), secrets masked to `[SECRET]` via `environments.MaskValues`.

### HAR Replay
Replaying a captured HAR via Reqly: `import har` materializes the captured traffic as a `har-import` collection, then `reqly collection run har-import` (or `reqly history replay <id>` for a prior `history.Entry` verbatim via `Client.Send`, `CONTEXT.md:154`). Exact replay only for M28 — no re-interpolation of `{{variables}}` against the HAR.

### Postman Import
Conversion of a Postman v2.1 collection JSON into a Git-native Reqly workspace (`internal/importer` + `reqly import postman`). Preserves requests, nested folders, collection/request variables, bodies (raw/urlencoded/form-data/graphql), and basic/bearer/apikey auth. Scripts, file-mode bodies, and unmappable auth types are reported as warnings — never silently dropped. Accepts both export shapes: bare collection object and the `{"collection": …}` envelope.

### Insomnia Import
Conversion of an Insomnia export — v4 JSON (`__export_format: 4`, flat resources linked by parentId) or v5 YAML (`collection.insomnia.rest/5.0`, hierarchical) — into a Git-native Reqly workspace (`reqly import insomnia`). Environments become native `environments/<name>.yaml` files (nested data flattened to dotted keys with warnings); cookie jars are dropped silently. Auth mapping and body conventions match Postman Import; unmappable auth types warn.

### Bruno Import
Conversion of a Bruno collection export JSON into a Git-native Reqly workspace (`reqly import bruno`). Preserves the items tree, body modes (json/xml/text/formUrlEncoded/multipartForm/graphql), and collection-level `root.request` auth/headers as collection descriptor defaults. Environment variables split by their secret flag into `variables:` vs `secrets:`. Scripts, assertions, docs, and unmappable auth types warn — never silently dropped. Directory imports are rejected with guidance to export a single JSON file.

### Import Dialog
The desktop GUI surface for import (GUI-5.1): a single staged modal — input (drop/paste), preview, results. One generic bridge call serves all formats; the format is auto-detected from content with a manual override. Preview shows what will be created plus the structured degradation report; commit writes the workspace only after explicit confirmation.

### Format Detection
Content sniffing that identifies which of the supported import formats (cURL, OpenAPI 3.x, HAR 1.2, Postman v2.1, Insomnia v4/v5, Bruno) a payload represents, independent of filename. Detection is advisory: the user can override it, and an unrecognized payload surfaces as "unknown" rather than guessing.

### Test Report
Machine-readable output of a collection run (`collection test --report-junit/--report-json`): JUnit XML for CI dashboards (testcase per step, failures carry request/assertion messages) and JSON for full detail. Reports are best-effort side artifacts — a report write failure warns and never changes the run's exit code. Secrets are masked in both.

### OpenAPI Export
Generation of an OpenAPI 3.0 YAML document from a collection or workspace (`export openapi`). Derives paths, parameters, request bodies, and security schemes from requests; response schemas are deliberately not invented (documented limitation). The inverse of Import OpenAPI.

### Endpoint Explorer
Read-only browsing of an OpenAPI spec's operations (`reqly openapi explore`) — method, path, operationId, tags, summary as a table or `--json`. Distinct from Import OpenAPI: nothing is written to the workspace; the spec is only read. Desktop explorer panel is M39b.

### Request Generation
Producing runnable native request files for selected operations of a spec (`reqly openapi generate`). Selectors are explicit (operationId, method+path, tag, or all); bodies and params are resolved to inline literals where the spec provides examples/defaults; unresolved required params stay literal placeholders rather than fake values. Sits between the Endpoint Explorer (read-only) and Import OpenAPI (whole-spec workspace write).

### JSON Schema Validation
Checking a JSON document (the *instance*) against a JSON Schema draft (`reqly schema validate`). Draft detected from `$schema` (2020-12 default), overridable per run. Violations are reported as instance paths with reasons, never silently truncated. Distinct from OpenAPI validation: it checks arbitrary payloads, not spec documents, and never sends network traffic.

### Instance Generation
Synthesizing a sample JSON document from a JSON Schema (`reqly schema generate`). Deterministic by default; explicit values win over synthesized ones (`const` > `enum` > `example` > `default`). Constraints are honored where mechanically possible; unresolvable ones degrade to warnings instead of fake precision. Feeds mock responses and test payloads.

### WSDL Import
Turning a WSDL 1.1 document into a Git-native workspace (`reqly import wsdl`): one request per operation, each a complete SOAP envelope skeleton (1.1 or 1.2 matched to the binding) POSTed to the port address with the binding's SOAPAction. Body children come from the operation's inline XSD element definitions; external schemas and exotic styles degrade to warnings rather than partial silence. Distinct from generic XML handling: Reqly has no general XML editor — the "XML builder" surface is exactly these generated envelopes.

### Import Preservation
Carrying a source collection's behavior — scripts, environments/variables, auth — into an imported Reqly workspace instead of dropping it (ADR 0026). Scripts are translated onto the `reqly.*` sandbox API; unmappable lines survive as `// TODO(reqly-import): …` comments in the file itself. Degradations are never silent: each becomes a structured Import Report entry. Distinct from import fidelity generally: preservation is about behavior that keeps working after import, not just fields landing in files.

### Script Translation
One-shot rewriting of a foreign scripting API (Postman `pm.*`, Bruno `bru.*`, Insomnia) onto Reqly's sandbox surface at import time. Variable get/set, test registration, and supported response reads translate; assertion libraries (`pm.expect`, chai) deliberately do not — they become TODO comments. Imported files contain only native syntax afterwards; there is no runtime dialect detection or compatibility shim.

### Import Report
The structured record of what an importer degraded, skipped, or translated: entries of `{item path, category, severity, message}` with categories `auth | script | body | environment | schema | other` and severities `translated | warned | dropped`. Replaces per-importer free-text warning strings; rendering belongs to callers (CLI grouped summary today, desktop dialog later).

### GraphQL Introspection
Schema discovery via the standard introspection query POSTed to an endpoint (`reqly graphql introspect`). Renders a text summary — root query/mutation/subscription fields first, remaining types alphabetically with wrapped type references (`[User!]`) — plus `--json` raw output and `--type` filtering. Distinct from sending queries (GraphQL body support, ADR 0013): introspection reads the schema, never user data.

### JWT Tooling
Offline JWT inspection and creation utilities (`internal/jwt` + `reqly jwt`). Distinct from `JWT Auth` (per-request HS signing via `auth.type: jwt`): tooling never sends a request, it decodes/inspects or locally signs a token string. M29 ships decode only; sign/verify are M29b.

### JWT Decode
`reqly jwt decode <token> [--json]` — base64url-decodes a JWT's header and payload without verification, pretty-prints each as JSON, reports `alg` + signature presence, and detects expiry via `exp`/`nbf`/`iat` (expired / not-yet-valid / time-to-expiry). Works for any `alg` (`HS*`, `RS*`, `ES*`, `none`) because it never checks the signature; malformed segments or non-JSON payload surface as explicit errors. No secret, no network, no masking needed — the token string itself is the input.

### JWT Verify (deferred)
`reqly jwt verify <token> --secret <s> [--alg HS256]` — HS256/384/512 HMAC verification reusing `internal/auth/jwt.go:79` `jwtHashes`, like `JWT Auth` but offline. Deferred to M29b; decode stays algorithm-agnostic.

### JWT Sign (deferred)
`reqly jwt sign --secret <s> [--alg HS256] [--claims '{"sub":"u1"}'] [--expiresIn 3600]` — produces a compact JWS via the same `signJWT` seam as `JWT Auth`. Deferred to M29b; no CLI flags in M29.

### Pagination Runner
Iteratively executes a paginated request, advancing the pagination variable(s) per step and collecting responses until a stop condition. M30 supports four strategies via declarative `pagination: {strategy, param, nextPath, maxPages}` on a request or collection: `page` (`?page=1` + `pageSize`), `offset` (`?offset=0` + `limit`), `cursor` (`?cursor=<nextCursor>` extracted via JSONPath `$.nextCursor` from prior body), and `link-header` (`Link: <url>; rel="next"`). Loop stops on empty body, missing next, status ≠2xx, or `maxPages` (default 100). No aggregation export for M30 — responses are streamed per step via the runner.

### Pagination Strategy
One of `page|offset|cursor|link-header`. `page` increments `page` param, `offset` adds `limit` to offset, `cursor` replaces the cursor param with the value at `nextPath`, `link-header` follows the `rel="next"` URL. All strategies reuse variable interpolation (`{{page}}` etc.) and the existing runner history/cookie seams.

### Pagination Stop Condition
When the loop terminates: no `next` value (cursor/link empty), empty array body, non-2xx, or `maxPages` reached. M30 no `while` expression (defer); stop is structural, not scripted.

### Pagination Aggregation (deferred)
Concatenating paginated JSON arrays into one result (`--out` / `aggregate: true`) — deferred to M30b; M30 streams per-step results only, matching the collection runner `OnStep` callback.

### Bulk Runner
Executes one request repeatedly against many input rows (CSV header→values or JSON array of objects), interpolating each row's fields as `{{var}}` via `variables` scopes. M31 supports CSV+JSON MVP, sequential default, parallel with `--parallel` + `--concurrency N` (semaphore, output ordered), `--continue-on-error` flag. Variable/generated dataset inputs deferred to M31b.

### Bulk Input Row
One map of variables per iteration (`map[string]string`), e.g. CSV row `id,name` → `{"id":"1","name":"a"}` or JSON object `{"id":1}` → stringified values. Interpolated per send via `variables.ScopeRuntime`.

### Bulk Concurrency
Parallel execution via semaphore `concurrency` (default 5 when `--parallel`, 1 sequential); results streamed via `OnStep` but collected in input order for reporting, matching `collection run` streaming.

### Retry Policy
Declarative per-request config (`request.retry: {count, delayMs, strategy, maxDelayMs, retryOn}`) governing automatic re-sending of a failed request. Lives in the request engine (`Client.Execute`) so every surface (CLI, desktop, runners) inherits it. Auth refresh/digest-challenge re-sends are orthogonal — they resolve within a single attempt and never consume retry budget.

### Retry Attempt
One full send of the request, including any in-attempt auth refresh or digest challenge. `count` is retries after the first attempt (total sends = count + 1). Only the final attempt is recorded in history, tagged with its attempt number.

### Backoff
Delay computation between attempts: `fixed` (constant `delayMs`) or `exponential` (default; `delayMs` doubled per attempt, capped at `maxDelayMs`, default 30000). A `Retry-After` header on 429/503 overrides the computed delay, clamped to `maxDelayMs`. No jitter for M32 (deterministic = testable). Context cancellation aborts mid-wait immediately.

### Retryable Response
Classification of a failure as worth retrying: network errors always, plus responses whose status is in `retryOn` (default `429, 502, 503, 504`; bare `500` excluded). All methods are retried regardless of idempotency — users opt in per request file.

### Execution Pipeline
The single path a send travels through the core: environment selection, variable layering, engine execution, history recording, secret masking, acquired-token capture. Owned by `core.RequestService.Run`; front-ends (CLI, Desktop, MCP) parse input and render output but never re-implement pipeline steps or receive unmasked secrets.

### Send Fidelity
The guarantee that a logical send behaves identically regardless of entry point (single run, runner step, replay): same environment precedence, same Cookie Jar attachment, same interpolation, same history recording. Divergent per-entry-point behavior is a defect, not a mode.


### Send Cancellation
User-initiated abort of a single in-flight send before its response arrives (desktop editor Stop button). The cancelled attempt leaves zero artifacts — no history row, no cookie ingest — because the execution pipeline records only successful engine returns. Cancel-after-finish is a silent no-op; a late response is discarded client-side by the tab's send token.

### gRPC Call
An invocation of a remote procedure addressed as `/package.Service/Method` against a `host:port` endpoint, carried in the request file's `grpc:` config block (service, method, JSON message, timeout) alongside the standard auth/headers/variables machinery. Unary and server-streaming shapes are supported; headers act as gRPC metadata; a non-OK gRPC status renders as a failed response (code + message), not a body.

### Server Reflection
A gRPC server's self-description service that lets the client discover services and message schemas at runtime without any local proto files. The primary schema source for a gRPC Call; explicit workspace-relative `.proto` paths in the `grpc:` config are the fallback for reflection-disabled servers.

### Message
The protobuf payload of a gRPC Call expressed as JSON in both directions (canonical protobuf-JSON mapping). The unit the editor edits, the response viewer renders, and scripting/assertions consume — never raw wire bytes.

### Design System
The desktop visual language defined in `DESIGN.md` and materialized as semantic CSS custom properties in `frontend/src/index.css` (`@theme` → Tailwind `bg-background` etc.). Covers tokens, typography, color, and status presentation; adding theme N+1 touches only `index.css` (`[data-theme="…"]` block) and the registry in `frontend/src/lib/themes.ts` — zero component changes (ADR 0029).

### Design Tokens
Semantic CSS custom properties on `:root` / `[data-theme="atlas-light|atlas-dark"]` and `.dark` appearance mirror: surfaces (`background`/`card`/`popover`/`muted`/`secondary`/`accent`), `border`/`input`/`ring`, `primary` (terracotta), `status-*` ramp, `radius` (6px base), fonts. `index.css` is the single source of truth; components consume only `var(--…)` / Tailwind `bg-*` — no hardcoded hex outside `index.css` (grep gate).

### Typography System
IBM Plex Sans (UI/body, 400/500/600) + IBM Plex Mono (data/code/numbers, 400/500 + `tabular-nums`), bundled offline via `@fontsource` (local-first, no CDN). Base `13px/1.45` on `body`; scale `xs 11 / sm 12 / base 13 / lg 15`; `.font-data` + `code/kbd/pre` force mono. Matches Atlas reference screenshots.

### Color System
Terracotta brand anchor (`--primary`) used sparingly (~10%): primary buttons, active accents, `ring`. Light `--primary: #c93517` is the AA-adjusted variant of brand `#e14b31` (4.0:1 → 4.5:1 on white with white foreground, documented in ADR 0029) — dark `--primary: #ff6f52` on `#0d1015`. Surfaces are near-neutral (`#fbfbfa` light, `#0d1015` blue-black dark). `prefers-contrast: more` bumps `--border` to `#8a9099` and `--muted-foreground` to foreground.

### Status Ramp
Semantic color scale for HTTP/method states, WCAG-checked pairs: `status-info` gray (1xx/neutral), `status-ok` green (2xx), `status-redirect` blue (3xx), `status-warn` amber (4xx/warnings), `status-error` red (5xx/failures). Method tints: GET green, POST blue, PUT/PATCH amber, DELETE red. Exposed as `--status-*` + Tailwind `bg-status-*` / `text-status-*` + `success`/`warning` aliases.

### StatusPill
The single memorable component (`frontend/src/components/status.tsx`) rendering dot + tabular-numeral code/text, never color alone — used identically in response header, run steps, and history rows. Variants `info|ok|redirect|warn|error` map to the Status Ramp; dot + literal satisfies a11y "never color alone."

### Hairline Border
Defined-edge separation via `1px solid var(--border)` (`#e4e6e9` light / `#252b35` dark; `#8a9099` in high-contrast), not shadows. Shell, cards, and panels use hairline borders only; `shadow-md/lg` + `ring-1 ring-foreground/10` is permitted solely on floating layers (`popover`, `dropdown-menu`, `select`, `toast`) to escape the page — lint forbids `shadow-*` elsewhere.

### Navigation Model
Two-axis desktop navigation defined in spec §4: horizontal axis = Tool Rail (top-level tool selection: Workspace / API Tools / Realtime / System), vertical axis = Context Sidebar (resource tree/search/actions scoped to the active tool). Main Workspace renders the active tool's full page; Bottom Utility Panel is cross-cutting. Toggle `⌘B` collapses sidebar, `⌘J` collapses bottom panel — both persisted.

### Navigation Map
The 15+ full pages of the app (spec §60), each a top-level route in Main Workspace: Home, Requests, Environments, History, Mocks, Diff, JWT, GraphQL, gRPC, Runners, Explorer, Docs, WebSocket, SSE, Settings — each with sub-panels (e.g. Requests → Params/Headers/Body/Auth/Tests/Response). Pages are lazy-loaded (`React.lazy`) per tool; state is preserved per tab where applicable.

### Page vs Panel Rules
Spec §62 distinction governing surface choice: **Page** = full Main Workspace route per tool (e.g. Request Builder, Environments Manager); **Context Panel** = sidebar resource navigation scoped to the active tool (e.g. Collections tree, History filters); **Bottom Panel** = cross-cutting inspectors (Console/Network/Tests/Variables/Cookies, `⌘J`); **Dialog** = transient action (Import/Export/Create Collection, confirm destructive). Determines routing, persistence, and focus management.

### Shared Patterns
Cross-tool interaction primitives (spec §61) reused everywhere: global Command Palette (`⌘K`) + per-tool filter input, primary (coral) / secondary (neutral) action hierarchy, StatusPill dot+code indicators, tab primitives (`components/ui/tabs.tsx`), and panel chrome (hairline borders, header + content). Ensures consistent density and keyboard parity.

### Final Layout Model
Canonical five-zone shell (spec §63, `DESIGN.md` Layout) as single source of truth: **TopBar** (workspace pill, `⌘K`, env, sync) → **Tool Rail** (48–56px, 4 groups) → **Context Sidebar** (220–280px, resizable/collapsible) → **Main Workspace** (tab-based page routing) → **Bottom Utility Panel** (resizable, `⌘J`). Persisted layout (sidebar collapsed, bottom height, active tool) via localStorage; all chrome consumes Design Tokens only.

