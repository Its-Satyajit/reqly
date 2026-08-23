# M44 T1 — Token & theme architecture

Blocked by: nothing (first ticket).
Blocks: T2–T7.

## Goal

Port the ui-demos design-9 token system into the React frontend as an extensible multi-theme architecture. The demo's 26 semantic CSS variables (`--bg`, `--layer1/2`, `--pop`, `--muted`, `--accent`, `--line`, `--line-strong`, `--input-bg`, `--text`, `--text-2`, `--primary`, …) become the canonical token contract. Atlas Dark is one named theme in a registry — NOT a hardcoded palette.

## Requirements

- `frontend/src/styles/tokens.css`: semantic variables only on `:root` / `[data-theme="…"]`. No component may hardcode a color, shadow, radius or font.
- Theme registry: named themes → token sets. Ship `atlas-dark` and `atlas-light`; registry is open (future themes = new token set + entry). `system` resolves to OS preference and tracks changes live.
- `ThemeProvider` + store slice: persisted selection (`atlas-dark | atlas-light | system`), applies `data-theme` on `<html>`.
- Port typography scale, spacing rhythm, radii, elevation/shadows from `design-9-git-native/styles.css`.

## Acceptance criteria

- [x] Switching theme at runtime re-skins the entire app without remount. — 2026-08-25
- [x] Adding a hypothetical third theme requires zero component changes (documented in ticket closure / ADR). — ADR 0029
- [x] No hardcoded hex values outside the token files (grep gate) — tokens live in `frontend/src/styles/tokens.css`. — 2026-08-25
- [x] typecheck + lint + React Doctor clean vs baseline; vitest 14/14. — 2026-08-25

## Reference

- `docs/internal/ui-demos/design-9-git-native/styles.css` (`:root` token block)
- `docs/internal/ui-demos/shared/engine.js` (`theme-${OPTS.theme}` shell pattern; boot config `{ shell:'atlas', theme:'dark' }`)
