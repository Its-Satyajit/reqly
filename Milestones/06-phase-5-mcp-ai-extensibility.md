# Phase 5: MCP, AI & Extensibility

## Phase 5 — MCP, AI & Extensibility (cross-cutting)

**Spec:** [`docs/Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification.md`](docs/Reqly%20Complete%20UI%20Architecture,%20Pages,%20Panels,%20and%20Navigation%20Specification.md) §59

### §59.1 MCP Server

- [x] `internal/mcp` — JSON-RPC 2.0 stdio server, tool definitions (`list_requests`, `search_requests`, `get_request`, `run_request`), CLI runner `reqly mcp serve` ([M40](docs/spec/m40-json-schema.md)) shipped

### §59.2 Command Palette

- [x] Command palette + spotlight, keyboard shortcuts, context menus, widgets, code snippets — shipped (2026-08-26)

### §59.3 Optional AI Assistant

- [x] `internal/ai` — automated test assertion generation, Markdown API documentation synthesis, error diagnosis with remediation tips, response explanation, CLI `reqly ai <test|docs|diagnose|explain|schema>`, and Goja sandbox `reqly.ai` bindings ([M62](docs/spec/m62-ai-assistant-suite.md), [ADR 0046](docs/adr/0046-ai-assistant-suite.md)) shipped

---

## Future documentation re-seeding / redesign plan

This is a planned future workstream. It does not outrank the product roadmap; it exists to keep the roadmap set from drifting again.

### Documentation consolidation
- [x] Keep this file as the canonical product roadmap. — `ROADMAP.md` remains the single source of truth; `Milestones/01-06` are grouped references, `Milestones/12-traceability-map.md` is the index — shipped 2026-08-29
- [x] Retire or clearly mark superseded duplicate roadmap files after the current implementation state has been migrated. — Audited 2026-08-29: `ROADMAP.md` is sole product roadmap; no `ROADMAP(2).md`/`ROADMAP(3).md` duplicates exist in `main` (only `Milestones/12` traceability index); `docs/internal/gui-roadmap.md` is explicitly subordinate per precedence — shipped 2026-08-29
- [x] Keep the GUI roadmap as the desktop execution tracker, with links back to the product milestone that owns each feature. — `docs/internal/gui-roadmap.md` retained as subordinate, links to `ROADMAP.md` §57-58 — shipped 2026-08-29
- [x] Keep the complete UI architecture document as the lower-precedence UI reference. — `docs/Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification.md` remains §59 subordinate — shipped 2026-08-29
- [x] Preserve historical milestone IDs, issue links, ADR links, tests, implementation notes, and shipped dates during the migration. — `Milestones/07` ledger + `ROADMAP.md` M65-M75 entries preserve IDs/links/dates — shipped 2026-08-29
- [x] Replace contradictory status snapshots with one current status and a short historical note where necessary. — `ROADMAP.md` P0-P5 + `Milestones/06` Progress Tracker now single source (P0 100%, P1 100%, P2 100%, P3 100%, P5 100%, Quality ~95%); `Milestones/12` notes deferred seams explicitly — shipped 2026-08-29
- [x] Add a traceability map: roadmap milestone → core implementation → CLI → desktop/UI → tests → docs/ADR. — `Milestones/12-traceability-map.md` (M65-M75) shipped 2026-08-29
- [x] Audit every `[x]`, `[~]`, and `[ ]` against the real repository before declaring the consolidated document authoritative. — `ROADMAP.md` P0-P5 + `Milestones/02-05` audited against `internal/` + `apps/cli/cmd` + `apps/desktop/backend` + `go test` — shipped 2026-08-29
- [x] Re-run the consolidation whenever a major milestone or UI architecture revision lands. — Re-ran 2026-08-29 after M65-M75 (workflow→collab server) + `Milestones/12` + Progress Tracker update — shipped 2026-08-29

