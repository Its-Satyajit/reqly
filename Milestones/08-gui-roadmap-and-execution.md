# Desktop GUI Roadmap & Execution (GUI-0 to GUI-5)

28. ~~**HAR import/export + replay**~~ — `internal/importer` HAR parse + `reqly import har <har-file> [--output <dir>] [--collection <name>]` ( `headers+cookies→Headers` `Cookie:` merged, `queryString→Query`, `postData.text→Body` base64 decoded, `mimeType→Content-Type`, >1MB spill `blobs/`, `pageref`/`timings`/`cache` warnings) + `reqly export har [--out <file.har>] [--env <name>] [--limit 500]` history→HAR via `internal/exporter/har.go` (`ExportHAR` pure, `timings` synthesized, base64 binary, secrets masked), replay via `har-import` collection + `history replay` ([ADR 0020](docs/adr/0020-har-import-export.md), CONTEXT `HAR`/`HAR Import`/`HAR Export`/`HAR Replay` grilling Q1–Q4 done, `docs/spec/m28-har-import-export.md`) — **shipped**
29. ~~**JWT tooling**~~ — `reqly jwt decode` (header/claims viewer, expiry detection) in `internal/jwt` + `reqly jwt decode [--json]` + `Bearer`/stdin (`internal/jwt.Decode` + `apps/cli/cmd/jwt.go`, expiry `exp`/`nbf`/`iat` → `expired`/`not_yet_valid`/`valid`/`no_expiry`, `Header:`/`Payload:` pretty + `--json`, [ADR 0021](docs/adr/0021-jwt-tooling-decode.md), CONTEXT `JWT Tooling`/`JWT Decode` grill Q1–Q5) — **shipped (decode MVP)**; `verify`/`sign` (HS via `jwtHashes`) + desktop inspector deferred to M29b
30. ~~**Pagination runner**~~ — `reqly pagination run <request-file> [--max-pages <n>]` ( `request.pagination: {strategy: page|offset|cursor|link-header, pageParam/pageSizeParam/offsetParam/limitParam/cursorParam, nextPath: $.nextCursor, maxPages: 100}` + `internal/pagination.Run` pure loop over `sendFn` `page`→`?page=1→2` `offset`→`?offset=0→10` `cursor`→`?cursor=<next>` via JSONPath `$.nextCursor` `link-header`→`Link: <url>; rel="next"` , stop empty/missing-next/non-2xx/maxPages, `--max-pages` overrides, `OnStep` streaming `step: status duration url`) ([ADR 0022](docs/adr/0022-pagination-runner.md), CONTEXT `Pagination Runner` `Strategy`/`Stop` grill Q1–Q4, `docs/spec/m30-pagination-runner.md`) — **shipped**
31. ~~**Bulk request execution**~~ — `reqly bulk run <request-file> --data <csv|json> [--parallel] [--concurrency <n>] [--continue-on-error]` (CSV header→`{{var}}`/JSON array stringified, `internal/bulk.Run` sequential default, parallel semaphore ordered `concurrency 5`, `ScopeRuntime` per row, stop first non-2xx unless `--continue-on-error`) ([ADR 0023](docs/adr/0023-bulk-runner.md), CONTEXT `Bulk Runner`/`Bulk Input Row`/`Bulk Concurrency` grill Q1–Q4, `docs/spec/m31-bulk-runner.md`) — **shipped**
32. ~~**Retry & resilience**~~ — engine-level `request.retry` (`count`/`delayMs`/`strategy`/`maxDelayMs`/`retryOn`) in `Client.Execute`; network errors + 429/502/503/504 default, `Retry-After` respected + clamped, exponential/fixed backoff capped, ctx-cancel aborts mid-wait, auth refresh stays inside one attempt, `response.Attempts` + `history show` attempts line + desktop attempts badge, `--retries`/`--retry-delay` flags, desktop collapsible Retry section in the request editor ([ADR 0024](docs/adr/0024-retry-resilience.md), `docs/spec/m32-retry-resilience.md`) — **shipped**
33. ~~**OpenAPI editor + endpoint explorer**~~ — in-app spec authoring + generate requests from spec + JSON Schema edit/validate (`reqly openapi validate/explore/generate`, Desktop explorer with Try in Builder + schema inspection) — **shipped**
34. ~~**API diff & breaking-change detection**~~ — endpoints/params/schemas/auth/response-types + spec/request/response/env diff polish (`reqly diff [--fail-on-breaking]`, severity classification in core + Desktop DiffView) — **shipped**
35. ~~**Contract testing + schema validation**~~ — OpenAPI/JSON Schema response validation pipeline (`internal/testing.AssertJSONSchema`, `internal/jsonschema.Validate`, response contract checks) — **shipped**
36. ~~**Advanced HTTP / Proxy & TLS controls**~~ — HTTP/2, per-env/per-request proxy, cert inspection, mTLS, custom CAs (`internal/request.Client` proxy & TLS support, Desktop Proxy & TLS settings) — **shipped**
37. ~~**Performance testing (lightweight)**~~ — RPS/latency P95/P99/error-rate/status-distribution (`internal/perf.Run`, `reqly perf run`, Desktop PerfView) — **shipped**
38. ~~**API Monitoring Dashboard (§57.1)**~~ — scheduled health checks + availability % / avg latency + CLI runner (`internal/monitor.Run`, `reqly monitor run`, Desktop MonitorView) ([ADR 0033](docs/adr/0033-monitor-scheduler.md), `docs/spec/m49-monitor-scheduler.md`) — **shipped**
39. ~~**Plugin Engine & Marketplace (§58.1)**~~ — Goja JS program compilation, manifest validation, CLI manager (`internal/plugin.Load`, `reqly plugin list/validate`) — **shipped**
40. ~~**Model Context Protocol Server & AI (§59.1, §59.3)**~~ — JSON-RPC 2.0 stdio server, tool definitions (`list_requests`, `search_requests`, `get_request`, `run_request`), CLI runner (`internal/mcp.Serve`, `reqly mcp serve`, `internal/ai`) — **shipped**


