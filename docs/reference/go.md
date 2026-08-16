# Go Code Conventions & Skill Mapping

Canonical Go conventions and skill mapping for Reqly (`internal/`, `apps/cli`).

## Primary Skill Location

Use Go skills located at `~/.agents/skills/cc-skills-golang/skills/`.

## Core Conventions

- **Code Style (`golang-code-style`):** Standard `gofmt` layout, explicit error handling, no panics in application logic.
- **Naming (`golang-naming`):** Short camelCase variable names, exported PascalCase, descriptive test names `Test<Target>_<Scenario>`.
- **Error Handling (`golang-error-handling`):** Return wrapped errors (`fmt.Errorf("...: %w", err)`). Never discard errors.
- **Concurrency (`golang-concurrency`, `golang-context`):** Pass `context.Context` as first parameter for network/IO operations. Prevent goroutine leaks using channel select and context cancellation.
- **CLI Development (`golang-spf13-cobra`):** Use Cobra for subcommands in `apps/cli/cmd/`.
- **Testing (`golang-testing`):** Table-driven tests with standard `testing` package or `stretchr/testify`.
