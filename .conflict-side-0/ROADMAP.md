# Reqly — UI Redesign Roadmap (slice-by-slice)

> **Source:** `docs/internal:DESIGN.md` (precise·dense·calm·engineered, “HTTP made legible”, StatusPill signature) + `docs/internal/gui-roadmap.md` GUI-0→GUI-5 + Complete UI Architecture §1–59 (quarantined on `docs/internal` — `git worktree add /tmp/docs-internal docs/internal`).
> **Status:** Core P0–P5 (`internal/`, `lib/` data layer, CLI) shipped on `main` `ae70f07a` and preserved (`lib/`+`stores/`+`internal/`). **UI is to be rebuilt from scratch** — this doc is the **new canonical UI roadmap** for the `frontend/` rewrite. All UI milestones below are `[ ]` not started for the redesign (prior `frontend/src/` chrome is discarded; only data layer is kept).
> **Archive:** `Milestones/` & old `ROADMAP.md` archived to Git history `f320d724` (P0–P5 100%, 13 [~] deferrals, slices 01–08). New UI slices start at `UI-01` below.

Checkboxes = shipped **for the redesign** (component + store + a11y + tests) per `docs/testing-strategy.md` DoD. `[x]` shipped in redesign, `[~]` partial, `[ ]` not started (all UI now `[ ]` — from scratch).

## Design direction (grounded — per `frontend-design` skill)

**Subject:** Git-native, local-first API client for developers who review requests in PRs. Page job: make HTTP legible at density without decoration.

**Palette (4+2, not cream/terracotta default):** `background #fbfbfa` / `foreground #191c21` (light), `background #0d1015` / `foreground #e8eaed` (dark) — near-neutral surfaces (~60%); `primary #c93517` (light, AA 4.5:1) / `#ff6f52` (dark) used ~10% (primary button, `ring`, active accent); **status ramp** GitHub-derived — `status-ok #1a7f37`/`#3fb950` 2xx, `status-redirect #0969da`/`#58a6ff` 3xx, `status-warn #9a6700`/`#d29922` 4xx, `status-error #cf222e`/`#f85149` 5xx, `status-info #57606a`/`#8b949e` 1xx; method tints map to ramp (GET green, POST blue, PUT amber, DELETE red). No gradient text.

**Type:** Display/UI `IBM Plex Sans 400/500/600` (engineered, bundled), Data/code `IBM Plex Mono 400/500` + `tabular-nums` for every URL/status/duration/table; `13px/1.45` base, `.font-data` mono discipline. No Inter/Roboto default.

**Layout:** 5-zone shell `TopBar` / `ToolRail` (48–56px) / `ContextSidebar` (220–280px) / `MainWorkspace` / `BottomPanel` + `StatusBar` — resizable via `react-resizable-panels`, `min-w-0` everywhere, `useDefaultLayout` persistence (`shellStorage`).

**Signature:** `StatusPill` (`frontend/src/components/status.tsx:1` dot + tabular code) — single memorable device used identically in `ResponseViewer` header, `RunView` steps, `HistoryView` rows; everything else stays quiet.

**Risk:** Density over decoration — keep 13px base, hairline `border-border` only (no shadows except `popover`/`select`/`toast`), `prefers-reduced-motion` collapses all, `prefers-contrast: more` bumps borders. One orchestrated motion (theme `0.15s ease`) not scattered.

---

## Tech Stack — `go 1.27` + `Wails v3` + `Goja` + `React 19`

> **Source:** `docs/internal:docs/technology-stack.md` (Go + Wails + Goja, React + Vite) via `git show origin/docs/internal:docs/technology-stack.md`. Versions from `go.mod:3` `go 1.27` + `frontend/package.json:1` + `go list -m`.

| Layer | Tech | Version | Purpose |
|-------|------|---------|---------|
| **Desktop** | **Wails v3** | `v3.0.0-beta.16` | Native shell via system WebView (no bundled Chromium) — Linux `WebKit`, macOS `WebKit`, Windows `WebView2` |
| **Core** | **Go** | `1.27` (`go.mod:3`) | Single execution pipeline (`internal/` — request, auth, vars, history, `sqlite`, Git, CLI, MCP) |
| **Scripting** | **Goja** | `8f1c069` | Embedded JS for `pre/post` scripts, `reqly.test()`, dynamic `{{$tag}}`, plugins — lazy init |
| **Frontend** | **React + TypeScript** | `19.2.8` + `5.x` | Highly interactive UI — `Vite` dev/build, `Tailwind v4` + `shadcn/ui` + `Base UI 1.7` |
| **State** | **Zustand** | `5.0.15` | Lightweight stores (`useWorkspaceStore`, `useRequestStore`, `useThemeStore` etc) — one file per domain |
| **Editor** | **CodeMirror 6** | `6.43` | JSON/JS/XML/YAML/Markdown/Text + `graphql` — `one-dark` theme |
| **Styling** | **Tailwind CSS** | `v4` | `@theme` tokens (`--background`/`--primary`/`--status-*`/`--radius 6px`) + `@custom-variant dark` |
| **Storage** | **Plain-text files + SQLite** | — | Collections/envs as YAML/JSON on disk (Git-native) + `modernc.org/sqlite v1.57` WAL/FTS5 `history.db` + `blobs/` |
| **Secrets** | **OS Keychain** | `go-keyring 0.2.8` | `secrets.Store` `FileStore` (`0600 .reqly/tokens.json`) vs `KeychainStore` |
| **OpenAPI** | **kin-openapi** | `0.149.0` | `internal/openapi` 3.x parse/validate + `oasdiff/yaml` |
| **Protocols** | **gRPC + WebSocket/SSE/MQTT** | `grpc 1.83.2`, `websocket 1.8.15` | `internal/grpc` reflection + `protocompile 0.14.1`, `coder/websocket` |
| **Build** | **Vite + Wails** | `6.x` + `Taskfile` | `nub` workspaces (`pnpm-lock.yaml`), `GoReleaser` OS matrix (`release.yml`) |
| **Test** | **Go + Vitest + Playwright** | `go test -race`, `vitest 4.1` | `20` frontend files `185` tests, `46` Go pkgs |

**Why Go+Wails+Goja:** small footprint, fast startup, native FS/Git, direct Go↔JS, no Node/Chromium bundled — target `linux`/`darwin`/`windows` via `Taskfile` `wails3 build`.

---

## Design Principles — `design-principles` skill (priority order is absolute)

