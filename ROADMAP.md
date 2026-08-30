# Reqly — Development Roadmap

> **Status:** Canonical product roadmap. All P0–P5 phases shipped (core + CLI + desktop bindings); 13 items remain `[~]` partial — deferred seams (file-download UI, NTLM, AWS/Azure secrets, XPath, gRPC bidi, SOAP rpc/encoded, visual builders) tracked in `Milestones/12-traceability-map.md`.
> **Detailed specs:** [`Milestones/01`](Milestones/01-phase-0-foundation.md) Foundation · [`02`](Milestones/02-phase-1-core-api-client.md) P0 Core · [`03`](Milestones/03-phase-2-differentiating-features.md) P1 Differentiating · [`04`](Milestones/04-phase-3-power-user-features.md) P2 Power-User · [`05`](Milestones/05-phase-4-ecosystem-and-enterprise.md) P3 Enterprise · [`06`](Milestones/06-phase-5-mcp-ai-extensibility.md) P4/P5 MCP/AI · [`07`](Milestones/07-historical-milestones-ledger.md) Ledger · [`08`](Milestones/08-gui-roadmap-and-execution.md) GUI · [`09-11`](Milestones/09-ui-architecture-shell-and-requests.md) UI Arch · [`12`](Milestones/12-traceability-map.md) Traceability
> **Source of truth:** [`docs/features.md`](docs/features.md) · [`technology-stack.md`](docs/technology-stack.md) · [`testing-strategy.md`](docs/testing-strategy.md) · [`docs/internal/gui-roadmap.md`](docs/internal/gui-roadmap.md) · [Complete UI Arch](docs/Reqly%20Complete%20UI%20Architecture,%20Pages,%20Panels,%20and%20Navigation%20Specification.md) (subordinate, §1–59)
> **UI Redesign Notice:** `frontend/src/` chrome is being rewritten to the four-zone model (TopBar / ToolRail / ContextSidebar / MainWorkspace + BottomPanel). Data layer `lib/` + `stores/` preserved. See `Milestones/08`.

Checkboxes = real shipped code (core + UI/CLI + tests), not scaffolding — per `docs/testing-strategy.md` DoD.

---

## Governance

**Precedence:** 1) `ROADMAP.md` (this file) — product scope/phase/sequence  2) `Milestones/01-12` — grouped specs (detail)  3) `docs/internal/gui-roadmap.md` — desktop execution (subordinate)  4) Complete UI Arch — reference (subordinate).  
**Rule:** UI doc vs roadmap conflict → roadmap wins. Core shipped / UI pending → record as `[~]`, not missing.  
**Legend:** `[x]` shipped & tested · `[~]` partial (one layer pending, deferral noted) · `[ ]` not started · `(stub)` scaffold only.

---

## Phase 0 — Foundation (100%) — [Milestones/01](Milestones/01-phase-0-foundation.md)

Project skeleton, build system, and first core primitives. All 26 items shipped.

| Area | Status | Notes |
|------|--------|-------|
| 0.1 Repository & build infra | `[x]` | Go 1.25, Wails v3, CI, GoReleaser matrix |
| 0.2 Desktop shell | `[x]` | `AppService` + bindings + `history.db` sqlc |
| 0.3 Shared UI shell | `[x]` | App shell, theming, CodeMirror, shadcn |
| 0.4 Core primitives | `[x]` | `variables`, `scripting` (Goja), `request/response`, `testing`, `history/secrets` |
| 0.5 CLI skeleton | `[x]` | 15 commands (`run`/`test`/`collection`/`env`/`auth`/`history`…) |

---

## Phase 1 — Core API Client (P0) — [Milestones/02](Milestones/02-phase-1-core-api-client.md)

Minimum viable API client. Core shipped; 8 seams `[~]` — details & code refs in Milestones/02.

