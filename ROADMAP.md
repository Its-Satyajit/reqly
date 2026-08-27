# Reqly — Development Roadmap

> **Status:** Canonical development roadmap. P0/core + CLI are shipped; P1 includes shipped backend/data-layer work plus ongoing desktop/UI completion; P2–P5 remain future work.
> **Overall completion:** Do not treat a single percentage as authoritative. Use the phase/milestone checkboxes and this document's status ledger.
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
- [~] **GraphQL** — query editor + variables via `BodyType: graphql` shipped (ADR 0013); introspection/autocomplete/schema browser deferred to P1
- [x] **gRPC** — proto files, reflection, service/method discovery, unary + server-streaming — `internal/grpc` (reflection via v1 protocol, protocompile `.proto` fallback, TLS/h2c, deadlines), `grpc:` request-file block, scripting/assertions parity, history, `reqly grpc services|invoke`, desktop gRPC view (ADR 0028, M43; client-stream/bidi deferred)
- [~] **SOAP** — WSDL import, operation discovery, envelope skeletons: `reqly import wsdl <file> [--output dir]` ([M41](docs/spec/m41-wsdl-import.md) — one runnable POST per operation with binding-matched SOAP 1.1/1.2 envelopes, SOAPAction, inline-XSD body placeholders; external schemas/rpc-encoded best-effort with warnings; the "XML builder" surface is these generated envelopes, no runtime builder)

### 1.9 Import / export

- [x] Import cURL — `reqly import curl` (method, headers, JSON/raw/data bodies, basic auth, user-agent, cookies, GET-style query data; unsupported features reported)
- [x] Import OpenAPI 3.x — `reqly import openapi` (servers, paths, operations, params, JSON bodies; writes a Git-native workspace)
- [x] Export Postman collection v2.1 — `reqly export postman` (flat list, inherited base URL/headers applied)
- [~] Import: Postman v2.1 ([M34](docs/spec/m34-postman-import.md) — requests, nested folders, variables, bodies raw/urlencoded/form-data/graphql, basic/bearer/apikey auth; scripts + file bodies warned), Insomnia v4/v5 ([M35](docs/spec/m35-insomnia-import.md) — both formats auto-detected, nested folders, environments as native `environments/*.yaml`, basic/bearer/apikey/digest auth; cookie jars + unsupported auth warned), Bruno ([M36](docs/spec/m36-bruno-import.md) — items tree, body modes, collection-level auth/headers defaults, secret-split environments), Swagger 2.x (via hand-rolled parser); HAR done ([M28](docs/spec/m28-har-import-export.md))
- [x] Export: requests ([`export workspace`](docs/adr/0017-workspace-save-export.md) + `export code`), OpenAPI 3.0 spec generation (`export openapi`, [M37](docs/spec/m37-export-reports-openapi.md)), responses (`export har` from history, M28 + desktop download), test results (`collection test --report-junit/--report-json`, M37); docs done (§1.11 `reqly docs`)
- [x] Import preservation (env/auth/scripts) + unsupported-feature reporting — [M42](docs/spec/m42-import-preservation.md) ([ADR 0026](docs/adr/0026-import-preservation-script-translation.md)): Postman/Bruno pre/post scripts translated onto the `reqly.*` sandbox into `preRequest`/`postRequest` (unmappable lines preserved as `TODO(reqly-import)` comments), Postman collection variables → `environments/<collection>.yaml`, structured `ImportReport` across all importers with CLI grouped summary (2026-08-24)

### 1.10 OpenAPI & JSON Schema

- [~] OpenAPI 3.x parse + validate — `internal/openapi` (kin-openapi, JSON/YAML, $ref resolution); OpenAPI 2.x import via hand-rolled parser; 3.1 partial
- [~] Endpoint explorer + generate requests from spec — `reqly openapi explore <spec> [--tag]... [--json]` (operation table / machine-readable list) and `reqly openapi generate <spec> [--operation]... | [--method --path] | [--tag]... | --all [--output dir]` ([M39](docs/spec/m39-openapi-explorer.md) — native request files, inline example/default bodies+params, bearer/basic/apikey-header → native auth blocks, unmappable features warned; desktop explorer panel deferred to M39b)
- [~] JSON Schema: validate, inspect, generate — `reqly schema validate <schema> [instance|-]` (draft detection + override, instance-path violations, stdin/--json), `reqly schema inspect <schema>` (tree summary, resolved $refs), `reqly schema generate <schema> [--seed] [--optional]` (deterministic synthesis honoring examples/defaults/constraints) ([M40](docs/spec/m40-json-schema.md); in-app *edit* deferred to M40b, test-assertion hook to §35)
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

## Phase 2 — Differentiating Features (P1)

Features that make Reqly more capable than a basic API client.
**Spec:** [`docs/Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification.md`](docs/Reqly%20Complete%20UI%20Architecture,%20Pages,%20Panels,%20and%20Navigation%20Specification.md) §56

> **GUI milestones:** Most Phase 2 features have core/CLI shipped but no desktop GUI.
> See [`docs/internal/gui-roadmap.md`](docs/internal/gui-roadmap.md) for GUI-specific tracking (GUI-5 through GUI-16).

### §56.1 OpenAPI Spec Editor

- [x] Spec editor tree + YAML CodeMirror editor — `features/spec-editor/SpecEditorView.tsx` + `stores/useSpecEditorStore.ts` — 2026-08-26
- [ ] Interactive endpoint editing with validation (P1 GUI pending)

### §56.2 Schema Visualization

- [x] Schema graph — `lib/schemaGraph.ts` + `lib/schemaGraph.test.ts` — 2026-08-26
- [ ] Interactive graph UI with zoom/pan/node selection (P1 GUI pending)

### §56.3 Request Templates

- [x] Request templates — zustand store + pure lib (search, instantiate, CRUD) + 21 tests — 2026-08-26
- [ ] Template picker UI in request builder (P1 GUI pending)

### §56.4 Proxy / TLS Controls

- [x] Proxy & TLS controls — zustand store + pure lib (validate, format, defaults) + 22 tests — 2026-08-26
- [ ] Proxy/TLS configuration panel UI (P1 GUI pending)

### §56.5 Data-driven Testing

- [x] Data-driven testing — CSV/JSON dataset lib + zustand store + 23 tests — 2026-08-26
- [ ] Dataset picker + runner integration UI (P1 GUI pending)

### §56.6 CI/CD Integration

- [x] CI/CD support — CLI command generation + GitHub Action YAML + zustand store + 13 tests — 2026-08-26
- [ ] CI/CD configuration panel UI (P1 GUI pending)

### §56.7 Full Mock Server GUI

- [~] Mock server — CLI `reqly mock` with path/method matching, schema/example generation, `--delay`, `--fail-every`; stateful mocks, scenarios, fault injection, and zustand store shipped — 2026-08-26
- [ ] Full mock server GUI with route editor, scenario manager, logs viewer (P1 GUI pending)

### §56.8 GraphQL / gRPC Documentation

- [x] GraphQL/gRPC docs — zustand store + pure lib (SDL parse, search, Markdown render) + 16 tests — 2026-08-26
- [ ] Documentation browser UI (P1 GUI pending)

### Other P1 Items

- [x] Code generation — `internal/exporter.Generate` (request → cURL/JS/Python/Go, header/body/auth, `[SECRET]` masked, `reqly export code` + desktop `Copy as`, golden files) — [ADR 0016](docs/adr/0016-code-generation.md)
- [x] API diff + breaking-change detection (endpoints, params, schemas, auth, response types) — `internal/diffing` (`OpenAPIFiles` structural diff + `breaking.go` severity classification), `reqly diff <file1> <file2>`, desktop Diff view
- [x] Request/response diff (JSON structural) — `diffing.JSON` + `reqly diff`
- [x] Environment diff — `reqly env diff` + desktop env tools panel
- [~] HAR import/export + replay — import (`internal/importer/har.go`) + export (`internal/exporter/har.go`) shipped; HAR-specific replay pending (history replay via `HistoryReplay` shipped)
- [~] JWT tooling (decode, claims viewer, signing) — decode/claims viewer + expiry detection (`reqly jwt decode`, ADR 0021) + per-request HS256/384/512 signing shipped; `verify`/`sign` CLI deferred to M29b
- [~] GraphQL introspection / gRPC reflection tooling — GraphQL schema introspection + summary shipped (`internal/graphql/introspect.go`, desktop GraphQL browser); gRPC reflection not started
- [ ] Advanced HTTP: HTTP/2, HTTP/3, streaming, chunked transfer, keep-alive
- [x] Pagination runner (page/offset/cursor/link-header, stop conditions, aggregation) — `internal/pagination` + `reqly pagination run` ([ADR 0022](docs/adr/0022-pagination-runner.md)) + desktop runners panel
- [x] Bulk request execution (CSV/JSON inputs, sequential/parallel, concurrency) — `internal/bulk` + `reqly bulk run --data` ([ADR 0023](docs/adr/0023-bulk-runner.md)) + desktop runners panel
- [x] Retry & resilience — engine-level `request.retry` block ([ADR 0024](docs/adr/0024-retry-resilience.md))
- [~] API documentation generation (REST + GraphQL + realtime) — REST shipped: `reqly docs generate` + desktop Docs panel (G-15); GraphQL SDL parser + zustand store shipped (2026-08-26); realtime doc output deferred

---

## Phase 3 — Power-User Features (P2)

Advanced functionality for experienced developers and teams.
**Spec:** [`docs/Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification.md`](docs/Reqly%20Complete%20UI%20Architecture,%20Pages,%20Panels,%20and%20Navigation%20Specification.md) §57

### §57.1 API Monitoring Dashboard

- [ ] Scheduled requests/collections, health checks, latency/availability, alerts

### §57.2 Performance Testing

- [ ] RPS, latency, P95/P99, error rate, status distribution

### §57.3 MQTT / Socket.IO

- [ ] MQTT publish/subscribe, topics, QoS, retained/will, auth, TLS
- [ ] Socket.IO connections, events, rooms, namespaces, debugging

### §57.4 Dependency Graph

- [ ] API dependency graph visualization

### §57.5 Request Replay

