# 03: Command palette core — registries plus ⌘K UI

**What to build:** ⌘K (or the top-bar search button) opens a command palette. v1 ships navigation commands (every workspace view), tab actions (new request, close active), import/export dialogs, and theme switching — all through a command registry (`{ id, title, hint?, keywords, run }`) and a data-provider registry so future features register entries without touching palette UI. Fuzzy filtering client-side (Fuse, already a dependency).

**Blocked by:** None (can start immediately). Parallel with 01 and 02.

**Status:** ready-for-agent

- [ ] Palette store with command registry and data-provider registry; registration API documented by usage
- [ ] ⌘K opens/closes; Esc closes; arrow keys + Enter run; input focused on open
- [ ] Top-bar search button opens the same palette
- [ ] Ships navigation, tab, import/export, and theme commands
- [ ] Every result row can show its shortcut hint when one exists
- [ ] Empty query shows top commands; no-match state directs ("No commands match “xyz”.")
- [ ] Store tests: register/filter/run contract, fuzzy ranking stability

Parent: #369
