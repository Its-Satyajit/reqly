# Reqly

A local-first, Git-native API development environment — collections, environments, tests, scripts, schemas, mocks, and documentation living together as version-controlled project files.

> **Status:** Scaffold / Work in progress
> **Product Name:** Reqly
> **Docs:** see [`FeatureSet.md`](FeatureSet.md), [`API Client — Technology Stack.md`](API%20Client%20—%20Technology%20Stack.md), and [`API Client — Testing Strategy & TDD.md`](API%20Client%20—%20Testing%20Strategy%20&%20TDD.md)

## Architecture

```
                    Go Core                    React + TypeScript + Vite
                       │                       Tailwind + Base UI
        ┌──────────────┼──────────────┐        CodeMirror 6 + Zustand
        │              │              │               │
        ▼              ▼              ▼          Wails v3 Bridge
     Desktop          CLI            MCP            │
    (Wails v3)      (cobra)                        ▼
                                             Go Core (Goja loaded lazily)
```

The Go core is independent of Wails and shared by the Desktop, CLI, and MCP front-ends. The frontend is a thin Vite + React + TypeScript app talking to the Go core over the Wails v3 bridge.

## Repository layout

```
apps/
├── desktop/
│   └── backend/          Go Wails v3 application + Taskfile + build/
│       ├── main.go       Wails v3 entry point (application.New)
│       ├── app.go        AppService (thin Go ↔ JS bindings)
│       ├── build/        Wails v3 build config & platform Taskfiles
│       └── frontend/     Wails-bound Vite host app (embed all:frontend/dist)
│           ├── src/      Host entry (main.tsx imports @reqly/frontend)
│           └── bindings/ Generated Go ↔ TypeScript bindings (wails3 generate bindings)
└── cli/                  Go CLI (cobra) using the same core

frontend/                 Shared React + TypeScript UI
├── app/                  Application shell
├── components/           UI components (Base UI primitives, shadcn CLI)
├── editors/              CodeMirror 6 editors
├── features/             Feature-scoped UI
├── stores/               Zustand state
└── lib/                  Utilities

internal/                 Go core packages
├── core/                 Application services
├── request/  response/   Request & response engine
├── auth/     secrets/    Authentication & secrets
├── variables/ environments/ collections/
├── scripting/            Goja runtime (lazy)
├── testing/  mocking/    Test & mock engines
├── openapi/  graphql/ grpc/ websocket/
├── git/      history/    importer/ exporter/ mcp/
```

## Prerequisites

- **Go 1.25+** — core, CLI, desktop backend. Run `go mod tidy` once installed (this scaffold pins Wails v3 and cobra in `go.mod`; Goja is resolved automatically by `go mod tidy` from the `internal/scripting` import).
- **Node 20+ / nub** — frontend (nub is the package manager; see `nubjs.com`).
- **Wails v3** — desktop builds (`wails3 install`), which additionally requires the system WebView/WebKit development packages for your platform. If `wails3` is not on your `PATH`, install it to `~/go/bin` and export it (e.g. `export PATH="$PATH:$HOME/go/bin"`).

> The Go module path is `github.com/reqly/reqly`. Update it to the real repository path before publishing if it differs.

## Development

```bash
nub install           # install frontend dependencies (nub.lock)
nub run dev           # run the Vite host app in a browser
nub run build         # production frontend build
nub run -r --if-present typecheck   # TypeScript checks

go mod tidy           # resolve Go dependencies (after installing Go)
go test ./...         # Go core tests
go test -race ./...   # race detector

cd apps/desktop/backend
wails3 generate bindings -d frontend/bindings -i -ts   # regenerate Go ↔ TS bindings after changing services
wails3 build          # build the desktop binary (bin/reqly)
wails3 dev            # run the full desktop app (requires Wails v3 + Go)
```

UI components are added with the shadcn CLI using the Base UI registry (`base-nova` style); aliases resolve through the package `imports` map (`#components/*`, `#lib/*`, `#hooks/*`) in `frontend/package.json`, so they work from the shared package and the desktop host app alike.

### Scripts

All scripts run through **nub** from the repository root. The root package orchestrates the desktop host app, which is where the Wails-bound UI lives.

| Command | Where | What it does |
| --- | --- | --- |
| `nub run dev` | root | Runs the desktop host app in dev mode (`@reqly/desktop`), plain Vite — browser-only, no Wails bridge |
| `nub run build` | root | Production build of the desktop host app (`@reqly/desktop`), output to `apps/desktop/backend/frontend/dist` for the Wails embed |
| `nub run -r --if-present typecheck` | root | `tsc --noEmit` across every workspace package |
| `nub run --filter @reqly/frontend dev` | root | Dev server for the **shared** UI package (`frontend/`) |
| `nub run --filter @reqly/frontend build` | root | Production build of the shared UI package |
| `nub run --filter @reqly/desktop dev` | root | Dev server for the desktop host app (`vite`, port 9245) |
| `nub run --filter @reqly/desktop build` | root | Host app production build (`tsc && vite build --mode production`) |
| `nub run --filter @reqly/desktop build:dev` | root | Host app dev-mode build (`vite build --minify false --mode development`), used by `wails3 dev` |
| `nub run --filter @reqly/desktop typecheck` | root | Type-check the host app (incl. generated bindings) |

Package-level scripts (run from the respective directory or via `nub run`):

- **`frontend/`** (shared UI): `dev` → `vite` · `build` → `vite build` · `preview` → `vite preview` · `typecheck` → `tsc --noEmit`
- **`apps/desktop/backend/frontend/`** (host app): `dev` → `vite` · `build:dev` → `tsc --noEmit && vite build --minify false --mode development` · `build` → `tsc --noEmit && vite build --mode production` · `preview` → `vite preview` · `typecheck` → `tsc --noEmit`

## Testing

Development follows **TDD** (see the testing strategy doc). The initial Go core demonstrates the pattern: `internal/variables` implements scope-precedence variable resolution with behavior-first tests.

- Go: `go test ./...`
- Race: `go test -race ./...`
- Benchmarks: `go test -bench=. ./...`
- Frontend: Vitest (TBD)
- E2E: Playwright (TBD)

## License

TBD