# Reqly — GUI Roadmap

> Desktop GUI milestones (GUI-0 through GUI-16) — tracks `apps/desktop/backend/frontend` delivery.
> Source of truth for `ROADMAP.md` Phase 1 GUI items.

## GUI-3 Shell Redesign — P0 UI Architecture (spec #369)

- [x] **G-3.1** Tool rail 52px grouped (Workspace/API Tools/Realtime/System) — 2026-08-26
- [x] **G-3.2** ContextSidebar 220–280px resizable/collapsible — 2026-08-26
- [x] **G-3.3** TopBar (workspace switcher, global search ⌘K, import/export, env selector) — 2026-08-26
- [x] **G-3.4** Theme registry `atlas-light`/`atlas-dark`/`system` with `resolvedTheme` — 2026-08-26
- [x] **G-3.5** Realtime `websocket`/`sse` single-connection pages + recents — 2026-08-26
- [x] **G-3.6** Command palette + data providers (Fuse) — 2026-08-26
- [x] **G-3.7** Settings view + bottom utility dock (Console/Network/Tests/Variables/Cookies, ⌘J) — 2026-08-26
- [x] **G-3.8** Spec editor tree + YAML + schema viz (§56.1–56.2) — 2026-08-26

## GUI-4 Design Quality

- [x] **G-4.0** Infinite-loop fix for palette `filtered` selector — 2026-08-26
- [x] **G-4.1** Race-safe `emitRunEvent` for `go test -race` — 2026-08-26

## GUI-5 P1 Data Layer (spec §56.3–56.8)

- [x] **G-5.1** Request templates — zustand store + pure lib (search, instantiate, CRUD) + 21 tests — 2026-08-26
- [x] **G-5.2** Proxy/TLS controls — zustand store + pure lib (validate, format, defaults) + 22 tests — 2026-08-26
- [x] **G-5.3** Data-driven testing — zustand store + pure lib (CSV/JSON parse, row vars, validate) + 23 tests — 2026-08-26
- [x] **G-5.4** CI/CD integration — zustand store + pure lib (CLI gen, GitHub Action YAML, report parse) + 13 tests — 2026-08-26
- [x] **G-5.5** Mock server GUI data — extended zustand store + pure lib (scenarios, fault injection, matchers, logs) + 20 tests — 2026-08-26
- [x] **G-5.6** GraphQL/gRPC docs — zustand store + pure lib (SDL parse, search, Markdown render) + 16 tests — 2026-08-26
