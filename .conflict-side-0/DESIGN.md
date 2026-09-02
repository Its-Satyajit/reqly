# DESIGN.md — Reqly Desktop

Single source of truth for the desktop UI's design system. CSS, Tailwind theme, and components are projections of this file; if they drift, this file wins.

## Discovery

- **Artifact**: data-dense developer tool (desktop API client, Wails webview). Not a marketing surface — density and state clarity outrank decoration.
- **Audience**: developers building/testing HTTP APIs; comfortable in terminals, Git, and monospace.
- **Adjectives** (committed): *precise, dense, calm, engineered*.
- **Aesthetic essence**: "HTTP made legible."
- **Positioning**: local-first, Git-native — plain-text honesty over glossy abstraction.

## Typography

| Role | Face | Notes |
|------|------|-------|
| UI / body | IBM Plex Sans 400/500/600 | engineered heritage; bundled offline via @fontsource |
| Data / code / numbers | IBM Plex Mono 400/500 + `tabular-nums` | every URL, header, status, duration, table cell |

Bundled locally (local-first boundary: no CDN font fetches). System stacks remain only as fallbacks.

## Color

Coral is the brand anchor (`--primary`), used sparingly (~10%): primary buttons, active accents, ring. Surfaces are near-neutral (barely-warm light `#fbfbfa`, blue-black dark `#0d1015`) carrying ~60%. Semantic meaning rides the **status ramp** (GitHub-derived, WCAG-checked pairs):

- `status-ok` green → 2xx · `status-redirect` blue → 3xx · `status-warn` amber → 4xx & warnings · `status-error` red → 5xx/failures · `status-info` gray → 1xx/neutral
- Method tints: GET green, POST blue, PUT/PATCH amber, DELETE red.
- Status is **never color alone** — always dot + literal code/text.

## Tokens

Defined in `frontend/src/index.css` as Tailwind v4 `@theme` custom properties: shadcn-compatible surfaces (background/card/popover/muted/secondary/accent/destructive/border/input/ring), the status ramp, method tints, radius scale (base 6px; two steps: sm/md), no shadows on flat chrome (defined-edge via hairline borders only). Accent = neutral hover surface (menus); coral stays out of hover states.

## Signature move

The **StatusPill** — one component (`components/status.tsx`) rendering colored dot + tabular-numeral code, used identically in the response header, run steps, and history rows. It is the app's single memorable device; everything else stays quiet.

## Craft-layer decisions

- **Layout**: resizable three-pane shell (sidebar | editor | viewer) with persisted layouts; `min-w-0` discipline everywhere; density-first spacing (13px base).
- **Components**: full state matrices on shared primitives (Base UI Button/Input/Select/Dialog); every dropdown = CompactSelect; destructive actions = AlertDialog, never native confirm; copy actions report failure via toast.
- **Motion**: 150ms ease for theme/surface transitions only; `prefers-reduced-motion` collapses all animation.
- **Iconography**: lucide-react exclusively, one grid/stroke; icon-only buttons carry accessible names.
- **Dark mode**: designed palette (blue-black surfaces, elevation by lightness, desaturated ramp) — not an inversion.
- **Accessibility floor**: visible focus-visible ring globally, keyboard tree navigation with aria-expanded, tab roles, 24px+ targets on primary controls.

## Slop audit

- Rejected AI-default looks: cream+serif+terracotta, near-black+acid-green, broadsheet-hairline.
- No Inter/Roboto/system-ui as primary face (Plex superfamily instead).
- No scattered hex classes — warnings route through `status-warn`; one token system.
- Gradient text, glassmorphism, blob radii, bounce easing: absent.
- Verified gates: `tsc --noEmit` both workspaces, oxlint 0/0, vite build.
