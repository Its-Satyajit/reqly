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
