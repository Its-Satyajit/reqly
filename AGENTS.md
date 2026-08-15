# Reqly — AI Agent & Developer Guidelines

Local-first, Git-native API development environment built with Go (`internal/`, `apps/cli`) and Wails v3 Desktop (`apps/desktop`).

## 1. Non-Negotiable Boundaries

- **Local-first & Git-native:** Store all collections, environments, request files, and tests in plain-text (JSON/YAML) versioned with Git. Cloud backends and accounts are forbidden.
- **Zero telemetry:** Payload, header, credential, secret, and traffic data strictly remains local.
- **Dual parity:** Implement core features in Go (`internal/`) and expose via CLI (`apps/cli`) and Desktop (`apps/desktop`).

## 2. Directory Layout

- `apps/cli`: Cobra CLI commands.
- `apps/desktop`: Wails v3 desktop (backend Go, frontend Svelte/TypeScript).
- `internal/`: Shared domain logic (core HTTP engine, runner, auth, mocking, mcp, requestfile parser).
- `docs/`: Specs, ADRs, and references.
- `CONTEXT.md`: Canonical domain glossary.

## 3. Skill Pipeline & Development Execution

When implementing features or architectural changes, execute the 5-stage pipeline in order (`~/.agents/skills/`):

1. `grill-with-docs`: Stress-test requirements against [`CONTEXT.md`](./CONTEXT.md) in an interactive session (one question per turn with recommended answer). Update [`CONTEXT.md`](./CONTEXT.md) and `docs/adr/` inline.
2. `to-spec`: Generate technical specification and design document.
3. `to-tickets`: Split specification into actionable, trackable task tickets.
4. `implement`: Execute tickets using strict TDD (write tests first).
5. `code-review`: Audit Go and Svelte/TypeScript code against conventions and verify clean test output.

## 4. Verification & Quality Gates

Completion criterion for any change: all unit and integration tests pass cleanly with zero dummy fallbacks or skipped assertions.

- **Go core & CLI:** `go test ./...`
- **CLI binary build:** `go build -o reqly ./apps/cli`
- **Frontend desktop:** `cd apps/desktop/frontend && npm test`
- **Go conventions:** Follow [`docs/Referance/go.md`](./docs/Referance/go.md).

## 5. Domain Glossary

Consult [`CONTEXT.md`](./CONTEXT.md) for canonical terms before naming entities or creating abstractions.