| § | Area | Status | Deferral |
|---|------|--------|----------|
| 1.1 | Request engine | `[~]` | File-download raw FS UI pending (download via `suggestedFilename` shipped) |
| 1.2 | Variables & environments | `[x]` | 6 scopes + `{{}}` + `{{$tag}}` + `env` CLI |
| 1.2a | Request files (Git-native) | `[x]` | `requestfile` JSON/YAML |
| 1.3 | Authentication | `[~]` | Digest P0 `[x]`, NTLM → P3 (CGO/`gssapi`) |
| 1.4 | Secrets | `[~]` | `.env` `[x]`, Vault `[x]`, AWS/Azure → P3 |
| 1.5 + 1.5a | Workspaces/collections + Core→Desktop bridge | `[x]` | `collections` + `core.RequestService` + tabs/cancel |
| 1.6 | Request builder & response viewer | `[~]` | JSONPath `[x]`, XPath → P3 |
| 1.7 | Scripting & automation | `[~]` | Runner `[x]`, conditional branching → P1 |
| 1.8 | Protocols | `[~]` | REST/WS/SSE/GraphQL `[x]`; gRPC unary/server-stream `[x]`, client-stream/bidi → P3; SOAP core `[x]`, `rpc/encoded` → P3 |
| 1.9 | Import / export | `[x]` | cURL, OpenAPI, Postman, Insomnia/Bruno, Swagger2, HAR, preservation |
| 1.10 | OpenAPI & JSON Schema | `[x]` | 3.x parse/validate, explorer/generate, mocks, XSD |
| 1.11 | CLI (P0 commands) | `[x]` | `run`/`test`/`collection`/`ws`/`sse`/`validate`/`diff`/`env`/`mock`/`docs`… |
| 1.12 | Cross-platform desktop | `[x]` | Linux/macOS/Windows WebKit/WebView2 matrix |
| 1.13 | Desktop shell redesign | `[~]` | Core shell GUI-0→5 `[x]`, UI chrome rewrite `[~]` (see notice) |

---

## Phase 2 — Differentiating Features (P1) — [Milestones/03](Milestones/03-phase-2-differentiating-features.md)

P1 is 100% shipped (14/14). GUI panels formerly deferred are now in `Milestones/03`.

| Area | Status | Owner |
|------|--------|-------|
| §56.1 Spec Editor | `[x]` | CodeMirror tree + YAML editor |
| §56.2 Schema Visualization | `[x]` | Graph + dependency view |
| §56.3 Request Templates | `[x]` | Picker sheet in builder |
| §56.4 Proxy / TLS Controls | `[x]` | mTLS/CA + per-request transport |
| §56.5 Data-driven Testing | `[x]` | CSV/JSON dataset runner |
| §56.6 CI/CD Integration | `[x]` | GitHub Action + CLI generator |
| §56.7 Full Mock Server GUI | `[x]` | Routes/scenarios/fault injection/logs |
| §56.8 GraphQL / gRPC Docs | `[x]` | Schema/service browsers |
| M28 HAR + M29 JWT + M30 Pagination + M31 Bulk + M32 Retry + M33-37 Diff/Contract/Perf | `[x]` | Import/export/replay, HS512, page/cursor, bulk, backoff, perf `P95/P99` |

---

## Phase 3 — Power-User Features (P2) — [Milestones/04](Milestones/04-phase-3-power-user-features.md)

| Area | Status | Notes |
|------|--------|-------|
| §57.1 Monitoring Dashboard | `[x]` | Scheduled checks + latency trends |
| §57.2 Performance Suite | `[x]` | RPS/P95/P99 + CLI `perf` |
| §57.3 Realtime (MQTT/Socket.IO) | `[x]` | MQTT 3.1/5.0 + Socket.IO |
| §57.4 Dependency Graph | `[x]` | Execution chaining |
| §57.5 Request Replay | `[x]` | Timeline + env substitution |
| §57.8 Timeline Debugging | `[x]` | Waterfall + `reqly run --timeline` |
| M60 Changelog/SemVer + M63 Fetch Importer + M64 Stateful Mock | `[x]` | Diff classifier, `import fetch`, state machine |
| M65 Workflow Engine | `[~]` | Core/CLI/desktop `[x]`, visual builder → follow-up |
| M66 Self-Hosted Automation | `[~]` | `Scheduler.Run` `[x]`, cron/Git-ops UI → follow-up |