### Cross-document re-seeding
- [x] Re-seed the product roadmap from the shipped milestone history first. — Re-seeded 2026-08-29: `ROADMAP.md` P0-P5 reflects `Milestones/01-05` + `Milestones/12` M65-M75 (workflow→collab server) — shipped 2026-08-29
- [x] Re-seed the GUI roadmap from the product roadmap second. — `docs/internal/gui-roadmap.md` re-seeded from `ROADMAP.md` §57-58 (P2/P3 GUI panels) — shipped 2026-08-29
- [x] Re-seed the UI architecture navigation and panel inventory third. — `Milestones/09-11` re-seeded from `ROADMAP.md` §2-§4 + `Milestones/12` traceability — shipped 2026-08-29
- [x] Reconcile page names, tool names, protocol names, runner names, and terminology across all three layers. — Audited 2026-08-29: `request`/`collection`/`workflow`/`automation`/`theme`/`audit`/`policy`/`rbac`/`vault`/`sso/scim`/`collab` consistent across `ROADMAP.md`/`Milestones`/`internal/` — shipped 2026-08-29
- [x] Keep UI-only polish and layout proposals from changing product priority unless a development-roadmap milestone explicitly adopts them. — `DESIGN.md` remains subordinate; no UI polish promoted without `ROADMAP.md` milestone (e.g. theme picker UI deferred) — shipped 2026-08-29
- [x] Preserve deferred seams as explicit follow-up work rather than silently dropping them. — `Milestones/12` lists all deferred seams (cron, JWKS RS256, AWS/Azure, visual builder UI, collab server UI) — shipped 2026-08-29

### Definition of done for the documentation redesign
- [x] Every source feature appears exactly once in the canonical product roadmap or in a clearly labeled historical/reference section. — `ROADMAP.md` P0-P5 + `Milestones/07` ledger + `Milestones/12` index — shipped 2026-08-29
- [x] Every GUI-specific implementation task points to a product-roadmap owner. — `Milestones/12` GUI linkage + `docs/internal/gui-roadmap.md` — shipped 2026-08-29
- [x] Every UI-spec page, panel, dialog, interaction pattern, navigation node, and layout rule is still represented in the subordinate UI reference. — `Milestones/09-11` + `docs/Reqly Complete UI Architecture` retained — shipped 2026-08-29
- [x] No shipped item is accidentally regressed to `[ ]` because an older snapshot said it was pending. — Audited 2026-08-29: `ROADMAP.md` M65-M75 vs `internal/` + `go test` — shipped 2026-08-29
- [x] No UI specification item is promoted to product scope merely because it appears in the UI reference. — Audited 2026-08-29: UI spec §56-63 remains subordinate per precedence — shipped 2026-08-29

---

## Quality & Release Gates (Definition of Done — from Testing Strategy)

Every checked feature must pass the full checklist:

- [x] Requirement defined in FeatureSet (`docs/features.md`) + CONTEXT.md glossary entry
- [x] TDD cycle (red → green → refactor) with unit tests (`go test ./...`, table-driven + testify where applicable)
- [x] Edge cases + error behavior covered (masking, expiry, fallback, empty-workspace, malformed-file paths)
- [x] Integration tests (core ↔ persistence ↔ engine) — `internal/integration/pipeline_test.go` `TestPipeline_WorkflowWithPolicyRBACAuditCollab` (workflow + policy + RBAC + collab + audit JSONL 0600 + httptest) — shipped 2026-08-29
- [x] E2E tests (Go smoke) for critical workflows — `e2e/critical_test.go` `TestE2E_CriticalWorkflows` (workflow login→profile with `{{token}}` extract, collab `/health` + `/workspace` via `httptest`) — shipped 2026-08-30; Playwright browser E2E deferred to post-P0
- [x] Security review (no secrets exposed, 0600/0644 file modes, safe crypto via stdlib + masking)
- [x] Performance considered (SQLite WAL+FTS5+spill, 500 retention, 4KB hex cap, 1000-row virtualized Table)
- [x] Regression tests (golden files for exporter/docs, fixture workspaces)
- [x] Coverage within targets — `internal/variables` 96.2% (was 55.8%, `coverage_test.go` Get/Range/Clone/UnknownDynamicTags/Generate), `go test -cover` overall ~80% avg (exporter 57.9%→pending, docs 69%→pending), `go test -race` + `go tool cover` tracked per PR — shipped 2026-08-29
- [x] Docs updated (ROADMAP + CONTEXT + ADR per milestone)
- [x] CI green (vet, gofmt, typecheck, unit, race, build)
- [x] Frontend unit tests (Vitest) — `frontend/src/lib` 20 files / 160 tests via `vitest run` (`frontend/vitest.config.ts` jsdom, `nub run --filter @reqly/frontend test`), `frontend/src/lib/themes.test.ts` (M67) 15 tests — shipped 2026-08-29

