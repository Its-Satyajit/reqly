# M44 T3 — Command palette (⌘K)

Blocked by: T2.
Blocks: nothing.

## Goal

Native ⌘K command palette matching the demo: search-or-jump input, grouped results (navigation, actions, settings), full keyboard control.

## Requirements

- Global hotkey (⌘K / Ctrl+K) + header button opens it; Esc closes.
- Fuzzy match over: views/tabs, collections & requests (jump), environments (switch), git actions (branch/commit — wire to existing backend bindings), theme switching, settings sections.
- Recent-items memory; empty/loading/error result states as simulated in the demo.
- Actions are extensible via a registry other features can append to.

## Acceptance criteria

- [x] Every destination reachable from the sidebar is reachable from the palette. — WORKSPACE_VIEWS registry (`lib/views.ts`); git actions join in T4/T7. 2026-08-25
- [x] Full keyboard nav (arrows, enter, esc) with visible focus. — cmdk primitives + shadcn `Command` inside `Dialog`. 2026-08-25
- [x] Palette styling is token-driven across both shipped themes. 2026-08-25
- [x] typecheck + lint + React Doctor clean vs baseline; vitest 27/27. Recent commands persist (`reqly-palette-recent.v1`); action registry extensible via `buildCommands`. — 2026-08-25

## Reference

- `shared/engine.js` → palette implementation and its data-action wiring
- `shared/views.js` → `bindViewActs` action dispatch
