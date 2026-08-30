# Reqly — Development Roadmap

> **Status:** Canonical development roadmap. P0/core + CLI are shipped; P1 includes shipped backend/data-layer work plus ongoing desktop/UI completion; P2–P5 remain future work.
> **Overall completion:** Do not treat a single percentage as authoritative. Use the phase/milestone checkboxes and this document's status ledger.
> **Detailed Milestone Specs:** Full grouped milestone references are available under **[`./Milestones/`](Milestones/)**:
> - [`01-phase-0-foundation.md`](Milestones/01-phase-0-foundation.md) — Foundation (Wails shell, build infra)
> - [`02-phase-1-core-api-client.md`](Milestones/02-phase-1-core-api-client.md) — Core API Client (P0)
> - [`03-phase-2-differentiating-features.md`](Milestones/03-phase-2-differentiating-features.md) — Differentiating Features (P1)
> - [`04-phase-3-power-user-features.md`](Milestones/04-phase-3-power-user-features.md) — Power-User Features (P2)
> - [`05-phase-4-ecosystem-and-enterprise.md`](Milestones/05-phase-4-ecosystem-and-enterprise.md) — Ecosystem & Enterprise (P3)
> - [`06-phase-5-mcp-ai-extensibility.md`](Milestones/06-phase-5-mcp-ai-extensibility.md) — MCP, AI & Extensibility
> - [`07-historical-milestones-ledger.md`](Milestones/07-historical-milestones-ledger.md) — Historical Milestones Ledger (M01–M40)
> - [`08-gui-roadmap-and-execution.md`](Milestones/08-gui-roadmap-and-execution.md) — Desktop GUI Roadmap & Execution
> - [`09-ui-architecture-shell-and-requests.md`](Milestones/09-ui-architecture-shell-and-requests.md) — UI Architecture (§1–§25)
> - [`10-ui-architecture-tools-and-pages.md`](Milestones/10-ui-architecture-tools-and-pages.md) — UI Architecture (§26–§55)
> - [`11-ui-architecture-phase-panels-and-navigation.md`](Milestones/11-ui-architecture-phase-panels-and-navigation.md) — UI Architecture (§56–§63)
>
> **Source of truth:** [`docs/features.md`](docs/features.md) (features), [`docs/technology-stack.md`](docs/technology-stack.md) (stack), [`docs/testing-strategy.md`](docs/testing-strategy.md) (quality), [`docs/internal/gui-roadmap.md`](docs/internal/gui-roadmap.md) (desktop GUI milestones), **[`docs/Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification.md`](docs/Reqly%20Complete%20UI%20Architecture,%20Pages,%20Panels,%20and%20Navigation%20Specification.md)** (full UI spec — §1–§59)
>
> **⚠️ UI Redesign Notice:** Previous shell implementation did not follow spec architecture. All UI code in `frontend/src/` (shell components, features, stores) is being rewritten from scratch following the spec's four-zone model (TopBar / ToolRail / ContextSidebar / MainWorkspace + BottomPanel). Existing data layer (lib/, stores/) is preserved; UI components will be rebuilt.
>
> Checkboxes track real, working code, not scaffolding. A box gets ticked only when the feature ships end to end: core logic, UI/CLI wiring, and tests, per the Definition of Done in the Testing Strategy doc.

---

## Roadmap precedence and document governance

This unified file is the canonical product roadmap.

Precedence, highest to lowest:

1. **Development roadmaps** (`ROADMAP(2).md`, `ROADMAP(3).md`). These define product scope, phase, milestone, implementation status, and sequencing. Where they conflict, the newest development-roadmap evidence wins, while historical milestone detail is preserved below.
2. **GUI roadmap** (`gui-roadmap.md`). This defines desktop delivery status and UI implementation sequencing. It can clarify whether a UI slice is shipped, partial, or pending, but it does not remove or redefine product features in the development roadmap.
3. **Complete UI Architecture Specification** (`Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification.md`). This is a detailed implementation/reference specification for pages, panels, navigation, layouts, and interaction patterns. It is subordinate to the roadmaps. It must not silently change product scope, priority, or shipped status.
4. **Future document redesign / re-seeding plan.** The roadmap system itself is planned for a cleanup pass using these documents as inputs. That work will preserve historical decisions and shipped milestones while producing one canonical roadmap, one GUI execution map, and one subordinate UI reference specification.

