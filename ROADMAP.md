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

## Milestones — by panel & page (each is a tracer bullet)

### UI-01 — Shell & Design System (foundational — rebuild from scratch)

**Goal:** Extract canonical tokens + shell chrome so every future panel inherits the same density/a11y floor.

- [ ] `frontend/src/styles/tokens.css:1` canonical contract (`:root`/`[data-theme='atlas-light']`/`.dark,[data-theme='atlas-dark']` — adding theme = one block + `lib/themes.ts:1` entry)
- [ ] `frontend/src/index.css:7` `@import "./styles/tokens.css"` (keep legacy `:root` until slice 08 removes duplication)
- [ ] `frontend/src/components/shell/storage.ts:1` `shellStorage` (localStorage adapter)
- [ ] `frontend/src/stores/useShellStore.ts:1` `inspectorOpen` + persistence
- [ ] `frontend/src/components/shell/AppShell.tsx:1` 5-zone shell (`topBar`/`toolRail`/`sidebar`/`children`/`bottom`/`statusBar`, `sidebarLayout`/`bottomLayout`/`⌘B`/`bottomCollapsed`)
- [ ] `frontend/src/app/App.tsx:57` 370→312 lines — thin wrapper (`TopBar`/`ToolRail`/`ContextSidebar`/`BottomPanel`/`StatusBar` via `AppShell`)
- [ ] Remove duplicate `:root` tokens from `index.css:68` (keep only `@theme` mapping) — follow-up after slice 08

**DoD:** `nub typecheck` + `oxlint` 0 + `go vet` + `vitest 185` green (already `33a1d1db`+`82d44305`).

### UI-02 — Request Workspace (Builder + Viewer)

**Stories:** 13–21 (Builder 4 tabs + overflow, URL bar method/URL/Send/Save, tabs persist/drag, Viewer status/timing/size/proto + Body/Headers/Cookies/Test Results/Timeline + split `↔/↕` + divider).

- [ ] `frontend/src/features/request-editor/RequestEditor.tsx:7` 4 tabs (`params`/`headers`/`body`/`auth`) + `overflowTabs` (`pre-request`/`tests`/`docs`/`settings`/`variables`), `TagPicker`, `methodTint`, `saveWarnings`
- [ ] `frontend/src/features/request-editor/RequestSettingsDialog.tsx:1` per-request `timeout`/`followRedirects` (`1882bd34`) + `settingsSummary` chip
- [ ] `frontend/src/features/response-viewer/ResponseViewer.tsx:1` `isTabular`/`binaryPreviewType`, `Timeline` waterfall
- [ ] `frontend/src/components/RequestTabs.tsx:1` `duplicateTab` (`SharedContextMenu`), `tabIsDirty` dot, `⌘W` guard
- [ ] Viewer `Response layout: split|inline` persisted via `useShellStore` (feat `m44-t5` has it, not yet in `main`) — **missing**
- [ ] `Body: binary` file picker drop handling already `RequestEditor.tsx:484` but no `shell` file-browse binding — **missing**

**Accept:** Send `GET` with `{{var}}` + `{{$uuid}}`, `followRedirects:false` shows `manual`, `duplicate` creates scratchpad copy, split toggle persists.

### UI-03 — Collections Explorer

**Stories:** 22–24 + 56.8 partial

- [ ] `frontend/src/components/CollectionTree.tsx:1` tree `expand/collapse`, `methodTintClass`, `TREE_KEYS` `ArrowDown/Up/Right/Left/Home/End` + `aria-expanded`, `draggable` + `onDrop`, `SharedContextMenu` `Duplicate request` (`1882bd34`)
- [ ] `frontend/src/lib/collections.ts:1` `EntryIdentity` + `WorkspaceRequest` + `CollectionsAdapter.duplicateRequest`
- [ ] `frontend/src/stores/useWorkspaceStore.ts:168` `duplicateRequestPath` (on-disk via `bridge` or scratchpad fallback)
- [ ] Search/filter `CollectionTree.tsx:148` + `useWorkspaceStore.workspaceTree`
- [ ] Context menu full set (`Rename`/`Move`/`Delete`/`Run`/`Import`/`Export`/`Generate Docs|Tests|Mock`) still `coming soon` stub — **missing** (only `Duplicate` wired)
- [ ] Collection `Run` button `RunControl` exists but no per-request `Run` in row — **missing**

