# 02: Workspace Home

**What to build:** The workspace landing page shown when no request is open. Displays stat cards, Quick Actions, Recent Requests, and an empty state onboarding for new workspaces.

**Blocked by:** #01 (Shell Foundation)

**Status:** ready-for-agent

## Acceptance Criteria

- [ ] Home Page renders when activeView is 'home'
- [ ] Stat cards show: Requests count, Environments count, Collections count, Recent Activity (requests today)
- [ ] Quick Actions render: New Request, Import, Open Collection, New Environment
- [ ] Recent Requests list shows last 10 requests with Method, URL, Status, Duration
- [ ] Empty state shows when workspace has no data: "Welcome to Reqly" with quick-start actions
- [ ] Empty state actions: "Create your first request", "Import an API spec", "Set up an environment"
- [ ] Stat cards are hidden when counts are zero (empty state replaces them)
- [ ] Quick Actions trigger the correct navigation or dialog
- [ ] Recent Requests are clickable to reopen the request
- [ ] Home Page uses theme tokens for all colors
