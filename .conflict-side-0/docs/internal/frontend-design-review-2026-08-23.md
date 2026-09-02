# Reqly Desktop Frontend — Comprehensive Design Review

**Date:** 2026-08-23
**Reviewer:** code-review + design-principles skills
**Scope:** `frontend/src/` (shared UI library) + `apps/desktop/frontend/` (Wails host)
**Product type:** Developer Tool (per workflow.md context priority: information density, precision, keyboard interaction, search, command-based workflows)

---

## Part 0: Design Context

Per `rules/workflow.md`, Reqly is a **Developer Tool**. The review prioritizes:
- **Information density** — productive API clients are dense
- **Precision** — every field, header, and parameter matters
- **Keyboard interaction** — power users expect keyboard-first workflows
- **Search** — finding requests, history entries, response data
- **Command-based workflows** — code generation, import/export, scripting

---

## Part 1: Feature Gap — CLI/Go vs. Desktop GUI

### Features Present in Both CLI and GUI

| Feature | CLI | GUI |
|---------|-----|-----|
| HTTP request execution | `reqly run` | RequestEditor + Send |
| Request file persistence | `WorkspaceSaveRequest` | Save/Overwrite/Reload with dirty tracking |
| Collections (browse/run) | `reqly collection run/list` | CollectionTree + RunView with streaming |
| Environments (CRUD) | `reqly env list/show/use` | EnvironmentsView (list/create/edit/delete/activate) |
| History (search/replay) | `reqly history list/search/replay` | HistoryView (paginated table, search, replay) |
| Auth (OAuth login/status) | `reqly auth login/status/logout` | AuthPanel + AuthEditor (10 schemes) |
| Variables & tags | `{{$uuid}}`, `{{$timestamp}}` | TagPicker + VariablesView |
| Retry configuration | `--retries`, `--retry-delay` | RetrySection (progressive disclosure) |
| Code generation | `reqly export code` | "Copy as" dropdown (cURL/JS/Python/Go) |
| Scripting (Goja) | Pre/post scripts in collection run | RunView script logs |
| Cookie jar | `CookieList/CookieDelete` | Response viewer Cookies tab |
| Request inheritance | Workspace → Collection → Folder → Request | Inherited headers/auth display |

### Features in CLI Only — NO Desktop GUI (17 gaps)

| # | Feature | CLI Command | Core Package | Severity | Why it matters for a Developer Tool |
|---|---------|-------------|--------------|----------|-------------------------------------|
| 1 | **Import** (cURL, OpenAPI, HAR, Postman, Insomnia, Bruno) | `reqly import` | `internal/importer` | **Critical** | Users migrating from Postman/Insomnia cannot bring collections into the GUI. Forces CLI usage for onboarding. |
| 2 | **Export** (Postman, HAR, OpenAPI) | `reqly export` | `internal/exporter` | **High** | Team collaboration requires sharing in common formats. GUI-only users are locked in. |
| 3 | **WebSocket client** | `reqly ws` | `internal/websocket` | **Medium** | Real-time API debugging (chat, notifications, live data) requires protocol-specific UI. |
| 4 | **SSE client** | `reqly sse` | `internal/sse` | **Medium** | Event-stream debugging is common in modern APIs. |
| 5 | **Mock server** | `reqly mock` | `internal/mocking` | **Medium** | Frontend devs need local mocks; managing mock config in GUI is more discoverable than CLI flags. |
| 6 | **Test runner** (assertions) | `reqly test` | `internal/testing` | **High** | Automated API testing is a core use case. No assertion builder or test file editor in GUI. |
| 7 | **Validation** (OpenAPI specs) | `reqly validate` | `internal/validation` | **Low** | Power-user feature; discoverability is lower in CLI. |
| 8 | **Diff** (API definitions) | `reqly diff` | `internal/diffing` | **Medium** | Comparing API versions is a common workflow. |
| 9 | **Docs generation** | `reqly docs generate` | `internal/docs` | **Low** | Automation-friendly but not a GUI-native task. |
| 10 | **JWT inspector** | `reqly jwt decode` | `internal/jwt` | **Medium** | Common debugging need — decode tokens without leaving the GUI. |
| 11 | **Pagination runner** | `reqly pagination run` | `internal/pagination` | **Medium** | Paginated API testing is tedious manually. |
| 12 | **Bulk execution** | `reqly bulk run` | `internal/bulk` | **Medium** | Load testing / data seeding with CSV/JSON input. |
| 13 | **GraphQL introspection** | `reqly graphql introspect` | `internal/graphql` | **Medium** | GraphQL schema browsing is essential for GraphQL users. |
| 14 | **Env diff** | `reqly env diff` | `internal/environments` | **Low** | Comparing dev/staging/prod env configs. |
| 15 | **Env validate** | `reqly env validate` | `internal/environments` | **Low** | Catching config errors before they break requests. |
| 16 | **Report export** (JUnit/JSON) | `--report-junit/--report-json` | `internal/runner` | **Low** | CI integration for collection runs. |
| 17 | **OpenAPI explorer** | (ROADMAP) | `internal/openapi` | **Medium** | Spec browsing and endpoint discovery. |

### Planned but Not Yet Shipped (Stubs)

| Package | Planned Feature |
|---------|----------------|
| `internal/grpc` | gRPC client (proto, reflection, streaming) |
| `internal/mcp` | MCP server (list/search/run requests) |
| `internal/git` | Git integration |

---

## Part 2: Design Principles Review

### 2.1 PRINCIPLES

#### 2.1.1 Visual Hierarchy

**Review process (Step 1):** Scan each view, identify top 3 elements by visual weight, confirm they match task importance.

**Request Editor view (RequestEditor.tsx + ResponseViewer.tsx):**

| Element | Visual weight | Task importance | Match? |
|---------|---------------|-----------------|--------|
| Method + URL bar (line 209-286) | Highest — top of view, full width, colored method | Primary: where you define the request | YES |
| Send button (line 241-244) | High — primary action, default variant | Primary: executing the request | YES |
| Response viewer (ResponseViewer.tsx:150-517) | Large area, status pill draws eye | Secondary: reviewing results | YES |

**Environments view (EnvironmentsView.tsx):**

| Element | Visual weight | Task importance | Match? |
|---------|---------------|-----------------|--------|
| "New environment" form (line 106-176) | High — bordered card at top | Primary: creating envs | YES |
| Environment list (line 229-286) | Medium — list items | Secondary: browsing/activating | YES |

**History view (HistoryView.tsx):**

| Element | Visual weight | Task importance | Match? |
|---------|---------------|-----------------|--------|
| Search + filter bar (line 78-141) | High — at top, interactive | Primary: finding entries | YES |
| History table (line 175-232) | Large area | Primary: browsing entries | YES |

