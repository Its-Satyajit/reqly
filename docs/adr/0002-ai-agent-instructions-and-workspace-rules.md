# ADR 0002: AI Agent Instructions and Workspace Rules

- **Status:** Accepted
- **Date:** 2026-08-15
- **Deciders:** Maintainers, AI Pair Programmers

## Context & Problem Statement

Multiple AI coding assistants (Antigravity, Gemini, Cursor, Claude, Copilot) contribute to the codebase. We require a single source of truth for repository principles (local-first, Git-native, dual GUI/CLI parity), test-driven development (TDD) quality gates, and the 5-stage skill pipeline.

## Decision Drivers

1. AI tools inspect distinct configuration file paths (`AGENTS.md`, `GEMINI.md`, `.cursorrules`, `.github/copilot-instructions.md`).
2. Duplicating rules across tool-specific files causes documentation drift.
3. Feature execution requires domain model alignment ([`CONTEXT.md`](../../CONTEXT.md)) and the 5-stage skill pipeline (`~/.agents/skills/`).

## Considered Options

- **Option 1:** Duplicate full rule sets inside each tool-specific file.
- **Option 2:** Establish `AGENTS.md` as the single source of truth, with thin pointer files delegating to `AGENTS.md`.

## Decision Outcome

Chosen **Option 2**. `AGENTS.md` is established at the repository root as the canonical AI guidelines document. Tool-specific files point directly to `AGENTS.md`.

### Consequences

- **Positive:** Single-file maintenance prevents instruction drift.
- **Positive:** Clear execution bounds for TDD (`go test ./...`) and skill pipeline invocation.
- **Negative:** AI tools unable to resolve markdown links must read `AGENTS.md` directly.
