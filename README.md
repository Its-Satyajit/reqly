# Reqly

Local-first, Git-native API development environment and client. Collections, environments, tests, scripts, schemas, and mocks live on disk as plain-text, version-controlled project files.

- **Roadmap:** [`ROADMAP.md`](./ROADMAP.md)
- **AI Agent Guidelines:** [`AGENTS.md`](./AGENTS.md)
- **Domain Glossary:** [`CONTEXT.md`](./CONTEXT.md)
- **Architecture & Tech Stack:** [`docs/API Client — Technology Stack.md`](<./docs/API Client — Technology Stack.md>)
- **Testing & TDD:** [`docs/API Client — Testing Strategy & TDD.md`](<./docs/API Client — Testing Strategy & TDD.md>)

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

The Go core (`internal/`) is shared across Desktop GUI (`apps/desktop`), Cobra CLI (`apps/cli`), and MCP server (`internal/mcp`).

## Prerequisites

- **Go 1.25+**
- **Node 20+ / nub**
- **Wails v3** (`wails3 install`)

## Quick Start & Verification

```bash
# Frontend setup & typecheck
nub install
nub run dev
nub run -r --if-present typecheck

# Go core & CLI testing
go mod tidy
go test ./...
go test -race ./...

# Desktop build
cd apps/desktop/backend
wails3 generate bindings -d frontend/bindings -i -ts
wails3 dev
```

## Mock API Server

Live test endpoint hosted at `https://reqly-test-api.vercel.app`:

```bash
reqly run https://reqly-test-api.vercel.app/api/users
reqly run -H "Authorization: Bearer admin-token" https://reqly-test-api.vercel.app/api/auth/me
```

## License

[GPL-3.0](./LICENSE)