> **Source:** `~/.agents/skills/design-principles` (`skill: design-principles`) — `visual hierarchy`, `color`, `typography`, `spacing`, `layout`, `responsiveness`, `consistency`, `accessibility`, `usability`, `navigation`. Rules in `rules/principles.md`/`quality.md`/`best-practices.md`/`safety.md`/`anti-patterns.md`.

**Priority (highest → lowest):** 1. **Functionality** (core purpose) → 2. **Usability** (minimal friction) → 3. **Accessibility** (all users) → 4. **Clarity** (hierarchy/state readable) → 5. **Consistency** (patterns/spacing/type predictable) → 6. **Responsiveness** (adapts across sizes) → 7. **Visual Polish** (aesthetics never override above).

| Clash | Resolution |
|-------|------------|
| Dense data vs whitespace | Density for productivity; whitespace for marketing |
| Animation vs performance | Performance wins — motion only where it communicates (`Motion` `motion@13.1.1` for `StatusPill` pulse, panel `0.15s ease` only) |
| Simplicity vs discoverability | Start simple; progressive disclosure (`overflowTabs`, `More`, `Advanced` in `Settings`) |
| Platform conventions vs design system | Conventions for core interactions (`⌘K`/`⌘W`/`Enter`), design system for identity (`StatusPill`, `methodTint`) |
| Density vs cognitive load | Group + hierarchy (`ContextSidebar` per-tool nav, `BottomPanel` tabs, `StatusPill` dot+code) — never dump everything at once |
| Aesthetics vs clarity | Clarity wins — `13px` dense, hairline `border-border`, `prefers-reduced-motion`/`prefers-contrast: more` |

**Core principle:** Users must: understand purpose, identify what matters, find what they need, see available actions, complete tasks minimally, recover without losing work (`tabIsDirty`/`changedOnDisk`/`pendingView` guard), navigate confidently (`ToolRail` + `ContextSidebar` + `⌘1-8`), adapt across sizes (`min-w-0`, `ResizablePanel`), meet a11y, build familiarity via consistency. **Clarity, usability, accessibility, consistency, efficiency > visual novelty.**

**Review checklist — every UI milestone ships only when all pass:**

- [ ] Purpose immediately understandable · [ ] Most important info visually obvious · [ ] All interactive elements clearly identifiable
- [ ] Navigation predictable; current location clear (`activeView` + `aria-expanded`/`aria-selected`) · [ ] Common tasks minimal steps (`Send` `⌘⏎`, `Save` `⌘S`, `Duplicate` right-click)
- [ ] Sensible defaults (`GET`, `No environment`, `Default (follow)`) · [ ] Errors explain + recover (`changedOnDisk` `Reload`/`Overwrite`, `Save` `saveWarnings`, `Console` `Copy`) · [ ] Input never silently lost (`dirtyEditors`, `pendingView` confirm)
- [ ] Works narrow/standard/wide (`280→220px` sidebar, `48→56px` rail, `ResizablePanel` `200–42%`/`6–42%`) · [ ] Keyboard functional (`TREE_KEYS` `ArrowDown/Up/Home/End`, `handleTabArrowKeys`, `useKeyboardMap`) · [ ] Focus visible (`:focus-visible` `ring`, `roving tabindex`)
- [ ] Color never sole carrier (`StatusPill` dot+code `never color alone`, method tints + `methodTint` class) · [ ] Realistic content handled (long `requestPath` `truncate`, empty `Connect to an endpoint…`, `No commands match`) · [ ] No decoration competes (`hairline` borders, no shadows except `popover`/`select`/`toast`, `Motion` one orchestrated moment) · [ ] Similar elements look/behave similarly (`CompactSelect` everywhere, `AlertDialog` for destructive, `cn`/`cva` variants)

*This checklist is the `DoD` for every `UI-NN` tracer bullet below — `frontend-design` handles aesthetics, `design-principles` is the quality gate that can veto it.*

---

## Writing for Agents — how this roadmap is written (so agents don't guess)

> **Source:** `writing-for-agents` — docs that agents can execute without asking a human. Every `UI-NN` below is a **tracer bullet** with explicit `Files`, `Change`, `Accept`, `Verify`, `Depends` — no pronouns without antecedents, no “etc.”, no “as needed”.

**Agent contract — each ticket must have:**

- **Title:** `UI-NN.T — verb + object` (e.g., `UI-01.1 — Extract tokens.css`)
- **Scope:** one file or one concern (max 2 files, max 50 lines changed if possible)
- **Files:** absolute `file_path:line_number` to touch + to not touch (`lib/`+`stores/` preserved, `internal/` untouched)
- **Change:** imperative, present tense, file:line ref (e.g., `Create frontend/src/styles/tokens.css:1 with :root atlas-light/dark`)
- **Accept:** observable behavior (e.g., `npm run typecheck` green, `Settings → Appearance` shows 4 themes + System)
- **Verify:** exact commands (`nub run typecheck`, `npx oxlint`, `go vet ./...`, `vitest run frontend/src/...`)
- **Depends:** explicit `UI-NN` id or `none`

**Writing rules for agents (enforced):**

- Use `file_path:line_number` for every ref (e.g., `frontend/src/components/shell/AppShell.tsx:1`), not `the shell`
- Name things by what the user controls (`Duplicate request` not `WorkspaceDuplicateRequest`)
- No “etc.”, “and so on”, “as needed” — list every item (`Params`/`Headers`/`Body`/`Auth` + `pre-request`/`tests`/`docs`/`settings`/`variables`)
- Keep one job per element (`CompactSelect` everywhere, `AlertDialog` for destructive, `cn`/`cva` variants)
- Empty states are invitations, not apologies (`Connect to an endpoint…` not `No data`)
- Every `as` needs `// SAFETY:` per `oxlint: require-safety-comment-for-type-assertion` (`as const` exempt)

**Example — `UI-01` broken down (this is the pattern for all `UI-NN` below):**

- `UI-01.1 — Extract tokens.css` — **Files:** `frontend/src/styles/tokens.css:1` (new) + `frontend/src/index.css:7` (edit) — **Change:** move `:root`/`[data-theme]` blocks to `tokens.css`, keep `frontend/src/styles/tokens.css:1` as source — **Accept:** `Settings → Appearance` still shows 4 themes, `data-theme` still toggles — **Verify:** `nub run typecheck` + `grep -r "#c93517" frontend/src --include="*.tsx" | wc -l` == 0 (no hardcoded hex) — **Depends:** `none`
- `UI-01.2 — Create shellStorage` — **Files:** `frontend/src/components/shell/storage.ts:1` — **Change:** export `shellStorage` (`getItem`/`setItem` via `localStorage`) — **Accept:** `useDefaultLayout` persists `reqly-shell-sidebar` — **Verify:** `vitest` `useDefaultLayout` + manual refresh keeps `17%` — **Depends:** `UI-01.1`

