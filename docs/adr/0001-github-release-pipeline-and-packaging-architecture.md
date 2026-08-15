# 1. GitHub Release Pipeline & Multi-Platform Packaging Architecture

* Status: accepted  
* Date: 2026-08-15  

## Context and Problem Statement

Reqly requires a reliable, automated GitHub release pipeline for cross-platform desktop application bundles (`macOS .dmg`, `Linux .deb/.AppImage`, `Windows .exe/.zip`) and standalone CLI binaries (`reqly`).

Open-source contributors must be able to build releases without requiring paid Apple Developer certificates or complex Cgo cross-compiler setups.

## Decision Drivers

* **Zero-Cost Open Source**: Support macOS building and installation without requiring a $99/year Apple Developer ID subscription.
* **Linux Multi-Distro Support**: Full support for Arch Linux (`pacman` / `PKGBUILD`), Debian/Ubuntu (`.deb`), and universal standalone `.AppImage`.
* **CI Reliability**: Avoid Cgo cross-compilation failures across macOS/Linux/Windows.
* **Instant Installation**: Support one-line terminal installation (`curl -fsSL ... | sh`) and instant SHA-256 hash verification.

## Considered Options

* **Option A**: Dual Matrix Workflow (GoReleaser for CLI + Wails v3 OS Matrix for Desktop App) with `.AppImage`, Arch `PKGBUILD`, and ad-hoc code signing.
* **Option B**: Monolithic GoReleaser build attempting Cgo cross-compilation across platforms.
* **Option C**: Manual release drafting in GitHub UI.

## Decision Outcome

Chosen Option: **Option A**.

### Positive Consequences

* Headless CLI builds complete in seconds via GoReleaser on Ubuntu runners.
* Desktop GUI app bundles (`.dmg`, `.deb`, `.AppImage`, `.exe`) build natively on OS-specific GitHub Actions runners (`macos-14`, `ubuntu-24.04`, `windows-2025`).
* Arch Linux users receive native `PKGBUILD` and Arch-aware `install.sh` detection.
* macOS users receive ad-hoc signed `.dmg` bundles with clear Gatekeeper instructions.
* Releases trigger automatically upon pushing `v*.*.*` tags.

### Negative Consequences

* Unsigned macOS binaries require users to run `xattr -d com.apple.quarantine` if Gatekeeper blocks initial launch.

## Pros and Cons of the Options

### Option A: Dual Matrix Workflow (GoReleaser + Wails v3 OS Matrix)

* Good, because CLI and Desktop builds are decoupled and compile natively without Cgo cross-compiler errors.
* Good, because Linux users get native `.AppImage`, `.deb`, `.tar.gz`, and Arch `PKGBUILD`.
* Bad, because it requires maintaining both GoReleaser config (`.goreleaser.yaml`) and GitHub Actions workflow (`.github/workflows/release.yml`).
