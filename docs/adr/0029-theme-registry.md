# ADR 0029 — Theme registry architecture (M44 T1)

Date: 2026-08-25 · Status: Accepted · Ticket: `.scratch/m44-design-port/issues/01-theme-token-architecture.md`

## Context

The approved ui-demos prototype themes orthogonally to layout: every shell renders `theme-${OPTS.theme}` with semantic CSS variables on `:root`. The user requires more themes over time — beyond `light`/`dark`/`system` and Atlas Dark. The pre-existing frontend had a binary light/dark store toggling a `.dark` class with shadcn tokens inline in `index.css`.

## Decision

1. **Named-theme registry** (`frontend/src/lib/themes.ts`): open list of `{ id, label, appearance }`; ships `atlas-dark`, `atlas-light`. Preference is `'system' | ThemeId`; a pure `resolveTheme()` maps preference + OS dark to a theme id.
2. **Token contract in one file** (`frontend/src/styles/tokens.css`): per-theme blocks selected by `[data-theme='<id>']`; `.dark` is kept as an appearance mirror so Tailwind's `@custom-variant dark` works for any number of dark themes without per-theme variants. Components consume tokens only; hex values live nowhere else (grep gate).
3. **Controller/store split** (`useThemeStore.ts`): framework-agnostic `createThemeController()` (DOM application, persistence, live OS tracking via `matchMedia('change')`) with the zustand store as a thin reactive wrapper. This keeps theme behavior testable without React and jsdom-free in principle.

Adding theme N+3 = one `[data-theme]` token block + one registry entry. Zero component changes.

## Consequences

- Old stored values (`'light'`/`'dark'`) are not migrated; unknown stored ids resolve to the default (`atlas-dark`) via `resolveTheme`.
- Toggle always resolves to an explicit theme id (persisted); choosing "system" again re-enables OS tracking (Settings surface arrives with T6's settings sweep).
- Test infra (vitest + jsdom) was stood up as part of T1 since no frontend tests existed; seams: `themes.test.ts`, `useThemeStore.test.ts`.

## References

- `docs/internal/ui-demos/shared/engine.js` (shell × theme boot config)
- `docs/spec/m44-design-port.md`