### UI-04 — Environments & History

**Stories:** 25–27 (Environments), 28–30 (History)

- [ ] `frontend/src/features/environments-view/EnvironmentsView.tsx:1` `Name`/`Value`/`Secret`/`Description`, tabs `Local`/`Development`/`Staging`/`Production`, `EnvDraft` + `EnvAdapter`
- [ ] `frontend/src/features/history-view/HistoryView.tsx:1` `Method`/`URL`/`Status`/`Duration`/`Env` + `filter` + `HistoryView` replay
- [ ] `frontend/src/stores/useHistoryStore.ts:1` `pool` capped 500, `loadPool`/`search`/`replay`
- [ ] History retention `30d/90d/1yr/forever` UI exists `SettingsView.tsx:21` but no `history.db` `DELETE ... WHERE createdAt <` pruning — **missing**
- [ ] History `filter` by `Method`/`Status`/`Env`/`Date` — current `HistoryView` only does text search — **missing**

### UI-05 — Tool Pages (Mocks, Diff, JWT, GraphQL, gRPC, Runners, Explorer, Docs)

**Stories:** 31–52

- [ ] `frontend/src/features/mock-view/MocksView.tsx:1` route editor + `scenario` + `fault injection` + `logs` + `MockStatus` dot
- [ ] `frontend/src/features/diff-view/DiffView.tsx:1` `A`/`B` `Current`/`History`/`Saved`/`Clipboard` + `Side by Side`/`Unified`/`Structural`/`Headers`
- [ ] `frontend/src/features/jwt-inspector/JwtInspector.tsx:1` `token` input + `Header`/`Payload`/`Signature` + `Valid/Expired` + `expiry`
- [ ] `frontend/src/features/graphql-browser/GraphqlBrowser.tsx:1` `Endpoint` + `Schema` sidebar (`Query`/`Mutation`/`Subscription`/`Types`/`Enums`...) + `Query` editor + `Response`
- [ ] `frontend/src/features/grpc-view/GrpcTab.tsx:1` `target` + `Services` tree + `Method` + `Request`/`Response`
- [ ] `frontend/src/features/runners-panel/RunnersPanel.tsx:1` `Collection`/`Pagination`/`Bulk`/`Dataset` tabs
- [ ] `frontend/src/features/openapi-explorer/OpenapiExplorer.tsx:1` `API tree` + `endpoint` + spec switch `dropdown`
- [ ] `frontend/src/features/docs-view/DocsView.tsx:1` Markdown `GraphQL`/`gRPC` docs + `export` `markdown`/`OpenAPI`
- [ ] Runners `Dataset` picker UI `DatasetPicker.tsx` exists but not wired to `RunnersPanel` dataset tab — **missing**
- [ ] `SpecEditorView.tsx:1` tree + YAML editor exists but `EndpointEditor` validation `patchEndpointInContent` not yet surfaced — **missing**

### UI-06 — Import / Export

**Stories:** 53–56

- [ ] `frontend/src/features/import-dialog/ImportDialog.tsx:1` `OpenAPI`/`Postman`/`Insomnia`/`cURL`/`HAR`/`Reqly` + preview `Collections`/`Requests`/`Envs`/`Conflicts`/`Warnings` + `Skip`/`Merge`/`Overwrite`
- [ ] `frontend/src/features/export-dialog/ExportDialog.tsx:1` `Collection`/`Workspace`/`OpenAPI`/`cURL`/`HAR`/`Environment`/`Documentation` + `reqly export`
- [ ] `frontend/src/lib/paletteProviders.ts:32` `import`/`export` commands (`useImportStore`/`useExportStore`) + `556b` fix
- [ ] `Import` conflict `Merge` is shallow (collection `upsert` only) — **missing** deep merge per ADR
- [ ] `Export` `Environment` as `environments/<name>.yaml` not yet in `ExportDialog` format list — **missing**

### UI-07 — Settings

**Stories:** 57–59

