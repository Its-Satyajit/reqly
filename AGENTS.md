# Reqly — AI Agent & Developer Guidelines

> **Target Audience:** AI Coding Assistants (Antigravity, Gemini, Cursor, Claude, Copilot) and Human Contributors.

Reqly is a local-first, Git-native API development environment and client built with a Go backend core, Wails v3 cross-platform desktop GUI (Svelte/TypeScript), and a Go CLI binary.

---

## 1. Core Principles & Philosophy

1. **Local-First & Git-Native:** All collections, environments, request files, and tests are stored on disk in standard plain-text formats (JSON/YAML) and versioned with standard Git. Never introduce mandatory cloud backends or online accounts.
2. **Privacy & Zero Telemetry:** Request/response payloads, headers, secrets, and traffic must never be collected or sent to telemetry servers.
3. **Dual GUI + CLI Parity:** Core features (request execution, collections, variables, mocking, testing, SSE, WebSocket) must be implemented in the Go core (`internal/`) and exposed via both the CLI (`apps/cli`) and Desktop GUI (`apps/desktop`).

---

## 2. Codebase Architecture & Directory Structure

```
.
├── apps/
│   ├── cli/             # Go CLI application (Cobra commands)
│   └── desktop/         # Wails v3 Desktop application
│       ├── backend/     # Go desktop app bindings & lifecycle
│       └── frontend/    # Svelte/TypeScript UI components & state
├── internal/            # Core Go domain packages (shared logic)
│   ├── auth/            # Auth strategies (Bearer, Basic, OAuth2, API Key)
│   ├── collections/     # Workspace & collection disk resolution
│   ├── core/            # Core HTTP engine & executor
│   ├── environments/    # Variable resolution & environment management
│   ├── importer/        # OpenAPI, Postman, Insomnia importers
│   ├── mcp/             # Model Context Protocol (MCP) server integration
│   ├── mocking/         # Mock API server engine (kin-openapi based)
│   ├── request/         # Request model definitions
│   ├── requestfile/     # Plain-text request file loader & parser
│   ├── runner/          # Collection & suite execution engine
│   └── scripting/       # Pre/post-request scripting & assertion engine
├── docs/                # Architectural specs, feature sets, & ADRs
└── CONTEXT.md           # Domain glossary and canonical terms
```

---

## 3. Workflow & AI Agent Rules

### 3.1 Feature Planning (`grill-with-docs` Skill)
When implementing new features or making architectural changes:
- Execute the `grill-with-docs` skill to stress-test designs against existing terms in [`CONTEXT.md`](./CONTEXT.md).
- Ask questions one at a time with recommended answers.
- Update [`CONTEXT.md`](./CONTEXT.md) inline as terms clarify, and create ADRs under `docs/adr/` for significant decisions.


### 3.2 Testing & Quality Gates (TDD)
Before declaring any feature complete:
- Run Go unit & integration tests: `go test ./...`
- Run frontend checks: `cd apps/desktop/frontend && npm test`
- Do not skip failing unit tests or introduce dummy fallback implementations. Fix the underlying root cause.

### 3.3 Code Style & Conventions
- **Go (`internal/`, `apps/cli`):** Idiomatic Go, error returning (no panics), clear interfaces, `gofmt`/`go vet` compliant. Follow rules in [`docs/Referance/go.md`](./docs/Referance/go.md).
- **TypeScript & Svelte (`apps/desktop/frontend`):** Clean component boundaries, fluid responsive design, strict type checking, HSL color tokens.

---

## 4. Useful Commands

| Action | Command |
| :--- | :--- |
| **Run Go Tests** | `go test ./...` |
| **Build CLI Binary** | `go build -o reqly ./apps/cli` |
| **Run CLI Commands** | `./reqly run <file/URL>` or `./reqly collection run <dir>` |
| **Desktop Dev Server** | `cd apps/desktop/backend && wails3 dev` |
| **Frontend Test/Lint** | `cd apps/desktop/frontend && npm test` |

---

## 5. Domain Terms Reference

Always consult [`CONTEXT.md`](./CONTEXT.md) for official terminology (e.g., Release Pipeline Architecture, Dual-Orchestration, Requestfile Format, AI Agent Protocol).
