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

> **Cleanup 2026-09-02:** §2.1–2.5 were granular spec §2 checklists duplicated by GUI-0.1/0.2/0.3 above (TopBar/ToolRail/Sidebar/Workspace/BottomPanel). Archived here as collapsed reference — source is GUI-0.1/0.2/0.3 `[x]` shipped 2026-08-27, not these `[ ]` items.
>
> - §2.1 TopBar: Logo, Workspace Switcher, Global Search ⌘K, Import/Export, Active Environment, Sync, Notifications, Settings, User Menu — **covered by GUI-0.1**
> - §2.2 Tool Rail (48–56px): Workspace (Home/Requests/Environments/History), API Tools (Mocks/Diff/JWT/GraphQL/gRPC/Runners/Explorer/Docs), Realtime (WebSocket/SSE), System (Settings) — **covered by GUI-0.1** (CLI parity adds MQTT/Socket.IO via GUI-7)
> - §2.3 Context Sidebar (220–280px): collapsible/resizable, per-tool tree, search, actions, recent/pinned, `⌘B` — **covered by GUI-0.1/0.3**
> - §2.4 Main Workspace: tab-based routing, full pages vs panels (§62) — **covered by GUI-0.2**
> - §2.5 Bottom Utility Panel: Console/Network/Tests/Variables/Cookies, `⌘J`, resizable — **covered by GUI-0.1** (`BottomPanel.tsx`)

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

## GUI-6 Theme-Adaptive OS Layouts & Headless Primitives (spec §3, M68)

- [x] **G-6.1** Headless & Unstyled UI Primitives — decouple `Button`, `Input`, `Tabs`, `Card`, `Select`, `Dialog` from hardcoded utility styles; drive purely via semantic CSS variables and slot classes — 2026-09-01
- [x] **G-6.2** Theme-Adaptive Shell Chrome — adapt `AppShell`, `TopBar`, `ToolRail`, `WorkspaceSidebar`, and `BottomPanel` layouts to active OS visual DNA (Windows 11 Centered Search + Mica, macOS Tahoe Floating Glass Toolbar + Pill Search, KDE Breeze Solid Desktop Chrome + 1px Dividers, GNOME Adwaita Integrated Headerbar) — 2026-09-01
- [x] **G-6.3** Theme-Adaptive Feature Views — refactor `SettingsView`, `HomeView`, `RequestEditor`, and `RealtimeTab` into OS-authentic layout patterns (GNOME boxed-list preference groups, macOS grouped glass containers, Windows 11 Fluent settings cards, KDE Breeze geometric desktop tables) — 2026-09-01

## GUI-7 CLI Parity — Realtime Expansion (MQTT + Socket.IO) — mirrors `ROADMAP.md:UI-13`

> **Gap:** `apps/cli/cmd/mqtt.go:1` `mqtt pub|sub` + `apps/cli/cmd/socketio.go:1` `socketio connect|emit` have no GUI. `apps/desktop/backend/realtime.go:1` only `ws`/`sse`.

- [ ] **G-7.1 MQTT Pub/Sub page** — `apps/desktop/backend/mqtt.go:1` `MqttPublish/Subscribe/Cancel` via `internal/mqtt` + `frontend/src/features/mqtt-view/MqttView.tsx:1` (broker/topic/QoS/retain + Publish/Subscribe log) + `frontend/src/lib/mqtt.ts:1` — mirrors CLI flags `--topic`/`--message`/`--qos`/`--retain`/`--username`/`--password` — 2026-09-02 gap
- [ ] **G-7.2 Socket.IO page** — `apps/desktop/backend/socketio.go:1` `SocketIOConnect/Emit/Close` via `internal/socketio` + `frontend/src/features/socketio-view/SocketIOView.tsx:1` (url + namespace + event/data) — mirrors `socketio connect <url> [--namespace]` / `emit <url> --event --data` — 2026-09-02 gap
- [ ] **G-7.3 Realtime rail unification** — `frontend/src/components/shell/ToolRail.tsx:1` `REALTIME_GROUP` + `frontend/src/app/App.tsx:1` `activeView` + `frontend/src/stores/useWorkspaceStore.ts:1` `WorkspaceView` `mqtt|socketio` — rail shows 4 entries (WS/SSE/MQTT/Socket.IO) — 2026-09-02 gap

## GUI-8 CLI Parity — Governance & Enterprise — mirrors `ROADMAP.md:UI-14`

> **Gap:** `policy`/`rbac`/`audit`/`sso`/`scim`/`collab` have `apps/desktop/backend/*.go:1` bindings but no frontend view.

