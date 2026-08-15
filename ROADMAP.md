# Reqly — Development Roadmap

> **Status:** Scaffold / Work in progress
> **Overall completion:** ~9% (Foundation + request engine + CLI `run`/`test` + request files)
> **Source of truth:** [`Docs/FeatureSet.md`](Docs/FeatureSet.md) (features), [`Docs/API Client — Technology Stack.md`](Docs/API%20Client%20—%20Technology%20Stack.md) (stack), [`Docs/API Client — Testing Strategy & TDD.md`](Docs/API%20Client%20—%20Testing%20Strategy%20&%20TDD.md) (quality)
>
> Checkboxes below track real, working code — not scaffolding. A box is only ticked when the feature ships end-to-end (core logic **and** UI/CLI wiring **and** tests) per the Definition of Done in the Testing Strategy doc.

---

## Legend

- `[x]` — shipped & tested (core + UI + tests)
- `[~]` — partial (some layers exist, not complete end-to-end)
- `[ ]` — not started
- **(stub)** — scaffold/file exists but no logic

---

## Phase 0 — Foundation (✅ ~95% complete)

The project skeleton, build system, and the first two core primitives.

### 0.1 Repository & build infra
- [x] Go module `github.com/Its-Satyajit/reqly` (Go 1.25)
- [x] npm workspaces + nub package manager (`nub.lock` committed)
- [x] Wails v3 desktop project (`apps/desktop/backend`) with Taskfile + build assets
- [x] CI workflow (frontend typecheck/build job; Go vet/gofmt/race/coverage job)
- [x] Makefile task aliases
- [x] GPL-3.0 license + SPDX headers on all Go sources

### 0.2 Desktop shell (Wails v3)
- [x] `main.go` — Wails v3 `application.New`, window (1280×800), dark background
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

### 0.4 Core primitives (first implementations, TDD)
- [~] `internal/variables` — scope-precedence resolution + interpolation (7 tests) — **missing** 5 of 6 scopes, env files, `.env`, validation, diff
- [~] `internal/scripting` — lazy Goja runtime (4 tests) — **missing** request/response API, pre/post-request wiring, dynamic values
- [~] `internal/request` + `internal/response` — request engine + response model (see §1.1)
- [~] `internal/testing` — assertion engine + JSONPath + suite runner + test-file loader (see §1.11)

### 0.5 CLI skeleton
- [x] Cobra command tree: `run`, `test`, `collection run`, `mock`, `validate`, `diff`, `docs`
- [~] 9 CLI commands wired to the Go core — `run`, `test`, `collection run`, `collection list`, `import curl`, `import openapi`, `export postman` done; `mock`, `validate`, `diff`, `docs` still stubs

---

## Phase 1 — Core API Client (P0)

The minimum set to make Reqly a serious API client.

### 1.1 Request engine (foundation for everything)
- [x] `internal/request` — full HTTP request model (URL, method, path/query params, headers, body, auth, certs, proxy, settings)
- [x] Request engine: HTTP/1.1 transport, timeouts, redirects, compression
- [x] Request execution shared by Desktop + CLI (single engine, no duplication)
- [x] Response model: status, headers, cookies, timing, size, raw body
- [~] Response body parsing (JSON done via `JSON()`/`JSONValue`; XML, HTML, text, CSV, binary pending)
- [ ] File upload / multipart / file download
- [ ] SQLite local metadata (history, search index, execution results, cache) + request replay

### 1.2 Variables & environments
- [~] All 8 variable scopes (global, environment, collection, folder, request, runtime, prompt, process env) — core has 6 scopes; request files carry `variables` maps
- [~] `{{key}}` interpolation wired through request builder + scripting — works in `run`/`test` via request files
- [ ] Environment management (create/edit/switch) — UI + core
- [ ] Environment validation (missing/invalid/unused/secrets)
- [ ] Dynamic values & template tags (UUID, timestamp, random, runtime)

### 1.2a Request files (plain-text, Git-native) 
- [x] `internal/requestfile` — JSON/YAML request file format (`name`, `variables`, `request`)
- [x] `reqly run <file>` — load request + variables from file, flags override file fields
- [x] `reqly test <file>` — test files accept YAML and `variables` (interpolated at runtime)
- [x] Shared file format for collections/folders (`internal/collections` descriptor format, see §1.5)