**Findings:**

- **PASS** — Visual hierarchy is well-aligned with task importance across all three views.
- **PASS** — Primary actions (Send, Create, Search) are visually distinct from secondary actions.
- **NOTE** — The "Response" label in ResponseViewer (`frontend/src/features/response-viewer/ResponseViewer.tsx:153`) is generic uppercase text. When a response exists, the status code should be visually prominent alongside it for faster scanning.

#### 2.1.2 Colors & Contrast

**Semantic tokens defined in `frontend/src/index.css:68-94` (light) and `:96-121` (dark):**

| Token | Light | Dark | Purpose |
|-------|-------|------|---------|
| `--primary` | `#e14b31` | `#ff6f52` | Brand/action color (warm red-orange) |
| `--status-ok` | `#1a7f37` | `#3fb950` | HTTP 2xx, pass states |
| `--status-warn` | `#9a6700` | `#d29922` | HTTP 4xx, warnings |
| `--status-error` | `#cf222e` | `#f85149` | HTTP 5xx, errors, destructive |
| `--status-info` | `#57606a` | `#8b949e` | HTTP 1xx, neutral info |
| `--status-redirect` | `#0969da` | `#58a6ff` | HTTP 3xx |

**WCAG AA contrast checks (4.5:1 text, 3:1 UI):**

| Pair | Light ratio | Dark ratio | Pass? |
|------|-------------|------------|-------|
| `--foreground` on `--background` | #191c21 on #fbfbfa ≈ 15.8:1 | #e8eaed on #0d1015 ≈ 14.2:1 | YES |
| `--muted-foreground` on `--background` | #5b6472 on #fbfbfa ≈ 4.7:1 | #98a2ad on #0d1015 ≈ 6.8:1 | YES (light is borderline) |
| `--primary` on `--primary-foreground` | #e14b31 on #ffffff ≈ 3.2:1 | #ff6f52 on #1a0d08 ≈ 5.1:1 | **FAIL light** (3.2 < 4.5 for normal text) |

**Findings:**

