# ADR 0003: Validation and Structural Diff Engines for Headless CLI and Core

## Status
Accepted

## Context
Reqly requires headless CLI capabilities for validating OpenAPI specifications / project descriptors (`reqly validate`) and structural diffing of JSON/YAML payloads and specs (`reqly diff`).

## Decision
1. Implement static validation in `internal/validation` using `kin-openapi` and project file inspection (`ValidateOpenAPIFile`, `ValidateProject`).
2. Implement structural diffing in `internal/diffing` using `github.com/r3labs/diff/v3` for key-aware structural comparison.
3. Expose both engines directly via Cobra commands in `apps/cli/cmd/validate.go` and `apps/cli/cmd/diff.go`.

## Consequences
- **Positive:** Dual parity; logic lives in reusable Go packages (`internal/`) accessible by CLI and future Desktop bindings.
- **Trade-off:** Introduced dependency on `github.com/r3labs/diff/v3` for AST/key-aware structural JSON diffing.