### 1.3 Authentication
- [ ] Basic, Bearer, API key
- [ ] JWT (encode/decode/claims viewer)
- [ ] Digest, NTLM
- [ ] OAuth 1.0 / OAuth 2.0 (Auth Code, Client Credentials, Password) with token storage/refresh/expiry
- [ ] AWS Signature, Akamai EdgeGrid, custom auth + auth inheritance

### 1.4 Secrets
- [ ] Encrypted-at-rest secret storage + OS keychain
- [ ] Secret variables + masking (UI, logs, test output, docs)
- [ ] `.env` support + external managers (Vault, AWS, Azure) — P3

### 1.5 Workspaces, collections & storage
- [x] `internal/collections` — workspaces, collections, nested folders
- [x] Plain-text, Git-native project files (mirror workspace → filesystem)
- [x] `internal/core` — application services layer shared by Desktop/CLI/MCP (`RequestService.Send`)
- [x] Inheritance: Workspace → Collection → Folder → Request (base URL, headers, auth, vars)
- [x] `reqly collection run <path>` + `reqly collection list` (CLI wired to the Go core)
- [ ] Environments: resolve the `environment` scope from `environments/` on disk
- [ ] Save/export a workspace (write descriptors + request files back to disk)

### 1.5a Core → Desktop bridge (from 0.2 `Greet` proof)
- [x] `internal/core` `RequestService` — wraps `request.Client`, bridge-friendly `SendResponse` DTO
- [x] Desktop `AppService.SendRequest` delegates to core (thin Wails boundary; `Greet` removed)
- [x] Regenerated Wails bindings → `appservice.ts` `SendRequest` + `models.ts` (`Request`, `SendResponse`)
- [x] Shared `useRequestStore` + pluggable `RequestSender` (Wails bridge in host; `fetchSender` fallback in browser dev)
- [x] `RequestEditor` Send → core; `ResponseViewer` renders status/headers/pretty body
- [ ] Per-tab request/response state (multiple tabs), cancel in-flight request

### 1.6 Request builder & response viewer (UI)
- [x] Method select, URL bar, Send → real response data flow
- [ ] Params/headers/body tabs in the builder
- [ ] Body editors: JSON, XML, form-data, URL-encoded, raw, binary, GraphQL
- [ ] Response viewer: metadata, raw/pretty/tree/table views, search
- [ ] JSONPath / XPath response querying
- [ ] Response actions: copy, download, format, save-as-example
- [ ] Cookies: view/edit/delete, persistence, domain/path matching

### 1.7 Scripting & automation
- [ ] Pre-request / post-request scripts (Goja)
- [~] Test scripts + assertion library (core assertion engine shipped: status, header, body, JSON, response-time)
- [ ] Request chaining (login → extract token → next request)
- [ ] Chain runner (sequential/conditional execution, variable passing, assertions, failure handling)
- [ ] Collection runner (sequential/parallel, variable passing, failure handling)

### 1.8 Protocols (P0: REST-first, then extended)
- [ ] **REST** — complete builder (see §1.1/§1.6)
- [ ] **WebSocket** — connection mgmt, message composer, in/out inspection
- [ ] **SSE** — live event stream, inspection, event history
- [ ] **GraphQL** — query editor, variables, introspection, autocomplete, schema browser
- [ ] **gRPC** — proto files, reflection, service/method discovery, unary + streaming
- [ ] **SOAP** — WSDL import, operation discovery, XML builder

### 1.9 Import / export
- [x] Import cURL — `reqly import curl` (method, headers, JSON/raw/data bodies, basic auth, user-agent, cookies, GET-style query data; unsupported features reported)
- [x] Import OpenAPI 3.x — `reqly import openapi` (servers, paths, operations, params, JSON bodies; writes a Git-native workspace)
- [x] Export Postman collection v2.1 — `reqly export postman` (flat list, inherited base URL/headers applied)
- [ ] Import: Postman, Insomnia, Swagger, HAR
- [ ] Export: requests, OpenAPI, responses, test results, docs
- [ ] Import preservation (env/auth/scripts) + unsupported-feature reporting

### 1.10 OpenAPI & JSON Schema
- [ ] OpenAPI 2.x/3.x/3.1 parse + validate
- [ ] Endpoint explorer + generate requests from spec
- [ ] JSON Schema: edit, validate, inspect, generate
- [ ] XML/XSD schema validation where applicable
- [ ] Generate mocks + docs from OpenAPI (see P1)

