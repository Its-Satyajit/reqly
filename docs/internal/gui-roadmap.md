# Reqly — GUI Roadmap

> Desktop GUI milestones — tracks `apps/desktop/backend/frontend` delivery.
> **Source of truth:** [`docs/Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification.md`](../Reqly%20Complete%20UI%20Architecture,%20Pages,%20Panels,%20and%20Navigation%20Specification.md) (full UI spec — §1–§59)

## GUI-0 Shell Redesign — RESTARTING FROM SCRATCH

> **⚠️ RESTARTING** — Previous implementation (G-3.x) did not follow spec §2 four-zone architecture. All UI components will be rewritten following the spec's TopBar / ToolRail / ContextSidebar / MainWorkspace / BottomPanel model.

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

## GUI-1 Design System (spec §3)

- [ ] Design tokens (spacing, radius, shadows)
- [ ] Typography system (IBM Plex Sans/Mono)
- [ ] Color system (terracotta accent #c93517/#ff6f52, BASE 6px radius)
- [ ] Status indicators (Connected/Running/Valid/Success/Warning/Error)
- [ ] Hairline borders, no shadows

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
