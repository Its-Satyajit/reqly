# ADR 0004: Environments as a Git-native Variable Layer

## Status
Accepted

## Context
Reqly needs first-class environment support (Development/Staging/Production variable sets) populating the dormant `environment` scope. The choice of on-disk format, precedence placement, and selection mechanism is hard to reverse once users have Git-tracked projects, so we record it.

## Decision
1. Environments are plain YAML files at `environments/<name>.yaml` (name from filename, optional `description`, `variables:` + `secrets:` maps). No inheritance, no `baseURL` — environments are purely a variable layer. Discovered by walking up to the nearest `environments/` directory.
2. New `process env` scope (OS env + nearest `.env`, OS wins) sits at the very bottom of precedence: process env → global → environment → collection → folder → request → runtime. `prompt` scope deferred.
3. Active-environment selection (highest wins): `REQLY_ENV` → `--env` flag → request/test file `environment:` field → workspace descriptor `environment:` field. No selection means an empty environment scope; a selected-but-missing environment is a hard error. Collection runs use one environment for the whole run.
4. `.env` parsed into the process-env scope with a hand-rolled parser — no new dotenv dependency, matching the project's minimal-dependency posture.
5. Secret values (and `.env` values) render as `[SECRET]` in CLI/test output via a `Mask` helper.

## Consequences
- **Positive:** Git-native, diffable, testable; reuses the existing precedence engine and runner seams; unblocks auth/secrets/dynamic values.
- **Trade-off:** Environments cannot carry baseURL/auth/headers (those stay collection/request-level); no inheritance means some duplication across env files.
- **Deferred:** dynamic values/template tags, reference-based secret-leakage detection, encryption at rest, and the desktop switcher UI.