*All `UI-NN` below follow this `UI-NN.T` sub-ticket pattern in the commit history (`33a1d1db` etc. are the original slices; new work will be `UI-01.1` etc.).*

---

## Redesign Gate — anti `find → see exists → tick` (so agents really redesign)

> **Problem this gates:** `agent start to redesign -> find the file -> check the contain -> see everything is there -> didn't edit anything -> give tick and move on` — a superficial tick where the file exists so the agent claims the milestone without changing the design.

**For every `UI-NN.T` sub-ticket, a tick is valid only if all 5 hold; `git diff` existence is not enough:**

1. **`git diff` non-empty for its `Files`** — `git diff main...HEAD --stat` must list each `Files` entry with `+`/`-` (not `0` or file-existence). `git status --short` `M`/`A` for those paths. **Fail if:** `ls frontend/src/styles/tokens.css` exists but `git diff --stat` is empty → not redesigned. Check: `git diff --numstat -- frontend/src/styles/tokens.css` `added >0` + `deleted >0` or new file `A`.
2. **Design delta is real, not a re-tick** — `frontend-design` `palette`/`type`/`layout`/`signature` must have at least one **opinionated change** from `main`’s prior chrome: a new `hex` in `tokens.css`/`index.css`, a new `font` role in `index.css` `@fontsource`, a new `ASCII wireframe` in the ticket, or a new `signature` element. **Fail if:** `grep -r "background #fbfbfa" frontend/src/styles/tokens.css` is unchanged from `ae70f07a` and no new `hex` added → copy-paste, not redesign.
3. **Before/after is observable** — ticket must include either (a) `frontend/src/styles/tokens.css:1` diff + screenshot `docs/internal/ui-demos/screenshots/*` before/after, or (b) type-level proof: `npx tsc --noEmit` still green **and** `grep -r "hardcoded.*#[0-9a-fA-F]" frontend/src --include="*.tsx" --include="*.ts" | wc -l` `==0` (no scattered hex, via `tokens.css` only) + `oxlint` 0. **Fail if:** `npx tsc` green but `git diff` shows only `ROADMAP.md` tick, no `frontend/src/` change.
4. **Behavior still holds via the gate, not via file existence** — the milestone’s `Verify` commands must be run and their **output** pasted in the PR (not “exists”): `nub run typecheck` `Done`, `npx oxlint` `0`, `go vet ./...` `0`, `vitest run frontend/src/...` `185` pass where applicable, plus manual `⌘K`/`⌘B`/`⌘J` or `Send GET {{var}}` as per `Accept`. **Fail if:** `ls` shows file but `vitest` not run.
5. **One-line design rationale is required** — PR body must state the single `Risk` choice for that milestone (e.g., “`Risk: density — keep 13px + hairline, one `StatusPill` pulse via `motion`”`) and why it fits the subject (Git-native, HTTP legible). **Fail if:** PR is `feat: ui-03` with no rationale.

**Reviewer checklist (human or `code-review` agent) before ticking `[ ]` → `[x]`:**

- [ ] `git diff main...HEAD --stat` shows that milestone’s `Files` with insertions
- [ ] `git diff` shows at least one new `hex`/`font`/`layout` token vs `ae70f07a` (not just reformat)
- [ ] `Verify` commands output pasted (typecheck/oxlint/vet/vitest) — not “file exists”
- [ ] Screenshot or type-proof pasted (before/after or `grep` hex ==0)
- [ ] One-line `Risk` rationale in PR body

*This gate is the `DoD` for every `UI-NN.T` — it vetoes `frontend-design` if the design is a no-op. Keep `ROADMAP.md:7` ticks honest: `[x]` means redesigned and verified, not found.*

---

## Milestones — by panel & page (each is a tracer bullet)

### UI-01 — Shell & Design System (foundational — rebuild from scratch)

**Goal:** Extract canonical tokens + shell chrome so every future panel inherits the same density/a11y floor.

**Sub-tickets (each is one `but commit` on `feat/shell-slice-N`):**

- [x] `UI-01.1 — Extract tokens.css` — **Files:** `frontend/src/styles/tokens.css:1` (new 60 lines `:root`/`[data-theme]`), `frontend/src/index.css:7` (add `@import "./styles/tokens.css"`; keep legacy block until `UI-13`) — **Accept:** `Settings → Appearance` shows 4 themes + System, `data-theme` toggles light/dark — **Verify:** `nub run typecheck` + `grep -r "#c93517" frontend/src --include="*.tsx" --include="*.ts" | wc -l` == 1 (only `tokens.css`) — **Depends:** `none`
- [x] `UI-01.2 — Create shellStorage` — **Files:** `frontend/src/components/shell/storage.ts:1` (8 lines `getItem`/`setItem`) — **Change:** export `shellStorage` — **Accept:** `useDefaultLayout` persists `reqly-shell-sidebar` `17%` after refresh — **Verify:** `vitest` `useDefaultLayout` — **Depends:** `UI-01.1`
- [x] `UI-01.3 — Create useShellStore` — **Files:** `frontend/src/stores/useShellStore.ts:1` (45 lines `inspectorOpen`/`inspectorTab` + `localStorage` `reqly-shell-inspector-*`) — **Accept:** `inspectorOpen` persists — **Verify:** `vitest` `useShellStore` `initialShellState` — **Depends:** `UI-01.2`
- [x] `UI-01.4 — Create AppShell` — **Files:** `frontend/src/components/shell/AppShell.tsx:1` (130 lines 5-zone `topBar`/`toolRail`/`sidebar`/`children`/`bottom`/`statusBar`, `sidebarLayout`/`bottomLayout`/`⌘B`/`bottomCollapsed` via `shellStorage`+`useBottomPanelStore`) — **Accept:** renders `TopBar` + `ToolRail` + `ContextSidebar` + `MainWorkspace` + `BottomPanel` + `StatusBar` at correct sizes — **Verify:** `nub run typecheck` + manual `⌘B` toggles sidebar — **Depends:** `UI-01.3`
- [x] `UI-01.5 — Thin App.tsx wrapper` — **Files:** `frontend/src/app/App.tsx:57` (370→312 lines) — **Change:** replace `ResizablePanelGroup` 30 lines with `<AppShell topBar={<TopBar/>} toolRail={<ToolRail/>} sidebar={<ContextSidebar/>} bottom={<BottomPanel/>} statusBar={<StatusBar/>}>` + keep `activeView` switch + `RequestTabs`/`ResponseViewer` split inside `children` — **Accept:** no visual diff, `⌘J` still toggles bottom — **Verify:** `nub run typecheck` + `go vet` + `vitest 185` — **Depends:** `UI-01.4`
- [ ] `UI-01.6 — Remove duplicate tokens` — **Files:** `frontend/src/index.css:68` (delete legacy `:root` block, keep only `@theme` `var(--background)` mapping) — **Accept:** no visual diff, single source `tokens.css` — **Verify:** `grep -n ":root" frontend/src/index.css` == 0 — **Depends:** `UI-13` (after all UI)

