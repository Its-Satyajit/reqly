# 06: Keyboard map — tool jumps and send shortcut

**What to build:** The app is fully mouse-optional: ⌘1–8 jump to the primary rail tools (home, requests, environments, history, then the API-tools group in rail order), and ⌘⏎ sends the active request, delegating to the request editor's send seam. Every shortcut is listed on the reference card in Settings → About.

**Blocked by:** 05 (the reference card lives there).

**Status:** ready-for-agent

- [ ] One window-level key handler owns ⌘K/⌘1–8/⌘⏎ alongside existing ⌘W/⌘B handling (no duplicate listeners)
- [ ] ⌘1–8 navigate through the shared view-switch guard
- [ ] ⌘⏎ sends the active request when a request tab is active; no-op otherwise
- [ ] Shortcuts never fire while typing in inputs where they'd conflict (verified against editor surfaces)
- [ ] Reference card matches implemented behavior exactly — no documented-but-dead shortcuts

Parent: #369
