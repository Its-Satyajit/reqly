# 03 — Desktop retry section (request editor)

**What to build:** Desktop users toggle and configure retries visually: a compact collapsible Retry section beside the timeout control in the request editor — count, base delay, fixed/exponential strategy dropdown, max-delay cap — collapsed by default showing "Retries: off" when unset, saving through the existing format-preserving seam; retry behavior itself flows free of charge through the shared engine.

Design pass per `/frontend-design` + `/design-principles`: progressive disclosure (collapsed until enabled), consistent with existing Settings-area controls, accessible labels/focus states, dark/light theming intact.

**Blocked by:** 01 — Engine retry loop

**Status:** ready-for-agent

- [x] Regenerate Wails bindings (Retry model + response Attempts surface in TS types)
- [x] Collapsible Retry section next to timeout: collapsed "Retries: off" summary when unset; expanded shows count, delayMs, strategy select (fixed/exponential), maxDelayMs fields
- [x] Dirty-tracking + save via existing request-editor save path; changed-on-disk reload interplay unaffected
- [x] UI follows design-principles: progressive disclosure, labeled controls, keyboard/focus support, theme-consistent styling matching neighboring settings controls
- [x] Frontend typecheck green (`npm run typecheck`)