**DoD:** `nub typecheck` + `oxlint` 0 + `go vet` + `vitest 185` green (already `33a1d1db`+`82d44305` for 01.1–01.5; 01.6 deferred).

### UI-02 — Request Workspace (Builder + Viewer)

**Stories:** 13–21 (Builder 4 tabs + overflow, URL bar method/URL/Send/Save, tabs persist/drag, Viewer status/timing/size/proto + Body/Headers/Cookies/Test Results/Timeline + split `↔/↕` + divider).

**Sub-tickets:**

- [x] `UI-02.1 — Builder 4 tabs` — **Files:** `frontend/src/features/request-editor/RequestEditor.tsx:7` (4 tabs `params`/`headers`/`body`/`auth` + `overflowTabs` 5) — **Accept:** `Params`/`Headers`/`Body`/`Auth` visible, `More` shows 5 — **Verify:** `vitest` `RequestEditor` — **Depends:** `UI-01.5`
- [x] `UI-02.2 — RequestSettingsDialog` — **Files:** `frontend/src/features/request-editor/RequestSettingsDialog.tsx:1` (124 lines `timeout`/`followRedirects` `RedirectValue` `default|on|off`) — **Change:** per-request `timeout`/`followRedirects` + `settingsSummary` chip `frontend/src/features/request-editor/RequestEditor.tsx:335` — **Accept:** chip shows `3000ms · no redirects` — **Verify:** `vitest` `RequestSettingsDialog` + `handleSend` forwards seam — **Depends:** `UI-02.1`
- [x] `UI-02.3 — ResponseViewer` — **Files:** `frontend/src/features/response-viewer/ResponseViewer.tsx:1` (`isTabular`/`binaryPreviewType` + `Timeline` waterfall `frontend/src/lib/response.ts:187`) — **Accept:** `Body`/`Headers`/`Cookies`/`Test Results`/`Timeline` tabs, `↔/↕` split — **Verify:** `vitest` `ResponseViewer` Table 1000 rows — **Depends:** `UI-02.2`
- [x] `UI-02.4 — RequestTabs` — **Files:** `frontend/src/components/RequestTabs.tsx:1` (`duplicateTab` `SharedContextMenu` `x,y`, `tabIsDirty` dot, `⌘W` guard `useRequestStore`) — **Change:** `onContextMenu` → `SharedContextMenu` `Duplicate request` → `duplicateTab` — **Accept:** right-click `Duplicate` creates `copy` — **Verify:** `vitest` `RequestTabs` — **Depends:** `UI-02.3`
- [x] `UI-02.5 — Response layout persist` — **Files:** `frontend/src/stores/useShellStore.ts:1` (`responseMode: split|inline` + `localStorage`), `frontend/src/app/App.tsx:94` `splitOrientation` → `useShellStore` — **Accept:** `↔/↕` persists after refresh — **Verify:** `vitest` `useShellStore` — **Depends:** `UI-02.4`
- [x] `UI-02.6 — Binary file picker` — **Files:** `frontend/src/features/request-editor/RequestEditor.tsx:484` (`drag` + `type` + `fixtures/`), `apps/desktop/frontend/src/bridge.ts:85` `file-browse` binding — **Accept:** drop `payload.bin` sets `./fixtures/payload.bin` — **Verify:** manual drop — **Depends:** `UI-02.5`

**Accept (milestone):** Send `GET` with `{{var}}` + `{{$uuid}}`, `followRedirects:false` shows `manual`, `duplicate` creates scratchpad copy, split toggle persists.

### UI-03 — Collections Explorer

**Stories:** 22–24 + 56.8 partial

- [x] Collection row `Run` action + `Open`/`Duplicate`/`Copy path` context menu (`CollectionTree.tsx:142`)

### UI-04 — Environments & History

**Stories:** 25–27 (Environments), 28–30 (History)

- [x] History `filter` by `Method`/`Status` multi-column selectors (`HistoryView.tsx:115`)
- [ ] History retention `30d/90d/1yr/forever` UI exists `SettingsView.tsx:21` but no `history.db` `DELETE ... WHERE createdAt <` pruning — **missing**

### UI-05 — Tool Pages (Mocks, Diff, JWT, GraphQL, gRPC, Runners, Explorer, Docs)

**Stories:** 31–52

- [x] `frontend/src/features/mock-view/MocksView.tsx:1` route editor + `scenario` + `fault injection` + `logs` + `MockStatus` dot
- [x] `frontend/src/features/diff-view/DiffView.tsx:1` `A`/`B` `Current`/`History`/`Saved`/`Clipboard` + `Side by Side`/`Unified`/`Structural`/`Headers`
- [x] `frontend/src/features/jwt-inspector/JwtInspector.tsx:1` `token` input + `Header`/`Payload`/`Signature` + `Valid/Expired` + `expiry`
- [x] `frontend/src/features/graphql-browser/GraphqlBrowser.tsx:1` `Endpoint` + `Schema` sidebar (`Query`/`Mutation`/`Subscription`/`Types`/`Enums`...) + `Query` editor + `Response`
- [x] `frontend/src/features/grpc-view/GrpcTab.tsx:1` `target` + `Services` tree + `Method` + `Request`/`Response`
- [x] `frontend/src/features/runners-panel/RunnersPanel.tsx:1` `Collection`/`Pagination`/`Bulk`/`Dataset` tabs
- [x] `frontend/src/features/openapi-explorer/OpenapiExplorer.tsx:1` `API tree` + `endpoint` + spec switch `dropdown`
- [x] `frontend/src/features/docs-view/DocsView.tsx:1` Markdown `GraphQL`/`gRPC` docs + `export` `markdown`/`OpenAPI`
- [x] Runners `Dataset` picker UI `DatasetPicker.tsx` exists but not wired to `RunnersPanel` dataset tab (wired via `DatasetPicker` in `RunnersPanel.tsx:253`)
- [x] `SpecEditorView.tsx:1` tree + YAML editor exists and `EndpointEditor` validation `patchEndpointInContent` is surfaced in `SpecEditorView.tsx:177`

