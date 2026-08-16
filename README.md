# Reqly

Local-first, Git-native API development environment and client. Collections, environments, tests, scripts, schemas, and mocks live on disk as plain-text, version-controlled project files — no cloud backend, no telemetry.

## What is Reqly?

Reqly is an API client where your work is a Git repository instead of an app database. Requests, collections, environments, and tests are plain-text files you can review, diff, and version alongside your code. A single Go core powers three interfaces: a Wails v3 desktop app, a Cobra CLI, and an MCP server.

## Documentation

- **Roadmap & milestones:** [`ROADMAP.md`](./ROADMAP.md)
- **Domain glossary:** [`CONTEXT.md`](./CONTEXT.md)
- **Agent guidelines:** [`AGENTS.md`](./AGENTS.md)
- **Architecture & tech stack:** [`docs/technology-stack.md`](./docs/technology-stack.md)
- **Testing strategy & TDD:** [`docs/testing-strategy.md`](./docs/testing-strategy.md)
- **Feature set:** [`docs/features.md`](./docs/features.md)

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

The Go core in [`internal/`](./internal/) is the single source of truth, shared by the Desktop GUI (`apps/desktop`), the Cobra CLI (`apps/cli`), and the MCP server (`internal/mcp`).

## Prerequisites

- **Go 1.25+**
- **Node 24+** and **nub** (`npm i -g nub`)
- **Wails v3** (`wails3 install`)

## Quick Start

```bash
# Install dependencies (nub is the package manager; npm workspaces)
nub install

# Run the tests
go test ./...
go test -race ./...

# Typecheck the frontend workspace(s)
nub run typecheck

# Run the CLI against the live demo API
go run ./apps/cli run https://reqly-test-api.vercel.app/api/users
go run ./apps/cli run -H "Authorization: Bearer admin-token" https://reqly-test-api.vercel.app/api/auth/me

# Serve a mock API from an OpenAPI spec
go run ./apps/cli mock path/to/openapi.yaml --port 4010

# Desktop app (Wails v3 dev mode)
cd apps/desktop/backend
wails3 generate bindings -d frontend/bindings -i -ts
wails3 dev
```

## License

[GPL-3.0](./LICENSE)