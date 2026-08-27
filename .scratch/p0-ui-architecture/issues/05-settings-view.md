# 05: Settings view — appearance, workspace, history, about

**What to build:** A Settings page in the System group with four sections: Appearance (theme picker over the registry from ticket 01), Workspace (name, path, open/switch folder), History (entry count from the adapter; retention shown only if exposed), About (app version). Everything is read-only where no backend binding exists — no placeholder toggles.

**Blocked by:** 01 (theme registry must exist for the Appearance picker).

**Status:** ready-for-agent

- [ ] `settings` view added to the union and reachable from the rail's system slot
- [ ] Appearance section lists registered themes incl. `system`; selection persists immediately
- [ ] Workspace section shows name/path and offers switch-folder via the existing bootstrap seam
- [ ] History section shows entry count from the adapter; graceful error text if unavailable
- [ ] About section shows version; includes the keyboard shortcut reference card
- [ ] No control renders without a working binding behind it

Parent: #369
