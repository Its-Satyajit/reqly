# Phase 1: Core API Client (P0)

## Phase 1 — Core API Client (P0)

The minimum set to make Reqly a serious API client.

### 1.1 Request engine (foundation for everything)

- [x] `internal/request` — full HTTP request model (URL, method, path/query params, headers, body, auth, certs, proxy, settings)
- [x] Request engine: HTTP/1.1 transport, timeouts, redirects, compression
- [x] Request execution shared by Desktop + CLI (single engine, no duplication)
- [x] Response model: status, headers, cookies, timing, size, raw body
- [x] Response body parsing — JSON (pretty/tree) + XML (pretty) + CSV (Table) + binary (image inline, PDF banner, hex 4KB) via `frontend/src/lib/response.ts:187` `isTabular/parseTable/binaryPreviewType` and `ResponseViewer` Table tab ([ADR 0014](docs/adr/0014-history-cookie-jar-table-view.md)); file download via `suggestedFilename`
- [x] File upload / multipart — `BodyType: binary` (single file, `application/octet-stream`) + file-aware `form-data` rows (`file` + `filename`, `multipart/form-data` via `boundaryFor`, [ADR 0013](docs/adr/0013-binary-graphql-body.md)); file download pending
- [x] SQLite local metadata — per-workspace `<workspace>/.reqly/history.db` (`modernc.org/sqlite` WAL, FTS5, `history` + `cookies` tables, 1MB spill to `blobs/`, 500 retention, `0600`), history search/replay (`internal/history` + `core.HistoryService` + `reqly history` + desktop History view), request replay exact via `HistoryReplay` ([ADR 0014](docs/adr/0014-history-cookie-jar-table-view.md))

### 1.2 Variables & environments

