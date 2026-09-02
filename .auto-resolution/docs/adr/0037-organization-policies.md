# ADR 0037: Organization Policies — Local Policy File (M70)

## Status
Accepted — grill Q1 (core + CLI + desktop, SSO deferred)

## Context
Enterprise §58.5 needs local policy enforcement without cloud. Workflow, RBAC, and audit are separate.

## Decision
`internal/policy` (M70): `Policy{RequireAudit,MaxWorkflowSteps,AllowedActions,RequireAuth,AllowCustomThemes}` + `DefaultPolicy` + `Validate` + `Enforce`/`EnforceWorkflow` + `Load`/`Save`/`DefaultPath` (`.reqly/policy.yaml` 0700/0600); CLI `reqly policy show/validate/enforce` + desktop `PolicyGet/Save/Enforce`.

## Consequences
Q1: Policy is static YAML, not dynamic — no hot reload.
Q2: SSO/SCIM deferred — policy does not handle identity.
