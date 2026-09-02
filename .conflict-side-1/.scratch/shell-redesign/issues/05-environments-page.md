# 05: Environments Page

**What to build:** The Environments page for managing environment variables. Shows environment tabs, variables table, and validation panel.

**Blocked by:** #01 (Shell Foundation)

**Status:** ready-for-agent

## Acceptance Criteria

- [ ] Environments Page renders when activeView is 'environments'
- [ ] Environment tabs show: Local, Development, Staging, Production (configurable)
- [ ] Active environment tab is highlighted
- [ ] Variables table shows: Name, Value, Secret (checkbox), Description
- [ ] Add variable button adds a new row
- [ ] Delete variable removes the row
- [ ] Secret values are masked in the UI (••••••••)
- [ ] Validation panel shows: Required variables, Malformed URLs, Unresolved variables, Duplicate variables, Unused variables
- [ ] Validation runs on save or on demand
- [ ] Save button writes changes to environment file
- [ ] Validate button runs validation checks
- [ ] Diff button opens environment diff (two dropdowns)
- [ ] Cross-check button shows variable usage across requests/collections
- [ ] All components use theme tokens
