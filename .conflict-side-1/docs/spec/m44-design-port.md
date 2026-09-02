# M44 Spec — Design port (G-17.2): Atlas design language + multi-theme architecture

Source prototype: `docs/internal/ui-demos/design-9-git-native/` + shared engine (`shared/{data,engine,views,views2}.js`, ~2,500 LOC vanilla JS).

## Context

The approved prototype is a generic engine: a `SHELLS` registry (workbench / terminal / blueprint / atlas / monograph / git) over one shared view+state layer, themed by orthogonal named themes rendered as `theme-${OPTS.theme}` with 26 semantic CSS variables on `:root`. `design-9` boots `{ shell: 'atlas', theme: 'dark' }`.

Decision: the React port adopts the **token + registry architecture**, not just the Atlas skin. Themes are an open set (`atlas-dark`, `atlas-light` now; more later); shells inform layout but only the atlas shell is implemented in M44.

## Data model / API surface

- Token contract: semantic CSS custom properties (bg, layer1/2, pop, muted, accent, line(+strong), input-bg, text(-2), primary, …). Components consume tokens only.
- Theme store slice: `{ theme: 'atlas-dark' | 'atlas-light' | 'system', resolvedTheme }`; persisted; `system` tracks OS.
- Command/action registries for palette extensibility.

## Edge cases

- System theme change while running → live re-skin without remount.
- Adding theme N+3 must not touch components (grep gate on hardcoded colors).
- Conflict/dangerous git actions confirm-guarded (T7).

## Testing strategy

- Unit: theme resolution/persistence, palette fuzzy match, gutter persistence.
- Component: shell states (loading/empty/error), sidebar status glyphs.
- E2E-ish manual gates per ticket against a scratch git repo (T4/T7).
- Every ticket: typecheck + lint + React Doctor clean vs baseline.

## Tickets (in `.scratch/m44-design-port/issues/`)

1. T1 token & theme architecture ← first
2. T2 shell & layout framework (blocks T4–T7)
3. T3 command palette
4. T4 git-native collections sidebar
5. T5 request builder & response viewer restyle
6. T6 secondary views sweep
7. T7 worktree panel & merge resolver (closer)

## Visual parity audit — 2026-08-25 (post-T7)

Compared the live app (wails3 dev + agent-browser screenshots + user screenshots of the real workspace) against the reference.

- Reference: [`ui-demos/design-9-git-native/screenshots/`](../ui-demos/design-9-git-native/screenshots/) — key frames: [`01-overview`](../ui-demos/design-9-git-native/screenshots/01-overview.png), [`02-rest-client`](../ui-demos/design-9-git-native/screenshots/02-rest-client.png), [`19-git-panel`](../ui-demos/design-9-git-native/screenshots/19-git-panel.png), [`20-settings`](../ui-demos/design-9-git-native/screenshots/20-settings.png)
- Current app (2026-08-25): `/home/satyajit/Pictures/Screenshots/app/` (14 captures of the real workspace)

**Verdict: architecture shipped, visuals did not.** Every T1–T7 mechanism works (theme registry, shell, palette, git sidebar, settings, commit strip, conflict resolver), but the views still render the pre-M44 flat layout. The app is not visually recognisable as the Atlas reference.

### Gap list (drives G-17.3)

| # | Area | Reference | Current |
|---|------|-----------|---------|
| 1 | Left edge | Icon activity rail (views as icons, settings pinned bottom) | none — views are a text list in the sidebar |
| 2 | Header | Workspace pill (name + env chip + branch chip), centered ⌘K search, icon actions | text logo + env dropdown + sun toggle |
| 3 | Sidebar | COLLECTIONS label, filter input, method-chip request tree, source-control panel with ±counts, Folder/Request buttons | flat nav list + basic tree + plain git panel |
| 4 | Builder | Prominent orange Send, Save + ⋯, tabs with count badges, KV table with column headers | small controls, cramped row |
| 5 | Response | "Ready to send" hero empty state (icon, CTA, recent chips), header with resolved URL | bare text line |
| 6 | Statusbar | branch / env / request count / zero-telemetry left, console slot right | text strip (view + env) |
| 7 | Home | Dashboard cards: quick actions, recent activity, environment, repo snapshot, protocol clients | none |

### Root cause

T1–T7 were scoped and reviewed against *mechanism* acceptance criteria (stores, bindings, states, gates). Nothing in the ticket gates compared rendered output to the reference. Process fix: visual-parity tickets must include a screenshot-diff step (agent-browser capture vs reference) as an acceptance criterion.

### Decision

- G-17.2 reverted to `[~]`; G-17.3 opened with the layered plan above (shell chrome → sidebar → builder → response → statusbar → overview).
- The M44 architecture branches remain valid groundwork; G-17.3 builds the visual layer on top of them.