- [ ] Exact / modified vars / other env / captured traffic replay

### §57.6 In-app Developer Tools / Debugger

- [ ] Request/auth/variables/script/runtime/network inspection

### §57.7 Git GUI

- [ ] Init/commit/branch/diff/history/pull/push/merge/conflicts

### §57.8 Network Interception / Timeline Debugging

- [ ] Capture/inspect/import/modify/replay network traffic
- [ ] Request timeline debugging (DNS/connect/TLS/request/server/response/transfer)

### Other P2 Items

- [ ] API changelog (from specs + Git changes)
- [ ] Browser integrations (DevTools import, cURL copy, Chrome/Firefox/Safari)
- [ ] Advanced mock state (multi-scenario state machines)
- [ ] Visual workflow builder
- [ ] Self-hosted automation

---

## Phase 4 — Ecosystem & Enterprise (P3)

Long-term ecosystem and organization features.
**Spec:** [`docs/Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification.md`](docs/Reqly%20Complete%20UI%20Architecture,%20Pages,%20Panels,%20and%20Navigation%20Specification.md) §58

### §58.1 Plugin Marketplace

- [ ] Plugin system (auth, template tags, request/response processing, protocols, UI)

### §58.2 Theme Marketplace

- [ ] Theme sharing + custom themes + UI extensions

### §58.3 Git Provider Integrations

- [ ] GitHub, GitLab, Bitbucket, Azure DevOps + PATs

### §58.4 Team / Shared Workspaces

- [ ] Multi-user collaboration, shared workspaces

### §58.5 Enterprise

- [ ] Self-hosted collaboration server
- [ ] Enterprise SSO, SCIM provisioning
- [ ] Audit logs, organization policies
- [ ] Enterprise secret management (Vault, AWS, Azure, role-based access)
- [ ] Advanced access control / permissions

---

## Phase 5 — MCP, AI & Extensibility (cross-cutting)

**Spec:** [`docs/Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification.md`](docs/Reqly%20Complete%20UI%20Architecture,%20Pages,%20Panels,%20and%20Navigation%20Specification.md) §59

### §59.1 MCP Server

- [ ] `internal/mcp` — list/search/run requests & collections, inspect schemas, retrieve responses, generate docs

### §59.2 Command Palette

- [x] Command palette + spotlight, keyboard shortcuts, context menus, widgets, code snippets — shipped (2026-08-26)

### §59.3 Optional AI Assistant

- [ ] Request generation, response explanation, test/docs generation, error analysis, schema assistance, breaking-change explanation

---

## Future documentation re-seeding / redesign plan

This is a planned future workstream. It does not outrank the product roadmap; it exists to keep the roadmap set from drifting again.

### Documentation consolidation
- [ ] Keep this file as the canonical product roadmap.
- [ ] Retire or clearly mark superseded duplicate roadmap files after the current implementation state has been migrated.
- [ ] Keep the GUI roadmap as the desktop execution tracker, with links back to the product milestone that owns each feature.
- [ ] Keep the complete UI architecture document as the lower-precedence UI reference.
- [ ] Preserve historical milestone IDs, issue links, ADR links, tests, implementation notes, and shipped dates during the migration.
- [ ] Replace contradictory status snapshots with one current status and a short historical note where necessary.
- [ ] Add a traceability map: roadmap milestone → core implementation → CLI → desktop/UI → tests → docs/ADR.
- [ ] Audit every `[x]`, `[~]`, and `[ ]` against the real repository before declaring the consolidated document authoritative.
- [ ] Re-run the consolidation whenever a major milestone or UI architecture revision lands.

### Cross-document re-seeding
- [ ] Re-seed the product roadmap from the shipped milestone history first.
- [ ] Re-seed the GUI roadmap from the product roadmap second.
- [ ] Re-seed the UI architecture navigation and panel inventory third.
- [ ] Reconcile page names, tool names, protocol names, runner names, and terminology across all three layers.
- [ ] Keep UI-only polish and layout proposals from changing product priority unless a development-roadmap milestone explicitly adopts them.
- [ ] Preserve deferred seams as explicit follow-up work rather than silently dropping them.

### Definition of done for the documentation redesign
- [ ] Every source feature appears exactly once in the canonical product roadmap or in a clearly labeled historical/reference section.
- [ ] Every GUI-specific implementation task points to a product-roadmap owner.
- [ ] Every UI-spec page, panel, dialog, interaction pattern, navigation node, and layout rule is still represented in the subordinate UI reference.
- [ ] No shipped item is accidentally regressed to `[ ]` because an older snapshot said it was pending.
- [ ] No UI specification item is promoted to product scope merely because it appears in the UI reference.

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

