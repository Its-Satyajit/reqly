# 04: Palette data providers — collections, environments, history

**What to build:** The palette finds your stuff, not just actions. Collections (from the workspace tree), environments, and recent history entries (from the FTS pool) appear as fuzzy-searchable results; running a collection/environment result navigates to it, a history result opens History filtered/replayed per existing seams.

**Blocked by:** 03 (needs the provider registry and palette UI).

**Status:** ready-for-agent

- [ ] Data providers registered for collections, environments, history pool
- [ ] Results ranked and labeled by kind; keyboard-selectable alongside commands
- [ ] Selecting a result performs the real navigation/open action through existing stores
- [ ] Providers read seeded stores in tests; large pools degrade gracefully (capped results)

Parent: #369