---

# Lower-precedence desktop GUI execution roadmap

This section preserves `gui-roadmap.md` in full. It is subordinate to the development roadmap above. GUI status can clarify implementation state, but it cannot redefine product scope or phase priority.

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

- [x] Logo — 2026-08-27
- [x] Workspace Switcher (folder open) — 2026-08-27
- [x] Global Search ⌘K — 2026-08-27
- [x] Import / Export buttons — 2026-08-27
- [x] Active Environment selector (Ticket #12) — 2026-08-27
- [x] Sync Status indicator (Git local-first save indicator) — 2026-08-27
- [x] Settings — 2026-08-27

### §2.2 Tool Rail (48–56px, left-most)

- [x] Workspace group: Home, Requests, Environments, History — 2026-08-27
- [x] API Tools group: Mocks, Diff, JWT, GraphQL, gRPC, Runners, Explorer, Docs — 2026-08-27
- [x] Realtime group: WebSocket, SSE — 2026-08-27
- [x] System group: Settings — 2026-08-27
- [x] Icon-based routing (top-level navigation) — 2026-08-27

### §2.3 Context Sidebar (220–280px)

- [x] Collapsible/resizable (drag handle) — 2026-08-27
- [x] Changes per active tool — 2026-08-27
- [x] Tree navigation — 2026-08-27
- [x] Search within tool — 2026-08-27
- [x] Contextual actions — 2026-08-27
- [x] Recent/pinned items (History recents & realtime endpoints) — 2026-08-27
- [x] `⌘B` toggle — 2026-08-27

### §2.4 Main Workspace

- [x] Tab-based content area — 2026-08-27
- [x] Page routing per active tool — 2026-08-27
- [x] Full pages vs context panels (§62 rules) — 2026-08-27

### §2.5 Bottom Utility Panel

- [x] Console tab — 2026-08-27
- [x] Network tab — 2026-08-27
- [x] Tests tab — 2026-08-27
- [x] Variables tab — 2026-08-27
- [x] Cookies tab — 2026-08-27
- [x] `⌘J` toggle — 2026-08-27
- [x] Resizable height — 2026-08-27

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

---

## Code Review Gate (`/code-review` — two-axis)

- [x] Standards: `oxlint` + `gofmt`/`go vet` + `anti-slop` + Fowler smell baseline — `git diff main...HEAD` (three-dot, merge-base) — no `as` without `// SAFETY:`, no hard violations
- [x] Spec: this milestone (`Milestones/` + Phase) vs implementation (`ROADMAP.md` DoD: core+UI/CLI+tests) — `git log main..HEAD` + `git diff main...HEAD` — both axes must be green before ticking `[x]` above; fix `main...HEAD` diff until green — run `/code-review`
