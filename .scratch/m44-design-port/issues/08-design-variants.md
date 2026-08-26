# M44 T2 — Six-variant design layer (UI exploration)

Blocked by: T1 (token architecture). Blocks: nothing (exploration outcome feeds T3–T7).

## Goal

Ship six switchable UI designs for the Reqly desktop app: the existing base design plus
five completely different, brand-colored designs (not subtle token tweaks). The first
exploration pass (Modern/IDE/Inspector/Minimal/Command Center) was rejected by the user
as too similar; it was collapsed and rebuilt as Current + five bold concepts. After all
six ship, the user picks one; the winner is improved and the rest are removed.

## The six designs (PRs #361–#363, #365, #367, #368 — 2026-08-26)

| # | Name | Personality / signature |
|---|------|-------------------------|
| 1 | Current | The token base itself; no overrides. |
| 2 | Ember | Orange activity-rail spine (solid gradient, dark glyphs), warm-tinted hairlines, brand focus rings. |
| 3 | Forge | Dark-first terminal energy; orange as structural light — glowing focus, luminous seams, square corners. |
| 4 | Blueprint | Technical drawing: drafting-grid canvas, hairline frames, uppercase micro-labels, orange annotation ink. |
| 5 | Signal | Status chroma as language: left-edge status bars on rows; orange reserved strictly for action. |
| 6 | Paper | Print-inspired warm paper canvas, orange ink for headings/active states, minimal borders, generous whitespace. |

## Architecture (unchanged from #367, extended in #368)

- Design axis orthogonal to themes: `data-design` on `<html>` (shape/micro-language),
  `data-theme` (colors). Registry: `frontend/src/lib/designs.ts`; per-design sheets in
  `frontend/src/styles/designs/<id>.css`, lazily imported per-case in
  `useDesignStore.loadDesignSheet()`; `styles/designs/shared.css` always loaded
  (density contract `--row`/`--gap`).
- Runtime switcher in the StatusBar (Base UI dropdown, radio items, persisted via
  `useDesignStore` → localStorage `reqly-design`). Switching only flips the attribute —
  tabs, drafts, environment, history all survive; verified in-browser.
- Dev-only browser demo (`vite` + `?demo`): in-memory adapters stand in for the Wails
  bridge so the full shell renders for design review. Compile-time gated.
- Fixed en route: infinite git-refresh loop in StatusBar when `repoFound:false`;
  **cyclic color-mix bug** — a custom property mixing into itself on the same element
  (`--border: color-mix(... var(--border))`) is cyclic and silently resolves to empty;
  all five design sheets now use literal anchors with `.dark` variants (fixed in #368,
  also repaired Ember's border/input/accent tints).

## Acceptance criteria

- [x] Six variants switchable at runtime, preference persisted. — 2026-08-26
- [x] Switching preserves active request, editor state, environment. — verified via agent-browser
- [x] No component duplication — one shared tree, token-level differentiation. — 2026-08-26
- [x] Five designs are recognizably different at a glance, each forging the brand orange. — 2026-08-26 (#368)
- [x] Every design legible in both atlas-light and atlas-dark (computed-style + screenshot verified). — 2026-08-26 (#368)
- [x] typecheck + oxlint + vitest 40/40 + react-doctor 100 + PR CI green (#368). — 2026-08-26

## Next: pick-one-then-prune

The user chooses one design; it gets improved, the other five are removed
(CSS files + registry entries + switch cases; if `current` loses, its token
set folds into `tokens.css` as the base).

## Reference

- `frontend/src/styles/designs/*.css`, `frontend/src/lib/designs.ts`,
  `frontend/src/stores/useDesignStore.ts`, `frontend/src/lib/demo-workspace.ts`
- Screenshots per design (light + dark) captured during review (browser demo at 1280×578).
