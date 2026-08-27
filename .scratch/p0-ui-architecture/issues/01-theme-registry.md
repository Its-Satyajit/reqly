# 01: Theme registry — named themes with system follow

**What to build:** The user switches appearance between named themes (`atlas-light`, `atlas-dark`) and `system` (follows the OS via `prefers-color-scheme`). The choice persists across restarts. The rail's theme button cycles light → dark → system. Themes are pure CSS-variable sets applied as a root class; components consume semantic tokens only, so adding theme #5 later is one stylesheet.

**Blocked by:** None (can start immediately).

**Status:** ready-for-agent

- [ ] Theme store holds `{ theme, resolvedTheme }` with `system` resolving through matchMedia; persisted
- [ ] Root class driven by the registry; existing light/dark token sets become the first two entries
- [ ] Rail theme button cycles light → dark → system and reflects the current mode (sun/moon/auto icon)
- [ ] Store tests: resolution, persistence, cycle order (matchMedia mocked)
- [ ] No component references concrete colors outside the token layer

Parent: #369
