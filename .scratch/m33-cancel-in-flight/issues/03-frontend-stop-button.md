# 03 — Stop button + cancelled state (shared frontend)

**What to build:** Per-send `sendId` flows through the sender adapter into `SendOptions.sendId`; while loading, Send toggles to **Stop**, invoking a new store `cancel()` action that calls the adapter's `cancelSend(sendId)` and drops the pending token; response pane renders a distinct neutral "Request cancelled" state.

**Blocked by:** 02 (bindings must exist; regenerate via `wails3 generate bindings` first)

**Status:** done

- [x] Adapter type gains `cancelSend(sendId): Promise<void>`; wailsSender wires `AppService.CancelSend`; fetchSender/browser-dev fallback resolves as no-op
- [x] `useRequestStore`: mint `crypto.randomUUID()` per send, pass in options; new `cancel(tabId)` action clears token + marks tab state `{status:'cancelled'}`
- [x] Bridge maps Go "context canceled" errors to the same cancelled outcome
- [x] `RequestEditor.tsx`: button shows **Stop** while loading (click = cancel), Send otherwise
- [x] ResponseViewer/store render: neutral "Request cancelled" message — visually distinct from error (no red)
- [x] Late response after cancel discarded via existing `sendTokens`
- [x] `npm run typecheck` green (frontend has no test script)