### UI-06 — Import / Export

**Stories:** 53–56

- [x] `Export` `Environment` as `environments/<name>.yaml` in `ExportDialog` format list (`export.ts:28`)
- [ ] `Import` conflict `Merge` is shallow (collection `upsert` only) — **missing** deep merge per ADR

### UI-07 — Settings

**Stories:** 57–59

- [x] `frontend/src/features/settings-view/SettingsView.tsx:1` full page `Appearance` (14 built-in themes + `system` `THEMES` `lib/themes.ts:1` — Atlas Light/Dark, Windows 11 Light/Dark, macOS Tahoe Light/Dark, Linux KDE Light/Dark, Linux GNOME Light/Dark, `setTheme`/`cycleTheme`), `Workspace` (`name`/`path`/`openFolder`), `Storage` (`History Retention` `30d`/`90d`/`1yr`/`forever` + `ProxyTlsPanel` `CicdPanel`), `About` `APP_VERSION` `lib/crash.ts:1` + `SHORTCUTS`
- [x] `frontend/src/features/settings-view/ProxyTlsPanel.tsx:1` `HTTP`/`HTTPS`/`SOCKS` + `insecureSkipVerify`/`caFile`/`TLS version`
- [x] `frontend/src/features/settings-view/CicdPanel.tsx:1` `GitHub Action YAML` + `CLI` generator
- [ ] `Keyboard Shortcuts` `SettingsView` shows table but not editable — spec wants `customizable` — **missing** (`useKeyboardMap` hard-coded)
- [ ] `Auth Settings` sub-page `saved credentials`/`OAuth clients` not yet — **missing** (per-request `AuthEditor` exists, but global `Auth Settings` empty)

### UI-08 — Global Panels

**Stories:** 60–64

- [x] `Console` `Copy JSON` and `Copy Text` export buttons (`BottomPanel.tsx:37`)
- [x] `Network` `Clear Activity` button (`BottomPanel.tsx:85`)

### UI-09 — Realtime

**Stories:** 65–66

- [x] `Realtime` `SendBinary` `base64` mode toggle surfaced in `RealtimeTab.tsx:212`

### UI-10 — Search & Palette

**Stories:** 67–68, 60–63 polish

- [x] `frontend/src/features/command-palette/CommandPalette.tsx:1` `⌘K` (`useCommandPaletteStore` `open`/`query`, `getFilteredResults` Fuse `threshold:0.4` capped 20), grouped `Navigation`/`Theme`/`Environment`/`Collection`/`History`/`command` (`groupByHint`), `recent 5` `localStorage` `RECENT_KEY`, directional empty (`Try Go to…`), `⌘K`/`↵`/`Esc` hints, `hint` pill
- [x] `frontend/src/lib/paletteProviders.ts:7` `navViews` 16 (`home`→`spec-editor`) + `import`/`export` + `theme-light`/`dark`/`windows-11-light`/`dark`/`macos-tahoe-light`/`dark`/`linux-kde-light`/`dark`/`linux-gnome-light`/`dark`/`system`
- [x] `Rail` dynamic order keyboard shortcuts (`useKeyboardMap.ts:89`)

### UI-11 — Body Editor & Auth

**Stories:** 69–72, 73–74, 75–76

- [x] `frontend/src/lib/body.ts:1` `BodyType` `none`/`json`/`raw`/`text`/`xml`/`html`/`form-data`/`urlencoded`/`binary`/`graphql` + `serializeBody`
- [x] `frontend/src/features/request-editor/RequestEditor.tsx:484` `Binary` file picker (`drag` + `type` + `fixtures/`)
- [x] `frontend/src/lib/request.ts:1` `KeyValueRow` (`file`/`filename` for `M21`), `sentRows`
- [x] `frontend/src/lib/authSchemes.ts:1` `RequestAuth` `Inherit`/`No Auth`/`Basic`/`Bearer`/`API Key`/`OAuth2` (3 flows)/`Digest`/`AWS`/`Custom` + `AuthEditor.tsx:1`
- [x] `Template` `Custom` categories and custom templates surfaced in `TemplatePickerSheet.tsx:32`

### UI-12 — Proxy / TLS & CI/CD

**Stories:** 77–82

- [x] `frontend/src/lib/proxyTls.ts:1` `validate`/`format` + `ProxyTlsPanel.tsx` global + per-request `RequestContext` override `App.tsx:57` `splitLayout`
- [x] `frontend/src/lib/cicd.ts:1` `GitHub Action YAML` + `CicdPanel` with recent pipeline runs summary (`CicdPanel.tsx:156`)

- [x] `Docs` `export` `Environment` as `yaml` supported via `ExportDialog` (`export.ts:28`)

### UI-13 — CLI Parity — Realtime Expansion (MQTT + Socket.IO)

**Gap:** `apps/cli/cmd/mqtt.go:1` `mqtt pub|sub` + `apps/cli/cmd/socketio.go:1` `socketio connect|emit` have no GUI — `apps/desktop/backend/realtime.go:1` only handles `ws`/`sse`, `frontend/src/features/realtime-pages/RealtimePage.tsx:1` only `WebSocketPage`/`SSEPage`.

**Sub-tickets (tracer bullets):**