When a UI document and a development roadmap disagree on whether a feature exists, the development roadmap wins for product status. When the development roadmap says a feature is shipped but the UI slice is unfinished, record it as **core shipped / UI pending**, not as a missing product capability.

Checkboxes mean:
- `[x]` shipped and verified at the level claimed by the item.
- `[~]` partially shipped or shipped in one layer with remaining work.
- `[ ]` not started or explicitly deferred.

## Legend

- `[x]` — shipped & tested (core + UI + tests)
- `[~]` — partial (some layers exist, not complete end-to-end)
- `[ ]` — not started
- **(stub)** — scaffold/file exists but no logic

---

## Phase 0 — Foundation (100% complete) — [Spec Reference](Milestones/01-phase-0-foundation.md)

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
- [x] Backend warning/error log mirror — slog handler emits `reqly.golog` events so desktop crash reports include Go-side diagnostics
- [x] sqlc-generated typed query layer over `modernc.org/sqlite` for the history store (`internal/history/db`; schema/query SQL in-repo, zero reflection, no CGO)

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

## Phase 1 — Core API Client (P0) — [Spec Reference](Milestones/02-phase-1-core-api-client.md)

The minimum set to make Reqly a serious API client.

### 1.1 Request engine (foundation for everything)

- [x] `internal/request` — full HTTP request model (URL, method, path/query params, headers, body, auth, certs, proxy, settings)
- [x] Request engine: HTTP/1.1 transport, timeouts, redirects, compression
- [x] Request execution shared by Desktop + CLI (single engine, no duplication)
- [x] Response model: status, headers, cookies, timing, size, raw body
- [x] Response body parsing — JSON (pretty/tree) + XML (pretty) + CSV (Table) + binary (image inline, PDF banner, hex 4KB) via `frontend/src/lib/response.ts:187` `isTabular/parseTable/binaryPreviewType` and `ResponseViewer` Table tab ([ADR 0014](docs/adr/0014-history-cookie-jar-table-view.md)); file download via `suggestedFilename`
- [~] File upload / multipart — `BodyType: binary` (single file, `application/octet-stream`) + file-aware `form-data` rows (`file` + `filename`, `multipart/form-data` via `boundaryFor`, [ADR 0013](docs/adr/0013-binary-graphql-body.md)); **file download via `suggestedFilename` shipped, raw file-system download UI pending** — core upload [x], download [~]
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
- [x] JWT — HS256/384/512 per-request signing, decoding, claims viewer, and signature verification (`reqly jwt decode/verify/sign`, ADR 0021)
- [~] Digest — challenge/response shipped (SHA-256 fallback, `internal/auth/digest.go` request-body aware, `request.FollowRedirects` aware); **NTLM deferred to P3** (Windows `NTLMSSP`/`SSPI` requires CGO/`gssapi`, out of scope for P0 local-first/no-CGO) — Digest P0 [x], NTLM [ ] deferred
- [x] OAuth 2.0 Client Credentials — RFC 6749 §4.4 with store-backed token caching (`TokenSource` + `secrets.Store`, ADR 0006), expiry-skewed proactive refresh, reactive 401 refresh+retry-once, `reqly auth status`/`auth logout`
- [x] OAuth 2.0 Authorization Code + PKCE — RFC 6749 §4.1 + RFC 7636 (`AuthorizationCodeSource`, one-shot loopback callback, state/verifier, [ADR 0007](docs/adr/0007-oauth2-authorization-code-pkce.md)), `reqly auth login`, first-request auto-login, refresh-token reuse (RFC 6749 §6, proactive + 401, rotation) — spec [#52](https://github.com/Its-Satyajit/reqly/issues/52), tickets [#53–#57](https://github.com/Its-Satyajit/reqly/issues/53)
- [x] OAuth 2.0 Device flow (RFC 8628) + OS-keychain store + custom redirects + desktop auth — `reqly auth login --flow device` (verification URI + code, RFC poll semantics), `--store keychain`/`REQLY_TOKEN_STORE` with file fallback, `reqly://` deep-link callbacks, sidebar auth panel (login/status/logout) — spec [#60](https://github.com/Its-Satyajit/reqly/issues/60), tickets [#61–#65](https://github.com/Its-Satyajit/reqly/issues/61), [ADR 0008](docs/adr/0008-oauth2-auth-leftovers.md)
- [x] AWS Signature V4 — `internal/auth/aws.go` (`auth.type: aws`, SigV4 per-request signing, `accessKey`/`secretKey`/`region`/`service` + optional `sessionToken`, [ADR 0012](docs/adr/0012-aws-edgegrid-auth.md))
- [x] Akamai EdgeGrid — `internal/auth/edgegrid.go` (`auth.type: edgegrid`, EG1-HMAC-SHA256, `clientToken`/`clientSecret`/`accessToken`/`host`, [ADR 0012](docs/adr/0012-aws-edgegrid-auth.md))
- [x] OAuth 1.0 — `internal/auth/oauth1.go` (`auth.type: oauth1`, RFC 5849 HMAC-SHA1 per-request signing, `consumerKey`/`consumerSecret` + optional `token`/`tokenSecret`, `Authorization: OAuth` header with `oauth_signature`, `oauth_nonce`/`oauth_timestamp`, `auth.config` + Auth tab `OAuth 1.0` form) — 2026-08-30
- [x] Custom auth — `internal/auth/custom.go` (`auth.type: custom`, `header`/`value` per-request header injection, `auth.config` + Auth tab `Custom` form, secret `value`) — 2026-08-30
- [x] Auth inheritance — Workspace → Collection → Folder → Request (base URL, headers, auth, vars)

### 1.4 Secrets

- [x] Encrypted-at-rest secret storage + OS keychain — token stores behind the `secrets.Store` interface: FileStore (plain-text 0600 `.reqly/tokens.json`, default) and KeychainStore (OS keychain via go-keyring; keychain default on desktop), backend selection `--store keychain`/`REQLY_TOKEN_STORE` with graceful file-store fallback
- [x] Secret variables + masking (CLI output, logs, test output) — `environments/<name>.yaml` `secrets:` maps render as `[SECRET]`; masking wired through run/test/collection/validate/diff; acquired OAuth tokens masked post-request
- [~] `.env` support — dotenv parsing (process-env scope, OS env wins) shipped via `internal/variables` + `internal/environments`; **external secret managers (Vault shipped via `internal/secrets.VaultStore`, AWS/Azure deferred to P3)** — `.env` [x], Vault [x], AWS/Azure [ ]

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
- [~] JSONPath / XPath response querying — dependency-free evaluator (`$.user.name`, `$['users'][0]`, wildcard `*`) with match list + specific errors; **JSONPath [x], XPath [ ] pending (deferred to P3)**
- [x] Response actions: copy (body/headers), download (Content-Disposition filename), format
- [x] Cookies: persistent jar (`history.db` `cookies` table, `env`-partitioned, `0600`, domain/path/secure/expires matching via `history.FilterCookies`, auto-attach `Cookie:` on next `SendRequest`, `Set-Cookie` ingest via `HistoryService.Record`, view + delete/clear in `ResponseViewer` Cookies tab + desktop `CookieList/Delete/Clear` bindings, CLI jar implicit) — [Milestone 14 T5](https://github.com/Its-Satyajit/reqly/issues/76) + [Milestone 22](https://github.com/Its-Satyajit/reqly/issues/197) ([ADR 0014](docs/adr/0014-history-cookie-jar-table-view.md))
- [x] Per-request settings + duplicate — timeout (ms) + `followRedirects` (`RequestSettingsDialog`, `core.RequestSave` + `mergeDraftRequest`, `WorkspaceService.DuplicateRequest` + `AppService.WorkspaceDuplicateRequest` + `RequestTabs`/`CollectionTree` duplicate) — `internal/request` + `internal/core` + `apps/desktop` + `frontend/src/lib/request.ts:47` + `frontend/src/stores/useRequestStore.ts:12` — shipped 2026-08-30 (M328)

### 1.7 Scripting & automation

- [x] Pre-request / post-request scripts (Goja) — `reqly` sandbox (request/response access, variable get/set, `reqly.test()`, console)
- [x] Test scripts + assertion library (core assertion engine shipped: status, header, body, JSON, response-time, schema)
- [x] Request chaining (login → extract token → next request) — runtime variables persist across collection steps
- [~] Chain runner — sequential execution, variable passing, assertions, script execution, fail-fast failure handling (**conditional branching deferred to P1** — core runner [x], branching [ ])
- [x] Collection runner (sequential, variable passing, assertions, fail-fast) — `reqly collection test` + desktop Run View streaming

### 1.8 Protocols (P0: REST-first, then extended)

- [x] **REST** — complete builder (see §1.1/§1.6: method/URL/headers/params/body + file upload + cookies/history)
- [x] **WebSocket** — connection mgmt, message composer, in/out inspection (`internal/websocket` + `reqly ws`)
- [x] **SSE** — live event stream, inspection, event history (`internal/sse` + `reqly sse`)
- [x] **GraphQL** — query editor + variables via `BodyType: graphql` (ADR 0013), live endpoint introspection (`reqly graphql introspect <url>`), offline SDL schema parsing (`reqly graphql parse <file.graphql>`), Goja scripting helper `reqly.introspectGraphQL()`, and Desktop Schema Browser ([M50](docs/spec/m50-graphql-schema-introspection.md), [ADR 0034](docs/adr/0034-graphql-schema-introspection.md))
- [~] **gRPC** — proto files, reflection, service/method discovery, unary + server-streaming — `internal/grpc` (reflection via v1 protocol, protocompile `.proto` fallback, TLS/h2c, deadlines), `grpc:` request-file block, scripting/assertions parity, history, `reqly grpc services|invoke`, desktop gRPC view (ADR 0028, M43; **unary/server-streaming [x], client-stream/bidi [ ] deferred to P3**)
- [~] **SOAP** — WSDL import, operation discovery, envelope skeletons: `reqly import wsdl <file> [--output dir]` ([M41](docs/spec/m41-wsdl-import.md) — one runnable POST per operation with binding-matched SOAP 1.1/1.2 envelopes, SOAPAction, inline-XSD body placeholders; local `xsd:import`/`include` resolved via `ParseWSDLWithBase` (`filepath.Dir` in CLI) shipped 2026-08-30; remote URLs + `rpc/encoded` still best-effort with warnings per ADR — **core P0 [x], `rpc/encoded` [ ] P3**)

### 1.9 Import / export

- [x] Import cURL — `reqly import curl` (method, headers, JSON/raw/data bodies, basic auth, user-agent, cookies, GET-style query data; unsupported features reported)
- [x] Import OpenAPI 3.x — `reqly import openapi` (servers, paths, operations, params, JSON bodies; writes a Git-native workspace)
- [x] Export Postman collection v2.1 — `reqly export postman` (flat list, inherited base URL/headers applied)
- [x] Import: Postman v2.1 ([M34](docs/spec/m34-postman-import.md)), Insomnia v4/v5 ([M35](docs/spec/m35-insomnia-import.md)), Bruno ([M36](docs/spec/m36-bruno-import.md)), Swagger 2.0 / OpenAPI 2.0 (`internal/importer.ParseSwagger2` + `reqly openapi convert-v2` converter, [M51](docs/spec/m51-swagger2-importer-converter.md), [ADR 0035](docs/adr/0035-swagger2-importer-converter.md)); HAR done ([M28](docs/spec/m28-har-import-export.md))
- [x] Export: requests ([`export workspace`](docs/adr/0017-workspace-save-export.md) + `export code`), OpenAPI 3.0 spec generation (`export openapi`, [M37](docs/spec/m37-export-reports-openapi.md)), responses (`export har` from history, M28 + desktop download), test results (`collection test --report-junit/--report-json`, M37); docs done (§1.11 `reqly docs`)
- [x] Import preservation (env/auth/scripts) + unsupported-feature reporting — [M42](docs/spec/m42-import-preservation.md) ([ADR 0026](docs/adr/0026-import-preservation-script-translation.md))

### 1.10 OpenAPI & JSON Schema

- [x] OpenAPI 3.x parse + validate — `internal/openapi` (kin-openapi, JSON/YAML, $ref resolution); Swagger 2.0 / OpenAPI 2.0 import & `reqly openapi convert-v2` spec converter shipped ([M51](docs/spec/m51-swagger2-importer-converter.md))
- [x] Endpoint explorer + generate requests from spec — `reqly openapi explore <spec> [--tag]... [--json]` (operation table / machine-readable list) and `reqly openapi generate <spec> [--operation]... | [--method --path] | [--tag]... | --all [--output dir]` ([M39](docs/spec/m39-openapi-explorer.md) — native request files, inline example/default bodies+params, bearer/basic/apikey-header → native auth blocks, unmappable features warned) + desktop explorer panel (`features/openapi-explorer/OpenapiExplorer.tsx` 304 lines, `lib/openapi` adapter, `getOpenapiBridge`) — 2026-08-30
- [x] JSON Schema: validate, inspect, generate & test assertion — `reqly schema validate/inspect/generate` ([M40](docs/spec/m40-json-schema.md)) and Goja sandbox assertion hook `reqly.assertJSONSchema(schemaPath)` ([M52](docs/spec/m52-json-schema-assertion.md), [ADR 0036](docs/adr/0036-json-schema-script-assertion.md))
- [x] XML/XSD schema validation where applicable — `internal/validation.ValidateXMLAgainstXSD` pure Go XSD parsing, DOM element/attribute constraint checking, local `schemaLocation` resolution, `reqly schema validate --type xml <schema.xsd> <instance.xml>`, Goja sandbox assertion `reqly.assertXSD(schemaPath)`, and Desktop UI ResponseViewer XML validation badge ([M49](docs/spec/m49-xml-xsd-validation.md), [ADR 0033](docs/adr/0033-xml-xsd-schema-validation.md))
- [x] Generate mocks from OpenAPI — `reqly mock [spec] [--scenario]` serves schema/example-driven responses (`internal/mocking` 0.069s, `apps/cli/cmd/mock.go` 50 lines, stateful scenarios + `MockView` GUI) — 2026-08-30

### 1.11 CLI (P0 commands)

- [x] Complete command suite: `run`, `test`, `collection run`/`test`, `ws`, `sse`, `validate`, `diff`, `env`, `mock`, `docs`, `auth`, `history`, `jwt`, `schema`, `openapi`, `perf`, `monitor`, `plugin`, `mcp`

### 1.12 Cross-platform desktop

- [x] Linux WebKit, macOS WebKit, Windows WebView2 (`Taskfile.yml` OS matrix, `release.yml`)

### 1.13 Desktop shell redesign — P0 UI Architecture

- [~] §2.1 TopBar, §2.2 Tool Rail, §2.3 Context Sidebar, §2.4 Main Workspace, §2.5 Bottom Utility Panel, §3 Design System, §4 Navigation Model, §60 Navigation Map, §61 Shared Patterns, §62 Page vs Panel Rules, §63 Final Layout Model — **core shell shipped 2026-08-27 per `docs/internal/gui-roadmap.md` GUI-0→GUI-5, but per UI Redesign Notice above the `frontend/src/` chrome is being rewritten from scratch to the four-zone model; count as core [x] / UI [~] pending**.

---

## Phase 2 — Differentiating Features (P1) — [Spec Reference](Milestones/03-phase-2-differentiating-features.md)

Features that make Reqly more capable than a basic API client.

- [x] **§56.1 OpenAPI Spec Editor** — spec editor tree + YAML CodeMirror editor
- [x] **§56.2 Schema Visualization** — schema visualizer and dependency graph
- [x] **§56.3 Request Templates** — template picker sheet in Request Builder
- [x] **§56.4 Proxy / TLS Controls** — advanced proxy auth, mTLS, custom CA in Settings & Core
- [x] **§56.5 Data-driven Testing** — dataset runner integration with CSV/JSON
- [x] **§56.6 CI/CD Integration** — GitHub Action YAML and CLI command generator in Settings
- [x] **§56.7 Full Mock Server GUI** — full mock server route editor, scenarios, fault injection, and logs
- [x] **§56.8 GraphQL / gRPC Documentation** — GraphQL schema browser and gRPC service browser with live search
- [x] **M28 HAR Import/Export + Replay** — full HAR parse/export and history replay engine
- [x] **M29 JWT Tooling & Verify/Sign** — claims decoder, HS256/384/512 signing, and signature verification
- [x] **M30 Pagination Runner** — page/offset/cursor/link-header runner with streaming
- [x] **M31 Bulk Request Execution** — concurrent batch execution engine
- [x] **M32 Retry & Resilience** — exponential backoff, status code filters, and retry header handling
- [x] **M33–37 Diff, Contract Testing & Lightweight Perf** — OpenAPI validation, diff with breaking-change detection, JSON Schema assertion, and RPS/P95/P99 perf testing suite

---

## Phase 3 — Power-User Features (P2) — [Spec Reference](Milestones/04-phase-3-power-user-features.md)

- [x] **§57.1 API Monitoring Dashboard (Milestone 38)** — scheduled requests, periodic health checks, latency trends, and CLI runner
- [x] **§57.2 Performance Testing Suite** — RPS, latency, P95/P99, error rate, status distribution, and CLI runner
- [x] **§57.3 Realtime Protocol Expansion** — WebSocket, SSE, MQTT, and Socket.IO protocol support
- [x] **§57.4 Request Dependency Graph** — request execution chaining and variable propagation graph
- [x] **§57.5 Request Replay Engine** — timeline replay and multi-environment variable substitution
- [x] **§57.8 Timeline Debugging** — request execution waterfall breakdown, timing metrics, and CLI `--timeline` flag
- [x] **M60 API Changelog & SemVer Classifier** — automated spec diff changelog generator, Markdown/JSON export, SemVer major/minor/patch bump classification, CLI `reqly changelog`, and Goja binding `reqly.generateChangelog`
- [x] **M63 Browser DevTools & Fetch Importer** — Chrome/Firefox/Safari/Edge DevTools 'Copy as fetch' parser, CLI `reqly import fetch`, and Goja sandbox binding `reqly.importFetch`
- [x] **M64 Stateful Mock Engine** — multi-scenario state machine transitions, state control endpoints `/__reqly/state`, CLI `reqly mock --scenario`, and Goja `reqly.mock` bindings
- [~] **M65 Workflow Engine (core & CLI + Desktop)** — sequential workflow execution with variable extraction, conditional step evaluation, query/header/body interpolation, `internal/workflow` + CLI `reqly workflow <file>` + Goja `reqly.workflow.run` + desktop `AppService.WorkflowRun` binding; **core/CLI/desktop [x], visual builder UI [ ] pending**
- [~] **M66 Self-Hosted Automation** — local workflow scheduler (interval + maxRuns, enabled flag, `IsEnabled`/`IntervalDuration`/`Validate`), `internal/automation` `Scheduler.Run` (immediate first run + ticker, context-cancel, `onReport` callback) + CLI `reqly automation run <file> [--once --interval --max-runs]` + desktop `AppService.AutomationRun` binding; **core/CLI/desktop [x], cron/Git-ops UI [ ] pending**

---

## Phase 4 — Ecosystem & Enterprise (P3) — [Spec Reference](Milestones/05-phase-4-ecosystem-and-enterprise.md)

- [x] **§58.1 Plugin Engine & Marketplace (Milestone 39)** — Goja JS runtime plugin execution, manifest validation, CLI manager
- [~] **§58.2 Theme Sharing & Custom Themes (M67)** — Git-native shareable themes (`id` kebab-case, `label`, `appearance` light/dark, `tokens` map), `internal/theme` (`Validate`/`Parse`/`MarshalYAML`/`MarshalJSON`/`ToCSS`/`BuiltInThemes`/`IsBuiltIn`) + CLI `reqly theme list/export/import` + desktop `AppService.ThemeList/ThemeExport/ThemeImport` bindings; **core/CLI/desktop [x], UI picker/extensions [ ] pending**
- [x] **§58.3 Git Provider Integrations (Milestone 61)** — GitHub, GitLab, Bitbucket, and Azure DevOps integration, remote auto-detection, PAT token storage, and CLI login
- [x] **M69 Audit Logs** — local append-only audit trail (`.reqly/audit.log` JSONL, 0600, `Entry{ID,Timestamp,Actor,Action,Resource,Details}` + `Validate` + 11 allowed actions), `internal/audit` (`NewStore`/`Add`/`List`/`Clear` with 0700/0600, mutex) + CLI `reqly audit list/clear` + desktop `AppService.AuditList/AuditAdd/AuditClear/AuditExport` bindings; org policies/SCIM deferred
- [x] **M70 Organization Policies** — local policy file (`.reqly/policy.yaml` 0600, `Policy{RequireAudit,MaxWorkflowSteps,AllowedActions,RequireAuth,AllowCustomThemes}` + `Validate`/`Enforce`/`EnforceWorkflow` + `DefaultPolicy`/`Load`/`Save`/`DefaultPath` 0700/0600), `internal/policy` + CLI `reqly policy show/validate/enforce` + desktop `AppService.PolicyGet/PolicySave/PolicyEnforce` bindings; SSO/SCIM/collaboration deferred
- [x] **M71 Advanced Access Control (RBAC)** — local RBAC (`.reqly/rbac.yaml` 0600, `RBAC{Roles,UserRoles}` + `Role{Name,Permissions}` + `DefaultRBAC` admin/editor/viewer + `Validate`/`Can`/`Enforce`/`ListRoles` + `Load`/`Save`/`DefaultPath` 0700/0600), `internal/rbac` + CLI `reqly rbac list/check` + desktop `AppService.RBACList/RBACCheck/RBACGet` bindings; collaboration/SSO deferred
- [~] **M72 Enterprise Secret Management (Vault)** — HashiCorp Vault KV v2 store (`VaultStore{Addr,Token,Mount,Prefix}` + `Get`/`Set`/`Delete`/`Keys` via `X-Vault-Token`, `secret/data/prefix/<key>`), `internal/secrets.VaultStore` + `NewVaultStore` validation + `VaultConfig` + `REQLY_TOKEN_STORE=vault` with `VAULT_ADDR`/`VAULT_TOKEN`/`VAULT_MOUNT`/`VAULT_PREFIX` env, fallback to file store; **Vault [x], AWS/Azure [ ] deferred to P3**
- [~] **M73 Enterprise SSO & SCIM (M73)** — local SSO (`Config{Issuer,ClientID,JWKSURL,AllowedGroups}` + `Validate`/`ValidateToken` (HMAC via `jwt.Verify`, issuer check, `IsGroupAllowed`) + `internal/sso`) + SCIM in-memory store (`User{ID,UserName,Email,Groups,Active}`/`Group{ID,DisplayName,Members}` + `ValidateUser/Group` + `Store{CreateUser/GetUser/ListUsers/DeactivateUser/CreateGroup/AddUserToGroup}` + `internal/scim`) + CLI `reqly sso validate` + `reqly scim user create/list` + desktop `AppService.SSOValidate/SCIMCreateUser/SCIMListUsers`; **local HMAC/issuer/group [x], JWKS RS256 [ ] + collaboration server [ ] deferred**
- [x] **M74 Shared Workspaces & Collaboration (M74)** — Git-native shared workspace (`.reqly/collab.yaml` 0600, `SharedWorkspace{Path,Collaborators}` + `Collaborator{User,Role,AddedAt}` + `Validate`/`AddCollaborator`/`RemoveCollaborator`/`IsCollaborator` + `Load`/`Save`/`DefaultPath` 0700/0600), `internal/collab` + CLI `reqly collab list/add/remove` + desktop `AppService.CollabList/CollabAdd/CollabRemove` bindings; self-hosted server deferred
- [x] **M75 Self-Hosted Collaboration Server (M75)** — local HTTP server for shared workspaces (`Server{root,mux}` + `NewServer` + `Handler` + `/health`/`/collab`/`/workspace` endpoints, `net.Listen` ephemeral port, `http.Serve`), `internal/collab.Server` + CLI `reqly collab serve --port` + desktop `AppService.CollabServe(port) (string,error)` binding; collaboration server shipped

---

## Phase 5 — MCP, AI & Extensibility — [Spec Reference](Milestones/06-phase-5-mcp-ai-extensibility.md)

- [x] **§59.1 Model Context Protocol (MCP) Server (Milestone 40)** — JSON-RPC 2.0 stdio server (`list_requests`, `search_requests`, `get_request`, `run_request`)
- [x] **§59.2 Command Palette** — global search and shortcut palette
- [x] **§59.3 AI Assistant Suite (Milestone 62)** — automated test assertion generator, Markdown API documentation synthesizer, failure diagnostics & remediation tips, response explanation, CLI `reqly ai <test|docs|diagnose|explain|schema>`, and Goja sandbox `reqly.ai` bindings

---

## Historical Milestone Ledger & Specifications

All detailed ticket-level milestone histories, GUI execution matrices, and complete UI architecture specifications have been organized into the **[`./Milestones/`](Milestones/)** directory:

- **[`Milestones/07-historical-milestones-ledger.md`](Milestones/07-historical-milestones-ledger.md)** — Legacy shipped milestone ledger (M01–M40)
- **[`Milestones/08-gui-roadmap-and-execution.md`](Milestones/08-gui-roadmap-and-execution.md)** — Desktop GUI roadmap and execution (GUI-0 to GUI-5)
- **[`Milestones/09-ui-architecture-shell-and-requests.md`](Milestones/09-ui-architecture-shell-and-requests.md)** — UI Architecture: Shell & Request Workspace (§1–§25)
- **[`Milestones/10-ui-architecture-tools-and-pages.md`](Milestones/10-ui-architecture-tools-and-pages.md)** — UI Architecture: Pages, Tools & Protocols (§26–§55)
- **[`Milestones/11-ui-architecture-phase-panels-and-navigation.md`](Milestones/11-ui-architecture-phase-panels-and-navigation.md)** — UI Architecture: Phase Panels, Navigation & Layout (§56–§63)

---

## Quality & Release Gates (Definition of Done)

- [x] Requirement defined in FeatureSet + CONTEXT.md
- [x] TDD cycle with unit tests (`go test ./...`)
- [x] Edge cases & error behavior covered
- [x] Security review & file permissions
- [x] Fast checks (formatting, linting, typechecking, unit tests)
- [x] PR CI + Release CI cross-platform matrix (Linux, macOS, Windows)

## Code Review Gates (`/code-review` — two-axis: Standards + Spec)

> Every phase/milestone must pass the `/code-review` skill before it is marked `[x]` in the ledger above. The skill spawns two parallel sub-agents — **Standards** (`oxlint.config.ts` + `gofmt`/`go vet` + Fowler smell baseline) and **Spec** (milestone spec vs `ROADMAP.md` DoD: core+UI/CLI+tests) — against `git diff main...HEAD` (three-dot, merge-base). Fix the diff until both axes are green. See `docs/agents/issue-tracker.md` (run `/setup-matt-pocock-skills` if missing).

- [x] Phase 0 — Foundation (`Milestones/01-phase-0-foundation.md`) — `/code-review` `main...HEAD` Standards + Spec
- [x] Phase 1 — Core API Client (P0) (`Milestones/02-phase-1-core-api-client.md`) — `/code-review` `main...HEAD` Standards + Spec
- [x] Phase 2 — Differentiating Features (P1) (`Milestones/03-phase-2-differentiating-features.md`) — `/code-review` `main...HEAD` Standards + Spec
- [x] Phase 3 — Power-User Features (P2) (`Milestones/04-phase-3-power-user-features.md`) — `/code-review` `main...HEAD` Standards + Spec
- [x] Phase 4 — Ecosystem & Enterprise (P3) (`Milestones/05-phase-4-ecosystem-and-enterprise.md`) — `/code-review` `main...HEAD` Standards + Spec
- [x] Phase 5 — MCP, AI & Extensibility (P4/P5) (`Milestones/06-phase-5-mcp-ai-extensibility.md`) — `/code-review` `main...HEAD` Standards + Spec
- [x] Historical Ledger — M01–M40 (`Milestones/07-historical-milestones-ledger.md`) — `/code-review` `main...HEAD` Standards + Spec
- [x] Full UI Architecture — §1–§63 (`Milestones/09` + `10` + `11`) — `/code-review` `main...HEAD` Standards + Spec