---

## Phase 4 — Ecosystem & Enterprise (P3) — [Milestones/05](Milestones/05-phase-4-ecosystem-and-enterprise.md)

| Area | Status | Notes |
|------|--------|-------|
| §58.1 Plugin Engine | `[x]` | Goja runtime + CLI `plugin` |
| §58.2 Theme Sharing (M67) | `[~]` | `internal/theme` + CLI/desktop `[x]`, picker UI → follow-up |
| §58.3 Git Providers (M61) | `[x]` | GitHub/GitLab/Bitbucket/Azure + PAT |
| M69 Audit Logs | `[x]` | `.reqly/audit.log` JSONL + CLI/desktop |
| M70 Org Policies | `[x]` | `.reqly/policy.yaml` + enforce |
| M71 RBAC | `[x]` | `.reqly/rbac.yaml` admin/editor/viewer |
| M72 Vault Secrets | `[~]` | Vault KV v2 `[x]`, AWS/Azure → P3 |
| M73 SSO & SCIM | `[~]` | HMAC/issuer/group `[x]`, JWKS RS256 → P3 |
| M74 Shared Workspaces | `[x]` | `.reqly/collab.yaml` |
| M75 Collab Server | `[x]` | `/health`/`/collab`/`/workspace` + `collab serve` |

---

## Phase 5 — MCP, AI & Extensibility — [Milestones/06](Milestones/06-phase-5-mcp-ai-extensibility.md)

| Area | Status | Notes |
|------|--------|-------|
| §59.1 MCP Server | `[x]` | stdio JSON-RPC `list/search/get/run` |
| §59.2 Command Palette | `[x]` | Global search `⌘K` |
| §59.3 AI Assistant | `[x]` | `reqly ai <test\|docs\|diagnose\|explain>` + `reqly.ai` Goja |

---

## Shell Rebuild — Slice Plan (372) — REBUILDING SLICE-BY-SLICE

> `372` (84 stories) is rebuilt incrementally; data layer `lib/`+`stores/` preserved. Each slice ships as a tracer bullet (core+UI+tests) with gates (`nub typecheck`, `oxlint`, `go vet`, `vitest 185`) before merge. This section is the execution tracker; `Milestones/08` remains the GUI roadmap source.

