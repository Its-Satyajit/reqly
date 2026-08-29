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

- [ ] Multi-user collaboration, shared workspaces

### §58.5 Enterprise

- [ ] Self-hosted collaboration server
- [ ] Enterprise SSO, SCIM provisioning
- [x] Audit logs — local append-only trail (`.reqly/audit.log` JSONL, 0600, `Entry{ID,Timestamp,Actor,Action,Resource,Details}` + `Validate` 11 actions) — `internal/audit` (`NewStore`/`Add`/`List`/`Clear` 0700/0600) + CLI `reqly audit list/clear` + desktop `AppService.AuditList/Add/Clear/Export` — `internal/audit` (M69) shipped (core + CLI + desktop shipped / org policies deferred)
- [ ] Organization policies
- [ ] Enterprise secret management (Vault, AWS, Azure, role-based access)
- [ ] Advanced access control / permissions

---

