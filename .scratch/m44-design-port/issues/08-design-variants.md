# M44 T2 — Six-variant design layer (UI exploration)

Blocked by: T1 (token architecture). Blocks: nothing (exploration outcome feeds T3–T7).

## Goal

Explore five substantially different visual designs for the Reqly desktop UI on top of the
shared component tree, keep the existing UI as the sixth option, and let the user switch
between all six at runtime without losing application state.

## Shipped (PRs #361–#363, #365, #367 — 2026-08-26)

- Design axis orthogonal to themes: `data-design` on `<html>` answers "how is it shaped",
  `data-theme` answers "which colors". Registry: `frontend/src/lib/designs.ts` +
  `[data-design]` blocks in `frontend/src/styles/designs.css` (radius, row density, gap,
  root type scale, shadows, input treatment, per-design micro-language).
- Variants: Current · Modern (soft, 15.5px, elevated popovers) · IDE (2px corners, 14px,
  shadow-free workbench) · Inspector (3px, uppercase protocol-trace eyebrows, tabular
  numerals) · Minimal (17px, borderless filled inputs) · Command Center (4px, accent kbd,
  tight focus rings).
- Runtime switcher in the StatusBar (Base UI dropdown, radio items, persisted via
  `useDesignStore` → localStorage `reqly-design`). Switching only flips the attribute —
  tabs, drafts, environment, history all survive; verified in-browser.
- Dev-only browser demo (`vite` + `?demo`): in-memory adapters (collections, env, history,
  bootstrap) stand in for the Wails bridge so the full shell renders for design review.
  Compile-time gated, tree-shaken from production.
- Fixed: infinite git-refresh loop in StatusBar when an adapter reports `repoFound:false`.

## Acceptance criteria

- [x] Six variants switchable at runtime, preference persisted. — 2026-08-26
- [x] Switching preserves active request, editor state, environment. — verified via agent-browser
- [x] No component duplication — one shared tree, token-level differentiation. — 2026-08-26
- [x] typecheck + oxlint + vitest 40/40 + react-doctor no new warnings; PR CI green. — 2026-08-26

## Reference

- `frontend/src/styles/designs.css`, `frontend/src/lib/designs.ts`,
  `frontend/src/stores/useDesignStore.ts`, `frontend/src/lib/demo-workspace.ts`
- Screenshots per variant captured during review (browser demo at 1280×578).