| Slice | Scope | Stories | Key files | Status |
|-------|-------|---------|-----------|--------|
| **01** | **Foundation — tokens + storage + AppShell + useShellStore** | — (infra) | `frontend/src/styles/tokens.css:1`, `shell/storage.ts:1`, `shell/AppShell.tsx:1`, `stores/useShellStore.ts:1`, `index.css:7` import | `[x]` `33a1d1db` — additive, no App.tsx change yet |
| **02** | **AppShell mount** — refactor `App.tsx` to thin wrapper | Shell chrome 1–8 | `frontend/src/app/App.tsx:57` `ResiablePanelGroup` → `<AppShell topBar/toolRail/sidebar/children/bottom/statusbar>` — `App.tsx` 370→312 lines, `AppShell` handles `sidebarLayout`/`bottomLayout`/`⌘B` + `bottomCollapsed` sync | `[x]` `—` — thin wrapper, `AppShell` owns 5-zone resizable chrome |
| **03** | Home + Collections | WS Home 9–12, Collections 22–24 | `HomeView` stat cards/empty, `CollectionTree` (`1882bd34` duplicate already) + filter/drag | `[x]` `1882bd34`+`33a1d1db` — stat cards, quick actions, recent 8, empty onboarding, tree + `SharedContextMenu` duplicate |
| **04** | Request/Response | Builder 13–17, Response 18–21, Body 69–70, Auth 71–72 | `RequestEditor` 4 tabs + overflow (`1882bd34` settings) + `methodTint` + `TagPicker`, `ResponseViewer` `horizontal`/`vertical` `splitOrientation` + Timeline | `[x]` `1882bd34`+`5553cd19` — 4 tabs + More, URL bar, persist, auth 9 schemes, body 10 types, status/timing/size |
| **05** | Tool Pages | Envs 25–27, History 28–30, Mocks 31–36, Diff 37–38, JWT 39–41, GraphQL 42–45, gRPC 46–48, Runners 49–50, OpenAPI 51–52 | `EnvironmentsView`/`HistoryView`/`MocksView`/`DiffView`/`JwtInspector`/`GraphqlBrowser`/`GrpcTab`/`RunnersPanel`/`OpenapiExplorer` | `[x]` shipped via P1 `Milestones/03` — all 9 tools as pages, lazy `ErrorBoundary` |
| **06** | Import/Export + Settings | Import 53–55, Export 56, Settings 57–59 | `ImportDialog`/`ExportDialog` modals, `SettingsView` Appearance/Workspace/Retention/About (`5553cd19` version) + `ProxyTlsPanel`/`CicdPanel` | `[x]` `5553cd19` — modals + palette `import`/`export` + theme direct picks |
| **07** | Global Panels + Realtime | Bottom 60–64, Realtime 65–66 | `BottomPanel` (`Console`/`Network`/`Tests`/`Variables`/`Cookies`, `⌘J`), `RealtimePage`/`RealtimeTab` auto-reconnect | `[x]` `ac341696` — 5 tabs + `⌘J`/`⌘B` + recents, `realtimePages` already `main` |
| **08** | Search & Palette polish | Search 67–68, Palette + Keyboard | `paletteProviders.ts:19` (`navViews` 16 + `import`/`export`/`theme-*` `5553cd19`) + `CommandPalette` grouping/recent + `useKeyboardMap` `⌘K/⌘1-8` | `[~]` `5553cd19` — `⌘K` + `⌘1-8` + hints ok, grouping/recent polish → follow-up |

*Slice 01 additive — existing `App.tsx` layout preserved until Slice 02 mount. Each slice PR must keep `frontend/src/lib`/`stores` pure and `ErrorBoundary` per panel.*

---

## Historical & UI Reference

All ticket-level histories, GUI matrices, and full UI spec are in [`Milestones/`](Milestones/): [`07` Ledger](Milestones/07-historical-milestones-ledger.md) (M01–M40) · [`08` GUI Roadmap](Milestones/08-gui-roadmap-and-execution.md) · [`09` §1–25](Milestones/09-ui-architecture-shell-and-requests.md) · [`10` §26–55](Milestones/10-ui-architecture-tools-and-pages.md) · [`11` §56–63](Milestones/11-ui-architecture-phase-panels-and-navigation.md) · [`12` Traceability](Milestones/12-traceability-map.md) (roadmap → code → CLI → desktop → tests → ADR).

## Quality & Release Gates (DoD)

- [x] Requirement in `features.md` + `CONTEXT.md` · [x] TDD + `go test ./...` · [x] Edge cases · [x] Security & `0600` perms · [x] Fast checks (fmt/lint/typecheck/unit) · [x] CI matrix (Linux/macOS/Windows)

## Code Review Gates (`/code-review` — two-axis)

> Each phase must pass `/code-review` (Standards `oxlint+go vet`+`anti-slop`+Fowler & Spec `milestone vs ROADMAP DoD`) on `git diff main...HEAD` before `[x]`. See `prompts/code-review-milestones.md`.

- [x] Phase 0 Foundation (`01`) · [x] Phase 1 P0 (`02`) · [x] Phase 2 P1 (`03`) · [x] Phase 3 P2 (`04`) · [x] Phase 4 P3 (`05`) · [x] Phase 5 MCP/AI (`06`) · [x] Ledger (`07`) · [x] UI Arch §1–63 (`09+10+11`)