- [ ] `UI-13.1 — MQTT Pub/Sub page` — **Files:** `apps/desktop/backend/mqtt.go:1` (new `MqttPublish`/`MqttSubscribe`/`MqttCancel` via `internal/mqtt`), `frontend/src/features/mqtt-view/MqttView.tsx:1` (new broker/topic/QoS/retain + Publish/Subscribe log), `frontend/src/lib/mqtt.ts:1` (bridge types) — **Change:** expose `internal/mqtt.MQTTOptions` through `AppService` and render publish/subscribe UI mirroring CLI flags `--topic`/`--message`/`--qos`/`--retain`/`--username`/`--password` — **Accept:** `MQTT` appears in `ToolRail` Realtime group, publish shows `published message to topic` toast, subscribe streams messages — **Verify:** `go vet ./...` + `nub run typecheck` + manual `mqtt pub localhost:1883 --topic test --message hi` — **Depends:** `none`
- [ ] `UI-13.2 — Socket.IO page` — **Files:** `apps/desktop/backend/socketio.go:1` (new `SocketIOConnect`/`SocketIOEmit`/`SocketIOClose` via `internal/socketio`), `frontend/src/features/socketio-view/SocketIOView.tsx:1` (new url + namespace + event/data + connect/emit/log), `frontend/src/lib/socketio.ts:1` — **Change:** expose `internal/socketio` through `AppService` and render `connect <url> [--namespace]` / `emit <url> --event --data` UI — **Accept:** `Socket.IO` appears in `ToolRail` Realtime group, connect shows live event stream — **Verify:** `go vet` + `nub typecheck` + manual `socketio connect ws://localhost:3000` — **Depends:** `UI-13.1`
- [ ] `UI-13.3 — Realtime rail unification` — **Files:** `frontend/src/components/shell/ToolRail.tsx:1` (extend `REALTIME_GROUP` with `MQTT`/`Socket.IO`), `frontend/src/app/App.tsx:1` (add `activeView === "mqtt"|"socketio"` routes), `frontend/src/stores/useWorkspaceStore.ts:1` (extend `WorkspaceView`) — **Change:** add `mqtt`/`socketio` to `WorkspaceView` union and `ToolRail` — **Accept:** rail shows `WebSocket`/`SSE`/`MQTT`/`Socket.IO` (4 items), `⌘K` palette lists them — **Verify:** `vitest` `useWorkspaceStore` + `nub typecheck` — **Depends:** `UI-13.2`

**DoD:** `Realtime` group has 4 entries, each streams via `internal/*` core with cancel, `go test -race ./internal/mqtt ./internal/socketio` green.

### UI-14 — CLI Parity — Governance & Enterprise (Policy/RBAC/Audit/SSO/SCIM/Collab)

**Gap:** `policy.go:1` `policy show|validate|enforce`, `rbac.go:1` `rbac list|check`, `audit.go:1` `audit list|clear`, `sso.go:1` `sso validate`, `scim.go:1` `scim user create|list`, `collab.go:1` `collab list|add|remove|serve` all have `apps/desktop/backend/*.go:1` bindings (`PolicyGet/Save/Enforce`, `RBACList/Check/Get`, `AuditList/Add/Clear/Export`, `SSOValidate`, `SCIMCreateUser/ListUsers`, `CollabList/Add/Remove/Serve`) but no `frontend/src/features/*` view, no `ToolRail` entry.

**Sub-tickets:**

- [ ] `UI-14.1 — Policy & RBAC page` — **Files:** `frontend/src/features/governance/PolicyRbacView.tsx:1` (new `PolicyEditor` + `RBACMatrix` + `enforce` dry-run), `frontend/src/lib/policy.ts:1`, `frontend/src/lib/rbac.ts:1` — **Change:** surface `PolicyGet/Save/Enforce` + `RBACList/Check` (0600, Git-native) with `validate` feedback — **Accept:** `Governance` view shows policy YAML + `enforce --action --resource` result — **Verify:** `go test ./internal/policy ./internal/rbac` + `nub typecheck` — **Depends:** `none`
- [ ] `UI-14.2 — Audit Log page` — **Files:** `frontend/src/features/audit-view/AuditView.tsx:1` (new append-only log table + `Clear` confirm + `Export`), `frontend/src/lib/audit.ts:1` — **Change:** surface `AuditList/Clear/Export` (0600) with `HistoryView`-style filter — **Accept:** audit trail renders, `Clear` requires `AlertDialog` confirm — **Verify:** `go test ./internal/audit` — **Depends:** `UI-14.1`
- [ ] `UI-14.3 — SSO & SCIM page` — **Files:** `frontend/src/features/sso-view/SsoScimView.tsx:1` (new `SSOValidate` issuer/client-id/token/secret form + `SCIM` user table `create/list`), `frontend/src/lib/sso.ts:1` — **Change:** surface `SSOValidate` (OIDC) + `SCIMCreateUser/ListUsers` (local in-memory) — **Accept:** SSO validate shows `valid/invalid` pill, SCIM create persists in table — **Verify:** `go test ./internal/sso ./internal/scim` — **Depends:** `UI-14.2`
- [ ] `UI-14.4 — Collaboration page` — **Files:** `frontend/src/features/collab-view/CollabView.tsx:1` (new `list` + `add --user --role` + `remove --user` + `serve` status), `frontend/src/lib/collab.ts:1` — **Change:** surface `CollabList/Add/Remove/Serve` (Git-native shared workspaces) — **Accept:** collaborators table + role `viewer|editor|admin` select — **Verify:** `go test ./internal/collab` — **Depends:** `UI-14.3`

**DoD:** `ToolRail` System/Governance group, every `0600` file respects `FileStore` perms, `Export` writes to `.reqly/`.

### UI-15 — CLI Parity — Automation & Orchestration (Automation/Workflow/Monitor/Changelog)

**Gap:** `automation.go:1` `automation run <yaml>` (local scheduler), `workflow.go:1` `workflow <yaml>` (multi-step), `changelog.go:1` `changelog <old> <new> --format --fail-on-breaking`, and `monitor.go:1` `monitor run --interval` (CLI) → GUI `MonitorView.tsx:1` is mock-only (random `Math.random()` data, no backend).

**Sub-tickets:**

