# T3 — Desktop Settings panel

> **Spec:** `docs/spec/m47-proxy-tls-per-request.md`
> **Blocks:** T1

- `frontend/src/features/request-editor/RequestSettings` — proxy input + insecure toggle + CA file picker, dirty tracking, `Spec` wiring
- Replaces `useProxyTlsStore` data-layer stub

**Done when:** `nub run typecheck` + panel dirty/save tests green