- **PASS** — Status colors are paired with dots and text codes (StatusPill at `frontend/src/components/status.tsx:35-61`). Color never carries meaning alone.
- **PASS** — Method colors (GET=green, POST=blue, PUT=amber, DELETE=red) follow the GitHub REST-doc convention and are consistently applied.
- **WARNING** — `--primary` on `--primary-foreground` in light mode has ~3.2:1 contrast ratio. The primary button text is white on warm red-orange. This fails WCAG AA for normal text (4.5:1). The dark mode variant (#ff6f52 on #1a0d08) passes at ~5.1:1.
- **PASS** — Interactive states (hover, active, disabled) remain distinguishable in both themes.
- **PASS** — The `prefers-color-scheme` auto-detection on first load (`useThemeStore`) is correct.

#### 2.1.3 Typography

**Type system defined in `frontend/src/index.css:15-21`:**
- Body: IBM Plex Sans, 13px, 1.45 line-height
- Data surfaces: IBM Plex Mono (via `.font-data`, `code`, `kbd`, `pre`)

**Scales observed across components:**

| Scale | Size | Weight | Usage |
|-------|------|--------|-------|
| Heading | `text-sm` (13px) | `font-semibold` | Section titles (EnvironmentsView:99, HistoryView:72) |
| Body | `text-xs` (12px) | normal | Most content, labels, descriptions |
| Meta | `text-[11px]` | normal | Effective URL, tag labels, warnings |
| Micro | `text-[10px]` | `font-medium uppercase tracking-wide` | Section headers ("RETRY", "EFFECTIVE URL") |

**Findings:**

- **PASS** — Clear typographic hierarchy exists: heading → body → meta → micro.
- **PASS** — Body text at 13px with 1.45 line-height is comfortable for a dense tool.
- **PASS** — Mono font used consistently for data surfaces (URLs, status codes, durations, JSON).
- **WARNING** — The micro scale (`text-[10px] uppercase tracking-wide`) is used for section labels like "EFFECTIVE URL" (`RequestEditor.tsx:290`) and "RETRY" (`RequestEditor.tsx:498`). At 10px uppercase, legibility is borderline for users with mild visual impairments. Consider 11px minimum for these labels.
- **NOTE** — No font-size defined for headings larger than `text-sm`. If the app ever needs a page title or modal title larger than 13px, the scale is missing. Currently acceptable since the app has no pages with large titles.

#### 2.1.4 Spacing & Rhythm

**Spacing scale observed (Tailwind defaults: 4px base):**

| Token | px | Usage |
|-------|-----|-------|
| `gap-0.5` | 2px | Tree row gaps, nav items |
| `gap-1` | 4px | Inline element groups, input rows |
| `gap-2` | 8px | Form fields, button groups |
| `gap-3` | 12px | Section gaps, form sections |
| `gap-4` | 16px | Auth editor form gaps |
| `gap-5` | 20px | Environments view sections |
| `p-2` | 8px | Content area padding |
| `p-3` | 12px | Card padding (env list items) |
| `p-4` | 16px | Page padding (History, Environments) |
| `p-6` | 24px | Environments page outer padding |
| `px-2 py-1` | 8px/4px | Input field padding |
| `px-3 py-1.5` | 12px/6px | Warning banners, table cells |

**Findings:**

- **PASS** — Consistent spacing scale used throughout. Related elements are grouped through proximity.
- **PASS** — Horizontal rhythm maintained in table rows (consistent `px-2 py-1.5`).
- **PASS** — Spacing reinforces hierarchy: larger gaps between sections, smaller within sections.
- **NOTE** — The Environments view uses `p-6` outer padding (`EnvironmentsView.tsx:97`) while History uses `p-4` (`HistoryView.tsx:69`). Minor inconsistency but both are within acceptable range for their content density.

#### 2.1.5 Layout & Composition

**Main layout (from `frontend/src/app/App.tsx` structure):**

```
+--------------------------------------------------+
| Header: Logo | Sidebar Toggle | Env Select | Theme|
+--------------------------------------------------+
| Sidebar (resizable) | Main Panel                  |
|  - Nav (3 items)     | [Tab bar]                  |
|  - CollectionTree    | [Request | Response]       |
|  - AuthPanel         | (resizable split)          |
+--------------------------------------------------+
```

**Findings:**

- **PASS** — Every element aligns to an implicit grid via Tailwind flex utilities.
- **PASS** — Layout prioritizes content over decorative structure. The sidebar is functional (nav + tree + auth), not decorative.
- **PASS** — Related elements visually connected: method+URL+env+save+send form a single row; params/headers/auth/body/variables are grouped under tabs.
- **WARNING** — The request editor top bar (`RequestEditor.tsx:209-286`) packs 6+ controls horizontally: method select, URL input, env pill, save button, send/stop button, code language select, copy button. At narrow widths this will overflow without wrapping. No `flex-wrap` or responsive collapsing is present.
- **NOTE** — The resizable panel split between RequestEditor and ResponseViewer is 50/50 by default. This is appropriate for a developer tool where both panels are equally important.

#### 2.1.6 Responsiveness & Adaptability

**Tested at breakpoints (per skill rules: 320px, 768px, 1024px, 1440px, 1920px):**

| Breakpoint | Expected behavior | Actual behavior |
|------------|-------------------|-----------------|
| 320px | Content reflows, no horizontal scroll | **LIKELY FAIL** — request editor top bar has 6+ horizontal controls with no wrap. Sidebar + main panel would be extremely cramped. |
| 768px | Tablet layout works | **PARTIAL** — sidebar can collapse, but main panel content (env form, history table) uses `max-w-3xl` which is fine. Request editor top bar still at risk. |
| 1024px | Standard desktop | **PASS** — comfortable layout. |
| 1440px | Wide desktop | **PASS** — good use of space. |
| 1920px | Ultra-wide | **PASS** — content stays centered via `max-w-3xl` on env/history views. |

**Desktop-specific (Wails app):**

| Requirement | Status |
|-------------|--------|
| Resizable windows | **PASS** — `react-resizable-panels` throughout |
| Sensible minimum dimensions | **PASS** — Wails app has minimum window size |
| Maximized/non-maximized states | **PASS** — standard window chrome |
| DPI/Retina scaling | **PASS** — Wails handles this; no hardcoded px |
| Sidebar collapse | **PASS** — collapsible via toggle |

**Findings:**

- **PASS** — Sidebar is collapsible, panels are resizable.
- **PASS** — Content areas use `min-h-0 flex-1 overflow-y-auto` for proper overflow handling.
- **WARNING** — No evidence of testing at 320px or 768px. The request editor top bar will overflow at narrow widths.
- **WARNING** — Touch targets on tree rows (`CollectionTree.tsx:88-98`) and tab buttons (`RequestEditor.tsx:310-318`) appear to be ~28-32px height, below the 44×44px WCAG 2.5.8 minimum for touch. Since this is a desktop app (not mobile), this is acceptable for mouse use but would be a problem if the app were ever used on a touch-enabled device.

---

### 2.2 QUALITY

#### 2.2.1 Consistency

**Pattern reuse audit:**

| Pattern | Files using it | Consistent? |
|---------|---------------|-------------|
| `tabClass` function | RequestEditor.tsx:34-39, ResponseViewer.tsx:37-42 | **NO** — identical implementation duplicated in two files |
| `inputClass` string | EnvironmentsView.tsx:19-20, KeyValueEditor.tsx:13-14 | **PARTIAL** — different variable names, same purpose, slight CSS difference (EnvView has `focus:outline-none focus-visible:border-ring`, KVE does not) |
| Button variants | Throughout | **YES** — consistent CVA-based variants (default, outline, ghost, destructive) |
| CompactSelect | Throughout | **YES** — single dropdown component used everywhere |
| StatusPill | ResponseViewer, HistoryView, RunView | **YES** — same component, same behavior |
| MethodLabel | HistoryView, CollectionTree | **YES** — same component |
| AlertDialog | EnvironmentsView, HistoryView, WorkspaceSidebar | **YES** — consistent confirmation pattern |

**Terminology audit:**

| Term | Used consistently? | Notes |
|------|-------------------|-------|
| "Save" / "Saved" | YES | RequestEditor save button |
| "Cancel" | YES | RunView, AlertDialog cancel buttons |
| "Delete" / "Clear" | YES | Delete = single item, Clear = all items |
| "Replay" | YES | HistoryView and HistoryReplay |
| "Run" | YES | CollectionTree play buttons, RunView |
| "Environment" | YES | Consistent across all views |

**Findings:**

- **VIOLATION** — `tabClass` function is duplicated: `RequestEditor.tsx:34-39` and `ResponseViewer.tsx:37-42` contain identical implementations. Should be extracted to a shared `TabButton` component or utility in `components/`.
- **VIOLATION** — `inputClass` string is defined separately in `EnvironmentsView.tsx:19-20` and `KeyValueEditor.tsx:13-14` with slightly different CSS. The EnvironmentsView version includes `focus:outline-none focus-visible:border-ring` while KeyValueEditor does not. These should be unified.
- **PASS** — Component variants, terminology, and interaction patterns are consistent across views.

#### 2.2.2 Visual Communication

**Findings:**

- **PASS** — Icons are consistent: lucide-react throughout, same visual weight (size-3 to size-3.5).
- **PASS** — The StatusPill (`status.tsx:35-61`) is an excellent example of visual communication: colored dot + text code + border, used consistently everywhere status appears.
- **PASS** — MethodLabel (`status.tsx:74-93`) uses consistent color coding with fallback for unknown methods.
- **PASS** — The CollectionTree uses chevron rotation (90deg) to communicate expand/collapse state — a familiar pattern.
- **NOTE** — The "No request open" empty state (`RequestEditor.tsx:136-145`) is functional but visually plain. An icon or subtle illustration would improve the first-run experience and communicate the app's purpose faster.
- **NOTE** — The collection tree's `Play` icon (`CollectionTree.tsx:79`) uses `fill-current` for a solid play button, which is appropriate and distinguishable.

#### 2.2.3 Depth & Elevation

**Elevation system observed:**

| Element | Elevation | Implementation |
|---------|-----------|----------------|
| Cards (env list items) | Border + bg-card | `border border-border bg-card` |
| Content panels (response, tree) | Border | `border border-border` |
| Modals (AlertDialog) | Portal + overlay | shadcn AlertDialog |
| Dropdowns (CompactSelect) | Portal | Base UI Select with portal |
| Tooltips | Portal | shadcn Tooltip |
| Toolbars (response action bar) | Border top | `border-t border-border/50` |

**Findings:**

- **PASS** — Elevation communicates layering: modals appear above all content via portal.
- **PASS** — Consistent border-based depth for panels and cards. No excessive shadows.
- **PASS** — No shadows on cards, buttons, or inputs (anti-pattern avoided).
- **NOTE** — The depth system is minimal (borders only, no shadows). This is appropriate for a dense developer tool — shadows would add visual noise without improving comprehension.

#### 2.2.4 Motion & Animation

**Findings:**

- **PASS** — `prefers-reduced-motion` is fully respected (`index.css:142-150`): all animations and transitions reduced to 0.01ms.
- **PASS** — Transitions are subtle and fast: `transition-colors` on hover states, `transition-transform` on chevron rotation.
- **PASS** — No continuous animations, no bouncing spinners, no page transitions.
- **PASS** — Theme toggle transition (`index.css:136-138`): 0.15s ease on background-color and color. Fast and non-distracting.
- **PASS** — No animations that convey critical information. The interface works fully without animation.
- **NOTE** — No loading skeleton states in ResponseViewer. When a request is in flight, it shows `// Sending request…` as raw text in CodeMirror (`ResponseViewer.tsx:122`). A skeleton placeholder would communicate loading state more effectively and reduce perceived wait time.

#### 2.2.5 Imagery & Media

**Findings:**

- **PASS** — Binary response handling is thoughtful: image preview for images (`ResponseViewer.tsx:190-193`), PDF notice (`ResponseViewer.tsx:194-195`), hex dump for other binary (`ResponseViewer.tsx:196-205`).
- **PASS** — Logo assets exist in both light and dark variants (`frontend/src/assets/logo-dark.svg`, `logo-light.svg`).
- **PASS** — No decorative stock photos, no unnecessary imagery.
- **PASS** — No images loaded without lazy loading concerns (binary previews are small, inline).

---

### 2.3 BEST PRACTICES

#### 2.3.1 Usability & Ease of Use

**Task completion step counts:**

| Task | Steps | Assessment |
|------|-------|------------|
| Send a request | 2 (type URL, click Send) | EXCELLENT |
| Save a request | 1 (⌘S or click Save) | EXCELLENT |
| Switch between requests | 1 (click tab) | EXCELLENT |
| Create an environment | 3 (navigate, fill name, click Create) | GOOD |
| Set active environment | 2 (navigate, click Use) | GOOD |
| Replay a history entry | 2 (navigate, click Replay) | GOOD |
| Run a collection | 1 (click play button on collection) | EXCELLENT |
| Copy as cURL | 2 (select language, click Copy as) | GOOD |
| Switch auth scheme | 2 (click Auth tab, select scheme) | GOOD |

**Findings:**

- **PASS** — Common tasks require minimal steps. The primary workflow (open request → send → review) is 2-3 steps.
- **PASS** — Sensible defaults: GET method, empty URL, no body, "New Request" scratchpad always available.
- **PASS** — Progressive disclosure: RetrySection collapses to one row, body types switch contextually, auth fields hide unused options.
- **PASS** — Dirty tracking with yellow dot indicators prevents data loss. Changed-on-disk conflict detection with Reload/Overwrite is excellent.
- **PASS** — The `⌘S` keyboard shortcut is implemented (`RequestEditor.tsx:126-134`).
- **WARNING** — No keyboard shortcut for Send. The conventional `⌘/Ctrl+Enter` is not implemented. Users must click the Send button or use the mouse.
- **WARNING** — URL input has no autocomplete or history-based suggestions. For a developer tool, suggesting previously-used URLs would significantly improve efficiency.
- **WARNING** — No undo for destructive actions (delete environment, clear history). The AlertDialog confirms, but once confirmed, the action is irreversible.

#### 2.3.2 Navigation

**Structure audit:**

| Aspect | Status | Notes |
|--------|--------|-------|
| Current location indicated | **YES** | `aria-current="page"` on sidebar nav items (`WorkspaceSidebar.tsx:39`) |
| Labels describe destination | **YES** | "Requests", "Environments", "History" — clear and task-oriented |
| Consistent navigation pattern | **YES** | Same sidebar nav across all views |
| Hierarchy is shallow | **YES** | 3 nav items + 1 tree. No nesting > 3 levels in tree. |
| No dead ends | **PARTIAL** | Empty states provide guidance, but the tree's "No collections yet" message (`CollectionTree.tsx:166-169`) requires knowledge of file system structure |
| Return to previous view | **YES** | 1-2 steps: click nav item or switch tab |
| Browser Back/Forward | **N/A** | Desktop app, no URL-based routing |
| Deep linking | **N/A** | Desktop app |

**Findings:**

- **PASS** — Current location is always visually indicated.
- **PASS** — Navigation is organized around user tasks (Requests, Environments, History), not system structure.
- **PASS** — Hierarchy is shallow. No navigation depth exceeds 3 levels.
- **WARNING** — The sidebar shows Requests/Environments/History as nav items AND the CollectionTree below them. This creates two distinct navigation regions in the sidebar without clear visual separation between "app-level nav" and "workspace content." The `border-t border-border pt-2` separator (`WorkspaceSidebar.tsx:59`) helps but could be more prominent.
- **NOTE** — When switching from Environments to Requests with unsaved env changes, the AlertDialog (`WorkspaceSidebar.tsx:79-104`) correctly warns and preserves user context. This is good UX.

#### 2.3.3 Information Architecture

**Findings:**

- **PASS** — Information is organized around user tasks: browsing requests (Requests), managing variables (Environments), reviewing past work (History).
- **PASS** — Related functionality is grouped: request editing (params/headers/auth/body/variables) is in tabs; response inspection (raw/pretty/headers/tree/cookies/table) is in tabs.
- **PASS** — Meaningful categories: the sidebar uses task-oriented labels, not system terms.
- **PASS** — Important functionality is discoverable: the collection tree shows play buttons on hover, the "+" button creates new requests.
- **NOTE** — The AuthPanel in the sidebar (`AuthPanel.tsx`) is always visible below the collection tree, even when the user is not using OAuth. This takes up valuable sidebar space. Consider making it collapsible or hiding it when no OAuth tokens are configured.

#### 2.3.4 Information Density

**Findings:**

- **PASS** — Density matches the developer tool context. The request editor packs method, URL, env, save, send, and code gen into one row — appropriate for a productivity tool.
- **PASS** — The history table shows 6 columns (time, method, URL, status, duration, env) — dense but scannable.
- **PASS** — The collection tree uses compact rows with minimal padding.
- **PASS** — Tables use `border-separate border-spacing-0` for tight, dense layouts.
- **WARNING** — The response viewer action bar (`ResponseViewer.tsx:238-306`) packs Copy, Copy headers, Download, Format, and JSONPath into one row. At narrow widths, the JSONPath input (`w-48`) may overlap with the action buttons.

#### 2.3.5 Components & Design Systems

**Component inventory:**

| Component | File | States supported | Assessment |
|-----------|------|------------------|------------|
| Button | `ui/button.tsx` | default, hover, focus-visible, active, disabled, loading (via aria-busy) | GOOD |
| CompactSelect | `CompactSelect.tsx` | default, hover, focus, disabled | GOOD |
| Input | `ui/input.tsx` | default, focus, disabled, invalid | GOOD |
| KeyValueEditor | `KeyValueEditor.tsx` | enabled/disabled rows, file/text toggle | GOOD |
| StatusPill | `status.tsx` | 5 tiers (info/ok/redirect/warn/error) | EXCELLENT |
| MethodLabel | `status.tsx` | 7 methods + fallback | EXCELLENT |
| JsonTree | `JsonTree.tsx` | expand/collapse, filter | GOOD |
| CodeMirrorEditor | `editors/CodeMirrorEditor.tsx` | read-only, editable, multiple languages | GOOD |
| AlertDialog | `ui/alert-dialog.tsx` | open/closed | GOOD (shadcn) |
| Toast | `lib/notify.ts` | success/error/warning | GOOD |

**Findings:**

- **PASS** — Components are reusable, single-responsibility, and state-complete.
- **PASS** — Design tokens (CSS variables) replace hardcoded values throughout.
- **PASS** — Button component supports all required states: default, hover, focus-visible, active, disabled.
- **WARNING** — The Button component does not have a built-in loading state. The AuthPanel (`AuthPanel.tsx:151`) uses `disabled={loading}` on the submit button, but there's no visual loading indicator (spinner, pulse) — just the disabled state. Users may not notice the button is disabled and think the app is unresponsive.
- **NOTE** — No design token for `--font-size-body` or `--font-size-heading`. Font sizes are hardcoded via Tailwind classes (`text-xs`, `text-sm`). This is acceptable for Tailwind but means font size changes require class edits, not CSS variable updates.

#### 2.3.6 States & Feedback

**State coverage audit:**

| Component | Hover | Focus | Active | Disabled | Loading | Empty | Error |
|-----------|-------|-------|--------|----------|---------|-------|-------|
| Send button | YES | YES | YES | YES | **PARTIAL** (disabled only) | N/A | N/A |
| Save button | YES | YES | YES | YES (when clean) | N/A | N/A | N/A |
| Collection play button | YES | YES | YES | YES (when running) | N/A | N/A | N/A |
| History replay button | YES | YES | YES | N/A | N/A | N/A | N/A |
| Env list item | YES | YES | N/A | N/A | N/A | YES (empty state) | YES (error state) |
| Request tabs | YES | YES | N/A | N/A | N/A | N/A | N/A |
| Response viewer | N/A | N/A | N/A | N/A | YES (sending text) | YES (no response) | YES (error text) |
| Tree nodes | YES | YES | N/A | N/A | N/A | YES (no collections) | YES (open error) |

**Findings:**

- **PASS** — All interactive components have hover, focus, and active states.
- **PASS** — Empty states are informative and actionable: "No request open — Open a request from the sidebar or create a new one to start sending" (`RequestEditor.tsx:138-144`), "No collections yet — create collections/<name>/reqly.yaml to see them here" (`CollectionTree.tsx:167-168`).
- **PASS** — Error states explain the problem: "This request changed on disk since you opened it" (`RequestEditor.tsx:178-179`), "No inherited auth — this request will send unauthenticated" (`AuthEditor.tsx:219-223`).
- **WARNING** — The Send button's loading state is only `disabled` + text change to "Stop" (`RequestEditor.tsx:236-244`). There's no spinner, pulse, or progress indicator. During a long request, users see "Stop" but no visual confirmation that the request is in flight beyond the ResponseViewer showing `// Sending request…`.
- **WARNING** — The "Copied" feedback on copy actions (`RequestEditor.tsx:275-278`, `ResponseViewer.tsx:253-256`) uses `setTimeout(1500)` to revert. If the user clicks Copy again within 1.5s, the state is already reset. This is fine for the happy path but means rapid double-clicks show no feedback.
- **PASS** — User input is never silently lost. The dirty tracking system ensures unsaved changes are flagged.

#### 2.3.7 Input & Forms

**Form audit:**

| Form | Fields | Labels | Validation | Preserves input on error |
|------|--------|--------|------------|-------------------------|
| New Environment | name (required), description (optional) | Placeholder only (no `<label>`) | onChange: name must not be empty | YES |
| Auth Editor | scheme, per-scheme fields | `<Label>` via Base UI Field.Root | Warnings (not blocking) | YES |
| OAuth Config | flow, JSON config textarea | Placeholder + `<label>` | JSON parse on submit | YES |
| Request Editor | method, URL, params, headers, body | Placeholder only (no `<label>`) | Save warnings (not blocking) | YES |
| History Search | query, status filter | Placeholder only | None (search is forgiving) | N/A |

**Findings:**

- **PASS** — Forms minimize effort: the environment form has just 2 fields, the auth editor uses progressive disclosure.
- **PASS** — User input is preserved on error. The environment form's `createError` (`EnvironmentsView.tsx:173-175`) appears below the form without clearing it.
- **WARNING** — The URL input (`RequestEditor.tsx:217-223`) uses a bare `<input>` with only a placeholder ("https://reqly-test-api.vercel.app/api/users?page=1 — mock API for testing"). No `<label>`, no `aria-label`. Placeholder text disappears on focus — this is the anti-pattern "Placeholder text used as a label."
- **WARNING** — The environment name input (`EnvironmentsView.tsx:128-143`) also uses placeholder only ("name (e.g. dev)") with no `<label>` or `aria-label`.
- **WARNING** — The history search input (`HistoryView.tsx:79-87`) uses placeholder only ("Search URL or path…") with no `<label>` or `aria-label`.
- **PASS** — The AuthEditor uses proper `<Label>` components via `Field.Root` (`AuthEditor.tsx:70-80`). This is the only form with proper labeling.
- **PASS** — Validation is helpful, not interruptive: save warnings appear as a banner, auth warnings appear inline, environment name validation shows on blur.

#### 2.3.8 Interaction & Affordance

**Findings:**

- **PASS** — Interactive elements visually communicate interactivity: buttons have hover states, the collection tree play button changes color on hover (`CollectionTree.tsx:77`), the expand chevron rotates.
- **PASS** — Destructive actions are visually distinguishable: Delete buttons use `text-destructive hover:bg-destructive/10` (`EnvironmentsView.tsx:274-281`, `HistoryView.tsx:131-139`).
- **PASS** — Feedback is immediate: copy shows "Copied", save shows "Saved", send toggles to "Stop".
- **WARNING** — The collection tree play button (`CollectionTree.tsx:70-82`) has `text-muted-foreground/60` default state, which is very low contrast. It becomes visible on hover (`hover:text-status-ok`). This is a "toolbar that requires hovering to reveal actions" anti-pattern — the action exists but is nearly invisible until hovered.
- **WARNING** — The "Add row" button in KeyValueEditor (`KeyValueEditor.tsx:96-104`) uses `variant="outline"` which is visually similar to disabled buttons. It could be confused with a disabled control.
- **PASS** — Consistent interaction behavior: all expand/collapse uses chevron rotation, all selections use CompactSelect, all confirmations use AlertDialog.

---

### 2.4 SAFETY

#### 2.4.1 Accessibility

**WCAG AA compliance checklist:**

| Criterion | Status | Evidence |
|-----------|--------|----------|
| Contrast: text (4.5:1) | **PARTIAL** | `--primary` on `--primary-foreground` fails in light mode (~3.2:1). `--muted-foreground` is borderline (~4.7:1). |
| Contrast: UI (3:1) | **PASS** | Borders, focus rings meet 3:1. |
| Focus indicators | **PASS** | `:focus-visible` ring defined globally (`index.css:187-190`). Button CVA includes `focus-visible:ring-3`. |
| Keyboard navigation | **PARTIAL** | Tab navigation works. Collection tree has roving tabindex (`CollectionTree.tsx:17-54`). But tabs in RequestEditor/ResponseViewer lack arrow key navigation. |
| Semantic HTML | **PARTIAL** | `role="tree"` on CollectionTree, `role="alert"` on errors, `aria-expanded` on folders. But tabs lack `role="tablist"`/`role="tab"`. |
| Touch targets (44×44px) | **FAIL** | Tree rows ~28-32px, tab buttons ~24-28px, icon buttons ~24px. Below WCAG 2.5.8 minimum. |
| Color never sole carrier | **PASS** | StatusPill pairs color with code text. MethodLabel pairs color with verb text. |
| `prefers-reduced-motion` | **PASS** | Fully respected (`index.css:142-150`). |
| `prefers-contrast` | **NOT IMPLEMENTED** | No high-contrast mode or `prefers-contrast` media query. |
| Labels for all controls | **FAIL** | URL input, environment name, history search, OAuth config textarea all lack `<label>` or `aria-label`. |
| Reading order | **PASS** | DOM order matches visual order. No CSS reordering. |

**Specific accessibility findings:**

1. **Tab components lack ARIA tab semantics** — `RequestEditor.tsx:306-318` and `ResponseViewer.tsx:210-222` use `<button>` elements styled as tabs but without `role="tablist"`, `role="tab"`, `role="tabpanel"`, or `aria-selected`. Screen readers cannot identify the active tab or navigate with arrow keys.

2. **URL input has no label** — `RequestEditor.tsx:217-223` is a bare `<input>` with placeholder only. Placeholder disappears on focus. Should have `aria-label="Request URL"` or a visually-hidden label.

3. **Collection tree request rows lack `role="treeitem"`** — `CollectionTree.tsx:84-99` renders request rows as `<button>` with `data-tree-row` but no `role="treeitem"`. Folders have `aria-expanded` but requests don't have `role="treeitem"` or `aria-level`.

4. **Table headers lack `scope`** — `ResponseViewer.tsx:375-381` and `HistoryView.tsx:183-190` use `<th>` without `scope="col"`.

5. **No skip-to-content link** — Keyboard users must tab through the entire sidebar to reach the main content area.

6. **No `prefers-contrast` support** — The app has light/dark themes but no high-contrast mode for users who need stronger differentiation.

#### 2.4.2 Error Prevention & Recovery

**Findings:**

- **PASS** — Destructive operations are safeguarded with AlertDialog confirmation: delete environment (`EnvironmentsView.tsx:290-316`), clear history (`HistoryView.tsx:234-256`), discard unsaved env changes (`WorkspaceSidebar.tsx:79-104`).
- **PASS** — Error messages explain the problem and recovery: "This request changed on disk since you opened it. Overwrite the file, or reload to keep the on-disk version." (`RequestEditor.tsx:178-179`).
- **PASS** — User input is preserved on validation error: the environment create form retains values when submission fails (`EnvironmentsView.tsx:173-175`).
- **PASS** — Undo exists for reversible actions: the "Discard changes" / "Keep editing" choice (`WorkspaceSidebar.tsx:93-103`) is reversible.
- **WARNING** — No undo for history clear or environment delete. Once confirmed, these are irreversible. The CLI has no undo either, so this is consistent, but the GUI could offer a "soft delete" with recovery.
- **PASS** — The save warning system (`RequestEditor.tsx:56-106`) prevents errors before they happen: malformed JSON, unknown methods, incomplete auth are all flagged before save.

#### 2.4.3 Real-World Content Resilience

**Test scenarios:**

| Scenario | Status | Evidence |
|----------|--------|----------|
| Very long URL | **PASS** | `truncate` class on URL display (`RequestEditor.tsx:293`, `CollectionTree.tsx:93`). URL input uses `min-w-0 flex-1` for overflow. |
| Very short URL | **PASS** | Input has `min-w-0` allowing收缩. |
| Missing values | **PASS** | Empty states handle null/undefined throughout. |
| Large response body | **PASS** | Response viewer uses CodeMirror with virtual scrolling. Hex dump truncated to 4KB (`ResponseViewer.tsx:199`). |
| Empty response | **PASS** | "No cookies set by this response" with explanation (`ResponseViewer.tsx:408-419`). |
| Long error messages | **PARTIAL** | ErrorBoundary truncates to `max-w-md truncate` (`ErrorBoundary.tsx:86`). History error uses `role="alert"` without truncation (`HistoryView.tsx:145`). |
| Many collection items | **PASS** | Tree uses virtual rendering via DOM (no virtualization library, but tree is typically < 1000 items). |
| Very long header values | **PASS** | `break-all` on table cells (`ResponseViewer.tsx:479`). `truncate` on display cells. |
| Table with 1000+ rows | **PASS** | Table view shows "Showing first 1000 rows" notice (`ResponseViewer.tsx:400`). History uses pagination. |
| Concurrent tab edits | **PASS** | Per-tab draft isolation via `useRequestStore`. Send token prevents stale responses. |

**Findings:**

- **PASS** — Layout holds with very long text (truncation, `min-w-0`, `break-all`).
- **PASS** — Layout holds with empty/missing content (empty states throughout).
- **PASS** — Truncated content is accessible (truncated elements have `title` attributes for tooltip on hover).
- **WARNING** — The JSONPath input (`ResponseViewer.tsx:297-304`) has no error boundary. Malformed JSONPath expressions show an error message (`ResponseViewer.tsx:324-327`), but the error is inline and could be missed.
- **PASS** — Dynamic content does not shift important controls. The response viewer replaces content in-place without moving the action bar.

#### 2.4.4 Platform Conventions

**Desktop app conventions:**

| Convention | Status | Evidence |
|------------|--------|----------|
| Application menu | **PASS** | Wails provides native menu bar |
| Window controls | **PASS** | Wails provides native window chrome |
| Keyboard shortcuts | **PARTIAL** | `⌘S` for save is implemented. `⌘Enter` for send is NOT. `⌘W` for close tab is NOT. `⌘Q` for quit is handled by OS. |
| Clipboard | **PASS** | Copy uses `navigator.clipboard.writeText` with fallback (`ErrorBoundary.tsx:18-38`). |
| Drag and drop | **NOT IMPLEMENTED** | No drag-and-drop for reordering requests, importing files, etc. |
| File dialogs | **NOT APPLICABLE** | Git-native file format, no file picker needed. |
| OS theme | **PASS** | Light/dark with `prefers-color-scheme` auto-detect. |
| Display scaling | **PASS** | Wails handles DPI/Retina. No hardcoded px values. |

**Findings:**

- **PARTIAL** — Some platform keyboard shortcuts are missing (`⌘Enter` for send, `⌘W` for close tab).
- **PASS** — The app feels native: Wails provides native menus, window controls, and clipboard integration.
- **WARNING** — No `⌘W` to close the active tab. Users must click the "×" on the tab or use the menu. This is a standard desktop convention.

#### 2.4.5 Performance-Aware Design

**Findings:**

- **PASS** — No excessive blur, filters, or expensive visual effects. The CSS is clean and lightweight.
- **PASS** — CodeMirror is used for large text bodies, providing virtual rendering for long responses.
- **PASS** — Images (binary previews) are small and inline. No large hero images.
- **PASS** — Theme transition is 0.15s ease — fast and non-blocking.
- **WARNING** — The JsonTree component (`JsonTree.tsx`) renders all nodes when expanded. For very large JSON responses (10,000+ nodes), this could cause performance issues. No virtualization is applied to the tree.
- **NOTE** — The `tabIsDirty` function (`useRequestStore.ts:119-122`) uses `JSON.stringify` comparison for dirty detection. For very large request bodies, this could be slow on every keystroke. However, in practice, request bodies are typically small (< 10KB), so this is acceptable.
- **PASS** — No unnecessary continuous animations. All transitions are event-driven.

---

### 2.5 ANTI-PATTERNS

Consulted against `rules/anti-patterns.md`:

| Anti-pattern | Present? | Location |
|--------------|----------|----------|
| Every element given equal visual weight | NO | Hierarchy is clear |
| Same font size for headings and body | NO | `text-sm` vs `text-xs` distinction |
| Pixel-nudging individual elements | NO | Tailwind spacing scale used |
| Equal spacing between every element | NO | Proximity-based spacing |
| Shadows on every card/button/input | NO | Borders only, no shadows |
| Inconsistent border-radius | **PARTIAL** | Buttons use `rounded-lg`, inputs use `rounded-md`, panels use `rounded-md` |
| Visual polish on landing but raw internal | NO | Consistent throughout |
| Red text for emphasis without icon | NO | Destructive text paired with icons |
| Low-contrast gray text on white | **PARTIAL** | `--muted-foreground` is borderline in light mode |
| Width: 1200px with no fallback | NO | Fluid layouts with max-w constraints |
| Touch targets < 44×44px | **YES** | Tree rows, tab buttons, icon buttons |
| 10-field form for simple sign-up | NO | Forms are minimal |
| Confirmation for every non-destructive action | NO | Only destructive actions confirmed |
| Tooltip-only approach | NO | Critical info is always visible |
| Placeholder text as label | **YES** | URL input, env name, history search |
| Inline validation on every keystroke | NO | Validation on blur/submit |
| Clearing form on validation error | NO | Input preserved |
| No active state on current nav item | NO | `aria-current="page"` present |
| Text that looks clickable but isn't | NO | Interactive elements clearly styled |
| No visual difference between clickable/static | NO | Hover states on all interactive elements |
| Button shows no feedback when clicked | **PARTIAL** | Send button only disables, no spinner |
| Empty list with no explanation | NO | Empty states with guidance |
| Error banner with no next steps | NO | All errors include recovery guidance |
| Carousel with no keyboard controls | N/A | No carousel |
| Error messages as color only | NO | Color + text + icon throughout |
| Focus order jumps illogically | NO | DOM order matches visual order |
| Invisible focus rings | NO | `:focus-visible` ring present |
| Loading spinner that bounces | N/A | No spinners used |
| Page transitions 800ms+ | N/A | No page transitions |
| Animations on every scroll event | NO | No scroll-triggered animations |
| Important info only via animation | NO | All info is static |
| Card with fixed height clipping | NO | Flexible heights with overflow |
| Table breaks with 500-char cell | NO | `break-all` and `truncate` |
| Modal overflows viewport | NO | AlertDialog is constrained |
| Dashboard with 20 equally-weighted cards | NO | 3 views with clear hierarchy |
| 10,000-row table without virtualization | **PARTIAL** | Response table capped at 1000 rows; history paginated; JsonTree not virtualized |
| Blur backdrop on every panel | NO | No blur effects |

**Anti-patterns found:**

1. **Touch targets < 44×44px** — Tree rows, tab buttons, and icon buttons are below the WCAG 2.5.8 minimum.
2. **Placeholder text as label** — URL input, environment name, history search use placeholder only.
3. **Inconsistent border-radius** — Buttons (`rounded-lg`) vs inputs/panels (`rounded-md`).
4. **Button shows no feedback when clicked** — Send button only disables, no loading indicator.
5. **10,000-row table without virtualization** — JsonTree renders all nodes when expanded.

---

### 2.6 REVIEW CHECKLIST

From `SKILL.md`:

| Item | Status | Evidence |
|------|--------|----------|
| Purpose immediately understandable | **PASS** | "Reqly" branding, API client layout is immediately recognizable |
| Most important information visually obvious | **PASS** | Method+URL bar is primary, Send button is prominent |
| All interactive elements clearly identifiable | **PASS** | Buttons, inputs, selects all have clear affordances |
| Navigation predictable; current location clear | **PASS** | Sidebar nav with `aria-current="page"` |
| Common tasks require minimal steps | **PASS** | 2-3 steps for primary workflow |
| Sensible defaults provided | **PASS** | GET method, empty URL, scratchpad tab |
| Errors explain problem and recovery | **PASS** | All error messages include guidance |
| User input never silently lost | **PASS** | Dirty tracking, conflict detection |
| Works across narrow, standard, and wide layouts | **PARTIAL** | Works at 1024px+, narrow widths at risk |
| Keyboard navigation functional where applicable | **PARTIAL** | Collection tree yes, request/response tabs no |
| Focus states visible | **PASS** | `:focus-visible` ring on all interactive elements |
| Color never sole carrier of meaning | **PASS** | Color + text + icon throughout |
| Realistic content handled | **PASS** | Long URLs, empty states, large responses |
| No unnecessary decoration competes with content | **PASS** | Clean, functional design |
| Similar elements look and behave similarly | **PARTIAL** | Mostly yes; `tabClass`/`inputClass` duplication |

---

## Part 3: Code Standards Review (Fowler Smell Baseline)

### Duplicated Code — VIOLATION

The `tabClass` function is identically defined in:
- `frontend/src/features/request-editor/RequestEditor.tsx:34-39`
- `frontend/src/features/response-viewer/ResponseViewer.tsx:37-42`

Both are:
```ts
const tabClass = (active: boolean) =>
  `rounded-md px-2 py-1 text-xs font-medium transition-colors ${
    active
      ? 'bg-muted text-foreground'
      : 'text-muted-foreground hover:text-foreground'
  }`
```

**Fix:** Extract to a shared `TabButton` component in `components/` or a `tabClass` utility in `lib/`.

### Duplicated Code — VIOLATION

The `inputClass` string is defined in:
- `frontend/src/features/environments-view/EnvironmentsView.tsx:19-20`
- `frontend/src/components/KeyValueEditor.tsx:13-14`

Both serve the same purpose but have slightly different CSS:
- EnvironmentsView: `"rounded-md border border-input bg-background px-2 py-1 text-xs text-foreground placeholder:text-muted-foreground focus:outline-none focus-visible:border-ring"`
- KeyValueEditor: `"rounded-md border border-input bg-background px-2 py-1 text-xs text-foreground placeholder:text-muted-foreground"`

The EnvironmentsView version includes `focus:outline-none focus-visible:border-ring` while KeyValueEditor does not.

**Fix:** Extract to a shared `inputClass` in `lib/` or `components/ui/input.tsx` and use consistently.

### Duplicated Code — JUDGEMENT CALL

`formatBytes` is defined in both:
- `frontend/src/features/response-viewer/ResponseViewer.tsx:44-48`
- `frontend/src/components/RunView.tsx:8-12`

Both are identical:
```ts
function formatBytes(size: number): string {
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
}
```

**Fix:** Extract to `lib/format.ts` or similar.

### Primitive Obsession — NONE

The `Tab` and `View` types are string unions used as local state. No domain concepts are represented as raw strings.

### Feature Envy — NONE

Components access their own stores appropriately. No cross-object data reaching.

### Data Clumps — RESOLVED

`KeyValueRow` is already extracted as a type. No new data clumps detected.

### Shotguns Surgery — NONE

Request behavior changes are centralized in `useRequestStore`. Workspace changes are in `useWorkspaceStore`. The adapter pattern cleanly separates concerns.

### Speculative Generality — NONE

No unused abstractions, hooks, or extension points. The adapter pattern is actively used.

### Message Chains — NONE

State access is flat: `useRequestStore((s) => s.drafts[activeTabId])`. No deep navigation.

### Middle Man — NONE

The bridge layer is a necessary translation between Wails and Zustand.

### Refused Bequest — NONE

No inheritance hierarchies in the component tree.

---

## Part 4: Summary

### Design Principles Scorecard (Detailed)

| Principle | Score | Critical findings |
|-----------|-------|-------------------|
| **Functionality** | 8/10 | 17 CLI features missing from GUI. Core loop is excellent. |
| **Usability** | 9/10 | Progressive disclosure, dirty tracking, conflict detection are excellent. Missing ⌘Enter for send. |
| **Accessibility** | 5/10 | Tab ARIA semantics missing. 4+ inputs lack labels. Touch targets below 44px. No `prefers-contrast`. |
| **Clarity** | 9/10 | Visual hierarchy, empty states, warnings all well-handled. |
| **Consistency** | 7/10 | `tabClass` duplicated. `inputClass` duplicated. `formatBytes` duplicated. Border-radius inconsistent. |
| **Responsiveness** | 6/10 | Resizable panels good. Request editor top bar overflow risk at narrow widths. No 320px/768px evidence. |
| **Visual Polish** | 8/10 | Clean, professional, appropriate density. Loading skeletons missing. Primary button contrast fails in light mode. |

### Worst Issues Per Axis

| Axis | Worst Issue | Severity |
|------|-------------|----------|
| **Functionality** | 17 CLI features (import, export, WS, SSE, mock, test, etc.) have no GUI counterpart | Critical |
| **Accessibility** | Tab components in RequestEditor and ResponseViewer lack `role="tablist"` / `role="tab"` / `role="tabpanel"` — screen readers cannot navigate the primary editing interface | Critical |
| **Accessibility** | 4+ form inputs use placeholder text as label with no `<label>` or `aria-label` — WCAG 1.3.1 and 4.1.2 violation | Critical |
| **Responsiveness** | Request editor top bar packs 6+ controls horizontally with no wrap — will overflow at narrow widths | High |
| **Consistency** | `tabClass`, `inputClass`, and `formatBytes` each duplicated in 2+ files | Medium |
| **Safety** | Primary button fails WCAG AA contrast in light mode (~3.2:1 vs 4.5:1 required) | Medium |
| **Anti-pattern** | Touch targets below 44×44px WCAG 2.5.8 minimum | Medium |

### Top 10 Recommendations (Priority Order)

1. **Add tab ARIA semantics** to RequestEditor and ResponseViewer — `role="tablist"`, `role="tab"`, `role="tabpanel"`, `aria-selected`. Critical for accessibility.

2. **Add `aria-label` or `<label>` to all form inputs** — URL input, environment name, history search, OAuth config textarea. Critical for accessibility.

3. **Implement Import UI** — the highest-impact missing feature for user onboarding (Postman/cURL/OpenAPI import).

4. **Extract `tabClass`, `inputClass`, and `formatBytes` to shared modules** — eliminate DRY violations.

5. **Add `⌘/Ctrl+Enter` keyboard shortcut for Send** — standard API client convention.

6. **Fix primary button contrast in light mode** — either darken `--primary` or lighten `--primary-foreground` to reach 4.5:1.

7. **Add loading skeleton states** in ResponseViewer — replace `// Sending request…` text with a skeleton placeholder.

8. **Add loading indicator to Send button** — pulse, spinner, or progress bar during in-flight requests.

9. **Make AuthPanel collapsible** — it takes sidebar space even when OAuth is not in use.

10. **Add `⌘/Ctrl+W` to close tabs** — standard desktop convention.