- [ ] `UI-15.1 — Automation scheduler` — **Files:** `frontend/src/features/automation-view/AutomationView.tsx:1` (new file picker + interval + run log), `apps/desktop/backend/automation.go:1` wire `AutomationRun` already exists — verify not-stub — **Change:** mount `AutomationRun` in UI with `workflow on interval` controls — **Accept:** pick `automation.yaml` → `Run` streams steps — **Verify:** `go test ./internal/automation` — **Depends:** `none`
- [ ] `UI-15.2 — Workflow runner` — **Files:** `frontend/src/features/workflow-view/WorkflowView.tsx:1` (new `workflow.yaml` editor + step graph + `RunView` reuse), `frontend/src/lib/workflow.ts:1` — **Change:** surface `WorkflowRun` (visual/programmatic multi-step) — **Accept:** workflow steps render as `RunView` chain — **Verify:** `go test ./internal/workflow` — **Depends:** `UI-15.1`
- [ ] `UI-15.3 — Monitor wiring (fix stub)` — **Files:** `frontend/src/features/monitor-view/MonitorView.tsx:1` (replace `Math.random()` mock with `getMonitorBridge`/`PerfRun` interval polling), `apps/desktop/backend/monitor.go:1` (new `MonitorRun` if missing) — **Change:** wire `monitor run <request-file> --interval 5m --json` to live `status/latency` chart, keep `MonitorView` chart but with real data — **Accept:** `Scheduled health checks` show live availability/latency, not seeded fake `00:00` points — **Verify:** `vitest` `MonitorView` + `go test ./internal/monitor` — **Depends:** `UI-15.2`
- [ ] `UI-15.4 — Changelog view` — **Files:** `frontend/src/features/changelog-view/ChangelogView.tsx:1` (new `old-spec`/`new-spec` pickers + `markdown|json` + `fail-on-breaking` toggle + SemVer bump pill), `frontend/src/lib/changelog.ts:1`, `apps/desktop/backend/changelog.go:1` (new `ChangelogGenerate` via `internal/diffing`/`oasdiff`) — **Change:** surface `changelog <old-spec> <new-spec>` human-readable changelog — **Accept:** shows `breaking/features/fixes` + `suggested bump major|minor|patch` — **Verify:** `go test ./internal/diffing` — **Depends:** `UI-15.3`

**DoD:** Automation/Workflow/Monitor/Changelog each have `AppService` binding + streamed logs, `MonitorView` no longer contains `Math.random()`.

### UI-16 — CLI Parity — Developer Tooling (AI/Schema/Plugin)

**Gap:** `ai.go:1` `ai explain|test|docs|diagnose|schema` (`internal/ai` local heuristics), `schema.go:1` `schema validate|inspect|generate` (`internal/jsonschema` + `validation`), `plugin.go:1` `plugin list|validate` (`internal/plugin`) — all CLI-only, no desktop backend, no GUI.

**Sub-tickets:**

- [ ] `UI-16.1 — AI assistant panel` — **Files:** `apps/desktop/backend/ai.go:1` (new `AiExplain`/`AiGenerateTests`/`AiGenerateDocs`/`AiDiagnose`/`AiExplainSchema` via `internal/ai`), `frontend/src/features/ai-view/AiView.tsx:1` (new `response.json` drop + `Explain`/`Test`/`Diagnose` tabs), `frontend/src/lib/ai.ts:1` — **Change:** expose `internal/ai` (local, zero telemetry) through `AppService` — **Accept:** drop `response.json` → `Explain` shows summary/latency, `Test` synthesizes Goja assertions — **Verify:** `go test ./internal/ai` + `nub typecheck` — **Depends:** `none`
- [ ] `UI-16.2 — JSON Schema workbench` — **Files:** `apps/desktop/backend/schema.go:1` (new `SchemaValidate`/`SchemaInspect`/`SchemaGenerate` via `internal/jsonschema`), `frontend/src/features/schema-view/SchemaView.tsx:1` (new schema/instance editors + `validate --json` + `inspect --json` + `generate --seed` + violation path list) — **Change:** surface `schema validate <schema> [instance|-]` with stdin support via file drop — **Accept:** violation paths render at instance path, `inspect` shows keywords, `generate` emits sample — **Verify:** `go test ./internal/jsonschema ./internal/validation` — **Depends:** `UI-16.1`
- [ ] `UI-16.3 — Plugin manager` — **Files:** `apps/desktop/backend/plugin.go:1` (new `PluginList`/`PluginValidate` via `internal/plugin`), `frontend/src/features/plugin-view/PluginView.tsx:1` (new `plugins/` table + `Validate` button + `Capabilities` badges) — **Change:** surface `plugin list|validate <name>` for `plugins/<name>` dirs — **Accept:** table shows `name version (capabilities)` or `invalid` — **Verify:** `go test ./internal/plugin` — **Depends:** `UI-16.2`

**DoD:** AI/Schema/Plugin each have isolated `frontend/src/features/*/AiView|SchemaView|PluginView` + `AppService` + `go test` + `vitest` parity, no cloud calls.

### UI-17 — CLI Parity — Parity Polish (JWT/OpenAPI/GraphQL/Theme/Validate + stale todos)

**Gap:** partial parity: `jwt.go:1` `verify|sign` missing in GUI (`JwtInspector.tsx:1` only `JwtDecode`), `openapi.go:1` `validate|convert-v2` missing (`OpenapiExplorer.tsx:1` only `explore|generate`), `graphql.go:1` `parse` missing (`GraphqlBrowser.tsx:1` only `introspect`), `theme.go:1` `import <file>` file picker missing (`SettingsView.tsx:1` only cycles built-ins), `validate.go:1` `project` missing (`SpecEditorView.tsx:1` only OpenAPI validate). Plus stale `ROADMAP` todos: history retention pruning, import deep merge, keyboard shortcuts editable, auth settings sub-page.

**Sub-tickets:**