- [ ] `frontend/src/features/settings-view/SettingsView.tsx:1` full page `Appearance` (4 themes + `system` `THEMES` `lib/themes.ts:1`, `setTheme`/`cycleTheme`), `Workspace` (`name`/`path`/`openFolder`), `Storage` (`History Retention` `30d`/`90d`/`1yr`/`forever` + `ProxyTlsPanel` `CicdPanel`), `About` `APP_VERSION` `lib/crash.ts:1` + `SHORTCUTS`
- [ ] `frontend/src/features/settings-view/ProxyTlsPanel.tsx:1` `HTTP`/`HTTPS`/`SOCKS` + `insecureSkipVerify`/`caFile`/`TLS version`
- [ ] `frontend/src/features/settings-view/CicdPanel.tsx:1` `GitHub Action YAML` + `CLI` generator
- [ ] `Keyboard Shortcuts` `SettingsView` shows table but not editable — spec wants `customizable` — **missing** (`useKeyboardMap` hard-coded)
- [ ] `Auth Settings` sub-page `saved credentials`/`OAuth clients` not yet — **missing** (per-request `AuthEditor` exists, but global `Auth Settings` empty)

### UI-08 — Global Panels

**Stories:** 60–64

- [ ] `frontend/src/components/shell/BottomPanel.tsx:1` 5 tabs `Console` (`INFO`/`ERROR` + `goLogs`/`breadcrumbs`)/`Network` (`Time`/`Method`/`URL`/`Status`/`Duration` + `history` adapter)/`Tests` (`Pass`/`Fail`/`Skipped`)/`Variables` (by scope `Global`→`Runtime`)/`Cookies` (`Domain`/`Path`/`Secure`/`HttpOnly`/`SameSite`/`Expires`) + `⌘J` toggle `useBottomPanelStore`
- [ ] `Console` `Copy` as `text` only, no `JSON` export — **missing**
- [ ] `Network` `Clear` button not yet — **missing**

### UI-09 — Realtime

**Stories:** 65–66

- [ ] `frontend/src/features/realtime-pages/RealtimePage.tsx:1` `WS_PAGE_ID`/`SSE_PAGE_ID` + `features/realtime-view/RealtimeTab.tsx:1` `url`/`headers`/`Messages`/`auto-reconnect` `exponential` `1s→30s` + `stores/useRealtimeStore.ts:1` + `useRealtimeRecentsStore.ts:1` `RECENTS_CAP:12`
- [ ] `frontend/src/components/shell/ContextSidebar.tsx:1` `RealtimeRecents` `Connect to an endpoint…` empty
- [ ] `Realtime` `SendBinary` `base64` not yet surfaced in `RealtimeTab` UI — **missing**

### UI-10 — Search & Palette

**Stories:** 67–68, 60–63 polish

- [ ] `frontend/src/features/command-palette/CommandPalette.tsx:1` `⌘K` (`useCommandPaletteStore` `open`/`query`, `getFilteredResults` Fuse `threshold:0.4` capped 20), grouped `Navigation`/`Theme`/`Environment`/`Collection`/`History`/`command` (`groupByHint`), `recent 5` `localStorage` `RECENT_KEY`, directional empty (`Try Go to…`), `⌘K`/`↵`/`Esc` hints, `hint` pill
- [ ] `frontend/src/lib/paletteProviders.ts:7` `navViews` 16 (`home`→`spec-editor`) + `import`/`export` + `theme-light`/`dark`/`system`
- [ ] `frontend/src/hooks/useKeyboardMap.ts:1` `⌘K` toggle, `⌘B` sidebar (`AppShell`), `⌘J` bottom, `⌘W` close guarded, `⌘⏎` send, `⌘1-8` rail order
- [ ] `Palette` `cmdk` a11y `CommandDialog`/`CommandGroup` not yet — still custom `Fuse` — **missing** (feat `m44-t3` has `cmdk` but adds dep)
- [ ] `Rail` `⌘1-8` only maps first 8 (`home`→`graphql`), later `grpc`/`runners`/`explorer`/`docs`/`spec-editor`/`websocket`/`sse`/`settings` unreachable — **missing** dynamic `WORKSPACE_VIEWS` order

### UI-11 — Body Editor & Auth

**Stories:** 69–72, 73–74, 75–76

