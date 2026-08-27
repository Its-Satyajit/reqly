# 06: History Page

**What to build:** The History page for reviewing and replaying past requests. Shows history table with filters and configurable retention.

**Blocked by:** #01 (Shell Foundation)

**Status:** ready-for-agent

## Acceptance Criteria

- [ ] History Page renders when activeView is 'history'
- [ ] History table shows: Method, URL, Status, Duration, Environment, Timestamp
- [ ] Table is sorted by timestamp (newest first)
- [ ] Filters: Method dropdown, Status dropdown, Environment dropdown, Date range
- [ ] Search input filters by URL
- [ ] Clear button removes all history (with confirmation)
- [ ] Row click opens request in Request Builder
- [ ] Context menu: Replay, Duplicate, Save to Collection, Delete
- [ ] Configurable retention in Settings → Storage → History Retention
- [ ] Retention options: 30 days, 90 days (default), 1 year, Forever
- [ ] History entries older than retention are cleaned up on app start
- [ ] All components use theme tokens