- [ ] `UI-17.1 — JWT verify/sign` — **Files:** `frontend/src/features/jwt-inspector/JwtInspector.tsx:1` (extend with `verify --secret` + `sign --secret --alg` tabs), `apps/desktop/backend/jwtdialog.go:1` (add `JwtVerify`/`JwtSign` via `internal/jwt`) — **Change:** add `verify` HMAC check pill + `sign` payload→JWT — **Accept:** verify shows `Valid/Invalid`, sign emits token — **Verify:** `go test ./internal/jwt` + `vitest` `JwtInspector` — **Depends:** `none`
- [ ] `UI-17.2 — OpenAPI polish` — **Files:** `frontend/src/features/openapi-explorer/OpenapiExplorer.tsx:1` (add `validate` badge + `convert-v2` Swagger 2.0 → 3.0.3 button), `apps/desktop/backend/openapiexplorer.go:1` (add `OpenapiValidate`/`OpenapiConvertV2` via `internal/openapi`) — **Change:** surface `openapi validate <spec>` + `convert-v2 <swagger2>` — **Accept:** validate shows errors at path, convert emits 3.0.3 YAML — **Verify:** `go test ./internal/openapi` — **Depends:** `UI-17.1`
- [ ] `UI-17.3 — GraphQL parse + Theme import + Validate project` — **Files:** `frontend/src/features/graphql-browser/GraphqlBrowser.tsx:1` (add `parse <schema.graphql> [--type] [--json]` local SDL tab), `frontend/src/features/settings-view/SettingsView.tsx:1` (add `Import theme` file picker → `ThemeImport`), `frontend/src/features/spec-editor/SpecEditorView.tsx:1` (add `Validate project [path]`), `apps/desktop/backend/graphqlbrowser.go:1` (add `GraphqlParse`), `apps/desktop/backend/theme.go:1` already `ThemeImport`, `apps/desktop/backend/spec.go:1` (add `ValidateProject`) — **Change:** wire remaining CLI sub-commands — **Accept:** parse shows SDL types, theme import persists YAML/JSON, project validate scans workspace — **Verify:** `nub typecheck` + `go vet` — **Depends:** `UI-17.2`
- [ ] `UI-17.4 — Stale todos closeout` — **Files:** `frontend/src/features/history-view/HistoryView.tsx:1` (pruning `DELETE ... WHERE createdAt < retention`), `frontend/src/features/import-dialog/ImportDialog.tsx:1` (deep merge per ADR, not shallow `upsert`), `frontend/src/hooks/useKeyboardMap.ts:1` + `frontend/src/features/settings-view/SettingsView.tsx:1` (editable shortcuts table, remove `hard-coded` guard), `frontend/src/features/settings-view/SettingsView.tsx:1` `Auth Settings` sub-page `saved credentials`/`OAuth clients` — **Change:** close `UI-04. retention (30d/90d/1yr/forever) pruning` + `UI-06. merge shallow` + `UI-07. Keyboard Shortcuts customizable` + `UI-07. Auth Settings` — **Accept:** retention actually deletes `history.db` rows, merge is deep, shortcuts persist to `localStorage`, Auth Settings lists `AuthStatus` tokens — **Verify:** `vitest` `useKeyboardMap` + `go test ./internal/history` — **Depends:** `UI-17.3`

**DoD:** `JwtInspector` has 3 tabs, `OpenapiExplorer` 4 actions, `GraphqlBrowser` 2 modes, `SettingsView` imports file themes, `SpecEditorView` validates project, and the 4 stale `[ ]` items in `UI-04`/`UI-06`/`UI-07` flip to `[x]`.

### UI-18 — Theme-Adaptive OS Layouts & Headless Primitives

**Milestone:** M68 (OS Theme Suite Refactor)

- [x] **UI-18.1 Headless & Unstyled UI Primitives:** Decouple `Button`, `Input`, `Tabs`, `Card`, `Select`, `Dialog` from hardcoded utility styles; drive purely via semantic CSS variables and slot classes — 2026-09-01
- [x] **UI-18.2 Theme-Adaptive Shell Chrome:** Adapt `AppShell`, `TopBar`, `ToolRail`, `WorkspaceSidebar`, and `BottomPanel` layouts to active OS visual DNA (Windows 11 Centered Search + Mica, macOS Tahoe Floating Glass Toolbar + Pill Search, KDE Breeze Solid Desktop Chrome + 1px Dividers, GNOME Adwaita Integrated Headerbar) — 2026-09-01
- [x] **UI-18.3 Theme-Adaptive Feature Views:** Refactor `SettingsView`, `HomeView`, `RequestEditor`, and `RealtimeTab` into OS-authentic layout patterns (GNOME boxed-list preference groups, macOS grouped glass containers, Windows 11 Fluent settings cards, KDE Breeze geometric desktop tables) — 2026-09-01

---

## Missing from `docs/internal` (gap vs spec) — cleaned 2026-09-02

- [x] `docs/internal/frontend-design-review-2026-08-23.md:1` — cream/serif/acid-green defaults rejected (use `tokens.css` + `IBM Plex` + `StatusPill`), no scattered hex, one token system — **done**; duplicate `:root` removed from `index.css`.
- [x] `docs/spec/m44-design-port.md:1` + `m45-openapi-editor-mvp.md` — `SpecEditorView` + `EndpointEditor` `validateEndpoint`/`patchEndpointInContent` (`frontend/src/lib/specTree.ts:1`) surfaced in `SpecEditorView.tsx:177` — **done**.
- [ ] `docs/internal/gui-roadmap.md:1` GUI-0 `StatusBar` empty placeholders — spec wants `Git branch`/`ahead/behind`/`dirty`/`active env` live via `ShellAdapter` (`internal/git` + `apps/desktop/backend/gitview.go` + `frontend/src/components/shell/GitSidebar` in `feat/m44-t4`) — **not in `main`** (intentionally local-first, no `git` binding yet) — tracked, not blocking CLI parity.
- [ ] `docs/adr/0030-navigation-model.md:1` — `AppShell` `inspectorOpen` (`frontend/src/stores/useShellStore.ts:1`) exists but no `inspector` content mounted — **missing** (intentional P3, deferred).
- [ ] `docs/adr/0029-theme-registry.md:1` — `THEMES` `atlas-light`/`dark` + `system` done, but `appearance: light|dark` derivation in `useThemeStore` not yet `themeById`/`firstWithAppearance` (feat `m44-t1` has test seam `createThemeController`) — **missing** polish.

**CLI gaps (now tracked as milestones, not missing):** former CLI-not-in-GUI gaps are no longer "missing" — they are `UI-13` (MQTT/Socket.IO) + `UI-14` (Governance: Policy/RBAC/Audit/SSO/SCIM/Collab) + `UI-15` (Automation/Workflow/Monitor/Changelog) + `UI-16` (AI/Schema/Plugin) + `UI-17` (Parity Polish: JWT verify/sign, OpenAPI validate/convert-v2, GraphQL parse, theme import, validate project, plus `UI-04`/`UI-06`/`UI-07` stale todos) above. `mcp serve` (`apps/cli/cmd/mcp.go:1`) is intentional headless (stdio) — no GUI.

## Execution (same as `372` slices 01–08, now expanded to 18)

Each UI milestone ships as a tracer bullet (component + store + a11y + test + `nub typecheck`/`oxlint`/`go vet`/`vitest 185` + `react-doctor` gates) on `feat/shell-slice-N` → `main`. Data layer `lib/`+`stores/` preserved. Keep `frontend/src/styles/tokens.css:1` as source; legacy block in `index.css` removed. `UI-01`–`UI-12` shipped, `UI-13`–`UI-17` are new CLI parity gaps (16 sub-tickets), `UI-18` shipped — total `18` milestones; `UI-01.6` deferred until after `UI-17`.
