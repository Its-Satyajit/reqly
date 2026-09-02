# 11: Global Panels

**What to build:** The Bottom Utility Panel tabs — Console, Network, Tests, Variables, Cookies. Each tab shows global activity across the workspace.

**Blocked by:** #01 (Shell Foundation)

**Status:** ready-for-agent

## Acceptance Criteria

### Console Panel
- [ ] Log table shows: Timestamp, Level (INFO/ERROR), Message
- [ ] INFO for normal operations (sending request, receiving response)
- [ ] ERROR for failures (timeout, connection refused, auth failure)
- [ ] Filter by level (INFO, ERROR)
- [ ] Clear button removes all logs
- [ ] Auto-scroll to bottom

### Network Panel
- [ ] Request table shows: Time, Method, URL, Status, Duration
- [ ] Sorted by time (newest first)
- [ ] Row click opens request in Request Builder
- [ ] Filter by method, status

### Tests Panel
- [ ] Summary: Passed count, Failed count, Skipped count
- [ ] Filter by: Request, Collection, File, Environment, Date
- [ ] Test list shows: Test name, Status (Pass/Fail/Skipped), Duration
- [ ] Failed tests show error message

### Variables Panel
- [ ] Variables grouped by scope: Global, Workspace, Environment, Collection, Request, Runtime
- [ ] Each variable shows: Name, Value, Source scope
- [ ] Secret values are masked (••••••••)
- [ ] Search/filter variables

### Cookies Panel
- [ ] Cookie table shows: Domain, Path, Name, Value, Secure, HttpOnly, SameSite, Expires
- [ ] Search/filter cookies
- [ ] Delete individual cookies
- [ ] Clear all cookies button
- [ ] Export cookies button

- [ ] All components use theme tokens