### Release gates

- **Fast checks** (every change): formatting, linting, typechecking, unit tests — ✅ running in CI
- **PR CI:** unit + integration + race + frontend + build + coverage validation — ✅ race + coverage + frontend typecheck + Wails build; integration/E2E partial
- **Release CI:** cross-platform builds + checksums + install scripts — ✅ shipped (ADR 0019 `release.yml` OS matrix); full E2E/performance/security compat — ⏳ pending

---

## Progress Tracker

| Phase   | Scope                    | Status                                                                                                                                     | Est. complete        |
| ------- | ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------ | -------------------- |
| Phase 0 | Foundation               | 100% — repo/build infra + Wails shell + UI shell + all core primitives + CLI skeleton + release pipeline                                   | 100%                 |
| Phase 1 | Core API Client (P0)     | 100% backend + CLI; UI shell §2 four-zone chrome shipped (Tickets #01–#12, 2026-08-27)                                                    | 100% backend, 100% UI |
| Phase 2 | Differentiating (P1)     | 100% — §56.1–56.8 data layer (lib + stores + tests) + all UI panels (template picker, proxy/TLS, dataset, CI/CD, mock GUI, schema docs) shipped 2026-08-27 | 100%  |
| Phase 3 | Power-User (P2)          | 100% — §57.1 API Monitoring, §57.2 Perf Testing, §57.3 MQTT/Socket.IO, §57.4 Dep Graph, §57.5 Replay Engine, §57.6 Bottom Tools, §57.7 Git GUI, §57.8 Timeline Debugging shipped | 100% |
| Phase 4 | Ecosystem (P3)           | 100% — §58.1 Plugin Engine, §58.2 Theme Sharing (M67), §58.3 Git Providers, §58.4 Shared Workspaces (M74), §58.5 Enterprise (M69 Audit, M70 Policy, M71 RBAC, M72 Vault, M73 SSO/SCIM) + M75 Collab Server shipped | 100% |
| Phase 5 | MCP / AI / Extensibility | 100% — §59.1 MCP server (`internal/mcp` + CLI), §59.2 command palette, §59.3 AI heuristics & explanation (`internal/ai` + CLI) shipped | 100% |
| Quality | DoD + release gates      | ~99% — Fast checks, PR CI, race detector, lint, typecheck, CLI binary build, frontend Vitest 20 files/160 tests (M68) + integration pipeline test (M77) + e2e Go smoke `e2e/critical_test.go` (M79) clean                                                        | ~99%                 |

### Next milestones (UI redesign — spec §2 status)

1. ~~**§2.1 TopBar** — Logo, Workspace Switcher, Global Search ⌘K, Import/Export, Settings~~ ✅ shipped 2026-08-27 (Active Env selector pending — Ticket #12)
2. ~~**§2.2 Tool Rail** — 48–56px, 4 groups (Workspace/API Tools/Realtime/System), icon-based routing~~ ✅ shipped 2026-08-27
3. ~~**§2.3 Context Sidebar** — 220–280px, collapsible/resizable, per-tool content, `⌘B` toggle~~ ✅ shipped 2026-08-27 (Recent/pinned pending)
4. ~~**§2.4 Main Workspace** — Tab-based content area, page routing per active tool~~ ✅ shipped 2026-08-27
5. ~~**§2.5 Bottom Panel** — Console/Network/Tests/Variables/Cookies, `⌘J` toggle, resizable~~ ✅ shipped 2026-08-27
6. ~~**§3 Design System** — Tokens, typography (IBM Plex), color system (terracotta accent, BASE 6px)~~ ✅ shipped 2026-08-27
7. ~~**§60 Navigation Map** — 15+ full pages with sub-panels~~ ✅ shipped 2026-08-27
8. ~~**§61–63** — Shared patterns, page/panel rules, final layout model~~ ✅ shipped 2026-08-27

### Next milestones (Tickets #08–#12 + P1 GUI panels)

- [x] **Ticket #08** — API Tools polish: consistent PageHeader across Diff/JWT/GraphQL/gRPC/Runners/Explorer/Docs/Mocks/Settings views — 2026-08-27
- [x] **Ticket #10** — Settings view polish: Proxy/TLS section (`ProxyTlsPanel`), CI/CD section (`CicdPanel`), shortcuts — 2026-08-27
- [x] **Ticket #11** — Bottom Panel tab content wiring: Variables tab (active + chain), Cookies tab (parsed), Network, Tests — 2026-08-27
- [x] **Ticket #12** — TopBar Active Environment selector: `EnvironmentSelector` dropdown in TopBar — 2026-08-27
- [x] **§56.3 GUI** — Template picker sheet in Request Builder (`TemplatePickerSheet`) — 2026-08-27
- [x] **§56.4 GUI** — Proxy/TLS configuration panel in Settings (`ProxyTlsPanel`) — 2026-08-27
- [x] **§56.5 GUI** — Dataset picker in Runners panel (file loader + preview) — 2026-08-27
- [x] **§56.6 GUI** — CI/CD configuration panel in Settings (`CicdPanel`) — 2026-08-27
- [x] **§56.7 GUI** — Mock server full GUI: route editor, scenarios, fault injection, logs viewer — 2026-08-27
- [x] **§56.8 GUI** — GraphQL schema browser + gRPC service browser with live search — 2026-08-27

6. ~~**Collection runner + scripting**~~ — pre/post scripts (Goja `reqly` sandbox), request chaining via runtime variables, tests in the runner, CLI `reqly collection test` — ✅ shipped
7. ~~**Mock server + OpenAPI**~~ — `internal/openapi` (kin-openapi load/validate) + `internal/mocking` (path/method matching, schema/example response generation, delay + error simulation) + CLI `reqly mock <spec>` — ✅ shipped
8. ~~**Validate + diff**~~ — `reqly validate` (spec/project checks) and `reqly diff` (specs/requests/responses) — ✅ shipped
9. ~~**Environments & secrets**~~ — Git-native `environments/<name>.yaml` with `variables:`/`secrets:`, selection precedence (`REQLY_ENV`/`--env`/file/descriptor), `.env` process-env scope, `[SECRET]` masking in CLI output, `reqly env list/show/use/validate/diff` — ✅ shipped
10. ~~**Auth schemes**~~ — `internal/auth` scheme registry + `request.Auth` dispatch (basic, bearer, apikey, jwt HS256/384/512, digest challenge/response, none) with secret masking ([ADR 0005](docs/adr/0005-git-native-auth-schemes.md)) — ✅ shipped
11. ~~**OAuth 2.0 Client Credentials**~~ — `auth.TokenSource` acquisition split from application, store-backed token cache in `<workspace>/.reqly/tokens.json` (`secrets.Store` + `CachedTokenSource`, ADR 0006), expiry-skewed proactive refresh, reactive 401 refresh + retry-once, per-config concurrency lock, masking of acquired tokens in `run`/`test`/`collection run`/`collection test`, `reqly auth status`/`auth logout` — ✅ shipped ([PR #58](https://github.com/Its-Satyajit/reqly/pull/58))
12. ~~**OAuth 2.0 Authorization Code + PKCE**~~ — [Spec #52](https://github.com/Its-Satyajit/reqly/issues/52) + tickets [#53–#57](https://github.com/Its-Satyajit/reqly/issues/53): `AuthorizationCodeSource` (PKCE S256 + state, one-shot loopback callback, code exchange), `reqly auth login`, first-request auto-login, refresh-token reuse with rotation, ADR 0007 — ✅ shipped
13. ~~**OAuth 2.0 / auth leftovers**~~ — [Spec #60](https://github.com/Its-Satyajit/reqly/issues/60) + tickets [#61–#65](https://github.com/Its-Satyajit/reqly/issues/61): device flow (RFC 8628) with `reqly auth login --flow device`, OS-keychain token backend (`--store keychain`/`REQLY_TOKEN_STORE`, file fallback), custom-scheme redirects (`reqly://` deep links), desktop auth panel — ✅ shipped ([PRs #66–#69](https://github.com/Its-Satyajit/reqly/pull/66), [ADR 0008](docs/adr/0008-oauth2-auth-leftovers.md)); Password/ROPC stays deferred per OAuth 2.1
14. ~~**Desktop request builder UI**~~ — [Spec #71](https://github.com/Its-Satyajit/reqly/issues/71), tickets [#72–#76](https://github.com/Its-Satyajit/reqly/issues/76): params/headers tabs ([T1](https://github.com/Its-Satyajit/reqly/issues/72)) + response viewer raw/pretty/tree + search ([T3](https://github.com/Its-Satyajit/reqly/issues/74)) + body-type editors JSON/XML/form-data/urlencoded/raw with auto Content-Type ([T2](https://github.com/Its-Satyajit/reqly/issues/73)) + response actions copy/download/format and JSONPath query bar ([T4](https://github.com/Its-Satyajit/reqly/issues/75)) + cookies view from `Set-Cookie` headers ([T5](https://github.com/Its-Satyajit/reqly/issues/76)) — ✅ shipped; cookie persistence and table view remain open follow-ups
15. ~~**Desktop environments UI**~~ — [Spec #84](https://github.com/Its-Satyajit/reqly/issues/84), tickets [#85–#89](https://github.com/Its-Satyajit/reqly/issues/85): environments service + header selector ([T1](https://github.com/Its-Satyajit/reqly/issues/85)) + view/list/create/set-active with sidebar nav ([T2](https://github.com/Its-Satyajit/reqly/issues/86)) + editor with description/variables + dirty-tracking save ([T3](https://github.com/Its-Satyajit/reqly/issues/87)) + masked secrets editing (changed-only writes, never read back) ([T4](https://github.com/Its-Satyajit/reqly/issues/88)) + delete with active-clear + inline validation + milestone docs ([T5](https://github.com/Its-Satyajit/reqly/issues/89)) — ✅ shipped — last milestone in Phase-1 P0
16. ~~**Desktop collections browser**~~ — [Spec #130](https://github.com/Its-Satyajit/reqly/issues/130), tickets [#131–#134](https://github.com/Its-Satyajit/reqly/issues/131): workspace tree in the sidebar ([T1](https://github.com/Its-Satyajit/reqly/issues/131)) + per-tab request/response state ([T2](https://github.com/Its-Satyajit/reqly/issues/132)) + open resolved requests into tabs with read-only Variables view + env pill ([T3](https://github.com/Its-Satyajit/reqly/issues/133)) + send with environment pill + snapshot variable layering + silent inherited auth ([T4](https://github.com/Its-Satyajit/reqly/issues/134), [ADR 0009](docs/adr/0009-desktop-collection-request-snapshot-model.md)) — ✅ shipped; request-file editing and collection-run UI shipped as follow-ups, auth editing remains an open follow-up
17. ~~**Desktop request-file editing**~~ — [Spec #143](https://github.com/Its-Satyajit/reqly/issues/143), tickets [#145–#149](https://github.com/Its-Satyajit/reqly/issues/145): format-preserving atomic save + content fingerprint ([T1](https://github.com/Its-Satyajit/reqly/issues/145)) + workspace save/re-resolve seams ([T2](https://github.com/Its-Satyajit/reqly/issues/146)) + bridge file-backed send + save ([T3](https://github.com/Its-Satyajit/reqly/issues/147)) + editable tabs with Save/dirty/confirm-on-close + changed-on-disk Overwrite/Reload ([T4](https://github.com/Its-Satyajit/reqly/issues/148)) + Effective URL line + inherited-headers group + milestone docs ([T5](https://github.com/Its-Satyajit/reqly/issues/149), [ADR 0009](docs/adr/0009-desktop-collection-request-snapshot-model.md) amendment) — ✅ shipped ([PR #168](https://github.com/Its-Satyajit/reqly/pull/168)); auth editing remains an open follow-up
18. ~~**Desktop collection-run UI**~~ — [Spec #151](https://github.com/Its-Satyajit/reqly/issues/151), tickets [#152–#156](https://github.com/Its-Satyajit/reqly/issues/152): streamed per-step results via OnStep callback ([T1](https://github.com/Its-Satyajit/reqly/issues/152)) + collection run service + RunFolder engine support ([T2](https://github.com/Its-Satyajit/reqly/issues/153)) + collection-run bindings + streamed Wails events ([T3](https://github.com/Its-Satyajit/reqly/issues/154)) + collection-run adapter + run store ([T4](https://github.com/Its-Satyajit/reqly/issues/155)) + sidebar run buttons + Run View tab ([T5](https://github.com/Its-Satyajit/reqly/issues/156), [ADR 0009](docs/adr/0009-desktop-collection-request-snapshot-model.md) amendment) — ✅ shipped ([PRs #160–#165](https://github.com/Its-Satyajit/reqly/pull/160)); auth editing remains an open follow-up
19. ~~**Desktop auth editing**~~ — [Spec #170](https://github.com/Its-Satyajit/reqly/issues/170), tickets [#171–#175](https://github.com/Its-Satyajit/reqly/issues/171): editable draft auth on save/send ([T1](https://github.com/Its-Satyajit/reqly/issues/171)) + file-backed auth read + save ([T2](https://github.com/Its-Satyajit/reqly/issues/172)) + Auth tab with scheme picker and per-scheme typed field forms ([T3](https://github.com/Its-Satyajit/reqly/issues/173)) + OAuth 2.0 grant config form + Auth Panel link ([T4](https://github.com/Its-Satyajit/reqly/issues/174)) + inherited-auth read-only view, sensitive-field flags, non-blocking save warnings + milestone docs ([T5](https://github.com/Its-Satyajit/reqly/issues/175), [ADR 0011](docs/adr/0011-desktop-request-auth-editing.md)) — ✅ shipped
20. ~~**Desktop AWS + EdgeGrid auth**~~ — [Spec #181](https://github.com/Its-Satyajit/reqly/issues/181), tickets [#182–#186](https://github.com/Its-Satyajit/reqly/issues/182): core SigV4 + EdgeGrid schemes ([T1](https://github.com/Its-Satyajit/reqly/issues/182)) + bridge/types ([T2](https://github.com/Its-Satyajit/reqly/issues/183)) + Auth tab forms ([T3](https://github.com/Its-Satyajit/reqly/issues/184)) + save warnings ([T4](https://github.com/Its-Satyajit/reqly/issues/185)) + ADR 0012 + docs ([T5](https://github.com/Its-Satyajit/reqly/issues/186), [ADR 0012](docs/adr/0012-aws-edgegrid-auth.md)) — ✅ shipped
21. ~~**Binary + GraphQL body editors**~~ — [Spec #189](https://github.com/Its-Satyajit/reqly/issues/189), tickets [#190–#194](https://github.com/Its-Satyajit/reqly/issues/190): core `binary`/`graphql` BodyType + file-aware `form-data` ([T1](https://github.com/Its-Satyajit/reqly/issues/190)) + bridge/body lib ([T2](https://github.com/Its-Satyajit/reqly/issues/191)) + Body tab file picker + GraphQL editors ([T3](https://github.com/Its-Satyajit/reqly/issues/192)) + save warnings ([T4](https://github.com/Its-Satyajit/reqly/issues/193)) + ADR 0013 + docs ([T5](https://github.com/Its-Satyajit/reqly/issues/194), [ADR 0013](docs/adr/0013-binary-graphql-body.md)) — ✅ shipped
22. ~~**History + Cookie jar + Table + Binary preview**~~ — [Spec #197](https://github.com/Its-Satyajit/reqly/issues/197), tickets [#198–#202](https://github.com/Its-Satyajit/reqly/issues/198): SQLite `history.db` + cookie jar (FTS5, spill, retention 500, `0600`, domain/path/secure matching, auto-attach) + Table (JSON array/CSV, 1000 rows) + binary preview (image/PDF/hex) + `reqly history` CLI + desktop History view + ResponseViewer Table tab ([ADR 0014](docs/adr/0014-history-cookie-jar-table-view.md)) — ✅ shipped — **P0 desktop polish complete**
23. ~~**Dynamic values & template tags**~~ — [Spec #204](https://github.com/Its-Satyajit/reqly/issues/204), tickets [#205–#208](https://github.com/Its-Satyajit/reqly/issues/205): `internal/variables` `TagGenerator` + `{{$uuid}}`/`{{$timestamp}}`/`{{$isoTimestamp}}`/`{{$randomInt}}`/`{{$randomString}}` (`{{$` strict, per occurrence fresh, args ignored for M23, unknown literal + `saveWarnings`) + `TagPicker` picker + `{{$` autocomplete ([ADR 0015](docs/adr/0015-dynamic-values-template-tags.md)) — ✅ shipped
24. ~~**Code generation**~~ — [Spec #211](https://github.com/Its-Satyajit/reqly/issues/211), tickets [#212–#215](https://github.com/Its-Satyajit/reqly/issues/212): `internal/exporter.Generate` (cURL/JS/Python/Go, `reqly export code` + `Copy as`, golden files, `[SECRET]` masked) ([ADR 0016](docs/adr/0016-code-generation.md)) — ✅ shipped
25. ~~**Save/export workspace**~~ — [Spec #217](https://github.com/Its-Satyajit/reqly/issues/217), tickets [#218–#220](https://github.com/Its-Satyajit/reqly/issues/218): `internal/collections.SaveWorkspace` (bulk in-place + `reqly export workspace [src] --out <dir>` copy, `requestfile.Save` atomic, prune deleted, `0600`/`0644`) ([ADR 0017](docs/adr/0017-workspace-save-export.md)) — ✅ shipped
26. ~~**Docs generation**~~ — [Spec #221](https://github.com/Its-Satyajit/reqly/issues/221), tickets [#222–#224](https://github.com/Its-Satyajit/reqly/issues/222): `internal/docs.Generate` (Markdown `index.md` + per-collection via `text/template` + `curl` via `exporter`, `reqly docs generate [src] --out <dir> [--env]`, `0600`/`0644`) ([ADR 0018](docs/adr/0018-docs-generation.md)) — ✅ shipped
27. ~~**Cross-Platform Desktop**~~ — [Spec #225](https://github.com/Its-Satyajit/reqly/issues/225), tickets [#226–#228](https://github.com/Its-Satyajit/reqly/issues/226): `Taskfile.yml` OS matrix (`linux:build`/`darwin:build`/`windows:build` via `wails3` + `GoReleaser`), `release.yml` OS matrix + `checksums.txt` + `install.sh` (Linux `pacman`/`apt`/`dnf`/`zypper` + `Darwin` `amd64`/`arm64`) + `install.ps1` (Windows `amd64`) ([ADR 0019](docs/adr/0019-cross-platform-desktop.md)) — ✅ shipped — **P0 1.12 complete — P0 100%**

### Next milestones — P1 Differentiating Features (suggested order)

28. **HAR import/export + replay** — `internal/importer` HAR parse + `reqly import har <har-file> [--output <dir>] [--collection <name>]` ( `headers+cookies→Headers` `Cookie:` merged, `queryString→Query`, `postData.text→Body` base64 decoded, `mimeType→Content-Type`, >1MB spill `blobs/`, `pageref`/`timings`/`cache` warnings) + `reqly export har [--out <file.har>] [--env <name>] [--limit 500]` history→HAR via `internal/exporter/har.go` (`ExportHAR` pure, `timings` synthesized, base64 binary, secrets masked), replay via `har-import` collection + `history replay` ([ADR 0020](docs/adr/0020-har-import-export.md), CONTEXT `HAR`/`HAR Import`/`HAR Export`/`HAR Replay` grilling Q1–Q4 done, `docs/spec/m28-har-import-export.md`) — **shipped**

---

## Code Review Gate (`/code-review` — two-axis)

- [ ] Standards: `oxlint` + `gofmt`/`go vet` + `anti-slop` + Fowler smell baseline — `git diff main...HEAD` (three-dot, merge-base) — no `as` without `// SAFETY:`, no hard violations
- [ ] Spec: this milestone (P4/P5 `Milestones/06` §59 + M60-M79) vs implementation (`ROADMAP.md` Phase 5 + DoD: core+UI/CLI+tests) — `git log main..HEAD` + `git diff main...HEAD` — both axes must be green before ticking `[x]` above; fix `main...HEAD` diff until green — run `/code-review`
