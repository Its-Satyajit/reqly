# Reqly — Development Roadmap

> **Status:** P0 complete, P1 in progress
> **Overall completion:** ~35%. All 27 milestones through M27 (cross-platform desktop) are shipped, plus the first few P1 items.
> **Source of truth:** [`docs/features.md`](docs/features.md) (features), [`docs/technology-stack.md`](docs/technology-stack.md) (stack), [`docs/testing-strategy.md`](docs/testing-strategy.md) (quality)
>
> Checkboxes track real, working code, not scaffolding. A box gets ticked only when the feature ships end to end: core logic, UI/CLI wiring, and tests, per the Definition of Done in the Testing Strategy doc.

---

## Legend

- `[x]` — shipped & tested (core + UI + tests)
- `[~]` — partial (some layers exist, not complete end-to-end)
- `[ ]` — not started
- **(stub)** — scaffold/file exists but no logic

---

## Phase 0 — Foundation (100% complete)

Project skeleton, build system, and the first core primitives.

### 0.1 Repository & build infra

- [x] Go module `github.com/Its-Satyajit/reqly` (Go 1.25)
- [x] npm workspaces + nub package manager (`pnpm-lock.yaml` committed)
- [x] Wails v3 desktop project (`apps/desktop/backend`) with Taskfile + build assets
- [x] CI workflow (frontend typecheck/build job; Go vet/gofmt/race/coverage job)
- [x] Makefile task aliases
- [x] Apache-2.0 license + SPDX headers on all Go sources
- [x] GoReleaser + Wails OS-matrix release pipeline (`release.yml`, `Taskfile.yml`, `install.sh`/`install.ps1`, ADR 0019)

### 0.2 Desktop shell (Wails v3)

- [x] `main.go` — Wails v3 `application.New`, window (1280×800), dark background (`NewAppService()` constructor)
- [x] `AppService` binding registered + `Greet` bridge proof → replaced by real `SendRequest` binding (see §1.5)
- [x] Go ↔ TypeScript bindings generated (`wails3 generate bindings`)
- [x] Host app (`apps/desktop/backend/frontend`) — Vite + React + Tailwind, wails vite plugin, port 9245
- [x] `wails3 build` produces `bin/reqly`

### 0.3 Shared UI shell (`frontend/`)

- [x] App shell (header, sidebar, split request/response panes)
- [x] Light/dark theming with Reqly brand colors + theme store + toggle
- [x] Dark/light logo in header; logo as app icon
- [x] Base UI via shadcn CLI (`button` component, `#`-alias imports)
- [x] CodeMirror 6 editor wrapper (json/js/xml/yaml/markdown/text)

### 0.4 Core primitives (shipped, TDD)

- [x] `internal/variables` — 6-scope resolution + `{{key}}`/`{{$tag}}` interpolation + `.env` process-env scope + env-file validation/diff
- [x] `internal/scripting` — Goja runtime with `reqly` sandbox (request/response access, variable get/set, `reqly.test()`, console) + pre/post wiring + dynamic values
- [x] `internal/request` + `internal/response` — request engine + response model (see §1.1)
- [x] `internal/testing` — assertion engine + JSONPath + suite runner + test-file loader (see §1.11)
- [x] `internal/history` + `internal/secrets` — SQLite history/cookie jar + token store (FileStore + KeychainStore)

### 0.5 CLI skeleton

- [x] Cobra command tree: `run`, `test`, `collection run`, `mock`, `validate`, `diff`, `docs` (+ `collection list`/`test`, `import`, `export`, `env`, `auth`, `history`, `ws`, `sse`)
- [x] 15 CLI commands wired to the Go core — `run`, `test`, `collection run`/`list`/`test`, `import curl`/`openapi`, `export postman`/`code`/`workspace`, `ws`, `sse`, `mock`, `validate`, `diff`, `docs generate`, `env` (list/show/use/validate/diff), `auth` (login/status/logout), `history` (list/show/search/clear/replay)

---

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
- [ ] OAuth 1.0, custom auth — deferred (reuses the same `auth.config` + Auth tab seams)
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
- [~] **GraphQL** — query editor + variables via `BodyType: graphql` shipped (ADR 0013); introspection/autocomplete/schema browser deferred to P1
- [ ] **gRPC** — proto files, reflection, service/method discovery, unary + streaming
- [ ] **SOAP** — WSDL import, operation discovery, XML builder

### 1.9 Import / export

