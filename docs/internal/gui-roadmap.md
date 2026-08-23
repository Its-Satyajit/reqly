# Reqly — Desktop GUI Roadmap

> **Status:** Phase 1 core GUI shipped, Phase 2 GUI gaps identified
> **Source of truth:** [`ROADMAP.md`](../../ROADMAP.md) (core + CLI), this doc (desktop GUI only)
> **Design review:** [`frontend-design-review-2026-08-23.md`](./frontend-design-review-2026-08-23.md)
> **Stack:** React 19, TypeScript, Vite, Tailwind v4, shadcn/ui (base-nova), Zustand, CodeMirror 6, Wails v3 bindings
>
> This roadmap covers only the desktop GUI layer (`frontend/` shared UI + `apps/desktop/frontend/` Wails host).
> Core logic, CLI commands, and Go packages are tracked in the main [`ROADMAP.md`](../../ROADMAP.md).
> Each milestone links back to the main ROADMAP section where the underlying feature was shipped.

---

## Legend

- `[x]` — GUI shipped & verified
- `[~]` — partial (GUI exists but incomplete)
- `[ ]` — not started
- `[!]` — design review finding (from `frontend-design-review-2026-08-23.md`)
- **(stub)** — scaffold exists but no logic

---

## GUI-0 — Foundation (shipped)

The shared UI shell and desktop host. All shipped in Phase 0 of the main ROADMAP.

- [x] App shell — header, sidebar, split request/response panes (`0.3`)
- [x] Light/dark theming — Reqly brand colors, theme store, toggle (`0.3`)
- [x] Dark/light logo in header; logo as app icon (`0.3`)
- [x] Base UI via shadcn CLI — button component, `#`-alias imports (`0.3`)
- [x] CodeMirror 6 editor wrapper — json/js/xml/yaml/markdown/text (`0.3`)
- [x] Wails v3 desktop shell — `main.go`, `AppService` binding, window 1280×800 (`0.2`)
- [x] Go ↔ TypeScript bindings — `wails3 generate bindings` (`0.2`)
- [x] Host app — Vite + React + Tailwind, wails vite plugin, port 9245 (`0.2`)

---

## GUI-1 — Core Request Workflow (shipped)

The essential request→response loop in the GUI. Maps to main ROADMAP Phase 1.

- [x] Method select + URL bar + Send (`1.6`)
- [x] Params/Headers/Auth/Body/Variables tabs (`1.6`)
- [x] Body editors — JSON/XML/raw/binary/GraphQL via CodeMirror, form-data/urlencoded via key-value rows (`1.6` + `0013`)
- [x] Response viewer — metadata, raw/pretty/table/tree views, binary preview, search (`1.6` + `0014`)
- [x] JSONPath response querying (`1.6`)
- [x] Response actions — copy (body/headers), download (`1.6`)
- [x] Cookie jar — view/delete/clear in ResponseViewer Cookies tab (`1.6` + `0014`)
- [x] Per-tab request/response state + dirty tracking (`1.5a`)
- [x] Save/Overwrite/Reload with conflict detection (`1.5a` + `0009`)
- [x] Cancel in-flight request — Stop button, send token (`1.5a` + M33)
- [x] Auth editor — 10 schemes, scheme picker, per-scheme fields, secret masking (`1.3` + `0011`)
- [x] Auth inheritance display — inherited from collection/folder (`1.3`)
- [x] Environment variables — tag picker, `{{$` autocomplete, variable interpolation (`1.2` + `0015`)
- [x] Retry configuration — progressive disclosure section, 4 fields (`1.1` + `0024`)
- [x] Code generation — "Copy as" dropdown (cURL/JS/Python/Go) (`1.9` + `0016`)

---

## GUI-2 — Workspace Navigation (shipped)

Browse and manage requests, environments, and history.

- [x] Collection tree — expand/collapse, click to open, refresh (`1.5`)
- [x] Environments list — create/edit/delete/set active (`1.2`)
- [x] Environment editor — in-memory drafts, secrets masking, dirty tracking (`1.2`)
- [x] History view — paginated table (50/page), search (FTS), status filter, replay (`1.1` + `0014`)
- [x] History replay — opens in new tab with pre-filled data (`1.1`)

---

## GUI-3 — Collection Execution (shipped)

Run and inspect collection execution in the GUI.

- [x] Run view — dedicated tab, live status, expandable tests/logs/response (`1.7`)
- [x] Streaming execution — real-time step updates via Wails bindings (`1.7`)
- [x] Fail-fast toggle + cancel (`1.7`)
- [x] Script logs — pre/post script output in run steps (`1.7`)

---

## GUI-4 — Design Quality (from design review)

Accessibility, consistency, and UX fixes identified in `frontend-design-review-2026-08-23.md`.

