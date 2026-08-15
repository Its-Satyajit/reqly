# ADR 0002: AI Agent Instructions and Workspace Rules

- **Status:** Accepted
- **Date:** 2026-08-15
- **Deciders:** Maintainers, AI Pair Programmers

## Context & Problem Statement

As AI coding agents (Antigravity, Gemini, Cursor, Claude, Copilot) become primary contributors to the codebase, we need a single, authoritative, plain-text instruction source that guides AI behavior, enforces architectural principles (local-first, Git-native, dual GUI/CLI parity), and enforces test-driven development (TDD) gates.

## Decision Drivers

1. Different AI tools look for different convention files (`AGENTS.md`, `GEMINI.md`, `.cursorrules`, `.github/copilot-instructions.md`).
2. Duplicating rules across multiple files risks documentation drift and outdated guidance.
3. Feature development must align with existing domain concepts documented in `CONTEXT.md` and undergo interactive grilling (`grill-with-docs`).

## Considered Options

1. **Option 1:** Add separate inline documentation in each tool's specific file format.
2. **Option 2:** Use `AGENTS.md` as the root single source of truth, with thin pointer files (`GEMINI.md`, `.cursorrules`, `.github/copilot-instructions.md`) delegating to `AGENTS.md`.

## Decision Outcome

Chosen **Option 2**. `AGENTS.md` is established at the root of the repository as the canonical AI guidelines document. Other AI tool configuration files point back to `AGENTS.md`.

### Consequences

- **Positive:** Single place to update AI rules, preventing drift across tools.
- **Positive:** Clear instructions on testing (`go test ./...`), domain model usage (`CONTEXT.md`), and codebase directory structure.
- **Negative:** AI tools that do not resolve markdown hyperlinks rely on reading the root `AGENTS.md` file explicitly.
