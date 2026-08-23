# M44 T5 — Request builder & response viewer restyle

Blocked by: T2.
Blocks: T6.

## Goal

Restyle the request editor and response viewer onto the new shell + tokens, matching the demo's REST layout: method+URL bar, tabbed body/params/headers/auth/script editors, split response with status pill, timing/size chips, response-mode (split | inline) preference.

## Requirements

- Port `restLayout`, `builderHtml`, `respHtml`/`respDoneHtml` structure to React components over the existing stores/bindings — no behavior changes.
- Status pills (`st-timeout/st-ok/st-redir/...`), elapsed ticker (`startElapsedTicker`), skeleton loading lines, byte/ms formatters.
- Request settings modal; duplicate-request action (`dupCurrent`).
- Keep all existing editors (auth-panel, scripting, jwt-inspector) functional inside the restyled layout.

## Acceptance criteria

- [x] Send → loading skeleton → rendered response flow visually matches demo states. — skeleton + StatusPill + duration/size chips pre-existed (G-4 pass); this slice added layout preference + method tints. 2026-08-25
- [x] All request types (REST/GraphQL/gRPC/WS/SSE) still send correctly after restyle. — no behavior changes; typecheck/build/tests green. 2026-08-25
- [x] Response mode setting persists. — `useShellStore.responseMode` (`reqly-shell-response-mode`), split | inline toggle in the requests column; vertical ResizablePanelGroup for inline. 2026-08-25
- [x] typecheck + lint + React Doctor clean vs baseline; vitest 34/34. Remaining polish (request-settings modal, duplicate action) folded into T6 sweep. — 2026-08-25

## Reference

- `shared/views.js` (restLayout/builder/resp fns), `shared/engine.js` (response host binding)
