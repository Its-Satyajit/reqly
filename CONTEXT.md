# Domain Glossary

### Release Pipeline Architecture
Dual-orchestration pipeline publishing releases to GitHub (`Its-Satyajit/reqly`). Separates headless Go CLI compilation (GoReleaser) from native GUI desktop app bundling (Wails v3 OS matrix across macOS, Linux, Windows).

### Linux Artifact Matrix
Multi-distro Linux distribution strategy for Reqly binaries. Covers Debian `.deb`, universal `.AppImage`, compressed `.tar.gz`, Arch Linux AUR (`reqly-bin`), and multi-distro detection in `install.sh` supporting `pacman`, `apt`, `dnf`, and binary fallback.

### Unsigned Release Fallback
Zero-cost release strategy for open-source distributions. macOS binaries use ad-hoc code signing (`codesign -s -`) with quarantine removal (`xattr -d com.apple.quarantine`), Windows binaries use unsigned `.exe`/`.zip`, and Linux binaries use native `.AppImage`/`.tar.gz`.

### Automated Tag Release Trigger
Event-driven GitHub Actions workflow triggered by pushing semver tags (`git push origin v*.*.*`). Executes GoReleaser and Wails v3 OS matrix, generates release notes from Conventional Commits, uploads binaries, and updates checksums.

### Multi-Distro Install Script
POSIX-compliant installation script (`curl -fsSL https://raw.githubusercontent.com/Its-Satyajit/reqly/main/install.sh | sh`). Detects OS and Linux package manager (`pacman`, `apt`, `dnf`, `zypper`), installing binaries or system packages to `/usr/local/bin/reqly`.

### AI Agent Protocol & Workspace Rules
Standardized instruction configuration (`AGENTS.md`, `GEMINI.md`, `.cursorrules`, `.github/copilot-instructions.md`) governing AI coding assistant contributions to Reqly. Enforces TDD verification (`go test ./...`), local-first/Git-native constraints, and the 5-stage skill pipeline (`/grill-with-docs` → `/to-spec` → `/to-tickets` → `/implement` → `/code-review`) in `~/.agents/skills/`.

### Specification & Project Validation Engine
Static verification subsystem (`internal/validation`) backing `reqly validate`. Audits OpenAPI 2.x/3.x specifications (`reqly validate openapi <path>`) and local Git-native project descriptors (`reqly validate project [path]`) for broken schema references, missing required attributes, and invalid descriptor syntax.

### Structural Diff Engine
AST and JSON/YAML key-aware diffing engine (`internal/diffing`) backing `reqly diff`. Computes semantic differences between API specifications (`reqly diff spec <spec1> <spec2>`), request descriptors (`reqly diff request <file1> <file2>`), and JSON/YAML response payloads (`reqly diff response <file1> <file2>`).

### Environment
A named set of variables (plus optional secrets) selectable per workspace and applied to requests through the variable precedence chain between the global and collection scopes.

### Active Environment
The environment currently selected for a workspace; resolved with the following precedence (highest wins): `REQLY_ENV` process variable, then the `--env` CLI flag, then a request/test file's `environment:` field, then the workspace descriptor's `environment:` field.

### Environment File
A YAML file under `environments/<name>.yaml` holding an environment's `variables:` and `secrets:` maps. Resolved from the nearest `environments/` directory — CWD (or the request file's directory) for standalone `run`/`test`, the workspace root for collection commands. The filesystem is the source of truth.

### Secret
An environment variable marked as sensitive; its value is masked in CLI and test output and never printed by `reqly env show`.

### Process Environment Scope
The lowest-precedence variable scope fed by the OS environment and the nearest-directory `.env` file, mirroring Node's `process.env`. When both define a key, the OS environment wins, so CI can override local `.env` values.

### Environment Validation
Workspace-aware static checks (`reqly env validate <name>`) over an environment: file syntax, secret-exposure warnings (name-heuristic for `key/token/secret/password/credential` plus duplicate keys in both `variables:` and `secrets:`), and undefined-variable detection across the workspace's requests under the active env.

### Environment Diff
Secret-aware comparison of two environments (`reqly env diff <nameA> <nameB>`): added/removed/changed keys via the structural diff engine, with changed secret values rendered as `[SECRET]` so the diff shows *which* secret changed without leaking it.

### Authentication Scheme
A named credential-application strategy (`internal/auth`) that mutates an outgoing request. Dispatched by a request's `auth.type` string through a registered `Scheme` interface; each scheme owns its `config` keys and applies them to the request before send.

### Auth Config
The flat string map on a request/collection descriptor (`auth.config`) holding a scheme's fields (e.g. `token`, `username`, `password`, `key`, `in`). Values are interpolated like variables and fed to the environment masker so secrets never leak in output.

### Auth Inheritance
A request resolves its auth from the nearest enclosing collection/folder that defines one; a request's own non-empty `auth.type` overrides it, and `auth.type: none` explicitly disables inherited auth for that request.

### No Auth
The `none` scheme: clears any inherited auth on a request so it is sent unauthenticated, even under an auth-bearing collection/folder.

### JWT Auth
A scheme that signs a JSON Web Token per request from `config` (secret, algorithm, claims) and sends it as `Authorization: Bearer <token>`. Distinct from JWT *tooling* (decode/claims viewer), which is a separate CLI feature.