- [x] 6 variable scopes shipped (global, environment, collection, folder, request, runtime + process-env via `.env`; prompt/runtime fully wired) — request files carry `variables` maps; 8-scope model tracked in CONTEXT.md
- [x] `{{key}}` interpolation wired through request builder + scripting — works in `run`/`test`/`collection` via request files
- [x] Environment management — `internal/environments` + `reqly env list/show/use` (Git-native `environments/<name>.yaml`, `REQLY_ENV`/`--env`/file/descriptor selection precedence) + desktop Environments UI ([Milestone 15](https://github.com/Its-Satyajit/reqly/issues/84))
- [x] Environment validation — `reqly env validate` (file syntax, secret-name + duplicate-key warnings, undefined-variable detection across workspace request/test files)
- [x] Dynamic values & template tags — `internal/variables` `TagGenerator` (`{{$uuid}}` v4 + `{{$timestamp}}` unix + `{{$isoTimestamp}}` ISO8601 + `{{$randomInt}}` 0-1000 + `{{$randomString}}` 8 alphanum), `{{$` strict vs `{{` variables, per occurrence fresh, unknown left literal with `saveWarnings`, `TagPicker` picker + `{{$` autocomplete ([ADR 0015](docs/adr/0015-dynamic-values-template-tags.md))

### 1.2a Request files (plain-text, Git-native)

- [x] `internal/requestfile` — JSON/YAML request file format (`name`, `variables`, `request`)
- [x] `reqly run <file>` — load request + variables from file, flags override file fields
- [x] `reqly test <file>` — test files accept YAML and `variables` (interpolated at runtime)
- [x] Shared file format for collections/folders (`internal/collections` descriptor format, see §1.5)

### 1.3 Authentication

- [x] Basic, Bearer, API key — `internal/auth` scheme registry, `request.Auth` dispatch, secret masking ([ADR 0005](docs/adr/0005-git-native-auth-schemes.md))
- [~] JWT — HS256/384/512 per-request signing + `reqly jwt decode` claims viewer shipped ([ADR 0021](docs/adr/0021-jwt-tooling-decode.md)); `verify`/`sign` deferred to M29b
- [~] Digest — challenge/response shipped (SHA-256 fallback, request-body aware); NTLM deferred
- [x] OAuth 2.0 Client Credentials — RFC 6749 §4.4 with store-backed token caching (`TokenSource` + `secrets.Store`, ADR 0006), expiry-skewed proactive refresh, reactive 401 refresh+retry-once, `reqly auth status`/`auth logout`
- [x] OAuth 2.0 Authorization Code + PKCE — RFC 6749 §4.1 + RFC 7636 (`AuthorizationCodeSource`, one-shot loopback callback, state/verifier, [ADR 0007](docs/adr/0007-oauth2-authorization-code-pkce.md)), `reqly auth login`, first-request auto-login, refresh-token reuse (RFC 6749 §6, proactive + 401, rotation) — spec [#52](https://github.com/Its-Satyajit/reqly/issues/52), tickets [#53–#57](https://github.com/Its-Satyajit/reqly/issues/53)
- [x] OAuth 2.0 Device flow (RFC 8628) + OS-keychain store + custom redirects + desktop auth — `reqly auth login --flow device` (verification URI + code, RFC poll semantics), `--store keychain`/`REQLY_TOKEN_STORE` with file fallback, `reqly://` deep-link callbacks, sidebar auth panel (login/status/logout) — spec [#60](https://github.com/Its-Satyajit/reqly/issues/60), tickets [#61–#65](https://github.com/Its-Satyajit/reqly/issues/61), [ADR 0008](docs/adr/0008-oauth2-auth-leftovers.md)
- [x] AWS Signature V4 — `internal/auth/aws.go` (`auth.type: aws`, SigV4 per-request signing, `accessKey`/`secretKey`/`region`/`service` + optional `sessionToken`, [ADR 0012](docs/adr/0012-aws-edgegrid-auth.md))
- [x] Akamai EdgeGrid — `internal/auth/edgegrid.go` (`auth.type: edgegrid`, EG1-HMAC-SHA256, `clientToken`/`clientSecret`/`accessToken`/`host`, [ADR 0012](docs/adr/0012-aws-edgegrid-auth.md))
- [x] OAuth 1.0 — `internal/auth/oauth1.go` (`auth.type: oauth1`, RFC 5849 HMAC-SHA1 per-request signing, `consumerKey`/`consumerSecret` + optional `token`/`tokenSecret`, `Authorization: OAuth` header with `oauth_signature`, `oauth_nonce`/`oauth_timestamp`, `auth.config` + Auth tab `OAuth 1.0` form) — 2026-08-30
- [ ] Custom auth — deferred (reuses same `auth.config` + Auth tab seams)
- [x] Auth inheritance — Workspace → Collection → Folder → Request (base URL, headers, auth, vars)

### 1.4 Secrets

- [x] Encrypted-at-rest secret storage + OS keychain — token stores behind the `secrets.Store` interface: FileStore (plain-text 0600 `.reqly/tokens.json`, default) and KeychainStore (OS keychain via go-keyring; keychain default on desktop), backend selection `--store keychain`/`REQLY_TOKEN_STORE` with graceful file-store fallback
- [x] Secret variables + masking (CLI output, logs, test output) — `environments/<name>.yaml` `secrets:` maps render as `[SECRET]`; masking wired through run/test/collection/validate/diff; acquired OAuth tokens masked post-request
- [x] `.env` support — dotenv parsing (process-env scope, OS env wins) shipped via `internal/variables` + `internal/environments`; external managers (Vault, AWS, Azure) — P3

### 1.5 Workspaces, collections & storage

- [x] `internal/collections` — workspaces, collections, nested folders
- [x] Plain-text, Git-native project files (mirror workspace → filesystem)
- [x] `internal/core` — application services layer shared by Desktop/CLI/MCP (`RequestService.Send`)
- [x] Inheritance: Workspace → Collection → Folder → Request (base URL, headers, auth, vars)
- [x] `reqly collection run <path>` + `reqly collection list` (CLI wired to the Go core)
- [x] Environments: resolve the `environment` scope from `environments/` on disk (workspace + file resolution, selection precedence)
- [x] Save/export a workspace — `internal/collections.SaveWorkspace` (bulk in-place + `reqly export workspace [src] --out <dir>` copy, `requestfile.Save` format-preserving atomic, prune deleted, `0600`/`0644`) + `reqly export workspace` CLI ([ADR 0017](docs/adr/0017-workspace-save-export.md))

### 1.5a Core → Desktop bridge (from 0.2 `Greet` proof)

- [x] `internal/core` `RequestService` — wraps `request.Client`, bridge-friendly `SendResponse` DTO
- [x] Desktop `AppService.SendRequest` delegates to core (thin Wails boundary; `Greet` removed)
- [x] Regenerated Wails bindings → `appservice.ts` `SendRequest` + `models.ts` (`Request`, `SendResponse`)
- [x] Shared `useRequestStore` + pluggable `RequestSender` (Wails bridge in host; `fetchSender` fallback in browser dev)
- [x] `RequestEditor` Send → core; `ResponseViewer` renders status/headers/pretty body
- [x] Per-tab request/response state (multiple tabs) + cancel in-flight request — per-tab state ([T2 #132](https://github.com/Its-Satyajit/reqly/issues/132)); Stop-button send cancellation via `SendOptions.sendId` + `CancelSend` binding, no history artifacts for cancelled sends ([M33](docs/spec/m33-cancel-in-flight-request.md))

### 1.6 Request builder & response viewer (UI)

> **GUI parity:** Import/Export, WebSocket/SSE, Test runner, Mock server, API diff, JWT inspector, env tools, GraphQL browser, pagination/bulk runners, and OpenAPI explorer all have desktop GUIs shipped (GUI-5–GUI-14, v1.4.0). See [`docs/internal/gui-roadmap.md`](docs/internal/gui-roadmap.md) for the full GUI milestone tracker.

- [x] Method select, URL bar, Send → real response data flow
- [x] Params/headers/body tabs in the builder
- [x] Body editors: JSON/XML/raw/binary/GraphQL via CodeMirror, form-data/urlencoded via key-value rows (file-aware `form-data` + `binary` file picker + `graphql` query+variables), auto Content-Type (manual wins) — [Milestone 14 T2](https://github.com/Its-Satyajit/reqly/issues/73) + [Milestone 21](https://github.com/Its-Satyajit/reqly/issues/189) ([ADR 0013](docs/adr/0013-binary-graphql-body.md))
- [x] Response viewer: metadata, raw/pretty/tree/table views (Table for JSON array-of-objects + CSV, 1000 rows virtualized, `isTabular` disabled hint), binary preview (image `data:` inline, PDF banner, hex 4KB), search — ([ADR 0014](docs/adr/0014-history-cookie-jar-table-view.md))
- [x] JSONPath / XPath response querying — dependency-free evaluator (`$.user.name`, `$['users'][0]`, wildcard `*`) with match list + specific errors; XPath pending
- [x] Response actions: copy (body/headers), download (Content-Disposition filename), format
- [x] Cookies: persistent jar (`history.db` `cookies` table, `env`-partitioned, `0600`, domain/path/secure/expires matching via `history.FilterCookies`, auto-attach `Cookie:` on next `SendRequest`, `Set-Cookie` ingest via `HistoryService.Record`, view + delete/clear in `ResponseViewer` Cookies tab + desktop `CookieList/Delete/Clear` bindings, CLI jar implicit) — [Milestone 14 T5](https://github.com/Its-Satyajit/reqly/issues/76) + [Milestone 22](https://github.com/Its-Satyajit/reqly/issues/197) ([ADR 0014](docs/adr/0014-history-cookie-jar-table-view.md))

### 1.7 Scripting & automation

- [x] Pre-request / post-request scripts (Goja) — `reqly` sandbox (request/response access, variable get/set, `reqly.test()`, console)
- [x] Test scripts + assertion library (core assertion engine shipped: status, header, body, JSON, response-time, schema)
- [x] Request chaining (login → extract token → next request) — runtime variables persist across collection steps
- [x] Chain runner — sequential execution, variable passing, assertions, script execution, fail-fast failure handling (conditional branching deferred to P1)
- [x] Collection runner (sequential, variable passing, assertions, fail-fast) — `reqly collection test` + desktop Run View streaming

### 1.8 Protocols (P0: REST-first, then extended)

- [x] **REST** — complete builder (see §1.1/§1.6: method/URL/headers/params/body + file upload + cookies/history)
- [x] **WebSocket** — connection mgmt, message composer, in/out inspection (`internal/websocket` + `reqly ws`)
- [x] **SSE** — live event stream, inspection, event history (`internal/sse` + `reqly sse`)
- [x] **GraphQL** — query editor + variables via `BodyType: graphql` (ADR 0013), live endpoint introspection (`reqly graphql introspect <url>`), offline SDL schema parsing (`reqly graphql parse <file.graphql>`), Goja scripting helper `reqly.introspectGraphQL()`, and Desktop Schema Browser ([M50](docs/spec/m50-graphql-schema-introspection.md), [ADR 0034](docs/adr/0034-graphql-schema-introspection.md))
- [x] **gRPC** — proto files, reflection, service/method discovery, unary + server-streaming — `internal/grpc` (reflection via v1 protocol, protocompile `.proto` fallback, TLS/h2c, deadlines), `grpc:` request-file block, scripting/assertions parity, history, `reqly grpc services|invoke`, desktop gRPC view (ADR 0028, M43; client-stream/bidi deferred)
- [~] **SOAP** — WSDL import, operation discovery, envelope skeletons: `reqly import wsdl <file> [--output dir]` ([M41](docs/spec/m41-wsdl-import.md) — one runnable POST per operation with binding-matched SOAP 1.1/1.2 envelopes, SOAPAction, inline-XSD body placeholders; external schemas/rpc-encoded best-effort with warnings; the "XML builder" surface is these generated envelopes, no runtime builder)

### 1.9 Import / export

- [x] Import cURL — `reqly import curl` (method, headers, JSON/raw/data bodies, basic auth, user-agent, cookies, GET-style query data; unsupported features reported)
- [x] Import OpenAPI 3.x — `reqly import openapi` (servers, paths, operations, params, JSON bodies; writes a Git-native workspace)
- [x] Export Postman collection v2.1 — `reqly export postman` (flat list, inherited base URL/headers applied)
- [x] Import: Postman v2.1 ([M34](docs/spec/m34-postman-import.md)), Insomnia v4/v5 ([M35](docs/spec/m35-insomnia-import.md)), Bruno ([M36](docs/spec/m36-bruno-import.md)), Swagger 2.0 / OpenAPI 2.0 (`internal/importer.ParseSwagger2` + `reqly openapi convert-v2` converter, [M51](docs/spec/m51-swagger2-importer-converter.md), [ADR 0035](docs/adr/0035-swagger2-importer-converter.md)); HAR done ([M28](docs/spec/m28-har-import-export.md))
- [x] Export: requests ([`export workspace`](docs/adr/0017-workspace-save-export.md) + `export code`), OpenAPI 3.0 spec generation (`export openapi`, [M37](docs/spec/m37-export-reports-openapi.md)), responses (`export har` from history, M28 + desktop download), test results (`collection test --report-junit/--report-json`, M37); docs done (§1.11 `reqly docs`)
- [x] Import preservation (env/auth/scripts) + unsupported-feature reporting — [M42](docs/spec/m42-import-preservation.md) ([ADR 0026](docs/adr/0026-import-preservation-script-translation.md)): Postman/Bruno pre/post scripts translated onto the `reqly.*` sandbox into `preRequest`/`postRequest` (unmappable lines preserved as `TODO(reqly-import)` comments), Postman collection variables → `environments/<collection>.yaml`, structured `ImportReport` across all importers with CLI grouped summary (2026-08-24)

### 1.10 OpenAPI & JSON Schema

- [x] OpenAPI 3.x parse + validate — `internal/openapi` (kin-openapi, JSON/YAML, $ref resolution); Swagger 2.0 / OpenAPI 2.0 import & `reqly openapi convert-v2` spec converter shipped ([M51](docs/spec/m51-swagger2-importer-converter.md))
- [~] Endpoint explorer + generate requests from spec — `reqly openapi explore <spec> [--tag]... [--json]` (operation table / machine-readable list) and `reqly openapi generate <spec> [--operation]... | [--method --path] | [--tag]... | --all [--output dir]` ([M39](docs/spec/m39-openapi-explorer.md) — native request files, inline example/default bodies+params, bearer/basic/apikey-header → native auth blocks, unmappable features warned; desktop explorer panel deferred to M39b)
- [x] JSON Schema: validate, inspect, generate & test assertion — `reqly schema validate/inspect/generate` ([M40](docs/spec/m40-json-schema.md)) and Goja sandbox assertion hook `reqly.assertJSONSchema(schemaPath)` ([M52](docs/spec/m52-json-schema-assertion.md), [ADR 0036](docs/adr/0036-json-schema-script-assertion.md))
- [x] XML/XSD schema validation where applicable — `internal/validation.ValidateXMLAgainstXSD` pure Go XSD parsing, DOM element/attribute constraint checking, local `schemaLocation` resolution, `reqly schema validate --type xml <schema.xsd> <instance.xml>`, Goja sandbox assertion `reqly.assertXSD(schemaPath)`, and Desktop UI ResponseViewer XML validation badge ([M49](docs/spec/m49-xml-xsd-validation.md), [ADR 0033](docs/adr/0033-xml-xsd-schema-validation.md))
- [~] Generate mocks from OpenAPI (see P1) — `reqly mock` serves schema/example-driven responses

### 1.11 CLI (P0 commands)

- [x] `reqly run` — send a request from the CLI (URL or JSON/YAML request file, flags override file)
- [x] `reqly test` — run tests against a request (JSON or YAML test file, variables interpolated)
- [x] `reqly collection run` — run a request in a collection with inherited config (workspace/collection/folder resolution); `collection list` shows the tree
- [x] `reqly collection test` — run every request in a collection with pre/post scripts and `reqly.test()` assertions; runtime variables chain across steps; `--fail-fast`
- [x] `reqly ws` — interactive WebSocket client (stdin sends text frames, incoming frames timestamped)
- [x] `reqly sse` — stream Server-Sent Events (named/ID'd events, multi-line data, retry hints, `--count`)
- [x] `reqly validate` — validate a project/spec (`internal/validation` + `reqly validate`)
- [x] `reqly diff` — diff specs/requests/responses (`internal/diffing` + `reqly diff`)
- [x] `reqly env` — manage environments (`internal/environments`): `list`, `show`, `use`, `validate`, `diff`
- [x] `reqly mock` — serve a mock API from an OpenAPI spec (kin-openapi parsing, path/method matching, schema/example response generation, `--delay`, `--fail-every`)
- [x] `reqly docs` — `internal/docs.Generate` (Markdown `index.md` + per-collection `<coll>.md` via `text/template` + `curl` via `exporter`, `reqly docs generate [src] --out <dir> [--env]`, `0600`/`0644`) ([ADR 0018](docs/adr/0018-docs-generation.md))

### 1.12 Cross-platform desktop

- [x] Linux build (WebKit) — `Taskfile.yml` `linux:build` via `wails3 build`, GoReleaser `.deb`/`.AppImage`/`.tar.gz`, `install.sh` pacman/apt/dnf/zypper — ADR 0019
- [x] macOS build (WebKit) — `darwin:build` (amd64/arm64), ad-hoc `codesign -s -` + `xattr -d com.apple.quarantine` fallback, `install.sh` — ADR 0019
- [x] Windows build (WebView2) — `windows:build` (amd64), unsigned `.exe`/`.zip`, `install.ps1` — ADR 0019
- [x] Release CI — `release.yml` OS matrix + `checksums.txt` on semver tags (`v*.*.*`), Conventional Commits notes

### 1.13 Desktop shell redesign — P0 UI Architecture

> **⚠️ RESTARTING FROM SCRATCH** — Previous implementation did not follow spec §2 four-zone architecture. All UI components will be rewritten following the spec's TopBar / ToolRail / ContextSidebar / MainWorkspace / BottomPanel model. **Progress 2026-08-27: Tickets #01–#04 shipped — shell chrome, Home, Request Builder, Collections Explorer — see `docs/internal/gui-roadmap.md` GUI-0.1–0.3.**

- [x] **§2.1** TopBar — Logo, Workspace Switcher, Global Search ⌘K, Import, Export, Active Environment, Sync Status, Settings — always visible — 2026-08-27
- [x] **§2.2** Tool Rail (48–56px) — 4 groups: Workspace (Home/Requests/Environments/History), API Tools (Mocks/Diff/JWT/GraphQL/gRPC/Runners/Explorer/Docs), Realtime (WebSocket/SSE), System (Settings) — 2026-08-27
- [x] **§2.3** Context Sidebar (220–280px) — Collapsible/resizable, changes per active tool, tree navigation, search, actions, `⌘B` toggle — 2026-08-27 (Collections Explorer + search/drag/context-menu)
- [x] **§2.4** Main Workspace — Tab-based content area, page/panel routing per active tool — 2026-08-27 (Request Builder tabs + Response Viewer)
- [x] **§2.5** Bottom Utility Panel — Console/Network/Tests/Variables/Cookies, `⌘J` toggle, resizable — 2026-08-27 (pre-existing, verified)
- [x] **§3** Design System — Tokens, typography, color system (IBM Plex, terracotta accent, BASE 6px radius) — GUI-1 shipped 2026-08-27
- [x] **§4** Navigation Model — Two-axis: horizontal (tool rail) + vertical (sidebar resource) — GUI-2 shipped 2026-08-27
- [x] **§60** Navigation Map — 15+ full pages with sub-panels (Params/Headers/Body/Auth/Tests/Response) — GUI-2 shipped 2026-08-27
- [x] **§61** Shared Patterns — Search, primary/secondary actions, status indicators, tabs, panels — GUI-2 shipped 2026-08-27
- [x] **§62** Page vs Panel Rules — Full pages vs context panels vs request/response panels vs bottom panels vs dialogs — GUI-2 shipped 2026-08-27
- [x] **§63** Final Layout Model — Canonical five-zone shell as single source of truth — GUI-2 shipped 2026-08-27
- [x] Infinite-loop fix for palette `filtered` selector (`useSyncExternalStore` new-array identity) — 2026-08-26
- [x] P1 spec editor (§56.1) tree + YAML `CodeMirror` + schema viz graph (§56.2) — 2026-08-26

---