- [ ] **G-8.1 Policy & RBAC page** — `frontend/src/features/governance/PolicyRbacView.tsx:1` + `frontend/src/lib/policy.ts:1`/`rbac.ts:1` — `PolicyGet/Save/Enforce` + `RBACList/Check/Get` (0600, Git-native) — 2026-09-02 gap
- [ ] **G-8.2 Audit Log page** — `frontend/src/features/audit-view/AuditView.tsx:1` + `frontend/src/lib/audit.ts:1` — `AuditList/Clear/Export` (append-only, 0600) — 2026-09-02 gap
- [ ] **G-8.3 SSO & SCIM page** — `frontend/src/features/sso-view/SsoScimView.tsx:1` + `frontend/src/lib/sso.ts:1` — `SSOValidate` (OIDC) + `SCIMCreateUser/ListUsers` — 2026-09-02 gap
- [ ] **G-8.4 Collaboration page** — `frontend/src/features/collab-view/CollabView.tsx:1` + `frontend/src/lib/collab.ts:1` — `CollabList/Add/Remove/Serve` (Git-native shared workspaces) — 2026-09-02 gap

## GUI-9 CLI Parity — Automation & Orchestration — mirrors `ROADMAP.md:UI-15`

> **Gap:** `automation run`, `workflow <yaml>`, `changelog <old> <new>`, and `monitor run --interval` (CLI) → GUI `MonitorView.tsx:1` is mock-only.

- [ ] **G-9.1 Automation scheduler** — `frontend/src/features/automation-view/AutomationView.tsx:1` — mounts existing `AutomationRun` (`apps/desktop/backend/automation.go:1`) — 2026-09-02 gap
- [ ] **G-9.2 Workflow runner** — `frontend/src/features/workflow-view/WorkflowView.tsx:1` + `frontend/src/lib/workflow.ts:1` — `WorkflowRun` multi-step via `RunView` — 2026-09-02 gap
- [ ] **G-9.3 Monitor wiring (fix stub)** — `frontend/src/features/monitor-view/MonitorView.tsx:1` replace `Math.random()` with `MonitorRun` + `apps/desktop/backend/monitor.go:1` `MonitorRun` — live availability/latency chart — 2026-09-02 gap
- [ ] **G-9.4 Changelog view** — `frontend/src/features/changelog-view/ChangelogView.tsx:1` + `apps/desktop/backend/changelog.go:1` `ChangelogGenerate` — `changelog <old> <new> --format --fail-on-breaking` + SemVer bump — 2026-09-02 gap

## GUI-10 CLI Parity — Developer Tooling — mirrors `ROADMAP.md:UI-16`

> **Gap:** `ai explain|test|docs|diagnose|schema`, `schema validate|inspect|generate`, `plugin list|validate` are CLI-only.

- [ ] **G-10.1 AI assistant panel** — `apps/desktop/backend/ai.go:1` `AiExplain/GenerateTests/GenerateDocs/Diagnose/ExplainSchema` via `internal/ai` + `frontend/src/features/ai-view/AiView.tsx:1` — local heuristics, zero telemetry — 2026-09-02 gap
- [ ] **G-10.2 JSON Schema workbench** — `apps/desktop/backend/schema.go:1` `SchemaValidate/Inspect/Generate` via `internal/jsonschema` + `frontend/src/features/schema-view/SchemaView.tsx:1` — violation paths, keywords, sample generation — 2026-09-02 gap
- [ ] **G-10.3 Plugin manager** — `apps/desktop/backend/plugin.go:1` `PluginList/Validate` via `internal/plugin` + `frontend/src/features/plugin-view/PluginView.tsx:1` — `plugins/<name>` table + capabilities — 2026-09-02 gap

## GUI-11 CLI Parity — Parity Polish — mirrors `ROADMAP.md:UI-17`

> **Gap:** partial parity — `jwt verify|sign`, `openapi validate|convert-v2`, `graphql parse`, `theme import`, `validate project`, plus stale `ROADMAP` todos.

- [ ] **G-11.1 JWT verify/sign** — `frontend/src/features/jwt-inspector/JwtInspector.tsx:1` extend + `apps/desktop/backend/jwtdialog.go:1` `JwtVerify/JwtSign` via `internal/jwt` — 2026-09-02 gap
- [ ] **G-11.2 OpenAPI polish** — `frontend/src/features/openapi-explorer/OpenapiExplorer.tsx:1` + `apps/desktop/backend/openapiexplorer.go:1` `OpenapiValidate/ConvertV2` via `internal/openapi` — 2026-09-02 gap
- [ ] **G-11.3 GraphQL parse + Theme import + Validate project** — `frontend/src/features/graphql-browser/GraphqlBrowser.tsx:1` `GraphqlParse` + `frontend/src/features/settings-view/SettingsView.tsx:1` `ThemeImport` file picker + `frontend/src/features/spec-editor/SpecEditorView.tsx:1` `ValidateProject` — 2026-09-02 gap
- [ ] **G-11.4 Stale todos closeout** — `frontend/src/features/history-view/HistoryView.tsx:1` retention pruning `DELETE WHERE createdAt <`, `frontend/src/features/import-dialog/ImportDialog.tsx:1` deep merge, `frontend/src/hooks/useKeyboardMap.ts:1` editable shortcuts, `frontend/src/features/settings-view/SettingsView.tsx:1` Auth Settings sub-page — closes `UI-04`/`UI-06`/`UI-07` `[ ]` — 2026-09-02 gap

> **Intentional gap (no GUI):** `apps/cli/cmd/mcp.go:1` `mcp serve` (stdio MCP) — headless, no GUI expected — not tracked as milestone.

