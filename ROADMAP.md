# Reqly — Development Roadmap

> **Status:** Scaffold / Work in progress
> **Overall completion:** ~4% (Foundation + core primitives only)
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
- [x] `AppService` binding registered + `Greet` bridge proof
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
- [~] `internal/request` — data model only (no engine, no transport, no tests)

### 0.5 CLI skeleton
- [x] Cobra command tree: `run`, `test`, `collection run`, `mock`, `validate`, `diff`, `docs`
- [ ] All 7 CLI commands wired to the Go core **(stub — all print "not implemented yet")**

---

## Phase 1 — Core API Client (P0)

The minimum set to make Reqly a serious API client.

### 1.1 Request engine (foundation for everything)
- [ ] `internal/request` — full HTTP request model (URL, method, path/query params, headers, body, auth, certs, proxy, settings)
- [ ] Request engine: HTTP/1.1 transport, timeouts, redirects, compression
- [ ] Request execution shared by Desktop + CLI (single engine, no duplication)
- [ ] Response model: status, headers, cookies, timing, size, raw body
- [ ] Response body parsing (JSON, XML, HTML, text, CSV, binary)
- [ ] File upload / multipart / file download
- [ ] Request history (SQLite) + replay

### 1.2 Variables & environments
- [ ] All 6 variable scopes (global, environment, collection, folder, request, runtime)
- [ ] `{{key}}` interpolation wired through request builder + scripting
- [ ] Environment management (create/edit/switch) — UI + core
- [ ] Environment validation (missing/invalid/unused/secrets)
- [ ] Dynamic values & template tags (UUID, timestamp, random, runtime)

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
- [ ] `internal/collections` — workspaces, collections, nested folders
- [ ] Plain-text, Git-native project files (mirror workspace → filesystem)
- [ ] `internal/core` — application services layer shared by Desktop/CLI/MCP
- [ ] Inheritance: Workspace → Collection → Folder → Request (base URL, headers, auth, vars)

### 1.6 Request builder & response viewer (UI)
- [ ] Method select, URL bar, params/headers/body tabs
- [ ] Body editors: JSON, XML, form-data, URL-encoded, raw, binary, GraphQL
- [ ] Send request → real response data flow (currently `console.log` only)
- [ ] Response viewer: metadata, raw/pretty/tree/table views, search
- [ ] Response actions: copy, download, format, save-as-example
- [ ] Cookies: view/edit/delete, persistence, domain/path matching

### 1.7 Scripting & automation
- [ ] Pre-request / post-request scripts (Goja)
- [ ] Test scripts + assertion library (status, headers, body, JSON/XML values, response time, schema)
- [ ] Request chaining (login → extract token → next request)
- [ ] Collection runner (sequential/parallel, variable passing, failure handling)

### 1.8 Protocols (P0: REST-first, then extended)
- [ ] **REST** — complete builder (see §1.1/§1.6)
- [ ] **WebSocket** — connection mgmt, message composer, in/out inspection
- [ ] **SSE** — live event stream, inspection, event history
- [ ] **GraphQL** — query editor, variables, introspection, autocomplete, schema browser
- [ ] **gRPC** — proto files, reflection, service/method discovery, unary + streaming
- [ ] **SOAP** — WSDL import, operation discovery, XML builder

### 1.9 Import / export
- [ ] Import: Postman, Insomnia, OpenAPI, Swagger, cURL, HAR
- [ ] Export: collections, requests, OpenAPI, responses, test results, docs
- [ ] Import preservation (env/auth/scripts) + unsupported-feature reporting

### 1.10 OpenAPI & JSON Schema
- [ ] OpenAPI 2.x/3.x/3.1 parse + validate
- [ ] Endpoint explorer + generate requests from spec
- [ ] JSON Schema: edit, validate, inspect, generate
- [ ] Generate mocks + docs from OpenAPI (see P1)

### 1.11 CLI (P0 commands)
- [ ] `reqly run` — send a request from the CLI
- [ ] `reqly test` — run tests against a request
- [ ] `reqly collection run` — run a collection
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
| Phase 1 | Core API Client (P0) | Request engine, storage, auth, env, UI wiring, CLI — all **not started** | ~2% |
| Phase 2 | Differentiating (P1) | Not started | 0% |
| Phase 3 | Power-User (P2) | Not started | 0% |
| Phase 4 | Ecosystem (P3) | Not started | 0% |
| Phase 5 | MCP / AI / Extensibility | Not started | 0% |
| Quality | DoD + release gates | Fast checks green; Vitest/E2E/integration TBD | ~20% |

### Next milestones (suggested order)
1. **Request engine end-to-end** — `internal/request` engine + transport + response parsing, with tests (the foundation everything else builds on)
2. **CLI `run`/`test`** — first real end-to-end feature: request file → execute → assert, via CLI
3. **Core → Desktop bridge** — replace `Greet` with real service bindings; wire `RequestEditor` Send → core
4. **Workspaces & collections on disk** — Git-native storage + inheritance
5. **Import/export** — cURL/OpenAPI import + collection export
6. **WebSocket + SSE** — first realtime protocols
7. **Collection runner + scripting** — pre/post scripts + tests in the runner
8. **Mock server + OpenAPI** — generate mocks from specs