- [ ] `frontend/src/lib/body.ts:1` `BodyType` `none`/`json`/`raw`/`text`/`xml`/`html`/`form-data`/`urlencoded`/`binary`/`graphql` + `serializeBody`
- [ ] `frontend/src/features/request-editor/RequestEditor.tsx:484` `Binary` file picker (`drag` + `type` + `fixtures/`)
- [ ] `frontend/src/lib/request.ts:1` `KeyValueRow` (`file`/`filename` for `M21`), `sentRows`
- [ ] `frontend/src/lib/authSchemes.ts:1` `RequestAuth` `Inherit`/`No Auth`/`Basic`/`Bearer`/`API Key`/`OAuth2` (3 flows)/`Digest`/`AWS`/`Custom` + `AuthEditor.tsx:1`
- [ ] Request Templates `lib/templates.ts:1` `search`/`instantiate` + `TemplatePickerSheet.tsx:1` (`5553cd19` already)
- [ ] Data-driven `lib/datasets.ts:1` `parseCsv`/`parseJsonDataset` + `DatasetPicker.tsx:1` file + inline
- [ ] `Auth` `OAuth2` `Device` flow `verification_uri` display in `AuthPanel` not yet — **missing** (core `internal/auth/device.go` ships, UI only shows `token` field)
- [ ] `Template` `Save as custom` not yet — **missing**

### UI-12 — Proxy / TLS & CI/CD

**Stories:** 77–82

- [ ] `frontend/src/lib/proxyTls.ts:1` `validate`/`format` + `ProxyTlsPanel.tsx` global + per-request `RequestContext` override `App.tsx:57` `splitLayout`
- [ ] `frontend/src/lib/cicd.ts:1` `GitHub Action YAML` + `CicdPanel` + global `CI runs` not yet — spec wants `recent CI runs` `pass/fail` in `Settings` — **missing**

### UI-13 — Documentation

**Stories:** 83–84

- [ ] `frontend/src/features/docs-view/DocsView.tsx:1` rendered `GraphQL`/`gRPC` + `export` `markdown`/`OpenAPI`
- [ ] `Docs` `export` `Environment` as `yaml` not yet — **missing**

---

## Missing from `docs/internal` (gap vs spec)

- `docs/internal/gui-roadmap.md:1` GUI-0 `StatusBar` empty placeholders — spec wants `Git branch`/`ahead/behind`/`dirty`/`active env` live via `ShellAdapter` (`internal/git` + `apps/desktop/backend/gitview.go` + `frontend/src/components/shell/GitSidebar` in `feat/m44-t4`) — **not in `main`** (intentionally local-first, no `git` binding yet)
- `docs/internal/frontend-design-review-2026-08-23.md:1` — cream/serif/acid-green defaults rejected (we use `tokens.css` + `IBM Plex` + `StatusPill`), no scattered hex, one token system — **done**; remaining gap: `index.css:68` still duplicates `:root` tokens (keep `tokens.css` as source, remove legacy block after slice 08)
- `docs/spec/m44-design-port.md:1` + `m45-openapi-editor-mvp.md` — `SpecEditorView` exists but `EndpointEditor` `validateEndpoint`/`patchEndpointInContent` (`lib/specTree.ts:1`) not yet surfaced in `SpecEditorView` toolbar — **missing**
- `docs/adr/0030-navigation-model.md:1` — `AppShell` `inspectorOpen` (`useShellStore.ts:1`) exists but no `inspector` content mounted — **missing** (intentional P3)
- `docs/adr/0029-theme-registry.md:1` — `THEMES` `atlas-light`/`dark` + `system` done, but `appearance: light|dark` derivation in `useThemeStore` not yet `themeById`/`firstWithAppearance` (feat `m44-t1` has test seam `createThemeController`) — **missing** polish

## Execution (same as `372` slices 01–08, now expanded to 13)

Each UI milestone ships as a tracer bullet (component + store + a11y + test + `nub typecheck`/`oxlint`/`go vet`/`vitest 185` + `react-doctor` gates) on `feat/shell-slice-N` → `main`. Data layer `lib/`+`stores/` preserved. Keep `frontend/src/styles/tokens.css:1` as source; remove `index.css:68` legacy block after UI-13.