| Phase   | Scope                    | Status                                                                                                                                     | Est. complete        |
| ------- | ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------ | -------------------- |
| Phase 0 | Foundation               | 100% — repo/build infra + Wails shell + UI shell + all core primitives + CLI skeleton + release pipeline                                   | 100%                 |
| Phase 1 | Core API Client (P0)     | 100% backend + CLI; UI shell §2 four-zone chrome shipped (Tickets #01–#12, 2026-08-27)                                                    | 100% backend, 100% UI |
| Phase 2 | Differentiating (P1)     | 100% — §56.1–56.8 data layer (lib + stores + tests) + all UI panels (template picker, proxy/TLS, dataset, CI/CD, mock GUI, schema docs) shipped 2026-08-27 | 100%  |
| Phase 3 | Power-User (P2)          | 0% — §57.1–57.8 not started                                                                                                                | 0%                   |
| Phase 4 | Ecosystem (P3)           | 0% — §58.1–58.5 not started                                                                                                                | 0%                   |
| Phase 5 | MCP / AI / Extensibility | ~10% — §59.2 command palette shipped (2026-08-26); §59.1 MCP server stub shipped (M56); §59.3 AI heuristic shipped (M58); full AI assistant pending | ~10%        |
| Quality | DoD + release gates      | ~55% — Fast + PR CI green; E2E/Playwright + Vitest + full perf/security compat pending                                                     | ~55%                 |

### Next milestones (UI redesign — spec §2 status)

1. ~~**§2.1 TopBar** — Logo, Workspace Switcher, Global Search ⌘K, Import/Export, Settings~~ ✅ shipped 2026-08-27 (Active Env selector pending — Ticket #12)
2. ~~**§2.2 Tool Rail** — 48–56px, 4 groups (Workspace/API Tools/Realtime/System), icon-based routing~~ ✅ shipped 2026-08-27
3. ~~**§2.3 Context Sidebar** — 220–280px, collapsible/resizable, per-tool content, `⌘B` toggle~~ ✅ shipped 2026-08-27 (Recent/pinned pending)
4. ~~**§2.4 Main Workspace** — Tab-based content area, page routing per active tool~~ ✅ shipped 2026-08-27
5. ~~**§2.5 Bottom Panel** — Console/Network/Tests/Variables/Cookies, `⌘J` toggle, resizable~~ ✅ shipped 2026-08-27
6. ~~**§3 Design System** — Tokens, typography (IBM Plex), color system (terracotta accent, BASE 6px)~~ ✅ shipped 2026-08-27
7. ~~**§60 Navigation Map** — 15+ full pages with sub-panels~~ ✅ shipped 2026-08-27
8. ~~**§61–63** — Shared patterns, page/panel rules, final layout model~~ ✅ shipped 2026-08-27

### Next milestones (Tickets #08–#12 + P1 GUI panels)

- [x] **Ticket #08** — API Tools polish: consistent PageHeader across Diff/JWT/GraphQL/gRPC/Runners/Explorer/Docs/Mocks/Settings views — 2026-08-27
- [x] **Ticket #10** — Settings view polish: Proxy/TLS section (`ProxyTlsPanel`), CI/CD section (`CicdPanel`), shortcuts — 2026-08-27
- [x] **Ticket #11** — Bottom Panel tab content wiring: Variables tab (active + chain), Cookies tab (parsed), Network, Tests — 2026-08-27
- [x] **Ticket #12** — TopBar Active Environment selector: `EnvironmentSelector` dropdown in TopBar — 2026-08-27
- [x] **§56.3 GUI** — Template picker sheet in Request Builder (`TemplatePickerSheet`) — 2026-08-27
- [x] **§56.4 GUI** — Proxy/TLS configuration panel in Settings (`ProxyTlsPanel`) — 2026-08-27
- [x] **§56.5 GUI** — Dataset picker in Runners panel (file loader + preview) — 2026-08-27
- [x] **§56.6 GUI** — CI/CD configuration panel in Settings (`CicdPanel`) — 2026-08-27
- [x] **§56.7 GUI** — Mock server full GUI: route editor, scenarios, fault injection, logs viewer — 2026-08-27
- [x] **§56.8 GUI** — GraphQL schema browser + gRPC service browser with live search — 2026-08-27

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
33. ~~**OpenAPI editor + endpoint explorer**~~ — in-app spec authoring + generate requests from spec + JSON Schema edit/validate (`reqly openapi validate/explore/generate`, Desktop explorer with Try in Builder + schema inspection) — **shipped**
34. **API diff & breaking-change detection** — endpoints/params/schemas/auth/response-types + spec/request/response/env diff polish
35. **Contract testing + schema validation** — OpenAPI/JSON Schema response validation pipeline
36. **Advanced HTTP / Proxy & TLS controls** — HTTP/2, per-env/per-request proxy, cert inspection, mTLS, custom CAs
37. **Performance testing (lightweight)** — RPS/latency P95/P99/error-rate/status-distribution

> **Companion:** [**reqly-test-api**](https://reqly-test-api.vercel.app) — a small ElysiaJS mock API (Vercel-hosted, hardcoded data) for exercising `reqly run`/`test`, auth, delay, and error-status flows against a real endpoint. Useful while the in-app mock server (milestone 7) is pending; see the README's "Mock API" section.

---

# Historical milestone detail retained from the development roadmap

The following ticket-level milestone history is preserved from the older development roadmap so the consolidation does not discard implementation detail, issue references, PR references, ADR references, or the exact sequencing that led to the current state.

> **Authority:** Historical detail does not override the current status in the canonical roadmap above.

## Legacy shipped milestone ledger
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

## Early P1 milestone detail

### Next milestones — P1 Differentiating Features (suggested order)

28. **HAR import/export + replay** — `internal/importer` HAR parse + `reqly import har <har-file> [--output <dir>] [--collection <name>]` ( `headers+cookies→Headers` `Cookie:` merged, `queryString→Query`, `postData.text→Body` base64 decoded, `mimeType→Content-Type`, >1MB spill `blobs/`, `pageref`/`timings`/`cache` warnings) + `reqly export har [--out <file.har>] [--env <name>] [--limit 500]` history→HAR via `internal/exporter/har.go` (`ExportHAR` pure, `timings` synthesized, base64 binary, secrets masked), replay via `har-import` collection + `history replay` ([ADR 0020](docs/adr/0020-har-import-export.md), CONTEXT `HAR`/`HAR Import`/`HAR Export`/`HAR Replay` grilling Q1–Q4 done, `docs/spec/m28-har-import-export.md`) — **shipped**
29. **JWT tooling** — `reqly jwt decode` (header/claims viewer, expiry detection) in `internal/jwt` + `reqly jwt decode [--json]` + `Bearer`/stdin (`internal/jwt.Decode` + `apps/cli/cmd/jwt.go`, expiry `exp`/`nbf`/`iat` → `expired`/`not_yet_valid`/`valid`/`no_expiry`, `Header:`/`Payload:` pretty + `--json`, [ADR 0021](docs/adr/0021-jwt-tooling-decode.md), CONTEXT `JWT Tooling`/`JWT Decode` grill Q1–Q5) — **shipped (decode MVP)**; `verify`/`sign` (HS via `jwtHashes`) + desktop inspector deferred to M29b
30. **Pagination runner** — `reqly pagination run <request-file> [--max-pages <n>]` ( `request.pagination: {strategy: page|offset|cursor|link-header, pageParam/pageSizeParam/offsetParam/limitParam/cursorParam, nextPath: $.nextCursor, maxPages: 100}` + `internal/pagination.Run` pure loop over `sendFn` `page`→`?page=1→2` `offset`→`?offset=0→10` `cursor`→`?cursor=<next>` via JSONPath `$.nextCursor` `link-header`→`Link: <url>; rel="next"` , stop empty/missing-next/non-2xx/maxPages, `--max-pages` overrides, `OnStep` streaming `step: status duration url`) ([ADR 0022](docs/adr/0022-pagination-runner.md), CONTEXT `Pagination Runner` `Strategy`/`Stop` grill Q1–Q4, `docs/spec/m30-pagination-runner.md`) — **shipped**
31. **Bulk request execution** — `reqly bulk run <request-file> --data <csv|json> [--parallel] [--concurrency <n>] [--continue-on-error]` (CSV header→`{{var}}`/JSON array stringified, `internal/bulk.Run` sequential default, parallel semaphore ordered `concurrency 5`, `ScopeRuntime` per row, stop first non-2xx unless `--continue-on-error`) ([ADR 0023](docs/adr/0023-bulk-runner.md), CONTEXT `Bulk Runner`/`Bulk Input Row`/`Bulk Concurrency` grill Q1–Q4, `docs/spec/m31-bulk-runner.md`) — **shipped**
32. ~~**Retry & resilience**~~ — engine-level `request.retry` (`count`/`delayMs`/`strategy`/`maxDelayMs`/`retryOn`) in `Client.Execute`; network errors + 429/502/503/504 default, `Retry-After` respected + clamped, exponential/fixed backoff capped, ctx-cancel aborts mid-wait, auth refresh stays inside one attempt, `response.Attempts` + `history show` attempts line + desktop attempts badge, `--retries`/`--retry-delay` flags, desktop collapsible Retry section in the request editor ([ADR 0024](docs/adr/0024-retry-resilience.md), `docs/spec/m32-retry-resilience.md`) — **shipped**
33. ~~**OpenAPI editor + endpoint explorer**~~ — in-app spec authoring + generate requests from spec + JSON Schema edit/validate (`reqly openapi validate/explore/generate`, Desktop explorer with Try in Builder + schema inspection) — **shipped**
34. **API diff & breaking-change detection** — endpoints/params/schemas/auth/response-types + spec/request/response/env diff polish
35. **Contract testing + schema validation** — OpenAPI/JSON Schema response validation pipeline
36. **Advanced HTTP / Proxy & TLS controls** — HTTP/2, per-env/per-request proxy, cert inspection, mTLS, custom CAs
37. **Performance testing (lightweight)** — RPS/latency P95/P99/error-rate/status-distribution


---

# Lower-precedence desktop GUI execution roadmap

This section preserves `gui-roadmap.md` in full. It is subordinate to the development roadmap above. GUI status can clarify implementation state, but it cannot redefine product scope or phase priority.

# Reqly — GUI Roadmap

> Desktop GUI milestones — tracks `apps/desktop/backend/frontend` delivery.
> **Source of truth:** [`docs/Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification.md`](../Reqly%20Complete%20UI%20Architecture,%20Pages,%20Panels,%20and%20Navigation%20Specification.md) (full UI spec — §1–§59)

## GUI-0 Shell Redesign — RESTARTING FROM SCRATCH

> **⚠️ RESTARTING** — Previous implementation (G-3.x) did not follow spec §2 four-zone architecture. All UI components will be rewritten following the spec's TopBar / ToolRail / ContextSidebar / MainWorkspace / BottomPanel model.
> **Progress 2026-08-27:** Tickets #01 Shell Foundation (`wzm`/`mny`), #02 Workspace Home (`swt`), #03 Request Builder + Response Viewer (`muu`), #04 Collections Explorer (`pvo`), #05–06 Environments/History (`kvs`), #07 Mocks (`syw`) shipped — spec §2 chrome + workspace + history + mocks complete; #08–12 (Tools/Import-Export/Settings/Panels/Search) existent — pending polish/review.

### GUI-0.1 Shell Chrome (Tickets #01–#02)

- [x] TopBar — Logo, Import/Export, Settings, Sync indicator — 2026-08-27
- [x] ToolRail — 4 groups (Workspace/API Tools/Realtime/System), collapsed 56/40px — 2026-08-27
- [x] StatusBar — theme tokens, empty placeholders — 2026-08-27
- [x] Workspace Home — stat cards spec-compliant + empty-state onboarding — 2026-08-27

### GUI-0.2 Request Workspace (Ticket #03)

- [x] Request tabs — open/close/pin/duplicate/drag-reorder, persist via localStorage — 2026-08-27
- [x] URL bar — methods GET…TRACE, Send/Save, theme tokens — 2026-08-27
- [x] Builder tabs Params/Headers/Body/Auth + overflow Pre-request/Tests/Docs/Settings — 2026-08-27
- [x] Body types None/JSON/XML/Text/HTML/Form/Binary/GraphQL — 2026-08-27
- [x] Auth Custom + OAuth2 three flows — 2026-08-27
- [x] Response Viewer — Body/Headers/Cookies/Test Results/Timeline, vertical/horizontal split — 2026-08-27

### GUI-0.3 Collections Explorer (Ticket #04)

- [x] Collections tree — expand/collapse, icons, keyboard nav — 2026-08-27
- [x] Search/filter — 2026-08-27
- [x] Drag-and-drop reordering — 2026-08-27
- [x] Context menu — Rename/Move/Duplicate/Delete/Run/Import/Export/Generate — 2026-08-27
- [x] New Collection/Folder/Request buttons — 2026-08-27

### §2.1 TopBar (always visible)

- [x] Logo — 2026-08-27
- [x] Workspace Switcher (folder open) — 2026-08-27
- [x] Global Search ⌘K — 2026-08-27
- [x] Import / Export buttons — 2026-08-27
- [x] Active Environment selector (Ticket #12) — 2026-08-27
- [x] Sync Status indicator (Git local-first save indicator) — 2026-08-27
- [x] Settings — 2026-08-27

### §2.2 Tool Rail (48–56px, left-most)

- [x] Workspace group: Home, Requests, Environments, History — 2026-08-27
- [x] API Tools group: Mocks, Diff, JWT, GraphQL, gRPC, Runners, Explorer, Docs — 2026-08-27
- [x] Realtime group: WebSocket, SSE — 2026-08-27
- [x] System group: Settings — 2026-08-27
- [x] Icon-based routing (top-level navigation) — 2026-08-27

### §2.3 Context Sidebar (220–280px)

- [x] Collapsible/resizable (drag handle) — 2026-08-27
- [x] Changes per active tool — 2026-08-27
- [x] Tree navigation — 2026-08-27
- [x] Search within tool — 2026-08-27
- [x] Contextual actions — 2026-08-27
- [x] Recent/pinned items (History recents & realtime endpoints) — 2026-08-27
- [x] `⌘B` toggle — 2026-08-27

### §2.4 Main Workspace

- [x] Tab-based content area — 2026-08-27
- [x] Page routing per active tool — 2026-08-27
- [x] Full pages vs context panels (§62 rules) — 2026-08-27

### §2.5 Bottom Utility Panel

- [x] Console tab — 2026-08-27
- [x] Network tab — 2026-08-27
- [x] Tests tab — 2026-08-27
- [x] Variables tab — 2026-08-27
- [x] Cookies tab — 2026-08-27
- [x] `⌘J` toggle — 2026-08-27
- [x] Resizable height — 2026-08-27

## GUI-1 Design System (spec §3) — ✅ shipped 2026-08-27

- [x] Design tokens (spacing, radius, shadows) — `frontend/src/index.css` `@theme` semantic vars (`--background`/`--border`/`--primary`/`--status-*`/`--radius: 0.375rem`), grep gate for hardcoded hex
- [x] Typography system (IBM Plex Sans/Mono) — `@fontsource` 400/500/600 + 13px/1.45 base, `.font-data` mono discipline
- [x] Color system (terracotta accent #c93517/#ff6f52, BASE 6px radius) — AA-adjusted `#c93517` (4.5:1) + `prefers-contrast: more` bump, `DESIGN.md` Color/Tokens sections
- [x] Status indicators (Connected/Running/Valid/Success/Warning/Error) — Status Ramp + `StatusPill` dot+code (never color alone), method tints
- [x] Hairline borders, no shadows — `border-border` shell/cards/panels; `shadow-md/lg` only on floating `popover`/`dropdown-menu`/`select`/`toast`

## GUI-2 Navigation Model (spec §4, §60–63) — ✅ shipped 2026-08-27

- [x] Two-axis navigation: horizontal (tool rail) + vertical (sidebar) — TopBar/ToolRail/Sidebar/Workspace/BottomPanel, `⌘B`/`⌘J` persisted
- [x] 15+ full pages with sub-panels — Home/Requests/Environments/History/Mocks/Diff/JWT/GraphQL/gRPC/Runners/Explorer/Docs/WS/SSE/Settings, lazy per-tool
- [x] Page vs panel rules (§62) — page=tool route, sidebar=resource nav, bottom=inspector, dialog=transient
- [x] Shared interaction patterns (§61) — `⌘K` palette, per-tool filter, StatusPill, tabs/button primitives
- [x] Final layout model (§63) — canonical five-zone shell as single source of truth

## GUI-5 P1 Data Layer (spec §56.3–56.8) — PRESERVED

> These data layer items (lib + stores + tests) are preserved from the previous implementation.

- [x] **G-5.1** Request templates — zustand store + pure lib (search, instantiate, CRUD) + 21 tests — 2026-08-26
- [x] **G-5.2** Proxy/TLS controls — zustand store + pure lib (validate, format, defaults) + 22 tests — 2026-08-26
- [x] **G-5.3** Data-driven testing — zustand store + pure lib (CSV/JSON parse, row vars, validate) + 23 tests — 2026-08-26
- [x] **G-5.4** CI/CD integration — zustand store + pure lib (CLI gen, GitHub Action YAML, report parse) + 13 tests — 2026-08-26
- [x] **G-5.5** Mock server GUI data — extended zustand store + pure lib (scenarios, fault injection, matchers, logs) + 20 tests — 2026-08-26
- [x] **G-5.6** GraphQL/gRPC docs — zustand store + pure lib (SDL parse, search, Markdown render) + 16 tests — 2026-08-26


---

# Lower-precedence UI architecture reference

This appendix preserves the complete `Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification` as a reference inventory. It is intentionally subordinate to the development roadmaps and the GUI execution roadmap.

Rules for using this appendix:
- Its page/panel/navigation details are implementation guidance.
- Its presence does not create a new product commitment.
- Its layout or naming proposals do not override roadmap priority.
- When it contains a UI idea that is absent from the development roadmap, treat that idea as a candidate/reference until a roadmap milestone adopts it.
- When roadmap status and UI reference status differ, roadmap status wins.

# Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification

## 1. Purpose

This document is the consolidated UI specification for Reqly.

It includes the full set of pages, panels, sub-panels, editors, inspectors, dialogs, navigation surfaces, workspace controls, debugging surfaces, and roadmap features discussed previously.

The architecture is based on four persistent UI layers:

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ TOP BAR                                                                      │
├──────┬───────────────────────┬───────────────────────────────────────────────┤
│      │                       │                                               │
│ TOOL │ CONTEXT SIDEBAR       │ MAIN WORKSPACE                               │
│ RAIL │                       │                                               │
│      │                       │                                               │
│      │                       │                                               │
├──────┴───────────────────────┴───────────────────────────────────────────────┤
│ BOTTOM UTILITY PANEL                                                         │
└──────────────────────────────────────────────────────────────────────────────┘
```

The model is:

```text
Workspace
   ↓
Tool Rail
   ↓
Context Sidebar
   ↓
Page / Main Workspace
   ↓
Page-specific Panels
   ↓
Bottom Utility Panels
```

---

# 2. Global Application Shell

## 2.1 Top Bar

The top bar is always available.

```text
┌──────────────────────────────────────────────────────────────────────────────────────────────┐
│ ◉ REQLY     Workspace ▾     ⌕ Search      Import  Export      Development ▾     ● Sync   ⚙  │
└──────────────────────────────────────────────────────────────────────────────────────────────┘
```

### Components

```text
Top Bar
├── Reqly Logo
├── Workspace Switcher
├── Global Search
├── Import
├── Export
├── Active Environment
├── Sync / Connection Status
├── Notifications
├── Settings
└── User / Account Menu
```

## 2.2 Workspace Switcher

```text
┌─────────────────────────────────────────┐
│ WORKSPACE                               │
├─────────────────────────────────────────┤
│ 🔍 Search workspaces...                 │
│                                         │
│ PERSONAL                                │
│ ● My Workspace                          │
│                                         │
│ PROJECTS                                │
│ ◈ Reqly API                             │
│ ◈ Payments                              │
│ ◈ Internal Services                     │
│                                         │
│ SHARED                                  │
│ 👥 Backend Team                         │
│ 👥 Engineering                          │
│                                         │
│ ──────────────────────────────────────  │
│ + Create Workspace                      │
│ ⚙ Manage Workspaces                     │
└─────────────────────────────────────────┘
```

## 2.3 Global Search

Global search should search across:

```text
Requests
Collections
Folders
Environments
Variables
History
Mocks
OpenAPI
GraphQL operations
gRPC methods
Tests
Documentation
Workspaces
Commands
```

Example:

```text
┌─────────────────────────────────────────────────────────────┐
│ ⌕ Search Reqly...                                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ Requests                                                    │
│   GET /users                                                │
│   POST /users                                               │
│                                                             │
│ Collections                                                │
│   Users                                                     │
│                                                             │
│ History                                                     │
│   GET /users?page=2                                        │
│                                                             │
│ Commands                                                    │
│   Open Environments                                        │
│   New Request                                               │
└─────────────────────────────────────────────────────────────┘
```

---

# 3. Tool Rail

The tool rail is approximately 48–56px wide.

```text
┌──────┐
│  ◉   │ Workspace
├──────┤
│  ⚡  │ Requests
│  ◈   │ Environments
│  ≋   │ History
├──────┤
│  ◫   │ Mocks
│  ⇄   │ Diff
│  ♢   │ JWT
│  ◎   │ GraphQL
│  ⌁   │ gRPC
│  ▶   │ Runners
│  ◇   │ Explorer
│  ▤   │ Docs
├──────┤
│  ◌   │ WebSocket
│  ◌   │ SSE
├──────┤
│  ⚙   │ Settings
└──────┘
```

## Tool groups

```text
WORKSPACE
├── Workspace
├── Requests
├── Environments
└── History

API TOOLS
├── Mocks
├── Diff
├── JWT Inspector
├── GraphQL
├── gRPC
├── Runners
├── Explorer
└── Docs

REALTIME
├── WebSocket
└── SSE

SYSTEM
└── Settings
```

---

# 4. Context Sidebar

The second sidebar changes according to the active tool.

It answers:

> What am I working with inside this tool?

Typical width:

```text
220–280px
```

It supports:

* collapse
* resize
* search
* tree navigation
* contextual actions
* recent items
* pinned items

Shortcut:

```text
Ctrl/Cmd + B
```

---

# 5. Workspace Home / Overview

The workspace itself should have a landing page.

```text
┌──────────────────────────────────────────────────────────────────────────────────────┐
│ WORKSPACE: Reqly API                                                                  │
├──────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                      │
│ Welcome back                                                                          │
│                                                                                      │
│ ┌────────────────┐ ┌────────────────┐ ┌────────────────┐ ┌────────────────────────┐ │
│ │ Requests       │ │ Environments   │ │ Collections    │ │ Recent Activity        │ │
│ │ 128            │ │ 4              │ │ 12             │ │ 18 requests today     │ │
│ └────────────────┘ └────────────────┘ └────────────────┘ └────────────────────────┘ │
│                                                                                      │
│ QUICK ACTIONS                                                                         │
│ [ New Request ] [ Import API ] [ Open Collection ] [ New Environment ]               │
│                                                                                      │
│ RECENT REQUESTS                                                                       │
│ GET /users                                      200     142 ms                        │
│ POST /users                                     201     183 ms                        │
│ GET /products                                   200      91 ms                        │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

---

# 6. Requests Page

Requests are Reqly's primary workspace.

The page contains:

```text
Requests
├── Request Tabs
├── Request URL Bar
├── Request Builder
├── Request Metadata
├── Response Viewer
└── Request Actions
```

---

# 7. Request Tabs

Requests should behave similarly to editor tabs.

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ ● GET Users    ● Create User    ● Get Product    +                           │
└──────────────────────────────────────────────────────────────────────────────┘
```

Tab features:

* open
* close
* pin
* duplicate
* reopen closed tab
* close others
* close all
* unsaved indicator
* drag reorder
* context menu
* restore previous session

---

# 8. Request URL Bar

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ GET ▾ │ https://api.example.com/v1/users/{{user_id}}           Send ▶ ▾       │
└──────────────────────────────────────────────────────────────────────────────┘
```

Features:

* method selector
* URL editor
* autocomplete
* path variable detection
* environment variables
* request history
* Send
* Send with options
* cancel request
* save request
* duplicate request

Methods:

```text
GET
POST
PUT
PATCH
DELETE
HEAD
OPTIONS
CONNECT
TRACE
```

---

# 9. Request Builder

## 9.1 Request Builder Navigation

```text
Params
Headers
Body
Auth
Pre-request
Tests
Docs
Settings
```

These are individual request panels.

---

# 10. Params Panel

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ QUERY PARAMETERS                                                             │
├────┬──────────────┬──────────────────────────────┬──────────────────────────┤
│ ☑  │ Key          │ Value                        │ Description              │
├────┼──────────────┼──────────────────────────────┼──────────────────────────┤
│ ☑  │ page         │ 1                            │ Current page             │
│ ☑  │ limit        │ 20                           │ Items per page            │
│ ☑  │ search       │ {{search}}                   │ Search term              │
└────┴──────────────┴──────────────────────────────┴──────────────────────────┘

+ Add parameter
```

Supports:

* query parameters
* path parameters
* parameter enable/disable
* descriptions
* variable interpolation
* generated values
* encoding preview

---

# 11. Headers Panel

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ HEADERS                                                                      │
├────┬────────────────┬───────────────────────────────────────────────────────┤
│ ☑  │ Content-Type   │ application/json                                      │
│ ☑  │ Accept         │ application/json                                      │
│ ☑  │ Authorization  │ Bearer {{access_token}}                              │
│ ☐  │ X-Debug        │ true                                                  │
└────┴────────────────┴───────────────────────────────────────────────────────┘

+ Add header
Import headers
```

Header sources:

```text
Request
Collection
Folder
Environment
Auth
Generated
```

---

# 12. Body Panel

Supported formats:

```text
None
JSON
Raw
Text
XML
HTML
Form URL Encoded
Multipart Form
Binary
GraphQL
```

JSON editor:

```text
1 │ {
2 │   "name": "John Doe",
3 │   "email": "john@example.com",
4 │   "role": "user"
5 │ }
```

Editor capabilities:

* syntax highlighting
* autocomplete
* formatting
* schema validation
* folding
* line numbers
* error markers
* search
* replace
* generated examples

---

# 13. Auth Panel

Authentication options:

```text
Inherit
No Auth
Basic Auth
Bearer Token
API Key
OAuth 2.0
Digest Auth
AWS Signature
Custom
```

Example:

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ AUTHENTICATION                                                               │
├──────────────────────────────────────────────────────────────────────────────┤
│ Type: Bearer Token ▾                                                         │
│                                                                              │
│ Token                                                                        │
│ {{access_token}}                                                             │
│                                                                              │
│ [ Generate ] [ Select Saved Token ]                                          │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

# 14. Pre-request Script Panel

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ PRE-REQUEST SCRIPT                                                           │
├──────────────────────────────────────────────────────────────────────────────┤
│ 1 │ const token = await reqly.auth.getToken();                              │
│ 2 │ reqly.variables.set("timestamp", Date.now());                           │
│ 3 │                                                                          │
└──────────────────────────────────────────────────────────────────────────────┘
```

Capabilities:

* JavaScript/TypeScript-style scripting
* autocomplete
* variables
* request mutation
* auth helpers
* console output
* execution errors
* script files

---

# 15. Tests Panel

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ TESTS                                                                         │
├──────────────────────────────────────────────────────────────────────────────┤
│ 1 │ expect(response.status).toBe(200);                                     │
│ 2 │ expect(response.body.data).toHaveLength(20);                            │
│ 3 │                                                                          │
└──────────────────────────────────────────────────────────────────────────────┘

[ Run Tests ]
```

Results:

```text
● status is 200
● response has users
● content type is JSON
○ pagination metadata
```

---

# 16. Request Docs Panel

The request-level docs panel shows:

* endpoint description
* generated API documentation
* OpenAPI metadata
* parameters
* schemas
* examples
* authentication requirements

---

# 17. Request Settings Panel

Request-specific settings:

```text
Timeout
Redirect handling
SSL verification
Proxy
Retry
Cookies
Compression
HTTP version
Streaming
```

---

# 18. Request Metadata Panel

A request metadata inspector should expose:

```text
Request
├── Collection
├── Folder
├── Created
├── Updated
├── Owner
├── Environment
├── Tags
├── Description
└── Source
```

Actions:

* rename
* move
* duplicate
* tag
* pin
* archive
* delete

---

# 19. Response Viewer

The response viewer is part of the Requests page.

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ RESPONSE                                                                     │
├──────────────────────────────────────────────────────────────────────────────┤
│ ● 200 OK    142 ms    2.4 KB    HTTP/2                        Copy  Save  ⋮ │
├──────────────────────────────────────────────────────────────────────────────┤
│ Body │ Headers │ Cookies │ Test Results │ Timeline                           │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

# 20. Response Body Panel

Modes:

```text
Pretty
Raw
Tree
Preview
```

Format modes:

```text
JSON
XML
HTML
Text
Binary
Image
```

JSON tree:

```text
▼ data
  ├── 0
  │   ├── id
  │   ├── name
  │   ├── email
  │   └── role
  └── 1
      └── ...
```

Actions:

* copy
* download
* expand all
* collapse all
* search
* generate schema
* generate tests
* send to diff
* replay

---

# 21. Response Headers Panel

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ RESPONSE HEADERS                                                             │
├──────────────────────────────────────────────────────────────────────────────┤
│ content-type      application/json                                           │
│ cache-control     no-cache                                                   │
│ content-length   2401                                                        │
│ server            nginx                                                      │
│ x-request-id      abc-123                                                    │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

# 22. Response Cookies Panel

Shows:

```text
Name
Value
Domain
Path
Expires
Secure
HttpOnly
SameSite
```

Cookie actions:

* inspect
* copy
* delete
* replay

---

# 23. Test Results Panel

Displays:

```text
● 8 passed
● 1 skipped
● 0 failed
```

Includes:

* assertion details
* execution duration
* stack trace
* console output
* failed value diff

---

# 24. Timeline Panel

```text
DNS       ███
TCP          ████
TLS              ██████
Upload                  ██
Server                    █████████████
Download                                ███
──────────────────────────────────────────────►
0ms                                        142ms
```

Breakdown:

```text
DNS
Connection
TLS
Upload
Server processing
Download
```

---

# 25. Response Actions

```text
Copy
Save
Download
Replay
Diff
Generate Schema
Generate Test
Generate Documentation
Open in Explorer
```

---

# 26. Collections Explorer

The Requests context sidebar contains the collection tree.

```text
┌──────────────────────────────┐
│ COLLECTIONS                  │
├──────────────────────────────┤
│ 🔍 Search                   │
│                              │
│ ▼ Reqly API                 │
│   ▼ Authentication          │
│     POST Login              │
│     POST Refresh            │
│                              │
│   ▼ Users                   │
│     GET List Users          │
│     GET Get User            │
│     POST Create User        │
│                              │
│   ▼ Products                │
│     GET Products            │
│                              │
│ + Collection                │
│ + Folder                    │
│ + Request                   │
└──────────────────────────────┘
```

Collection actions:

```text
Rename
Move
Duplicate
Delete
Run
Import
Export
Generate Docs
Generate Tests
Generate Mock
```

---

# 27. Collections Page

A collection can also open as a full page.

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ COLLECTION: Users                                      Run Collection ▶      │
├──────────────────────────────────────────────────────────────────────────────┤
│ Description                                                                    │
│ User management API                                                           │
│                                                                              │
│ Requests                                                                       │
│ GET     /users                                                                │
│ GET     /users/:id                                                            │
│ POST    /users                                                                │
│ PATCH   /users/:id                                                            │
│ DELETE  /users/:id                                                            │
│                                                                              │
│ Tests: 18     Requests: 5     Docs: Generated                               │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

# 28. Environments Page

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ ENVIRONMENTS                                           + New Environment      │
├──────────────────────────────────────────────────────────────────────────────┤
│ Local   Development ●   Staging   Production                                  │
├──────────────────────────────────────────────────────────────────────────────┤
│ Variables                                                                     │
│                                                                              │
│ Name             Value                         Secret       Description       │
│ base_url         https://api.dev.com                       API URL            │
│ api_version      v1                                                    │
│ access_token     •••••••••••••                ✓            Auth              │
│ user_id          42                                             Test user    │
│                                                                              │
│ + Add variable                                                               │
│                                                                              │
│ [ Save ] [ Validate ] [ Diff ] [ Cross-check ]                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

# 29. Environment Diff Panel

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ ENVIRONMENT DIFF                                                             │
├──────────────────────────────────────────────────────────────────────────────┤
│ Development ▾                          Production ▾                           │
├──────────────────────────────────────────────────────────────────────────────┤
│ Variable         Development              Production            Result      │
│ base_url         api.dev.example.com      api.example.com       ≠           │
│ api_version      v1                       v1                     =           │
│ timeout          5000                     10000                  ≠           │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

# 30. Environment Validate Panel

Validation categories:

```text
Required variables
Malformed URLs
Unresolved variables
Duplicate variables
Unused variables
Invalid values
Circular references
Missing secrets
```

Results:

```text
● 24 checks passed
● 2 warnings
● 1 error
```

---

# 31. Environment Cross-check Panel

Cross-checks variables against:

```text
Requests
Collections
Scripts
Tests
OpenAPI
Mocks
Documentation
```

Example:

```text
Variable           Used By                     Result
base_url           42 requests                ✓
access_token       18 requests                ✓
legacy_host        0 requests                 ⚠ unused
missing_token      POST /orders               ✕ missing
```

---

# 32. History Page

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ HISTORY                                                Search   Clear         │
├──────────────────────────────────────────────────────────────────────────────┤
│ All ▾   Method ▾   Status ▾   Environment ▾   Date ▾                         │
├──────────────────────────────────────────────────────────────────────────────┤
│ TODAY                                                                        │
│                                                                              │
│ 10:42 GET     /v1/users              200   142ms   Development               │
│ 10:39 POST    /v1/users              201   183ms   Development               │
│ 10:32 GET     /v1/products           200    91ms   Development               │
│ 10:27 DELETE  /v1/users/42           204   112ms   Development               │
└──────────────────────────────────────────────────────────────────────────────┘
```

History detail panel:

```text
Request
Response
Headers
Body
Timing
Environment
Variables
```

Actions:

```text
Reopen
Duplicate
Save to Collection
Replay
Diff
Export
Delete
```

---

# 33. Mocks Page

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ MOCK SERVERS                                          + New Mock Server       │
├──────────────────────────────────────────────────────────────────────────────┤
│ Demo API                                                                     │
│ ● Running    http://localhost:4010                                          │
├──────────────────────────────────────────────────────────────────────────────┤
│ Routes                                                                        │
│ GET    /users            200      users.json                                │
│ POST   /users            201      create-user.json                          │
│ GET    /users/:id        200      user.json                                 │
│ DELETE /users/:id        204      empty                                     │
└──────────────────────────────────────────────────────────────────────────────┘
```

Mock route editor:

```text
Method
Path
Matching rules
Status
Headers
Body
Latency
Scenario
```

---

# 34. Diff Page

The Diff tool compares:

```text
Requests
Responses
Headers
JSON
Text
Environments
Saved versions
```

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ DIFF                                                                         │
├──────────────────────────────┬───────────────────────────────────────────────┤
│ A                            │ B                                             │
├──────────────────────────────┼───────────────────────────────────────────────┤
│ GET /users?page=1            │ GET /users?page=2                             │
│                              │                                               │
│ "total": 42                  │ "total": 44                                  │
│ "page": 1                    │ "page": 2                                    │
└──────────────────────────────┴───────────────────────────────────────────────┘
```

Modes:

```text
Side by Side
Unified
Structural JSON
Headers
```

---

# 35. JWT Inspector Page

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ JWT INSPECTOR                                                                │
├──────────────────────────────────────────────────────────────────────────────┤
│ Paste JWT                                                                     │
│ ┌──────────────────────────────────────────────────────────────────────────┐ │
│ │ eyJhbGciOiJIUzI1NiIs...                                                  │ │
│ └──────────────────────────────────────────────────────────────────────────┘ │
│ [ Decode ] [ Clear ]                                                        │
├──────────────────────────────────────────────────────────────────────────────┤
│ HEADER                                                                       │
│ { "alg": "HS256", "typ": "JWT" }                                             │
│                                                                              │
│ PAYLOAD                                                                      │
│ { "sub": "123", "name": "John", "iat": 123, "exp": 456 }                     │
│                                                                              │
│ SIGNATURE                                                                    │
│ Present                                                                      │
│                                                                              │
│ TOKEN STATUS                                                                 │
│ ● Valid                                                                      │
│ Expires in 42 minutes                                                        │
└──────────────────────────────────────────────────────────────────────────────┘
```

Inspect:

```text
Header
Payload
Signature
Claims
Expiration
Issued At
Not Before
Issuer
Audience
Subject
Algorithm
```

---

# 36. GraphQL Browser Page

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ GRAPHQL                                                                       │
├──────────────────────────────────────────────────────────────────────────────┤
│ Endpoint: https://api.example.com/graphql                         Run ▶     │
├───────────────────────┬──────────────────────────────────────────────────────┤
│ SCHEMA                │ QUERY                                                │
│                       │                                                      │
│ Query                 │ query Users {                                        │
│ ├── users             │   users {                                            │
│ ├── user              │     id                                               │
│ └── search            │     name                                             │
│                       │     email                                            │
│ Mutation              │   }                                                  │
│ ├── createUser        │ }                                                    │
│ └── deleteUser        │                                                      │
├───────────────────────┴──────────────────────────────────────────────────────┤
│ RESPONSE                                                                     │
└──────────────────────────────────────────────────────────────────────────────┘
```

Panels:

```text
Schema
Query
Mutation
Variables
Headers
Fragments
Response
Documentation
```

---

# 37. GraphQL Schema Browser

The schema sidebar should allow:

```text
Query
Mutation
Subscription
Types
Enums
Interfaces
Input Types
Scalars
Directives
```

---

# 38. Runners Page

```text
Runners
├── Collection Runner
├── Pagination Runner
├── Bulk Runner
├── Dataset Runner
└── Run Results
```

---

# 39. Pagination Runner

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ PAGINATION RUNNER                                                            │
├──────────────────────────────────────────────────────────────────────────────┤
│ Request: GET /users                                                          │
│ Strategy: Offset ▾                                                           │
│ Parameter: page                                                              │
│ Start: 1                                                                     │
│ Max pages: 100                                                               │
│                                                                              │
│ Stop condition: Empty response                                               │
│                                                                              │
│ [ Run ]                                                                      │
├──────────────────────────────────────────────────────────────────────────────┤
│ Page │ Status │ Duration │ Items                                             │
│ 1    │ 200    │ 124ms   │ 20                                                │
│ 2    │ 200    │ 118ms   │ 20                                                │
│ 3    │ 200    │ 121ms   │ 20                                                │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

# 40. Bulk Runner

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ BULK RUNNER                                                                  │
├──────────────────────────────────────────────────────────────────────────────┤
│ Source: Users Collection                                                      │
│ Dataset: users.csv                                                            │
│ Environment: Development                                                      │
│ Concurrency: 4                                                               │
│ Delay: 100 ms                                                                 │
│ Retries: 2                                                                    │
│                                                                              │
│ [ Run ]                                                                      │
├──────────────────────────────────────────────────────────────────────────────┤
│ Progress: ███████████████████░░░░ 78%                                        │
│                                                                              │
│ Passed  82     Failed  3     Skipped 1                                      │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

# 41. Explorer Page

OpenAPI Explorer:

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ OPENAPI EXPLORER                                        Import Spec           │
├───────────────────┬──────────────────────────────────────────────────────────┤
│ API               │ GET /users                                               │
│                   │                                                          │
│ ▼ Users           │ List users                                               │
│   GET /users      │                                                          │
│   POST /users     │ Parameters                                               │
│   GET /users/{id} │ page     query      integer                             │
│                   │ limit    query      integer                             │
│ ▼ Products        │                                                          │
│ ▼ Orders          │ Responses                                                │
│                   │ 200 UserList                                             │
│                   │ 400 Error                                                │
│                   │                                                          │
│                   │ [ Open in Request Builder ]                             │
└───────────────────┴──────────────────────────────────────────────────────────┘
```

Panels:

```text
API Tree
Endpoint Details
Parameters
Request Schema
Response Schema
Responses
Security
Examples
```

---

# 42. REST Documentation Page

```text
┌──────────────────────┬───────────────────────────────────────────────────────┐
│ CONTENTS             │ GET /users                                            │
├──────────────────────┼───────────────────────────────────────────────────────┤
│ Overview             │ Returns a list of users.                             │
│ Authentication      │                                                       │
│ Users                │ Parameters                                            │
│ Products             │ page     integer                                     │
│ Orders               │ limit    integer                                     │
│ Errors               │                                                       │
│                      │ Example                                               │
│                      │ curl ...                                              │
│                      │                                                       │
│                      │ Response                                              │
└──────────────────────┴───────────────────────────────────────────────────────┘
```

---

# 43. gRPC Page

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ gRPC                                                                         │
├──────────────────────────────────────────────────────────────────────────────┤
│ grpc.example.com:443                                            Connect ●    │
├───────────────────────┬──────────────────────────────────────────────────────┤
│ SERVICES              │ METHOD                                               │
│ UserService           │ GetUser                                              │
│ ├── GetUser           │                                                      │
│ ├── ListUsers         │ Request                                              │
│ └── CreateUser        │ { "id": 42 }                                         │
│                       │                                                      │
│ OrderService          │ Metadata                                             │
│ ├── GetOrder          │ authorization: Bearer ...                           │
│ └── CreateOrder       │                                                      │
│                       │ [ Invoke ]                                           │
├───────────────────────┴──────────────────────────────────────────────────────┤
│ RESPONSE                                                                     │
└──────────────────────────────────────────────────────────────────────────────┘
```

Panels:

```text
Servers
Services
Methods
Request
Metadata
Response
Status
Streaming
Timeline
```

---

# 44. Import Dialog

```text
┌────────────────────────────────────────────────────────────────────┐
│ IMPORT                                                             │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│ ○ OpenAPI / Swagger                                               │
│ ○ Postman Collection                                              │
│ ○ Insomnia                                                        │
│ ○ cURL                                                            │
│ ○ HAR                                                              │
│ ○ Reqly                                                            │
│                                                                    │
│ ┌────────────────────────────────────────────────────────────────┐ │
│ │ Drop file here or Browse                                      │ │
│ └────────────────────────────────────────────────────────────────┘ │
│                                                                    │
│ Destination: My Workspace ▾                                       │
│                                                                    │
│ [ Cancel ]                                      [ Import ]         │
└────────────────────────────────────────────────────────────────────┘
```

Import preview should show:

```text
Collections found
Requests found
Environments found
Variables found
Conflicts
Warnings
```

---

# 45. Export Dialog

```text
Export
├── Collection
├── Workspace
├── OpenAPI
├── cURL
├── HAR
├── Environment
└── Documentation
```

Export options:

```text
Include secrets
Include tests
Include scripts
Include docs
Normalize variables
```

---

# 46. Auth / OAuth Tokens Panel

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ AUTHENTICATION                                                               │
├──────────────────────────────────────────────────────────────────────────────┤
│ SAVED CREDENTIALS                                                            │
│ ● Development OAuth                                                          │
│ ● GitHub Token                                                               │
│ ● API Key                                                                    │
├──────────────────────────────────────────────────────────────────────────────┤
│ OAuth 2.0                                                                    │
│ Authorization URL                                                            │
│ Token URL                                                                    │
│ Client ID                                                                     │
│ Client Secret                                                                │
│ Scopes                                                                       │
│                                                                              │
│ [ Authorize ] [ Refresh ] [ Revoke ]                                        │
├──────────────────────────────────────────────────────────────────────────────┤
│ ● Valid                                                                      │
│ Expires in 52 minutes                                                        │
└──────────────────────────────────────────────────────────────────────────────┘
```

Token management:

```text
Access Tokens
Refresh Tokens
API Keys
OAuth Clients
Saved Credentials
Token Expiry
Token Revocation
```

---

# 47. Test Files Page

```text
┌───────────────────┬──────────────────────────────────────────────────────────┐
│ TEST FILES        │ users.test.ts                                            │
├───────────────────┼──────────────────────────────────────────────────────────┤
│ users.test.ts     │ describe("Users API", () => {                            │
│ auth.test.ts      │   it("returns users", async () => {                      │
│ products.test.ts  │     const response = await reqly.send(...);             │
│                   │     expect(response.status).toBe(200);                  │
│                   │   });                                                    │
└───────────────────┴──────────────────────────────────────────────────────────┘
```

Features:

```text
File browser
Editor
Syntax highlighting
Run test
Run file
Debug
Test output
Test history
Failures
```

---

# 48. WebSocket Page

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ WEBSOCKET                                                                     │
├──────────────────────────────────────────────────────────────────────────────┤
│ ws://localhost:8080/socket                                    Connected ●    │
├───────────────────────────┬──────────────────────────────────────────────────┤
│ CONNECTION                │ MESSAGES                                         │
│ Headers                   │                                                  │
│ Authorization: Bearer...  │ → {"type":"subscribe"}                          │
│                           │ ← {"type":"connected"}                          │
│ Protocols                 │ ← {"type":"message"}                            │
│ json                      │                                                  │
│                           │ ┌──────────────────────────────────────────────┐ │
│                           │ │ {"type":"ping"}                              │ │
│                           │ └──────────────────────────────────────────────┘ │
│                           │ [ Send ]                                         │
└───────────────────────────┴──────────────────────────────────────────────────┘
```

Additional tabs:

```text
Connection
Headers
Messages
Events
Timeline
```

---

# 49. SSE Page

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ SERVER-SENT EVENTS                                                            │
├──────────────────────────────────────────────────────────────────────────────┤
│ https://api.example.com/events                                 Connected ●    │
├──────────────────────────────────────────────────────────────────────────────┤
│ 10:42:31  event: connected                                                   │
│           data: {"client":"123"}                                             │
│                                                                              │
│ 10:42:34  event: update                                                      │
│           data: {"status":"processing"}                                      │
│                                                                              │
│ 10:42:39  event: complete                                                    │
│           data: {"status":"done"}                                            │
│                                                                              │
│ [ Clear ] [ Pause ] [ Save Stream ]                                         │
└──────────────────────────────────────────────────────────────────────────────┘
```

Tabs:

```text
Connection
Headers
Event Stream
Timeline
```

---

# 50. Global Console Panel

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ CONSOLE                                                                       │
├──────────────────────────────────────────────────────────────────────────────┤
│ 10:42:31 INFO Sending GET /v1/users                                         │
│ 10:42:31 INFO DNS lookup                   4 ms                             │
│ 10:42:31 INFO TCP connection               8 ms                             │
│ 10:42:31 INFO TLS handshake               21 ms                             │
│ 10:42:31 INFO Response received           106 ms                            │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

# 51. Global Network Panel

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ NETWORK                                                                       │
├──────────────────────────────────────────────────────────────────────────────┤
│ Time      Method  URL                  Status   Duration                     │
│ 10:42     GET     /users               200      142ms                        │
│ 10:39     POST    /users               201      183ms                        │
│ 10:32     GET     /products            200       91ms                        │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

# 52. Global Tests Panel

Shows test activity across the current workspace:

```text
Passed    84
Failed     3
Skipped    2
```

Filters:

```text
Request
Collection
File
Environment
Date
```

---

# 53. Global Variables Panel

Shows effective variables.

```text
Variable
├── Global
├── Workspace
├── Environment
├── Collection
├── Request
└── Runtime
```

Example:

```text
base_url       https://api.dev.com
access_token   •••••••••
user_id        42
timestamp      1720000000
```

The UI should show which scope produced each value.

---

# 54. Global Cookies Panel

Displays:

```text
Domain
Path
Name
Value
Secure
HttpOnly
SameSite
Expires
```

Supports:

* inspect
* search
* delete
* clear
* export

---

# 55. Settings

Settings should be a full-page utility rather than a miscellaneous modal.

```text
Settings
├── General
├── Appearance
├── Editor
├── Network
├── Proxy
├── TLS
├── Environments
├── Auth
├── Storage
├── Keyboard Shortcuts
├── Notifications
├── Extensions
└── Advanced
```

---

# 56. Phase 2: P1

## 56.1 OpenAPI Spec Editor

```text
┌──────────────────────┬───────────────────────────────────────────────────────┐
│ SPEC TREE            │ openapi.yaml                                          │
├──────────────────────┼───────────────────────────────────────────────────────┤
│ Info                 │ openapi: 3.1.0                                       │
│ Servers              │                                                       │
│ Paths                │ paths:                                                │
│ ├── /users           │   /users:                                             │
│ ├── /products        │     get:                                              │
│ └── /orders          │       summary: List users                            │
│ Components           │                                                       │
│ ├── Schemas          │ components:                                           │
│ └── Security         │   schemas:                                            │
└──────────────────────┴───────────────────────────────────────────────────────┘
```

## 56.2 Schema Visualization

```text
                    ┌─────────────┐
                    │    User     │
                    └──────┬──────┘
                           │
             ┌─────────────┼─────────────┐
             ▼             ▼             ▼
         Address         Order        Profile
                            │
                            ▼
                         Product
```

## 56.3 Request Templates

```text
Templates
├── REST
│   ├── CRUD
│   ├── Pagination
│   ├── Upload
│   └── Authentication
├── GraphQL
├── gRPC
└── Realtime
```

## 56.4 Proxy / TLS Controls

```text
Proxy
├── HTTP Proxy
├── HTTPS Proxy
├── SOCKS
└── Authentication

TLS
├── Verification
├── Custom CA
├── Client Certificates
└── TLS Version
```

## 56.5 Data-driven Testing

Dataset UI:

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ DATASET                                                                       │
├──────────────────────────────────────────────────────────────────────────────┤
│ users.csv                                                                     │
│                                                                              │
│ id    name      role                                                          │
│ 1     John      user                                                          │
│ 2     Jane      admin                                                         │
│ 3     Mike      user                                                          │
│                                                                              │
│ Iterations: 3    Concurrency: 2                                             │
│ [ Run Dataset ]                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

## 56.6 CI/CD Integration

Provides:

```text
CLI command
Pipeline configuration
Environment selection
Secrets
Reports
Exit codes
Test artifacts
```

## 56.7 Full Mock Server GUI

Adds:

```text
Routes
Scenarios
State
Dynamic Data
Latency
Fault Injection
Request Matching
Logs
Server Configuration
```

## 56.8 GraphQL / gRPC Documentation

Generated documentation from:

```text
GraphQL Schema
Protobuf
gRPC Reflection
```

---

# 57. Phase 3: P2

## 57.1 API Monitoring Dashboard

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ API MONITORING                                                               │
├──────────────────────────────────────────────────────────────────────────────┤
│ Requests 48,291 │ Success 99.7% │ Avg 184ms │ Availability 99.99%           │
├──────────────────────────────────────────────────────────────────────────────┤
│ Availability                  Latency                                        │
│ 100% ──────────────           400ms ┤       ╭──╮                             │
│  99% ───────╮──────           200ms ┤──╭────╯  ╰──                           │
│              ╰────              0ms └────────────────                         │
└──────────────────────────────────────────────────────────────────────────────┘
```

## 57.2 Performance Testing

Panels:

```text
Load Configuration
Concurrency
Request Rate
Duration
Latency
Throughput
Errors
Scenarios
Results
Comparison
```

## 57.3 MQTT / Socket.IO

Shared realtime interface:

```text
Connection
Topics / Events
Messages
Headers
Payload
Timeline
```

## 57.4 Dependency Graph

```text
Frontend
   │
   ▼
Users API ─────► Auth API
   │
   ├───────────► Orders API
   │                 │
   │                 ▼
   └───────────► Payments API
```

## 57.5 Request Replay

Replay sources:

```text
History
Network
Monitoring
Timeline
Diff
Saved Request
```

Replay controls:

```text
Original Environment
Current Environment
Modified Headers
Modified Body
Modified Query
```

## 57.6 In-app Developer Tools / Debugger

Panels:

```text
Console
Network
Variables
Scripts
Breakpoints
Runtime Errors
Request Inspector
Response Inspector
```

## 57.7 Git GUI

```text
Repository
├── Status
├── Branches
├── Commits
├── Diff
├── Staging
├── Pull
├── Push
├── Merge
└── Conflict Resolution
```

## 57.8 Network Interception / Timeline Debugging

```text
DNS       ███
TCP          ████
TLS              ██████
Upload                  ██
Server                    █████████████
Download                                ███
──────────────────────────────────────────────►
0ms                                        142ms
```

---

# 58. Phase 4: P3

## 58.1 Plugin Marketplace

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ PLUGIN MARKETPLACE                                                           │
├──────────────────────────────────────────────────────────────────────────────┤
│ Search plugins...                                                            │
│                                                                              │
│ ┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐              │
│ │ GraphQL Tools    │ │ AWS Toolkit      │ │ Database Explorer│              │
│ │ ★ 4.9            │ │ ★ 4.8            │ │ ★ 4.7            │              │
│ │ [ Install ]      │ │ [ Install ]      │ │ [ Install ]      │              │
│ └──────────────────┘ └──────────────────┘ └──────────────────┘              │
└──────────────────────────────────────────────────────────────────────────────┘
```

## 58.2 Theme Marketplace

Controls:

```text
Color Tokens
Typography
Density
Editor Style
Icons
Panel Style
```

## 58.3 Git Provider Integrations

Providers:

```text
GitHub
GitLab
Bitbucket
```

Functions:

```text
Authenticate
Select Repository
Select Branch
Pull
Push
Sync
Commit
Review
```

## 58.4 Team / Shared Workspaces

```text
Workspace
├── Members
├── Roles
├── Collections
├── Environments
├── Shared Secrets
├── Activity
└── Settings
```

## 58.5 Enterprise

```text
Enterprise
├── SSO
├── SCIM
├── Audit Logs
├── Organization Policies
├── Permissions
├── Secret Policies
├── IP Restrictions
└── Compliance
```

---

# 59. Phase 5

## 59.1 MCP Server

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ MCP SERVER                                                                   │
├──────────────────────────────────────────────────────────────────────────────┤
│ Status: ● Running                                                            │
│ Endpoint: localhost:xxxx                                                     │
│                                                                              │
│ TOOLS                                                                        │
│ send_request                                                                 │
│ inspect_response                                                             │
│ list_collections                                                             │
│ get_environment                                                              │
│ run_collection                                                               │
│ inspect_openapi                                                              │
│                                                                              │
│ RESOURCES                                                                    │
│ collections                                                                  │
│ environments                                                                 │
│ documentation                                                                │
└──────────────────────────────────────────────────────────────────────────────┘
```

## 59.2 Command Palette

```text
┌─────────────────────────────────────────────────────────────┐
│ > Search commands...                                        │
├─────────────────────────────────────────────────────────────┤
│ REQUEST                                                     │
│   New Request                                    Ctrl+N     │
│   Send Request                                  Ctrl+Enter  │
│   Duplicate Request                              Ctrl+D     │
│                                                             │
│ NAVIGATION                                                  │
│   Open History                                  Ctrl+H      │
│   Open Environments                             Ctrl+E      │
│   Open Collections                              Ctrl+K      │
│                                                             │
│ TOOLS                                                       │
│   Open JWT Inspector                                        │
│   Open GraphQL                                              │
│   Open gRPC                                                 │
└─────────────────────────────────────────────────────────────┘
```

## 59.3 Optional AI Assistant

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ REQUEST WORKSPACE                                      AI ASSISTANT           │
├─────────────────────────────────────────────────────┬────────────────────────┤
│ GET /users                                          │ CONTEXT                │
│                                                     │ Request                │
│ Params                                              │ Response               │
│ Headers                                             │ Environment            │
│ Body                                                │ OpenAPI                │
│ Response                                            │ Collection             │
│                                                     │ Tests                  │
│                                                     │                        │
│                                                     │ SUGGESTIONS            │
│                                                     │ Explain response       │
│                                                     │ Generate tests         │
│                                                     │ Find issue             │
│                                                     │ Generate schema        │
│                                                     │                        │
│                                                     │ Ask about request…    │
└─────────────────────────────────────────────────────┴────────────────────────┘
```

---

# 60. Complete Navigation Map

```text
REQLY
│
├── WORKSPACE
│   ├── Home / Overview
│   ├── Collections
│   │   ├── Collection Details
│   │   ├── Request Tree
│   │   └── Collection Runner
│   │
│   ├── Requests
│   │   ├── Request Tabs
│   │   ├── URL Bar
│   │   ├── Params
│   │   ├── Headers
│   │   ├── Body
│   │   ├── Auth
│   │   ├── Pre-request
│   │   ├── Tests
│   │   ├── Docs
│   │   ├── Settings
│   │   ├── Metadata
│   │   └── Response
│   │       ├── Body
│   │       ├── Headers
│   │       ├── Cookies
│   │       ├── Test Results
│   │       └── Timeline
│   │
│   ├── Environments
│   │   ├── Variables
│   │   ├── Diff
│   │   ├── Validate
│   │   └── Cross-check
│   │
│   └── History
│       ├── Request History
│       └── Request Detail
│
├── API TOOLS
│   ├── Mocks
│   │   ├── Servers
│   │   ├── Routes
│   │   ├── Route Editor
│   │   └── Logs
│   │
│   ├── Diff
│   │   ├── Request Diff
│   │   ├── Response Diff
│   │   ├── JSON Diff
│   │   ├── Headers Diff
│   │   └── Environment Diff
│   │
│   ├── JWT Inspector
│   │   ├── Header
│   │   ├── Payload
│   │   ├── Claims
│   │   └── Token Status
│   │
│   ├── GraphQL
│   │   ├── Schema
│   │   ├── Query
│   │   ├── Mutation
│   │   ├── Variables
│   │   ├── Headers
│   │   └── Response
│   │
│   ├── gRPC
│   │   ├── Servers
│   │   ├── Services
│   │   ├── Methods
│   │   ├── Request
│   │   ├── Metadata
│   │   ├── Response
│   │   └── Streaming
│   │
│   ├── Runners
│   │   ├── Collection
│   │   ├── Pagination
│   │   ├── Bulk
│   │   ├── Dataset
│   │   └── Results
│   │
│   ├── Explorer
│   │   ├── API Tree
│   │   ├── Endpoint Details
│   │   ├── Parameters
│   │   ├── Schemas
│   │   ├── Responses
│   │   └── Security
│   │
│   └── Docs
│       ├── Overview
│       ├── Authentication
│       ├── Endpoints
│       ├── Schemas
│       ├── Examples
│       └── Errors
│
├── REALTIME
│   ├── WebSocket
│   │   ├── Connection
│   │   ├── Headers
│   │   ├── Messages
│   │   ├── Events
│   │   └── Timeline
│   │
│   └── SSE
│       ├── Connection
│       ├── Headers
│       ├── Event Stream
│       └── Timeline
│
├── DEVELOPMENT
│   ├── Test Files
│   ├── Console
│   ├── Network
│   ├── Variables
│   ├── Cookies
│   └── Test Results
│
├── SETTINGS
│   ├── General
│   ├── Appearance
│   ├── Editor
│   ├── Network
│   ├── Proxy
│   ├── TLS
│   ├── Auth
│   ├── Storage
│   ├── Keyboard Shortcuts
│   ├── Notifications
│   └── Advanced
│
├── PHASE 2
│   ├── OpenAPI Editor
│   ├── Schema Visualization
│   ├── Request Templates
│   ├── Proxy / TLS Controls
│   ├── Data-driven Testing
│   ├── CI/CD
│   ├── Full Mock GUI
│   └── GraphQL / gRPC Docs
│
├── PHASE 3
│   ├── Monitoring
│   ├── Performance Testing
│   ├── MQTT
│   ├── Socket.IO
│   ├── Dependency Graph
│   ├── Request Replay
│   ├── Developer Tools
│   ├── Git GUI
│   └── Network Timeline
│
├── PHASE 4
│   ├── Plugin Marketplace
│   ├── Theme Marketplace
│   ├── Git Providers
│   ├── Team Workspaces
│   └── Enterprise
│
└── PHASE 5
    ├── MCP
    ├── Command Palette
    └── AI Assistant
```

---

# 61. Shared Interaction Patterns

All pages should reuse the same interaction vocabulary.

## Search

Every large collection should expose:

```text
⌕ Search
```

## Primary action

Usually appears top-right:

```text
[ New ]
[ Run ]
[ Send ]
[ Save ]
[ Connect ]
```

## Secondary actions

Use:

```text
⋮
```

for:

```text
Rename
Duplicate
Move
Export
Delete
Archive
```

## Status

Use consistent states:

```text
● Connected
● Running
● Valid
● Success
● Warning
● Error
```

## Tabs

Use tabs for different views of the same resource.

Examples:

```text
Body
Headers
Cookies
Tests
Timeline
```

## Panels

Use panels for related configuration.

Examples:

```text
Params
Headers
Auth
Body
Tests
```

---

# 62. Page vs Panel Rules

Reqly should distinguish between pages and panels.

### Full pages

Use for:

```text
Requests
Environments
History
Mocks
Diff
JWT
GraphQL
gRPC
Runners
Explorer
Docs
WebSocket
SSE
Settings
```

### Context panels

Use for:

```text
Collection Tree
Request Metadata
Environment Variables
Schema Tree
Service Tree
Saved Credentials
```

### Request panels

Use for:

```text
Params
Headers
Body
Auth
Pre-request
Tests
Docs
Settings
```

### Response panels

Use for:

```text
Body
Headers
Cookies
Test Results
Timeline
```

### Bottom utility panels

Use for:

```text
Console
Network
Tests
Variables
Cookies
```

### Dialogs

Use for:

```text
Import
Export
Create Workspace
Create Collection
Create Environment
Authentication
Connection Setup
Confirmation
```

This keeps the product from turning every small piece of functionality into a separate page, which is one of the faster ways to make an application feel like enterprise software from 2011.

---

# 63. Final Layout Model

The final Reqly application should read like this:

```text
┌──────────────────────────────────────────────────────────────────────────────────────────────┐
│ TOP BAR                                                                                      │
│ Reqly │ Workspace │ Search │ Import │ Export │ Environment │ Status │ Settings               │
├──────┬───────────────────────┬───────────────────────────────────────────────────────────────┤
│      │                       │                                                               │
│ TOOL │ CONTEXT               │ MAIN WORKSPACE                                                │
│ RAIL │ SIDEBAR               │                                                               │
│      │                       │ Tabs / Editor / Inspector / Results                           │
│ ⚡   │ Collections            │                                                               │
│ ◈    │ Requests               │                                                               │
│ ≋    │ History                │                                                               │
│ ◫    │ Mocks                  │                                                               │
│ ⇄    │ Diff                   │                                                               │
│ ♢    │ JWT                    │                                                               │
│ ◎    │ GraphQL                │                                                               │
│ ⌁    │ gRPC                   │                                                               │
│ ▶    │ Runners                │                                                               │
│ ◇    │ Explorer               │                                                               │
│ ▤    │ Docs                   │                                                               │
│ ◌    │ WebSocket              │                                                               │
│ ◌    │ SSE                    │                                                               │
│      │                       │                                                               │
├──────┴───────────────────────┴───────────────────────────────────────────────────────────────┤
│ Console │ Network │ Tests │ Variables │ Cookies                                  Ready ●    │
└──────────────────────────────────────────────────────────────────────────────────────────────┘
```

The key architectural distinction is:

```text
TOP BAR
    = Workspace / global actions

TOOL RAIL
    = switch major tool

CONTEXT SIDEBAR
    = switch resource inside that tool

MAIN WORKSPACE
    = perform the actual API task

PAGE PANELS
    = configure or inspect that task

BOTTOM PANEL
    = observe execution and debugging state
```

This consolidated structure preserves the missing surfaces from the earlier specs while keeping the two-sidebar model intact.


# Source traceability

## Primary source files used
- `ROADMAP(2).md` — newest development-roadmap snapshot used as the canonical base.
- `ROADMAP(3).md` — older development roadmap used to preserve detailed milestone/ticket history and implementation notes that would otherwise be lost.
- `gui-roadmap.md` — desktop GUI execution tracker used for UI delivery state.
- `Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification.md` — full lower-precedence UI reference preserved in the appendix.

## Conflict policy
- Newer development-roadmap state beats older development-roadmap state.
- Development-roadmap scope beats GUI-roadmap scope.
- GUI roadmap delivery state can clarify desktop status but cannot delete product-roadmap work.
- UI architecture is reference material and cannot promote itself into product scope.
- A feature may legitimately have mixed status, such as core `[x]`, desktop `[~]`, and UI polish `[ ]`.
- Historical entries remain for traceability even when current status has moved on.

## Completeness rule
No source detail should be deleted merely because it is duplicated. Duplicate statements are consolidated into the current roadmap, while unique ticket-level or UI-reference detail is retained in the historical or lower-precedence sections.