### 4.1 Accessibility (Critical — blocks WCAG compliance)

- [!] **G-4.1.1** Add tab ARIA semantics — `role="tablist"`, `role="tab"`, `role="tabpanel"`, `aria-selected` on RequestEditor and ResponseViewer tab bars (`RequestEditor.tsx`, `ResponseViewer.tsx`)
- [!] **G-4.1.2** Add `aria-label` or `<label>` to all form inputs — URL input, environment name, history search, OAuth config textarea (`RequestEditor.tsx:209`, `EnvironmentsView.tsx:106`, `HistoryView.tsx:88`, `AuthEditor.tsx:557`)
- [!] **G-4.1.3** Fix primary button contrast in light mode — `--primary` (#e14b31) on `--primary-foreground` (#ffffff) = ~3.2:1, needs 4.5:1 (`index.css:76-77`)
- [!] **G-4.1.4** Add `prefers-contrast` media query handling (`index.css:142-150`)

### 4.2 Consistency (DRY violations)

- [!] **G-4.2.1** Extract `tabClass` to shared module — identical in `RequestEditor.tsx:34-39` and `ResponseViewer.tsx:37-42`
- [!] **G-4.2.2** Extract `inputClass` to shared module — different CSS in `EnvironmentsView.tsx:19-20` (has `focus:outline-none focus-visible:border-ring`) vs `KeyValueEditor.tsx:13-14` (no `focus-visible:border-ring`). Reconcile and deduplicate.
- [!] **G-4.2.3** Extract `formatBytes` to shared module — identical in `ResponseViewer.tsx:44-48` and `RunView.tsx:8-12`

### 4.3 Keyboard Shortcuts

- [!] **G-4.3.1** Add `⌘/Ctrl+Enter` for Send — standard API client convention, not in any ADR but expected per design-principles platform conventions
- [!] **G-4.3.2** Add `⌘/Ctrl+W` to close tabs — standard desktop convention
- [!] **G-4.3.3** Add arrow key navigation for request/response tab bars — currently no keyboard arrow support

### 4.4 Loading & Feedback States

- [!] **G-4.4.1** Add loading skeleton in ResponseViewer — replace `// Sending request…` text with skeleton placeholder matching design-principles motion rules
- [!] **G-4.4.2** Add loading indicator to Send button — pulse, spinner, or progress bar during in-flight requests (currently only disables + shows "Stop")

### 4.5 Documentation Fixes

- [!] **G-4.5.1** Update CONTEXT.md Response View entry (line 129-130) — add Table as 6th tab (added in M22)
- [!] **G-4.5.2** Update CONTEXT.md Copy as entry (line 225-226) — ResponseViewer does NOT have "Copy as", only Copy/Copy headers/Download

---

## GUI-5 — Import & Export (P1 gap — highest-impact missing feature)

The design review identifies Import as the highest-impact missing GUI feature. Users migrating from Postman/Insomnia cannot bring collections into the GUI without CLI.

Core logic: `internal/importer`, `internal/exporter` (shipped in main ROADMAP `1.9`).

### 5.1 Import

- [ ] **G-5.1.1** Import dialog — modal with format auto-detection (cURL, OpenAPI 3.x, HAR, Postman v2.1, Insomnia v4/v5, Bruno)
- [ ] **G-5.1.2** cURL import — paste cURL command, parse, create request tab (wraps `reqly import curl`)
- [ ] **G-5.1.3** OpenAPI import — file picker for JSON/YAML, preview operations, import to workspace (wraps `reqly import openapi`)
- [ ] **G-5.1.4** HAR import — file picker, preview requests, import (wraps `reqly import har`)
- [ ] **G-5.1.5** Postman import — file picker, preview collections, import with env vars (wraps `reqly import postman`)
- [ ] **G-5.1.6** Insomnia import — file picker, auto-detect format, import (wraps `reqly import insomnia`)
- [ ] **G-5.1.7** Bruno import — file picker, import items tree (wraps `reqly import bruno`)
- [ ] **G-5.1.8** Import results summary — show imported collections/requests/envs with warnings

### 5.2 Export

- [ ] **G-5.2.1** Export dialog — modal with format selection and scope (current request, collection, workspace)
- [ ] **G-5.2.2** Export as Postman — wraps `reqly export postman`
- [ ] **G-5.2.3** Export as HAR — wraps `reqly export har`
- [ ] **G-5.2.4** Export as OpenAPI — wraps `reqly export openapi`
- [ ] **G-5.2.5** Export workspace — wraps `reqly export workspace`

---

## GUI-6 — Test Runner (P1 gap)

No assertion builder or test file editor in GUI. Core logic: `internal/testing` (shipped in main ROADMAP `1.7`).

- [ ] **G-6.1** Test file editor — CodeMirror YAML/JSON editor for `.reqly-test.yaml` files
- [ ] **G-6.2** Assertion builder — visual form for status/header/body/schema assertions (wraps `internal/testing` assertion types)
- [ ] **G-6.3** Test results view — pass/fail per assertion, response diff, timing
- [ ] **G-6.4** Test file tab — open/edit/save test files alongside request tabs
- [ ] **G-6.5** Run tests from GUI — trigger `reqly test` or inline test execution, show results in panel

---

## GUI-7 — WebSocket & SSE Client (P1 gap)

Protocol-specific UIs for real-time APIs. Core logic: `internal/websocket`, `internal/sse` (shipped in main ROADMAP `1.8`).

### 7.1 WebSocket

- [ ] **G-7.1.1** WebSocket tab — dedicated request tab type for `ws://` / `wss://` URLs
- [ ] **G-7.1.2** Connection panel — connect/disconnect, status indicator, protocol options
- [ ] **G-7.1.3** Message composer — send text/binary frames, JSON formatting
- [ ] **G-7.1.4** Message inspector — incoming/outgoing message log with timestamps, expand for full payload

### 7.2 SSE

- [ ] **G-7.2.1** SSE tab — dedicated request tab type for SSE endpoints
- [ ] **G-7.2.2** Event stream view — live event list with named/ID'd events, auto-scroll
- [ ] **G-7.2.3** Event inspector — expand event to see full payload, retry hints
- [ ] **G-7.2.4** Event history — buffer last N events for review

---

## GUI-8 — Mock Server (P2 gap)

Manage mock configs from the GUI. Core logic: `internal/mocking` (shipped in main ROADMAP `P1`).

- [ ] **G-8.1** Mock server panel — sidebar section or dedicated view
- [ ] **G-8.2** Mock configuration — create/edit/delete mocks from OpenAPI specs or manual route definitions
- [ ] **G-8.3** Mock route list — show routes with method/path/status, enable/disable per route
- [ ] **G-8.4** Mock responses — edit response bodies, status codes, headers, delays
- [ ] **G-8.5** Mock server controls — start/stop, port config, `--delay`, `--fail-every` sliders

---

## GUI-9 — API Diff & Comparison (P2 gap)

Compare API definitions and responses. Core logic: `internal/diffing` (shipped in main ROADMAP `1.11`).

- [ ] **G-9.1** Diff view — side-by-side or inline diff of two OpenAPI specs or request files
- [ ] **G-9.2** Breaking change highlights — color-code additions/removals/changes by severity
- [ ] **G-9.3** Response diff — compare two responses (JSON structural diff)
- [ ] **G-9.4** Diff from history — select two history entries, compare responses

---

## GUI-10 — JWT Inspector (P2 gap)

Decode and inspect JWT tokens in the GUI. Core logic: `internal/jwt` (shipped in main ROADMAP M29).

- [ ] **G-10.1** JWT decode panel — paste or auto-capture token from auth flow
- [ ] **G-10.2** Header/claims viewer — formatted JSON with field highlighting
- [ ] **G-10.3** Expiry indicator — visual warning for expired tokens, time-to-expiry badge
- [ ] **G-10.4** Auto-capture — decode JWT from Bearer token in response headers

---

## GUI-11 — GraphQL Schema Browser (P2 gap)

Browse GraphQL schemas in the GUI. Core logic: `internal/graphql` (shipped in main ROADMAP M38, desktop deferred to M38b).

- [ ] **G-11.1** Schema browser panel — tree view of types, queries, mutations, subscriptions
- [ ] **G-11.2** Type inspector — click type to see fields, args, return types
- [ ] **G-11.3** Introspection trigger — button to run introspection query, cache result
- [ ] **G-11.4** Doc explorer — inline documentation for types and fields

---

## GUI-12 — Pagination & Bulk Runner (P2 gap)

Paginated and bulk execution in the GUI. Core logic: `internal/pagination`, `internal/bulk` (shipped in main ROADMAP M30/M31).

### 12.1 Pagination Runner

- [ ] **G-12.1.1** Pagination config panel — strategy selector (page/offset/cursor/link-header)
- [ ] **G-12.1.2** Stop condition editor — max pages, stop-on-empty, custom condition
- [ ] **G-12.1.3** Results aggregation — merged response view across pages

### 12.2 Bulk Runner

- [ ] **G-12.2.1** Bulk execution panel — file picker for CSV/JSON input
- [ ] **G-12.2.2** Concurrency config — sequential/parallel, max concurrent
- [ ] **G-12.2.3** Progress view — per-row status, aggregate stats, failure details

---

## GUI-13 — Environment Diff & Validation (P2 gap)

Compare and validate environment configs. Core logic: `internal/environments` (shipped in main ROADMAP `1.2`).

- [ ] **G-13.1** Env diff view — side-by-side comparison of two environment files
- [ ] **G-13.2** Env validation panel — run `reqly env validate` from GUI, show warnings/errors inline
- [ ] **G-13.3** Cross-env validation — check variable consistency across environments

---

## GUI-14 — OpenAPI Explorer (P2 gap)

Browse and generate from OpenAPI specs in the GUI. Core logic: `internal/openapi` (shipped in main ROADMAP M39, desktop deferred to M39b).

- [ ] **G-14.1** Spec browser — tree view of paths/operations from loaded OpenAPI file
- [ ] **G-14.2** Operation preview — method, params, request body, responses
- [ ] **G-14.3** Generate request — click operation to create a pre-filled request tab
- [ ] **G-14.4** Schema viewer — inline JSON Schema display for request/response bodies

---

## GUI-15 — Documentation Generation (P3 gap)

Generate API docs from the GUI. Core logic: `internal/docs` (shipped in main ROADMAP M26, desktop deferred to M26b).

- [ ] **G-15.1** Docs generation panel — select collections, configure output
- [ ] **G-15.2** Preview generated docs — Markdown preview of output
- [ ] **G-15.3** Export docs — save to file or copy to clipboard

---

## GUI-16 — Report Export (P3 gap)

Export test results from the GUI. Core logic: `internal/runner` (shipped in main ROADMAP M37).

- [ ] **G-16.1** Report export from Run View — JUnit XML / JSON export button
- [ ] **G-16.2** Report preview — formatted test results summary

---

## Priority Matrix

| Priority | Milestone | Impact | Effort | Depends on |
|----------|-----------|--------|--------|------------|
| **P0** | GUI-4 (Design Quality) | Accessibility + consistency | Low-Medium | — |
| **P1** | GUI-5 (Import & Export) | Onboarding + collaboration | High | — |
| **P1** | GUI-6 (Test Runner) | Core developer tool gap | High | GUI-3 (Run View) |
| **P1** | GUI-7 (WebSocket & SSE) | Real-time API debugging | Medium | GUI-1 (Request) |
| **P2** | GUI-8 (Mock Server) | Frontend dev workflow | Medium | GUI-5 (Import) |
| **P2** | GUI-9 (API Diff) | Version management | Medium | GUI-5 (Import) |
| **P2** | GUI-10 (JWT Inspector) | Auth debugging | Low | GUI-1 (Auth Editor) |
| **P2** | GUI-11 (GraphQL Schema) | GraphQL users | Medium | GUI-1 (Body Editor) |
| **P2** | GUI-12 (Pagination & Bulk) | Power-user workflows | Medium | GUI-1 (Request) |
| **P2** | GUI-13 (Env Diff & Validate) | Config management | Low | GUI-2 (Environments) |
| **P2** | GUI-14 (OpenAPI Explorer) | Spec-driven development | Medium | GUI-5 (Import) |
| **P3** | GUI-15 (Docs Generation) | Documentation automation | Low | — |
| **P3** | GUI-16 (Report Export) | CI integration | Low | GUI-3 (Run View) |

---

## Cross-Reference to Main ROADMAP

| GUI Milestone | Main ROADMAP Section | Core Package | Status |
|---------------|---------------------|--------------|--------|
| GUI-0 | `0.2`, `0.3` | Wails shell, shared UI | Shipped |
| GUI-1 | `1.1`, `1.3`, `1.5a`, `1.6` | request, auth, core, variables | Shipped |
| GUI-2 | `1.2`, `1.5` | collections, environments, history | Shipped |
| GUI-3 | `1.7` | scripting, collections | Shipped |
| GUI-4 | — (design review) | frontend components | Open |
| GUI-5 | `1.9` | importer, exporter | Core shipped, GUI open |
| GUI-6 | `1.7`, `1.11` | testing | Core shipped, GUI open |
| GUI-7 | `1.8` | websocket, sse | Core shipped, GUI open |
| GUI-8 | `P1` | mocking | Core shipped, GUI open |
| GUI-9 | `1.11` | diffing | Core shipped, GUI open |
| GUI-10 | `M29` | jwt | Core shipped, GUI open |
| GUI-11 | `M38` | graphql | Core shipped, GUI deferred M38b |
| GUI-12 | `M30`, `M31` | pagination, bulk | Core shipped, GUI open |
| GUI-13 | `1.2` | environments | Core shipped, GUI open |
| GUI-14 | `M39` | openapi | Core shipped, GUI deferred M39b |
| GUI-15 | `M26` | docs | Core shipped, GUI deferred M26b |
| GUI-16 | `M37` | runner | Core shipped, GUI open |