- [x] Import cURL — `reqly import curl` (method, headers, JSON/raw/data bodies, basic auth, user-agent, cookies, GET-style query data; unsupported features reported)
- [x] Import OpenAPI 3.x — `reqly import openapi` (servers, paths, operations, params, JSON bodies; writes a Git-native workspace)
- [x] Export Postman collection v2.1 — `reqly export postman` (flat list, inherited base URL/headers applied)
- [~] Import: Postman v2.1 ([M34](docs/spec/m34-postman-import.md) — requests, nested folders, variables, bodies raw/urlencoded/form-data/graphql, basic/bearer/apikey auth; scripts + file bodies warned), Insomnia v4/v5 ([M35](docs/spec/m35-insomnia-import.md) — both formats auto-detected, nested folders, environments as native `environments/*.yaml`, basic/bearer/apikey/digest auth; cookie jars + unsupported auth warned), Bruno ([M36](docs/spec/m36-bruno-import.md) — items tree, body modes, collection-level auth/headers defaults, secret-split environments), Swagger 2.x (via hand-rolled parser); HAR done ([M28](docs/spec/m28-har-import-export.md))
- [x] Export: requests ([`export workspace`](docs/adr/0017-workspace-save-export.md) + `export code`), OpenAPI 3.0 spec generation (`export openapi`, [M37](docs/spec/m37-export-reports-openapi.md)), responses (`export har` from history, M28 + desktop download), test results (`collection test --report-junit/--report-json`, M37); docs done (§1.11 `reqly docs`)
- [ ] Import preservation (env/auth/scripts) + unsupported-feature reporting

### 1.10 OpenAPI & JSON Schema

- [~] OpenAPI 3.x parse + validate — `internal/openapi` (kin-openapi, JSON/YAML, $ref resolution); OpenAPI 2.x import via hand-rolled parser; 3.1 partial
- [~] Endpoint explorer + generate requests from spec — `reqly openapi explore <spec> [--tag]... [--json]` (operation table / machine-readable list) and `reqly openapi generate <spec> [--operation]... | [--method --path] | [--tag]... | --all [--output dir]` ([M39](docs/spec/m39-openapi-explorer.md) — native request files, inline example/default bodies+params, bearer/basic/apikey-header → native auth blocks, unmappable features warned; desktop explorer panel deferred to M39b)
- [ ] JSON Schema: edit, validate, inspect, generate
- [ ] XML/XSD schema validation where applicable
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

---

## Phase 2 — Differentiating Features (P1)

Features that make Reqly more capable than a basic API client.

- [ ] OpenAPI editor (spec authoring in-app)
- [ ] Schema validation + contract testing
- [ ] Schema visualization (relationships between types, objects, endpoints, schemas)
- [x] Code generation — `internal/exporter.Generate` (request → cURL/JS/Python/Go, header/body/auth, `[SECRET]` masked, `reqly export code` + desktop `Copy as`, golden files) — [ADR 0016](docs/adr/0016-code-generation.md)
- [~] Mock server (from OpenAPI/examples, request matching, dynamic responses, delay/error simulation, stateful mocks) — CLI `reqly mock` with path/method matching, schema/example generation, `--delay`, `--fail-every`; stateful mocks pending
- [ ] API diff + breaking-change detection (endpoints, params, schemas, auth, response types)
- [ ] Request/response diff (JSON structural)
- [ ] Environment diff
- [ ] Request inheritance & templates (reusable request templates)
- [ ] HAR import/export + replay
- [ ] JWT tooling (decode, claims viewer, signing)
- [ ] GraphQL introspection / gRPC reflection tooling
- [ ] Advanced HTTP: HTTP/2, HTTP/3, streaming, chunked transfer, keep-alive
- [ ] Proxy & TLS controls: system/HTTP/HTTPS/SOCKS5, per-env/per-request, cert inspection, mTLS, custom CAs
- [ ] Pagination runner (page/offset/cursor/link-header, stop conditions, aggregation)
- [ ] Bulk request execution (CSV/JSON inputs, sequential/parallel, concurrency)
- [ ] Data-driven testing (same test suite against multiple datasets)
- [x] Retry & resilience — engine-level `request.retry` block (`count`/`delayMs`/`strategy: fixed|exponential`/`maxDelayMs`/`retryOn`) inside `Client.Execute`, network errors + 429/502/503/504 default set, `Retry-After` honored + clamped, auth refresh orthogonal within one attempt, `response.Attempts` + history carry, `reqly run --retries/--retry-delay`, desktop Retry section ([ADR 0024](docs/adr/0024-retry-resilience.md))
- [ ] API documentation generation (REST + GraphQL + realtime)
- [ ] CI/CD support (run collections/tests in CI, mock deployment, env validation, docs generation)

