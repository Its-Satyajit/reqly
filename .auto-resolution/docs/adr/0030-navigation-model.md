# ADR 0030 — Navigation model (GUI-2, spec §4, §60–63)

Date: 2026-08-27 · Status: Accepted

## Context

Spec §4 defines a two-axis navigation (§4), with §60 enumerating 15+ full pages, §61 shared patterns, §62 page-vs-panel rules, and §63 the canonical five-zone shell. The pre-M44 frontend used a flat sidebar list for tool selection and a three-pane request editor — neither matched the Atlas reference (`design-9` screenshots) which has an icon activity rail + workspace pill header + context sidebar + main workspace + status bar.

GUI-0 (Tickets #01–#04, 2026-08-27) shipped the Atlas chrome (TopBar, Tool Rail, Sidebar, Workspace, Bottom Panel) against `index.css` Atlas tokens. GUI-1 locked the Design System contract. GUI-2 must freeze the navigation semantics so Phase 2 tools (Explorer, Docs, Runners, Mocks, etc.) have a consistent routing and panel model.

## Decision

1. **Two-axis (§4):** Horizontal = Tool Rail (4 groups: Workspace {Home, Requests, Environments, History}, API Tools {Mocks, Diff, JWT, GraphQL, gRPC, Runners, Explorer, Docs}, Realtime {WebSocket, SSE}, System {Settings}); vertical = Context Sidebar (resource tree/search/actions scoped to the active tool). Rail drives route `/tool/:id`; sidebar content switches per tool. `⌘B` toggles sidebar, persisted.

2. **Navigation Map (§60):** 15 tool pages in Main Workspace, each a top-level route, lazy-loaded (`React.lazy`): Home, Requests, Environments, History, Mocks, Diff, JWT, GraphQL, gRPC, Runners, Explorer, Docs, WebSocket, SSE, Settings. Sub-panels (Requests → Params/Headers/Body/Auth/Tests/Response) are tabs/panels inside the page, not separate routes.

3. **Page vs Panel (§62):** Page = full Main Workspace route per tool; Context Panel = sidebar resource nav per tool (Collections tree, History filters); Bottom Panel = cross-cutting inspectors (Console/Network/Tests/Variables/Cookies, `⌘J` resizable); Dialog = transient action (Import/Export/Create, confirm destructive). Determines routing, persistence, and focus.

4. **Shared Patterns (§61):** Global Command Palette (`⌘K`, already `lib/themes` + palette) + per-tool filter input; primary (coral) / secondary (neutral) action hierarchy; StatusPill dot+code (never color alone); shared `tabs.tsx`/`button`/`status.tsx` primitives; hairline panel chrome.

5. **Final Layout (§63):** Canonical five-zone shell TopBar → Tool Rail (48–56px) → Context Sidebar (220–280px, resizable/collapsible) → Main Workspace (tab-based) → Bottom Panel (resizable, `⌘J`). Single source of truth; layout (sidebar collapsed, bottom height, active tool) persisted to localStorage. All chrome consumes Design Tokens only (ADR 0029).

## Consequences

- Adding a new tool = one route + one sidebar renderer + one page component; no shell changes.
- Bottom panel is never tool-specific; tool-specific inspectors live in the page's own panels.
- Visual regression for shell must screenshot-diff against `design-9` reference per phase gate.

## References

- `docs/Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification.md` §4, §60–63
- `docs/internal/gui-roadmap.md` GUI-0, GUI-1, GUI-2
- `DESIGN.md` Layout; `frontend/src/index.css` tokens (ADR 0029)
- `docs/internal/ui-demos/design-9-git-native/screenshots/`
