# Reqly

> [!CAUTION]
> ## Alpha, not stable. Expect breakage.
> Reqly is pre-1.0. It may crash, corrupt local data, or break between commits. There are no stability guarantees, no migration path, and no semver until v1.0.0. If you need a stable API client today, use Postman, Insomnia, Bruno, or HTTPie, and come back when the P0/P1 milestones in [ROADMAP.md](./ROADMAP.md) are done.

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="frontend/src/assets/logo-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="frontend/src/assets/logo-light.svg">
    <img alt="Reqly Logo" src="frontend/src/assets/logo-light.svg" width="180">
  </picture>
</p>

<p align="center">
  <b>A local-first, Git-native API client.</b> Alpha quality.
</p>

<p align="center">
  <a href="#overview">Overview</a> •
  <a href="#what-it-does">What it does</a> •
  <a href="#why-reqly">Why Reqly?</a> •
  <a href="#architecture">Architecture</a> •
  <a href="#quick-start">Quick start</a> •
  <a href="#docs">Docs</a> •
  <a href="#license">License</a>
</p>

## Overview

Reqly is an API client that keeps your work on disk. Collections, environments, tests, scripts, and request history are plain-text files in your repository. You review them in pull requests, diff them in Git, and back them up like any other code.

There is no cloud backend and no account. There is also no telemetry: payloads, headers, tokens, and traffic never leave your machine.

## What it does

- Request files as JSON or YAML, so collections diff and merge in Git like source code.
- A Go core drives every interface, so the CLI and the desktop app behave identically on the same files.
- Desktop app built on Wails v3, React, and Vite.
- CLI for scripts and CI pipelines (`run`, `test`, `collection`, `mock`, `import`, `export`, `env`, `auth`, `history`, and more).
- Auth: basic, bearer, API key, JWT signing, digest, OAuth 2.0 (client credentials, authorization code with PKCE, device flow), AWS SigV4, Akamai EdgeGrid.
- WebSocket and SSE clients.
- A mock server generated from an OpenAPI spec, including delay and error injection.
- Retry policies, pagination runners, bulk execution, HAR import/export, and cURL/OpenAPI import with Postman export.

Planned but not shipped yet: gRPC, SOAP, the OpenAPI editor, contract testing, and the MCP server (currently a stub). See the roadmap for status.

## Why Reqly?

| | Cloud clients | Reqly |
| :--- | :--- | :--- |
| Storage | Proprietary cloud database | Plain-text files on your disk |
| Version control | App-internal history, paid team sync | Git commits, branches, PRs |
| Telemetry | Usage metrics, sync traffic | None |
| Interfaces | Desktop GUI | Desktop app, CLI |
| Account | Required | Not needed |

The trade-off is real: cloud clients give you polished sync and collaboration out of the box. Reqly bets that developers would rather keep requests in their repo and use Git for sharing. If that bet is wrong for you, the table above is the honest summary of what you'd give up.

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

The Go core ([`internal/`](./internal/)) owns one execution pipeline. Environment resolution, variable interpolation, auth, retries, cookie handling, masking, and history recording happen once, in the core. The desktop app and CLI parse input and render output. That is why behavior matches across interfaces without anyone maintaining two copies.

## Quick start

Prerequisites:

- Go 1.25+
- Node 24+ and nub (`npm i -g nub`)
- Wails v3 (`wails3 install`)

```bash
git clone https://github.com/Its-Satyajit/reqly.git
cd reqly
nub install

# Run the Go test suite and frontend typecheck
go test ./...
go test -race ./...
nub run typecheck

# Try the CLI against the companion mock API
go run ./apps/cli run https://reqly-test-api.vercel.app/api/users
go run ./apps/cli run -H "Authorization: Bearer admin-token" https://reqly-test-api.vercel.app/api/auth/me

# Serve a mock API from an OpenAPI spec
go run ./apps/cli mock path/to/openapi.yaml --port 4010

# Run the desktop app in development mode
cd apps/desktop/backend
wails3 generate bindings -d frontend/bindings -i -ts
wails3 dev
```

## Docs

- Roadmap and milestone status: [`ROADMAP.md`](./ROADMAP.md)
- Domain glossary: [`CONTEXT.md`](./CONTEXT.md)
- Agent guidelines: [`AGENTS.md`](./AGENTS.md)
- Architecture and tech stack: [`docs/technology-stack.md`](./docs/technology-stack.md)
- Testing strategy: [`docs/testing-strategy.md`](./docs/testing-strategy.md)
- Full feature set: [`docs/features.md`](./docs/features.md)

## Stability

Reqly is alpha-quality research software. Until v1.0:

- Crashes, hangs, data loss in `.reqly` directories and workspace YAML, and silent misbehavior are expected.
- CLI flags, the request file schema, and IPC bindings can change without deprecation notices. Pin a commit if you build on it.
- There is no SLA, no security audit, and no stable import/export round-trip guarantee.

Keep using Postman, Insomnia, Bruno, HTTPie, or Restish for daily work. Check back when the roadmap's P1 phase is near complete.

## License

Apache 2.0. See [LICENSE](./LICENSE).