---

## Phase 3 — Power-User Features (P2)

Advanced functionality for experienced developers and teams.

- [ ] API monitoring (scheduled requests/collections, health checks, latency/availability, alerts)
- [ ] Performance testing (RPS, latency, P95/P99, error rate, status distribution)
- [ ] MQTT (publish/subscribe, topics, QoS, retained/will, auth, TLS)
- [ ] Socket.IO (connections, events, rooms, namespaces, debugging)
- [ ] API dependency graph
- [ ] Request replay (exact / modified vars / other env / captured traffic)
- [ ] API changelog (from specs + Git changes)
- [ ] Browser integrations (DevTools import, cURL copy, Chrome/Firefox/Safari)
- [ ] In-app developer tools (app-level debugging: request/auth/variables/script/runtime/network inspection)
- [ ] Advanced mock state (multi-scenario state machines)
- [ ] Visual workflow builder
- [ ] Advanced network interception (capture/inspect/import/modify/replay)
- [ ] Self-hosted automation
- [ ] Git GUI integration (init/commit/branch/diff/history/pull/push/merge/conflicts)
- [ ] Request timeline debugging (DNS/connect/TLS/request/server/response/transfer)

---

## Phase 4 — Ecosystem & Enterprise (P3)

Long-term ecosystem and organization features.

- [ ] Plugin marketplace (auth, template tags, request/response processing, protocols, UI)
- [ ] Theme marketplace + custom themes + UI extensions
- [ ] Git provider integrations (GitHub, GitLab, Bitbucket, Azure DevOps) + PATs
- [ ] Shared workspaces / team collaboration
- [ ] Self-hosted collaboration server
- [ ] Enterprise SSO
- [ ] SCIM provisioning
- [ ] Audit logs
- [ ] Organization policies
- [ ] Enterprise secret management (Vault, AWS, Azure, role-based access)
- [ ] Advanced access control / permissions

---

## Phase 5 — MCP, AI & Extensibility (cross-cutting)

- [ ] `internal/mcp` — MCP server (list/search/run requests & collections, inspect schemas, retrieve responses, generate docs)
- [ ] Command palette / spotlight, keyboard shortcuts, context menus, widgets, code snippets
- [ ] Optional AI: request generation, response explanation, test/docs generation, error analysis, schema assistance, breaking-change explanation

---

## Quality & Release Gates (Definition of Done — from Testing Strategy)

Every checked feature must pass the full checklist:

- [x] Requirement defined in FeatureSet (`docs/features.md`) + CONTEXT.md glossary entry
- [x] TDD cycle (red → green → refactor) with unit tests (`go test ./...`, table-driven + testify where applicable)
- [x] Edge cases + error behavior covered (masking, expiry, fallback, empty-workspace, malformed-file paths)
- [~] Integration tests (core ↔ persistence ↔ engine) — runner + history/sqlite + env precedence covered; full E2E pending
- [ ] E2E tests (Playwright) for critical workflows — deferred to post-P0
- [x] Security review (no secrets exposed, 0600/0644 file modes, safe crypto via stdlib + masking)
- [x] Performance considered (SQLite WAL+FTS5+spill, 500 retention, 4KB hex cap, 1000-row virtualized Table)
- [x] Regression tests (golden files for exporter/docs, fixture workspaces)
- [~] Coverage within targets — CI enforces `go test -race` + coverage; thresholds tracked per PR
- [x] Docs updated (ROADMAP + CONTEXT + ADR per milestone)
- [x] CI green (vet, gofmt, typecheck, unit, race, build)
- [ ] Frontend unit tests (Vitest) — **TBD** (typecheck only: `nub run typecheck`)

### Release gates

- **Fast checks** (every change): formatting, linting, typechecking, unit tests — ✅ running in CI
- **PR CI:** unit + integration + race + frontend + build + coverage validation — ✅ race + coverage + frontend typecheck + Wails build; integration/E2E partial
- **Release CI:** cross-platform builds + checksums + install scripts — ✅ shipped (ADR 0019 `release.yml` OS matrix); full E2E/performance/security compat — ⏳ pending

---

## Progress Tracker

