# 12: Search

**What to build:** Global Search triggered by Cmd+K/Ctrl+K. Searches across requests, collections, environments, history, and commands.

**Blocked by:** #01 (Shell Foundation)

**Status:** ready-for-agent

## Acceptance Criteria

- [ ] Cmd+K / Ctrl+K opens search modal
- [ ] Search modal shows input with focus
- [ ] Escape closes the modal
- [ ] Search results grouped by type: Requests, Collections, History, Commands
- [ ] Each result shows: Icon, Name, Context (collection name, timestamp, etc.)
- [ ] Arrow keys navigate results
- [ ] Enter selects the highlighted result
- [ ] Click selects a result
- [ ] Selecting a Request opens it in Request Builder
- [ ] Selecting a Collection opens it in Collections Explorer
- [ ] Selecting a History item opens it in Request Builder
- [ ] Selecting a Command executes the command
- [ ] Search is fuzzy (typo-tolerant)
- [ ] Recent searches are shown when modal opens (before typing)
- [ ] All components use theme tokens
