# 02: Realtime pages — dedicated WebSocket and SSE views

**What to build:** WebSocket and SSE get their own pages in the Realtime tool group instead of living in request tabs. Each page hosts one active connection: URL + headers editor, connect/disconnect, live frame log with direction markers. The context sidebar lists session-recent endpoints for the active kind so reconnecting is one click. View switches go through the shared unsaved-env guard.

**Blocked by:** None (can start immediately). Parallel with 01 and 03.

**Status:** ready-for-agent

- [ ] `websocket` and `sse` views added to the workspace view union; rail items switch to them (no longer open tabs)
- [ ] Page body reuses the existing realtime tab internals; one active connection per page, keyed per kind
- [ ] Context sidebar shows session-recent endpoints derived from tabs seen this session, capped; empty state directs ("Connect to an endpoint and it will show up here.")
- [ ] Existing realtime tab launcher paths keep working or are removed cleanly (no dead code)
- [ ] Store tests for the session-recents slice (derived, capped, per kind)

Parent: #369