### 1.11 CLI (P0 commands)
- [x] `reqly run` — send a request from the CLI (URL or JSON/YAML request file, flags override file)
- [x] `reqly test` — run tests against a request (JSON or YAML test file, variables interpolated)
- [x] `reqly collection run` — run a request in a collection with inherited config (workspace/collection/folder resolution); `collection list` shows the tree
- [ ] `reqly validate` — validate a project/spec
- [ ] `reqly diff` — diff specs/requests
- [ ] `reqly mock` — serve a mock server
- [ ] `reqly docs` — generate documentation

### 1.12 Cross-platform desktop
- [ ] Linux build (WebKit) — verified
- [ ] macOS build (WebKit) — needs CI/signing
- [ ] Windows build (WebView2) — needs CI/signing

---

## Phase 2 — Differentiating Features (P1)

Features that make Reqly more capable than a basic API client.

- [ ] OpenAPI editor (spec authoring in-app)
- [ ] Schema validation + contract testing
- [ ] Schema visualization (relationships between types, objects, endpoints, schemas)
- [ ] Code generation (request → cURL, JS, Python, Go snippets)
- [ ] Mock server (from OpenAPI/examples, request matching, dynamic responses, delay/error simulation, stateful mocks)
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
- [ ] Retry & resilience (count, delay, backoff, status/network error, 429)
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

- [ ] Requirement defined in FeatureSet
- [ ] TDD cycle (red → green → refactor) with unit tests
- [ ] Edge cases + error behavior covered
- [ ] Integration tests (core ↔ persistence ↔ engine)
- [ ] E2E tests (Playwright) for critical workflows
- [ ] Security review (no secrets exposed, safe crypto)
- [ ] Performance considered
- [ ] Regression tests
- [ ] Coverage within targets
- [ ] Docs updated
- [ ] CI green (vet, gofmt, typecheck, unit, race, build)
- [ ] Frontend unit tests (Vitest) — **currently TBD**

### Release gates
- **Fast checks** (every change): formatting, linting, typechecking, unit tests — ✅ running in CI
- **PR CI:** unit + integration + race + frontend + build + coverage validation — ⏳ partial (no integration/E2E yet)
- **Release CI:** full E2E, performance, security, cross-platform builds, import/export compat, install/upgrade — ❌ not set up

---

## Progress Tracker

| Phase | Scope | Status | Est. complete |
| --- | --- | --- | --- |
| Phase 0 | Foundation | Foundation done; CLI commands + core primitives partial | ~95% |
| Phase 1 | Core API Client (P0) | Request engine + response model + CLI `run`/`test` + request files + core services + desktop bridge + workspace/collection storage with inheritance + cURL/OpenAPI import + Postman export shipped; auth, env, full UI pending | ~22% |
| Phase 2 | Differentiating (P1) | Not started | 0% |
| Phase 3 | Power-User (P2) | Not started | 0% |
| Phase 4 | Ecosystem (P3) | Not started | 0% |
| Phase 5 | MCP / AI / Extensibility | Not started | 0% |
| Quality | DoD + release gates | Fast checks green; Vitest/E2E/integration TBD | ~20% |

### Next milestones (suggested order)
1. ~~**Request file loading**~~ — `reqly run`/`test` from a plain-text request file (JSON/YAML request format + vars) — ✅ shipped
2. ~~**Core → Desktop bridge**~~ — `AppService.SendRequest` → core `RequestService`; `RequestEditor` Send wired to `useRequestStore` — ✅ shipped
3. ~~**Workspaces & collections on disk**~~ — Git-native storage + inheritance (build on `requestfile`) — ✅ shipped
4. ~~**Import/export**~~ — cURL/OpenAPI import + Postman collection export — ✅ shipped
5. **WebSocket + SSE** — first realtime protocols
6. **Collection runner + scripting** — pre/post scripts + tests in the runner
7. **Mock server + OpenAPI** — generate mocks from specs

> **Companion:** [**reqly-test-api**](https://reqly-test-api.vercel.app) — a small ElysiaJS mock API (Vercel-hosted, hardcoded data) for exercising `reqly run`/`test`, auth, delay, and error-status flows against a real endpoint. Useful while the in-app mock server (milestone 7) is pending; see the README's "Mock API" section.