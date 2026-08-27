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

- [ ] Logo
- [ ] Workspace Switcher
- [ ] Global Search ⌘K
- [ ] Import / Export buttons
- [ ] Active Environment selector
- [ ] Sync Status indicator
- [ ] Notifications
- [ ] Settings
- [ ] User Menu

### §2.2 Tool Rail (48–56px, left-most)

- [ ] Workspace group: Home, Requests, Environments, History
- [ ] API Tools group: Mocks, Diff, JWT, GraphQL, gRPC, Runners, Explorer, Docs
- [ ] Realtime group: WebSocket, SSE
- [ ] System group: Settings
- [ ] Icon-based routing (top-level navigation)

### §2.3 Context Sidebar (220–280px)

- [ ] Collapsible/resizable (drag handle)
- [ ] Changes per active tool
- [ ] Tree navigation
- [ ] Search within tool
- [ ] Contextual actions
- [ ] Recent/pinned items
- [ ] `⌘B` toggle

### §2.4 Main Workspace

- [ ] Tab-based content area
- [ ] Page routing per active tool
- [ ] Full pages vs context panels (§62 rules)

### §2.5 Bottom Utility Panel

- [ ] Console tab
- [ ] Network tab
- [ ] Tests tab
- [ ] Variables tab
- [ ] Cookies tab
- [ ] `⌘J` toggle
- [ ] Resizable height

## GUI-1 Design System (spec §3) — ✅ shipped 2026-08-27

- [x] Design tokens (spacing, radius, shadows) — `frontend/src/index.css` `@theme` semantic vars (`--background`/`--border`/`--primary`/`--status-*`/`--radius: 0.375rem`), grep gate for hardcoded hex
- [x] Typography system (IBM Plex Sans/Mono) — `@fontsource` 400/500/600 + 13px/1.45 base, `.font-data` mono discipline
- [x] Color system (terracotta accent #c93517/#ff6f52, BASE 6px radius) — AA-adjusted `#c93517` (4.5:1) + `prefers-contrast: more` bump, `DESIGN.md` Color/Tokens sections
- [x] Status indicators (Connected/Running/Valid/Success/Warning/Error) — Status Ramp + `StatusPill` dot+code (never color alone), method tints
- [x] Hairline borders, no shadows — `border-border` shell/cards/panels; `shadow-md/lg` only on floating `popover`/`dropdown-menu`/`select`/`toast`

## GUI-2 Navigation Model (spec §4, §60–63)

- [ ] Two-axis navigation: horizontal (tool rail) + vertical (sidebar)
- [ ] 15+ full pages with sub-panels
- [ ] Page vs panel rules (§62)
- [ ] Shared interaction patterns (§61)
- [ ] Final layout model (§63)

## GUI-5 P1 Data Layer (spec §56.3–56.8) — PRESERVED

> These data layer items (lib + stores + tests) are preserved from the previous implementation.

- [x] **G-5.1** Request templates — zustand store + pure lib (search, instantiate, CRUD) + 21 tests — 2026-08-26
- [x] **G-5.2** Proxy/TLS controls — zustand store + pure lib (validate, format, defaults) + 22 tests — 2026-08-26
- [x] **G-5.3** Data-driven testing — zustand store + pure lib (CSV/JSON parse, row vars, validate) + 23 tests — 2026-08-26
- [x] **G-5.4** CI/CD integration — zustand store + pure lib (CLI gen, GitHub Action YAML, report parse) + 13 tests — 2026-08-26
- [x] **G-5.5** Mock server GUI data — extended zustand store + pure lib (scenarios, fault injection, matchers, logs) + 20 tests — 2026-08-26
- [x] **G-5.6** GraphQL/gRPC docs — zustand store + pure lib (SDL parse, search, Markdown render) + 16 tests — 2026-08-26
