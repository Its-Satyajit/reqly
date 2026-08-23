# M44 T2 — Shell & layout framework

Blocked by: T1.
Blocks: T4, T5, T6, T7.

## Goal

Rebuild the app shell natively per the atlas shell in the demo engine: sidebar / tabstrip / viewhost / inspector / statusbar / commit strip, with resizable gutters.

## Requirements

- `AppShell` layout component mirroring `engine.js` `SHELLS.atlas` structure (header with brand + workspace pill + ⌘K button + branch chip + environment pill + settings; body = sidebar | gutter | tabstrip+viewhost | gutter | inspector; statusbar footer).
- Resizable gutters (`data-split="side" | "inspector"`), persisted widths.
- Inspector is a mount point views can populate (request/response context); collapsible.
- Commit strip slot reserved (populated by T7).
- Existing routing/navigation maps onto the tabstrip (open/close/duplicate tabs, `st-*` state classes for loading/ok/redirect/error tabs).

## Acceptance criteria

- [x] All current feature views render inside the new shell without visual regressions in behavior. — AppShell extraction, 2026-08-25
- [x] Gutter resize + persistence works; keyboard accessible separators. — sidebar + main split persisted (`data-split` handles); inspector width persistence lands with its first consumer (T5). 2026-08-25
- [x] Theme tokens only — no hardcoded colors (T1 gate). — 2026-08-25
- [x] typecheck + lint + React Doctor clean vs baseline; vitest 19/19. Inspector mount point ships scaffolded (useShellStore), first consumer in T5; commit-strip slot reserved per spec. — 2026-08-25

## Reference

- `shared/engine.js` → `SHELLS` registry (atlas shell template), tab management, gutter logic
- `design-9-git-native/styles.css` → `.atl*` classes
