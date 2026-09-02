# 01: Shell Foundation

**What to build:** The persistent application shell chrome — TopBar, ToolRail, ContextSidebar, BottomPanel, StatusBar — and the App.tsx layout using ResizablePanelGroups. All shell components render, navigate between pages, and respond to user interactions.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

## Acceptance Criteria

- [ ] TopBar renders with 7 components: Logo, Workspace Switcher, Global Search, Import, Export, Settings
- [ ] ToolRail renders with page icons grouped into Workspace, API Tools, Realtime, System sections
- [ ] ToolRail collapses from 56px (icons+labels) to 40px (icons-only) via toggle button
- [ ] ContextSidebar renders at 220–280px, collapses via Ctrl/Cmd+B
- [ ] ContextSidebar content changes based on active page (placeholder for each page)
- [ ] BottomPanel renders with 5 tabs: Console, Network, Tests, Variables, Cookies
- [ ] StatusBar renders with Git branch, ahead/behind, dirty state, active environment
- [ ] App.tsx uses two ResizablePanelGroups: sidebar|main and main|bottom
- [ ] Layout persists across restarts via useDefaultLayout()
- [ ] ToolRail navigation switches activeView in useWorkspaceStore
- [ ] All shell chrome uses theme tokens (no hardcoded colors)
- [ ] All interactive elements support keyboard navigation
- [ ] Error boundaries wrap every panel