| Phase   | Scope                    | Status                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | Est. complete |
| ------- | ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------- |
| Phase 0 | Foundation               | 100% — repo/build infra + Wails shell + UI shell + all core primitives + CLI skeleton + release pipeline (GoReleaser + Wails OS matrix + install scripts)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | 100%          |
| Phase 1 | Core API Client (P0)     | 100% — Request engine + response model + CLI `run`/`test` + request files + core services + desktop bridge + workspace/collection storage with inheritance + cURL/OpenAPI import + Postman export + WebSocket/SSE clients + scripting sandbox + collection runner + OpenAPI parsing + mock server + env management/validation/masking/diff + auth schemes (basic/bearer/apikey/jwt/digest/none + masking) shipped + OAuth 2.0 Client Credentials + Authorization Code/PKCE + Device flow with cached tokens, refresh-token reuse, `reqly auth login`/`status`/`logout`, auto-login, OS-keychain store, custom-scheme redirects, desktop auth panel + desktop request builder (params/headers/body tabs, response views/search/actions/JSONPath, cookies view) + desktop environments manager (list/create/edit/set-active/delete, masked secret editing, inline validation) + desktop collections browser (workspace tree, per-tab request/response state, env pill + snapshot variable layering, send) + desktop request-file editing (format-preserving atomic save, changed-on-disk Overwrite/Reload) + desktop collection-run UI (sidebar run buttons, streamed Run View) + desktop auth editing (Auth tab, scheme forms, oauth2 grant config, inherited-auth view, save warnings) + AWS SigV4 + Akamai EdgeGrid auth (per-request signing, Auth tab forms, save warnings) + binary + GraphQL body editors (file upload, multipart file rows, GraphQL query+variables) + history + cookie jar + table + binary preview (SQLite `history.db` + FTS5 + spill + replay, persistent jar + auto-attach, Table + image/PDF/hex) + dynamic values & template tags (`{{$uuid}}`/`{{$timestamp}}`/`{{$isoTimestamp}}`/`{{$randomInt}}`/`{{$randomString}}`, picker + autocomplete) + save/export workspace (bulk `SaveWorkspace` + `reqly export workspace`) + `reqly docs` (Markdown `index.md` + per-collection, `text/template` + `curl`) + cross-platform desktop (OS matrix + checksums + install.sh/ps1) — **P0 1.1-1.12 100%** | 100%          |
| Phase 2 | Differentiating (P1)     | Code generation (cURL/JS/Python/Go) via `internal/exporter` — first P1 shipped; HAR import/export + JWT `reqly jwt decode` (`internal/jwt`) + Pagination `reqly pagination run` (`internal/pagination`) + Bulk `reqly bulk run --data` (`internal/bulk` CSV/JSON, sequential/parallel) + engine-level Retry & resilience shipped via `docs/adr/0020`+`0021`+`0022`+`0023`+`0024`; remaining P1 (API diff, OpenAPI editor, contract testing) queued | ~35%          |
| Phase 3 | Power-User (P2)          | Not started                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | 0%            |
| Phase 4 | Ecosystem (P3)           | Not started                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | 0%            |
| Phase 5 | MCP / AI / Extensibility | Not started (`internal/mcp` stub only)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | 0%            |
| Quality | DoD + release gates      | Fast + PR CI green (vet/gofmt/race/coverage/typecheck/Wails build + cross-platform release); E2E/Playwright + Vitest + full perf/security compat pending                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | ~55%          |

### Next milestones (suggested order)

