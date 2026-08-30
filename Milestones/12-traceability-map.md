# Traceability Map — Roadmap Milestone → Implementation

**Source of truth:** `ROADMAP.md` (canonical product roadmap), `Milestones/01-06` (grouped specs), `docs/internal/gui-roadmap.md` (desktop), `docs/Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification.md` (subordinate UI reference).

**Legend:** `core` = `internal/` Go package, `CLI` = `apps/cli/cmd`, `desktop` = `apps/desktop/backend` + `frontend/src`, `tests` = unit/integration, `ADR` = `docs/adr/`.

| Milestone | ROADMAP § | Core | CLI | Desktop | Tests | ADR / Spec |
|-----------|-----------|------|-----|---------|-------|------------|
| M65 Workflow Engine | §57.4 | `internal/workflow` (`Workflow`, `WorkflowExecutor`, `Condition` Goja, `Extract`, `Interpolate`) | `reqly workflow <file>` | `AppService.WorkflowRun` + Goja `reqly.workflow.run` | `workflow_test.go` (7), `workflow_test.go` CLI, `workflow_script_test.go` | ADR 0033, Spec M65 |
| M66 Self-Hosted Automation | §57.5 | `internal/automation` (`Automation`, `Scheduler.Run` ticker) | `reqly automation run` | `AppService.AutomationRun` | `automation_test.go` (6), CLI, desktop | ADR 0034, Spec M66 |
| M67 Theme Sharing | §58.2 | `internal/theme` (`Validate`, `Parse`, `ToCSS`, `BuiltInThemes`) | `reqly theme list/export/import` | `AppService.ThemeList/Export/Import` + `frontend/src/lib/themes.ts` (`CustomTheme`, `validateCustomTheme`, `parseCustomTheme`, `customThemeToCSS`) | `theme_test.go` (5), `theme_test.go` CLI, `theme_test.go` desktop, `frontend/src/lib/themes.test.ts` (15) | Spec M67 |
| M68 Frontend Vitest | Quality | — | — | — | `frontend/vitest.config.ts` jsdom, `nub run --filter @reqly/frontend test` (20 files/160 tests) | Milestones/06 update |
| M69 Audit Logs | §58.5 | `internal/audit` (`Entry`, `Store` JSONL 0600) | `reqly audit list/clear` | `AppService.AuditList/Add/Clear/Export` | `audit_test.go` (4), CLI, desktop | Spec M69 |
| M70 Organization Policies | §58.5 | `internal/policy` (`Policy`, `Validate`, `Enforce`, `Load/Save`) | `reqly policy show/validate/enforce` | `AppService.PolicyGet/Save/Enforce` | `policy_test.go` (5), CLI, desktop | Spec M70 |
| M71 RBAC | §58.5 | `internal/rbac` (`RBAC`, `Role`, `DefaultRBAC`, `Validate`, `Can`, `Enforce`) | `reqly rbac list/check` | `AppService.RBACList/Check/Get` | `rbac_test.go` (5), CLI, desktop | Spec M71 |
| M72 Vault Secrets | §58.5 | `internal/secrets.VaultStore` (`X-Vault-Token`, `secret/data/prefix`) | `REQLY_TOKEN_STORE=vault` via `VAULT_ADDR/TOKEN` env | — (backend) | `vault_test.go` (2) + `secrets` suite | — |
| M73 SSO/SCIM | §58.5 | `internal/sso` (`Config`, `ValidateToken`, `IsGroupAllowed`) + `internal/scim` (`User/Group`, `Store`) | `reqly sso validate`, `reqly scim user create/list` | `AppService.SSOValidate/SCIMCreateUser/ListUsers` | `sso_test.go` (3), `scim_test.go` (4), CLI, desktop | Spec M73 |
| M74 Shared Workspaces | §58.4 | `internal/collab` (`SharedWorkspace`, `Add/Remove/IsCollaborator`, `Load/Save`) | `reqly collab list/add/remove` | `AppService.CollabList/Add/Remove` | `collab_test.go` (3), CLI, desktop | Spec M74 |
| M75 Collab Server | §58.4 | `internal/collab.Server` (`/health`, `/collab`, `/workspace`, `Handler`, `net.Listen`) | `reqly collab serve --port` | `AppService.CollabServe(port)` | `server_test.go` (4) + collab suite | Spec M75 |

**Coverage:** Every P2/P3 feature shipped since M65 appears exactly once in `ROADMAP.md` Phase 2/3/4 and once in `Milestones/04-05`. No shipped item regressed to `[ ]`. No UI-spec item promoted to product scope without a roadmap milestone. Deferred seams (cron, JWKS RS256, AWS/Azure, visual builder UI, collab server UI) remain explicit follow-ups in each ADR/spec.

**GUI linkage:** Each desktop binding above is the product-roadmap owner for its GUI counterpart (e.g. `WorkflowRun` → future visual DAG, `ThemeList` → future picker, `CollabServe` → future server panel). See `docs/internal/gui-roadmap.md` for GUI execution tracking.

**Verification:** All 12 milestones pass `go test ./...` (46 pkgs), `go test -race`, `go vet`, `gofmt -l`, `go build -o reqly`, `nub run typecheck`, `oxlint`, `vitest run` before `but commit`.

---

## Code Review Gate (`/code-review` — two-axis)

- [x] Standards: `oxlint` + `gofmt`/`go vet` + `anti-slop` + Fowler smell baseline — `git diff main...HEAD` (three-dot, merge-base) — no `as` without `// SAFETY:`, no hard violations
- [x] Spec: this milestone (`Milestones/` + Phase) vs implementation (`ROADMAP.md` DoD: core+UI/CLI+tests) — `git log main..HEAD` + `git diff main...HEAD` — both axes must be green before ticking `[x]` above; fix `main...HEAD` diff until green — run `/code-review`
