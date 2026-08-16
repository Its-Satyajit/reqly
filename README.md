# Reqly

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="frontend/src/assets/logo-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="frontend/src/assets/logo-light.svg">
    <img alt="Reqly Logo" src="frontend/src/assets/logo-light.svg" width="180">
  </picture>
</p>

<p align="center">
  <b>The Local-First, Git-Native API Development Environment</b>
</p>

<p align="center">
  <a href="#overview">Overview</a> •
  <a href="#key-features">Key Features</a> •
  <a href="#why-reqly">Why Reqly?</a> •
  <a href="#architecture">Architecture</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#documentation">Docs</a> •
  <a href="#license">License</a>
</p>

---

## Overview

**Reqly** is a high-performance, local-first API client designed for developers who demand privacy, speed, and Git integration. 

Unlike traditional cloud-locked API tools, **Reqly stores your entire workflow as version-controlled project files directly on disk**. Your collections, environments, tests, scripts, schemas, and mocks live in your repository alongside your code—**no cloud backend, no account mandatory, and zero telemetry**.

---

## Key Features

- **Git-Native Project Files** — Plain-text JSON/YAML formats designed to be reviewed, diffed, and merged seamlessly in your existing pull request workflow.
- **Zero Telemetry & 100% Privacy** — Your request payloads, headers, tokens, secrets, and traffic data stay entirely on your local machine.
- **Unified Engine Across 3 Interfaces** — A single, high-speed Go core drives all interfaces:
  - **Desktop App** — Sleek, reactive GUI powered by Wails v3 + React + Vite.
  - **CLI** — Powerful Cobra-based CLI for developer workflows and CI/CD pipelines.
  - **MCP Server** — Built-in Model Context Protocol server enabling direct AI assistant integration.
- **Multi-Protocol & Spec Support** — Built to handle REST, GraphQL, gRPC, WebSockets, SSE, SOAP, and OpenAPI 2.0/3.0/3.1 specs out of the box.
- **Mocking & Contract Validation** — Serve instant mock servers from OpenAPI definitions and enforce response schema validation effortlessly.

---

## Why Reqly?

| Feature | Standard Cloud Clients | **Reqly** |
| :--- | :--- | :--- |
| **Data Storage** | Proprietary cloud database | **Local plain-text files on disk** |
| **Version Control** | Internal app history / Paid team sync | **Native Git commits, branches, & PRs** |
| **Telemetry & Privacy** | Usage metrics & sync traffic | **Zero telemetry. 100% local-first** |
| **Interfaces** | Desktop GUI only | **Desktop App, CLI, & MCP Server** |
| **Account Required** | Yes (forcing login) | **No account needed** |

---

## Architecture

```
                    ┌─────────────────────────┐
                    │         Go Core         │
                    │       (internal/)       │
                    └────────────┬────────────┘
         ┌───────────────────────┼───────────────────────┐
         ▼                       ▼                       ▼
   Desktop GUI                Cobra CLI              MCP Server
 (Wails v3 + React)         (apps/cli)             (internal/mcp)
```

The Go core ([`internal/`](./internal/)) acts as the single source of truth, delivering rock-solid performance, deterministic behavior, and dual feature-parity across all native interfaces.

---

## Quick Start

### Prerequisites

- **Go 1.25+**
- **Node 24+** & **nub** (`npm i -g nub`)
- **Wails v3** (`wails3 install`)

### Installation & Execution

```bash
# Clone the repository and install dependencies
git clone https://github.com/Its-Satyajit/reqly.git
cd reqly
nub install

# Run core tests & typecheck
go test ./...
go test -race ./...
nub run typecheck

# Execute CLI against demo endpoints
go run ./apps/cli run https://reqly-test-api.vercel.app/api/users
go run ./apps/cli run -H "Authorization: Bearer admin-token" https://reqly-test-api.vercel.app/api/auth/me

# Spin up an instant mock API from an OpenAPI spec
go run ./apps/cli mock path/to/openapi.yaml --port 4010

# Launch Desktop App in development mode (Wails v3)
cd apps/desktop/backend
wails3 generate bindings -d frontend/bindings -i -ts
wails3 dev
```

---

## Documentation & Resources

- **Roadmap & Milestones:** [`ROADMAP.md`](./ROADMAP.md)
- **Domain Glossary:** [`CONTEXT.md`](./CONTEXT.md)
- **Agent Guidelines:** [`AGENTS.md`](./AGENTS.md)
- **Architecture & Tech Stack:** [`docs/technology-stack.md`](./docs/technology-stack.md)
- **Testing Strategy:** [`docs/testing-strategy.md`](./docs/testing-strategy.md)
- **Full Feature Set:** [`docs/features.md`](./docs/features.md)

---

## License

Distributed under the **[GPL-3.0 License](./LICENSE)**. Built for developers who care about code ownership.