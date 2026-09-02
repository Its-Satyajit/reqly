# shadcn/ui Component Gap Analysis & Adoption

_Date: 2026-08-26 · Scope: `frontend/` (Wails v3 desktop, React 19 + Vite + Tailwind v4, base-nova / Base UI)_

## Baseline (pre-migration)

Installed: `button, field, input, input-group, textarea, select, badge, alert,
alert-dialog, dialog, tooltip, separator, spinner, toast (custom), resizable,
command, label, cycle-toggle (custom)` + custom `CompactSelect`, `ContextMenu`,
`toastManager`.

## Installed this pass

`tabs, context-menu, dropdown-menu, skeleton, switch, checkbox, scroll-area,
table, card, sidebar, popover, sheet, kbd, empty, progress, toggle-group,
accordion, hover-card, sonner, toggle`, hook `use-mobile`.

Note: local `button.tsx` and `input.tsx` are intentionally customized
(dense compact styling: `h-8`, `text-xs`) and were preserved over upstream.

## Migrated call sites

| Change | Where |
|---|---|
| Hand-rolled `role="tablist"` → shadcn `Tabs` | ResponseViewer, RequestEditor, EnvironmentsView, GraphqlBrowser, DocsView |
| Hand-rolled ARIA menu → `ContextMenu` primitives | `ContextMenu.tsx` internals (API unchanged for CollectionTree/RequestTabs) |
| Overflow menu → `DropdownMenu` | RequestEditor codegen menu |
| `cycle-toggle` deleted → `ToggleGroup` | ResponseModeToggle |
| `CompactSelect` deleted → `Select` | 11 call sites across 10 files |
| Custom toast system deleted → Sonner | `lib/notify.ts` rewritten; `<Toaster />` from `ui/sonner` in App |
| `Empty` | HistoryView (×2 states), TestsView, OverviewHome |
| `Kbd` | CommandPalette ⌘K, SettingsView shortcuts, ResponseViewer hint |
| `Card` | JwtInspector token/header/payload panels |
| `ScrollArea` | WorkspaceSidebar primary scroller, ResponseViewer headers + JSON body |
| `HoverCard` | WorkspaceSidebar import/export rich hints |
| `Switch` | RunnersPanel parallel toggle, RealtimeTab binary/auto-scroll |
| `Checkbox` | KeyValueEditor enabled column, TestTab exact/invert, DocsView include toggle |
| `Popover` | RequestEditor retry section disclosure |
| `Skeleton` | ResponseViewer loading rows |

Deliberately **not** converted: `RequestTabs.tsx` (close buttons/middle-click
inside tab chrome conflict with single-button `TabsTrigger`; no drag-reorder
exists).

## Deferred (follow-up tickets)

- `Sidebar` — rewrite `WorkspaceSidebar`/`ActivityRail` onto shadcn Sidebar.
- `Sheet` — GitPanel / import-export as side panels.
- `Accordion`/`Collapsible` — GraphQL browser sections, docs view.
- `Progress` — needs a real numeric progress source first (none exists today).
- `Table` — history/test results once a data-grid need is confirmed.
- `ToggleGroup` for response-mode variants beyond the two-state switch.

## Verification status (2026-08-26)

- `tsc --noEmit`: clean
- `oxlint src`: clean
- vitest: 36/36 passed
- react-doctor: remaining findings are pre-existing app issues
  (`no-children-prop` ×6, field.tsx memoization, workspace localStorage key,
  pnpm hardening ×2) plus the deferred unused components above.

## Conventions noted during adoption

- Base UI build ⇒ use `render={...}` (not Radix `asChild`).
- Lint (`anti-slop`) requires `// SAFETY:` comments directly above any type
  assertion; avoid runtime `typeof` narrowing (use `instanceof`).
- Aliases: `#components/ui`, `#lib/utils`. Icons: lucide-react. Keep dense
  compact styling when adopting primitives.
