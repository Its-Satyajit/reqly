# Phase 4: Ecosystem & Enterprise (P3)

## Phase 4 — Ecosystem & Enterprise (P3)

Long-term ecosystem and organization features.
**Spec:** [`docs/Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification.md`](docs/Reqly%20Complete%20UI%20Architecture,%20Pages,%20Panels,%20and%20Navigation%20Specification.md) §58

### §58.1 Plugin Marketplace

- [x] Plugin engine (Goja JS runtime, manifest validation, CLI manager `reqly plugin list/validate`) (`internal/plugin`, [M39](docs/spec/m39-openapi-explorer.md)) shipped

### §58.2 Theme Marketplace

- [x] Theme sharing + custom themes — Git-native shareable themes (`id` kebab-case, `label`, `appearance` light/dark, `tokens` primary/background etc.), `internal/theme` (`Validate`/`Parse`/`MarshalYAML`/`MarshalJSON`/`ToCSS`/`BuiltInThemes`/`IsBuiltIn`) + CLI `reqly theme list/export/import` + desktop `AppService.ThemeList/ThemeExport/ThemeImport` bindings; UI picker/extensions pending — `internal/theme` (M67) shipped (core + CLI + desktop shipped / UI extensions pending)

### §58.3 Git Provider Integrations

- [x] GitHub, GitLab, Bitbucket, and Azure DevOps integration, remote host auto-detection, PAT token management, CLI `reqly auth login --provider <github|gitlab|bitbucket|azure-devops>`, and Goja `reqly.git.detectProvider` binding (`internal/git/provider`, [M61](docs/spec/m61-git-providers.md), [ADR 0045](docs/adr/0045-git-provider-integrations.md)) shipped

### §58.4 Team / Shared Workspaces

- [x] Shared workspaces & collaboration — Git-native shared workspace (`.reqly/collab.yaml` 0600, `SharedWorkspace{Path,Collaborators}` + `Collaborator{User,Role,AddedAt}` + `Validate`/`AddCollaborator`/`RemoveCollaborator`/`IsCollaborator` + `Load`/`Save`/`DefaultPath` 0700/0600) — `internal/collab` + CLI `reqly collab list/add/remove` + desktop `AppService.CollabList/CollabAdd/CollabRemove` — `internal/collab` (M74) shipped (core + CLI + desktop shipped / self-hosted server deferred)

### §58.5 Enterprise

- [x] Self-hosted collaboration server — local HTTP server for shared workspaces (`Server{root,mux}` + `NewServer` + `Handler` + `/health`/`/collab`/`/workspace` (collaborators+collections), `net.Listen` ephemeral, `http.Serve`) — `internal/collab.Server` + CLI `reqly collab serve --port` + desktop `AppService.CollabServe(port)` — `internal/collab` (M75) shipped (core + CLI + desktop shipped)
- [x] Enterprise SSO & SCIM — local SSO (`Config{Issuer,ClientID,JWKSURL,AllowedGroups}` + `Validate`/`ValidateToken` HMAC via `jwt.Verify` + `IsGroupAllowed`) + SCIM in-memory store (`User{ID,UserName,Email,Groups,Active}`/`Group{ID,DisplayName,Members}` + `ValidateUser/Group` + `Store{CreateUser/GetUser/ListUsers/DeactivateUser/CreateGroup/AddUserToGroup}`) — `internal/sso` + `internal/scim` + CLI `reqly sso validate` + `reqly scim user create/list` + desktop `AppService.SSOValidate/SCIMCreateUser/SCIMListUsers` — `internal/sso`+`scim` (M73) shipped (core + CLI + desktop shipped)
- [x] Audit logs — local append-only trail (`.reqly/audit.log` JSONL, 0600, `Entry{ID,Timestamp,Actor,Action,Resource,Details}` + `Validate` 11 actions) — `internal/audit` (`NewStore`/`Add`/`List`/`Clear` 0700/0600) + CLI `reqly audit list/clear` + desktop `AppService.AuditList/Add/Clear/Export` — `internal/audit` (M69) shipped (core + CLI + desktop shipped / org policies deferred)
- [x] Organization policies — local policy file (`.reqly/policy.yaml` 0600, `Policy{RequireAudit,MaxWorkflowSteps,AllowedActions,RequireAuth,AllowCustomThemes}` + `Validate`/`Enforce`/`EnforceWorkflow` + `DefaultPolicy`/`Load`/`Save`/`DefaultPath` 0700/0600) — `internal/policy` + CLI `reqly policy show/validate/enforce` + desktop `AppService.PolicyGet/PolicySave/PolicyEnforce` — `internal/policy` (M70) shipped (core + CLI + desktop shipped / SSO/SCIM deferred)
- [x] Enterprise secret management (Vault) — HashiCorp Vault KV v2 store (`VaultStore{Addr,Token,Mount,Prefix}` + `Get`/`Set`/`Delete`/`Keys` via `X-Vault-Token`, `secret/data/prefix/<key>`) — `internal/secrets.VaultStore` + `REQLY_TOKEN_STORE=vault` with `VAULT_ADDR`/`VAULT_TOKEN`/`VAULT_MOUNT`/`VAULT_PREFIX` env, fallback to file store; AWS/Azure deferred — `internal/secrets` (M72) shipped (core + backend shipped / AWS/Azure deferred)
- [x] Advanced access control (RBAC) — local RBAC (`.reqly/rbac.yaml` 0600, `RBAC{Roles,UserRoles}` + `Role{Name,Permissions}` + `DefaultRBAC` admin/editor/viewer + `Validate`/`Can`/`Enforce`/`ListRoles` + `Load`/`Save`/`DefaultPath` 0700/0600) — `internal/rbac` + CLI `reqly rbac list/check` + desktop `AppService.RBACList/RBACCheck/RBACGet` — `internal/rbac` (M71) shipped (core + CLI + desktop shipped / collaboration/SSO deferred)

---

## Code Review Gate (`/code-review` — two-axis)

- [x] Standards: `oxlint` + `gofmt`/`go vet` + `anti-slop` + Fowler smell baseline — `git diff main...HEAD` (three-dot, merge-base) — no `as` without `// SAFETY:`, no hard violations
- [x] Spec: this milestone (P3 `Milestones/05` §58 + M67-M75) vs implementation (`ROADMAP.md` Phase 4 + DoD: core+UI/CLI+tests) — `git log main..HEAD` + `git diff main...HEAD` — both axes must be green before ticking `[x]` above; fix `main...HEAD` diff until green — run `/code-review`

