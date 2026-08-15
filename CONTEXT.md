# Domain Glossary

### Release Pipeline Architecture
The dual-orchestration pipeline for publishing Reqly releases to GitHub (`Its-Satyajit/reqly`). It separates headless Go CLI compilation (managed via GoReleaser) from native GUI desktop app bundling (managed via Wails v3 OS build matrix across macOS, Linux, and Windows).

### Linux Artifact Matrix
The multi-distro Linux distribution strategy for Reqly binaries. Includes Debian `.deb`, universal standalone `.AppImage`, compressed `.tar.gz`, Arch Linux `PKGBUILD` (AUR `reqly-bin`), and multi-distro detection in `install.sh` supporting `pacman`, `apt`, `dnf`, and binary fallback.

### Unsigned Release Fallback
The zero-cost release strategy for open-source distributions. macOS binaries use ad-hoc code signing (`codesign -s -`) with CLI instructions (`xattr -d com.apple.quarantine`) for Gatekeeper, Windows binaries use unsigned `.exe`/`.zip`, and Linux binaries use native `.AppImage` and `.tar.gz` without requiring paid Apple Developer or Microsoft certificates.

### Automated Tag Release Trigger
The event-driven GitHub Actions workflow triggered by pushing semver tags (`git push origin v*.*.*`). Automatically executes the GoReleaser + Wails v3 OS matrix, generates release notes from Conventional Commits, uploads binaries, and updates checksums.

### Multi-Distro Install Script
The POSIX-compliant installation one-liner (`curl -fsSL https://raw.githubusercontent.com/Its-Satyajit/reqly/main/install.sh | sh`). Detects operating system and Linux package manager (`pacman`, `apt`, `dnf`, `zypper`), installing appropriate binaries or system packages to `/usr/local/bin/reqly`.

### AI Agent Protocol & Workspace Rules
The standardized instruction setup (`AGENTS.md`, `GEMINI.md`, `.cursorrules`, `.github/copilot-instructions.md`) governing AI coding assistant contributions to Reqly. Enforces TDD verification (`go test ./...`), domain model alignment via the `grill-with-docs` skill, and local-first/Git-native architectural constraints.