1. ~~**Request file loading**~~ — `reqly run`/`test` from a plain-text request file (JSON/YAML request format + vars) — ✅ shipped
2. ~~**Core → Desktop bridge**~~ — `AppService.SendRequest` → core `RequestService`; `RequestEditor` Send wired to `useRequestStore` — ✅ shipped
3. ~~**Workspaces & collections on disk**~~ — Git-native storage + inheritance (build on `requestfile`) — ✅ shipped
4. ~~**Import/export**~~ — cURL/OpenAPI import + Postman collection export — ✅ shipped
5. ~~**WebSocket + SSE**~~ — realtime protocols: `internal/websocket` (connection mgmt, text/binary messages) + `internal/sse` (event stream parser) + CLI `reqly ws`/`reqly sse` — ✅ shipped
6. ~~**Collection runner + scripting**~~ — pre/post scripts (Goja `reqly` sandbox), request chaining via runtime variables, tests in the runner, CLI `reqly collection test` — ✅ shipped
7. ~~**Mock server + OpenAPI**~~ — `internal/openapi` (kin-openapi load/validate) + `internal/mocking` (path/method matching, schema/example response generation, delay + error simulation) + CLI `reqly mock <spec>` — ✅ shipped
8. ~~**Validate + diff**~~ — `reqly validate` (spec/project checks) and `reqly diff` (specs/requests/responses) — ✅ shipped
9. ~~**Environments & secrets**~~ — Git-native `environments/<name>.yaml` with `variables:`/`secrets:`, selection precedence (`REQLY_ENV`/`--env`/file/descriptor), `.env` process-env scope, `[SECRET]` masking in CLI output, `reqly env list/show/use/validate/diff` — ✅ shipped
10. ~~**Auth schemes**~~ — `internal/auth` scheme registry + `request.Auth` dispatch (basic, bearer, apikey, jwt HS256/384/512, digest challenge/response, none) with secret masking ([ADR 0005](docs/adr/0005-git-native-auth-schemes.md)) — ✅ shipped
11. ~~**OAuth 2.0 Client Credentials**~~ — `auth.TokenSource` acquisition split from application, store-backed token cache in `<workspace>/.reqly/tokens.json` (`secrets.Store` + `CachedTokenSource`, ADR 0006), expiry-skewed proactive refresh, reactive 401 refresh + retry-once, per-config concurrency lock, masking of acquired tokens in `run`/`test`/`collection run`/`collection test`, `reqly auth status`/`auth logout` — ✅ shipped ([PR #58](https://github.com/Its-Satyajit/reqly/pull/58))
12. ~~**OAuth 2.0 Authorization Code + PKCE**~~ — [Spec #52](https://github.com/Its-Satyajit/reqly/issues/52) + tickets [#53–#57](https://github.com/Its-Satyajit/reqly/issues/53): `AuthorizationCodeSource` (PKCE S256 + state, one-shot loopback callback, code exchange), `reqly auth login`, first-request auto-login, refresh-token reuse with rotation, ADR 0007 — ✅ shipped
13. ~~**OAuth 2.0 / auth leftovers**~~ — [Spec #60](https://github.com/Its-Satyajit/reqly/issues/60) + tickets [#61–#65](https://github.com/Its-Satyajit/reqly/issues/61): device flow (RFC 8628) with `reqly auth login --flow device`, OS-keychain token backend (`--store keychain`/`REQLY_TOKEN_STORE`, file fallback), custom-scheme redirects (`reqly://` deep links), desktop auth panel — ✅ shipped ([PRs #66–#69](https://github.com/Its-Satyajit/reqly/pull/66), [ADR 0008](docs/adr/0008-oauth2-auth-leftovers.md)); Password/ROPC stays deferred per OAuth 2.1
14. ~~**Desktop request builder UI**~~ — [Spec #71](https://github.com/Its-Satyajit/reqly/issues/71), tickets [#72–#76](https://github.com/Its-Satyajit/reqly/issues/76): params/headers tabs ([T1](https://github.com/Its-Satyajit/reqly/issues/72)) + response viewer raw/pretty/tree + search ([T3](https://github.com/Its-Satyajit/reqly/issues/74)) + body-type editors JSON/XML/form-data/urlencoded/raw with auto Content-Type ([T2](https://github.com/Its-Satyajit/reqly/issues/73)) + response actions copy/download/format and JSONPath query bar ([T4](https://github.com/Its-Satyajit/reqly/issues/75)) + cookies view from `Set-Cookie` headers ([T5](https://github.com/Its-Satyajit/reqly/issues/76)) — ✅ shipped; cookie persistence and table view remain open follow-ups
15. ~~**Desktop environments UI**~~ — [Spec #84](https://github.com/Its-Satyajit/reqly/issues/84), tickets [#85–#89](https://github.com/Its-Satyajit/reqly/issues/85): environments service + header selector ([T1](https://github.com/Its-Satyajit/reqly/issues/85)) + view/list/create/set-active with sidebar nav ([T2](https://github.com/Its-Satyajit/reqly/issues/86)) + editor with description/variables + dirty-tracking save ([T3](https://github.com/Its-Satyajit/reqly/issues/87)) + masked secrets editing (changed-only writes, never read back) ([T4](https://github.com/Its-Satyajit/reqly/issues/88)) + delete with active-clear + inline validation + milestone docs ([T5](https://github.com/Its-Satyajit/reqly/issues/89)) — ✅ shipped — last milestone in Phase-1 P0
16. ~~**Desktop collections browser**~~ — [Spec #130](https://github.com/Its-Satyajit/reqly/issues/130), tickets [#131–#134](https://github.com/Its-Satyajit/reqly/issues/131): workspace tree in the sidebar ([T1](https://github.com/Its-Satyajit/reqly/issues/131)) + per-tab request/response state ([T2](https://github.com/Its-Satyajit/reqly/issues/132)) + open resolved requests into tabs with read-only Variables view + env pill ([T3](https://github.com/Its-Satyajit/reqly/issues/133)) + send with environment pill + snapshot variable layering + silent inherited auth ([T4](https://github.com/Its-Satyajit/reqly/issues/134), [ADR 0009](docs/adr/0009-desktop-collection-request-snapshot-model.md)) — ✅ shipped; request-file editing and collection-run UI shipped as follow-ups, auth editing remains an open follow-up
17. ~~**Desktop request-file editing**~~ — [Spec #143](https://github.com/Its-Satyajit/reqly/issues/143), tickets [#145–#149](https://github.com/Its-Satyajit/reqly/issues/145): format-preserving atomic save + content fingerprint ([T1](https://github.com/Its-Satyajit/reqly/issues/145)) + workspace save/re-resolve seams ([T2](https://github.com/Its-Satyajit/reqly/issues/146)) + bridge file-backed send + save ([T3](https://github.com/Its-Satyajit/reqly/issues/147)) + editable tabs with Save/dirty/confirm-on-close + changed-on-disk Overwrite/Reload ([T4](https://github.com/Its-Satyajit/reqly/issues/148)) + Effective URL line + inherited-headers group + milestone docs ([T5](https://github.com/Its-Satyajit/reqly/issues/149), [ADR 0009](docs/adr/0009-desktop-collection-request-snapshot-model.md) amendment) — ✅ shipped ([PR #168](https://github.com/Its-Satyajit/reqly/pull/168)); auth editing remains an open follow-up
18. ~~**Desktop collection-run UI**~~ — [Spec #151](https://github.com/Its-Satyajit/reqly/issues/151), tickets [#152–#156](https://github.com/Its-Satyajit/reqly/issues/152): streamed per-step results via OnStep callback ([T1](https://github.com/Its-Satyajit/reqly/issues/152)) + collection run service + RunFolder engine support ([T2](https://github.com/Its-Satyajit/reqly/issues/153)) + collection-run bindings + streamed Wails events ([T3](https://github.com/Its-Satyajit/reqly/issues/154)) + collection-run adapter + run store ([T4](https://github.com/Its-Satyajit/reqly/issues/155)) + sidebar run buttons + Run View tab ([T5](https://github.com/Its-Satyajit/reqly/issues/156), [ADR 0009](docs/adr/0009-desktop-collection-request-snapshot-model.md) amendment) — ✅ shipped ([PRs #160–#165](https://github.com/Its-Satyajit/reqly/pull/160)); auth editing remains an open follow-up
19. ~~**Desktop auth editing**~~ — [Spec #170](https://github.com/Its-Satyajit/reqly/issues/170), tickets [#171–#175](https://github.com/Its-Satyajit/reqly/issues/171): editable draft auth on save/send ([T1](https://github.com/Its-Satyajit/reqly/issues/171)) + file-backed auth read + save ([T2](https://github.com/Its-Satyajit/reqly/issues/172)) + Auth tab with scheme picker and per-scheme typed field forms ([T3](https://github.com/Its-Satyajit/reqly/issues/173)) + OAuth 2.0 grant config form + Auth Panel link ([T4](https://github.com/Its-Satyajit/reqly/issues/174)) + inherited-auth read-only view, sensitive-field flags, non-blocking save warnings + milestone docs ([T5](https://github.com/Its-Satyajit/reqly/issues/175), [ADR 0011](docs/adr/0011-desktop-request-auth-editing.md)) — ✅ shipped
20. ~~**Desktop AWS + EdgeGrid auth**~~ — [Spec #181](https://github.com/Its-Satyajit/reqly/issues/181), tickets [#182–#186](https://github.com/Its-Satyajit/reqly/issues/182): core SigV4 + EdgeGrid schemes ([T1](https://github.com/Its-Satyajit/reqly/issues/182)) + bridge/types ([T2](https://github.com/Its-Satyajit/reqly/issues/183)) + Auth tab forms ([T3](https://github.com/Its-Satyajit/reqly/issues/184)) + save warnings ([T4](https://github.com/Its-Satyajit/reqly/issues/185)) + ADR 0012 + docs ([T5](https://github.com/Its-Satyajit/reqly/issues/186), [ADR 0012](docs/adr/0012-aws-edgegrid-auth.md)) — ✅ shipped
21. ~~**Binary + GraphQL body editors**~~ — [Spec #189](https://github.com/Its-Satyajit/reqly/issues/189), tickets [#190–#194](https://github.com/Its-Satyajit/reqly/issues/190): core `binary`/`graphql` BodyType + file-aware `form-data` ([T1](https://github.com/Its-Satyajit/reqly/issues/190)) + bridge/body lib ([T2](https://github.com/Its-Satyajit/reqly/issues/191)) + Body tab file picker + GraphQL editors ([T3](https://github.com/Its-Satyajit/reqly/issues/192)) + save warnings ([T4](https://github.com/Its-Satyajit/reqly/issues/193)) + ADR 0013 + docs ([T5](https://github.com/Its-Satyajit/reqly/issues/194), [ADR 0013](docs/adr/0013-binary-graphql-body.md)) — ✅ shipped
22. ~~**History + Cookie jar + Table + Binary preview**~~ — [Spec #197](https://github.com/Its-Satyajit/reqly/issues/197), tickets [#198–#202](https://github.com/Its-Satyajit/reqly/issues/198): SQLite `history.db` + cookie jar (FTS5, spill, retention 500, `0600`, domain/path/secure matching, auto-attach) + Table (JSON array/CSV, 1000 rows) + binary preview (image/PDF/hex) + `reqly history` CLI + desktop History view + ResponseViewer Table tab ([ADR 0014](docs/adr/0014-history-cookie-jar-table-view.md)) — ✅ shipped — **P0 desktop polish complete**
23. ~~**Dynamic values & template tags**~~ — [Spec #204](https://github.com/Its-Satyajit/reqly/issues/204), tickets [#205–#208](https://github.com/Its-Satyajit/reqly/issues/205): `internal/variables` `TagGenerator` + `{{$uuid}}`/`{{$timestamp}}`/`{{$isoTimestamp}}`/`{{$randomInt}}`/`{{$randomString}}` (`{{$` strict, per occurrence fresh, args ignored for M23, unknown literal + `saveWarnings`) + `TagPicker` picker + `{{$` autocomplete ([ADR 0015](docs/adr/0015-dynamic-values-template-tags.md)) — ✅ shipped
24. ~~**Code generation**~~ — [Spec #211](https://github.com/Its-Satyajit/reqly/issues/211), tickets [#212–#215](https://github.com/Its-Satyajit/reqly/issues/212): `internal/exporter.Generate` (cURL/JS/Python/Go, `reqly export code` + `Copy as`, golden files, `[SECRET]` masked) ([ADR 0016](docs/adr/0016-code-generation.md)) — ✅ shipped
25. ~~**Save/export workspace**~~ — [Spec #217](https://github.com/Its-Satyajit/reqly/issues/217), tickets [#218–#220](https://github.com/Its-Satyajit/reqly/issues/218): `internal/collections.SaveWorkspace` (bulk in-place + `reqly export workspace [src] --out <dir>` copy, `requestfile.Save` atomic, prune deleted, `0600`/`0644`) ([ADR 0017](docs/adr/0017-workspace-save-export.md)) — ✅ shipped
26. ~~**Docs generation**~~ — [Spec #221](https://github.com/Its-Satyajit/reqly/issues/221), tickets [#222–#224](https://github.com/Its-Satyajit/reqly/issues/222): `internal/docs.Generate` (Markdown `index.md` + per-collection via `text/template` + `curl` via `exporter`, `reqly docs generate [src] --out <dir> [--env]`, `0600`/`0644`) ([ADR 0018](docs/adr/0018-docs-generation.md)) — ✅ shipped
27. ~~**Cross-Platform Desktop**~~ — [Spec #225](https://github.com/Its-Satyajit/reqly/issues/225), tickets [#226–#228](https://github.com/Its-Satyajit/reqly/issues/226): `Taskfile.yml` OS matrix (`linux:build`/`darwin:build`/`windows:build` via `wails3` + `GoReleaser`), `release.yml` OS matrix + `checksums.txt` + `install.sh` (Linux `pacman`/`apt`/`dnf`/`zypper` + `Darwin` `amd64`/`arm64`) + `install.ps1` (Windows `amd64`) ([ADR 0019](docs/adr/0019-cross-platform-desktop.md)) — ✅ shipped — **P0 1.12 complete — P0 100%**

### Next milestones — P1 Differentiating Features (suggested order)

28. **HAR import/export + replay** — `internal/importer` HAR parse + `reqly import har <har-file> [--output <dir>] [--collection <name>]` ( `headers+cookies→Headers` `Cookie:` merged, `queryString→Query`, `postData.text→Body` base64 decoded, `mimeType→Content-Type`, >1MB spill `blobs/`, `pageref`/`timings`/`cache` warnings) + `reqly export har [--out <file.har>] [--env <name>] [--limit 500]` history→HAR via `internal/exporter/har.go` (`ExportHAR` pure, `timings` synthesized, base64 binary, secrets masked), replay via `har-import` collection + `history replay` ([ADR 0020](docs/adr/0020-har-import-export.md), CONTEXT `HAR`/`HAR Import`/`HAR Export`/`HAR Replay` grilling Q1–Q4 done, `docs/spec/m28-har-import-export.md`) — **shipped**
29. **JWT tooling** — `reqly jwt decode` (header/claims viewer, expiry detection) in `internal/jwt` + `reqly jwt decode [--json]` + `Bearer`/stdin (`internal/jwt.Decode` + `apps/cli/cmd/jwt.go`, expiry `exp`/`nbf`/`iat` → `expired`/`not_yet_valid`/`valid`/`no_expiry`, `Header:`/`Payload:` pretty + `--json`, [ADR 0021](docs/adr/0021-jwt-tooling-decode.md), CONTEXT `JWT Tooling`/`JWT Decode` grill Q1–Q5) — **shipped (decode MVP)**; `verify`/`sign` (HS via `jwtHashes`) + desktop inspector deferred to M29b
30. **Pagination runner** — `reqly pagination run <request-file> [--max-pages <n>]` ( `request.pagination: {strategy: page|offset|cursor|link-header, pageParam/pageSizeParam/offsetParam/limitParam/cursorParam, nextPath: $.nextCursor, maxPages: 100}` + `internal/pagination.Run` pure loop over `sendFn` `page`→`?page=1→2` `offset`→`?offset=0→10` `cursor`→`?cursor=<next>` via JSONPath `$.nextCursor` `link-header`→`Link: <url>; rel="next"` , stop empty/missing-next/non-2xx/maxPages, `--max-pages` overrides, `OnStep` streaming `step: status duration url`) ([ADR 0022](docs/adr/0022-pagination-runner.md), CONTEXT `Pagination Runner` `Strategy`/`Stop` grill Q1–Q4, `docs/spec/m30-pagination-runner.md`) — **shipped**
31. **Bulk request execution** — `reqly bulk run <request-file> --data <csv|json> [--parallel] [--concurrency <n>] [--continue-on-error]` (CSV header→`{{var}}`/JSON array stringified, `internal/bulk.Run` sequential default, parallel semaphore ordered `concurrency 5`, `ScopeRuntime` per row, stop first non-2xx unless `--continue-on-error`) ([ADR 0023](docs/adr/0023-bulk-runner.md), CONTEXT `Bulk Runner`/`Bulk Input Row`/`Bulk Concurrency` grill Q1–Q4, `docs/spec/m31-bulk-runner.md`) — **shipped**
32. ~~**Retry & resilience**~~ — engine-level `request.retry` (`count`/`delayMs`/`strategy`/`maxDelayMs`/`retryOn`) in `Client.Execute`; network errors + 429/502/503/504 default, `Retry-After` respected + clamped, exponential/fixed backoff capped, ctx-cancel aborts mid-wait, auth refresh stays inside one attempt, `response.Attempts` + `history show` attempts line + desktop attempts badge, `--retries`/`--retry-delay` flags, desktop collapsible Retry section in the request editor ([ADR 0024](docs/adr/0024-retry-resilience.md), `docs/spec/m32-retry-resilience.md`) — **shipped**
33. **OpenAPI editor + endpoint explorer** — in-app spec authoring + generate requests from spec + JSON Schema edit/validate
34. **API diff & breaking-change detection** — endpoints/params/schemas/auth/response-types + spec/request/response/env diff polish
35. **Contract testing + schema validation** — OpenAPI/JSON Schema response validation pipeline
36. **Advanced HTTP / Proxy & TLS controls** — HTTP/2, per-env/per-request proxy, cert inspection, mTLS, custom CAs
37. **Performance testing (lightweight)** — RPS/latency P95/P99/error-rate/status-distribution

> **Companion:** [**reqly-test-api**](https://reqly-test-api.vercel.app) — a small ElysiaJS mock API (Vercel-hosted, hardcoded data) for exercising `reqly run`/`test`, auth, delay, and error-status flows against a real endpoint. Useful while the in-app mock server (milestone 7) is pending; see the README's "Mock API" section.
