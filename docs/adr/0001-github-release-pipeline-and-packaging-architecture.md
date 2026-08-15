# ADR 0001: GitHub Release Pipeline & Multi-Platform Packaging Architecture

- **Status:** Accepted
- **Date:** 2026-08-15
- **Deciders:** Maintainers

## Context & Problem Statement

Reqly requires automated cross-platform releases for desktop bundles (`macOS .dmg`, `Linux .deb/.AppImage`, `Windows .exe/.zip`) and standalone CLI binaries (`reqly`). Builds must run on free GitHub runners without requiring paid Apple Developer certificates or Cgo cross-compiler setups.

## Decision Drivers

- **Zero-Cost Open Source:** Support macOS installation without paid Developer ID subscriptions.
- **Linux Multi-Distro Support:** Target Arch Linux (`pacman`/`PKGBUILD`), Debian/Ubuntu (`.deb`), and universal `.AppImage`.
- **CI Reliability:** Eliminate Cgo cross-compilation failures across OS targets.
- **Instant Installation:** Support one-line installation (`curl -fsSL ... | sh`) with hash verification.

## Considered Options

- **Option A:** Dual-orchestration pipeline (GoReleaser for CLI + Wails v3 OS matrix for Desktop GUI) with ad-hoc signing.
- **Option B:** Single GoReleaser job attempting Cgo cross-compilation.
- **Option C:** Manual GitHub release uploads.

## Decision Outcome

Chosen **Option A**.

### Consequences

- **Positive:** CLI compiles in seconds; GUI apps build natively on OS-specific runners (`macos-14`, `ubuntu-24.04`, `windows-2025`).
- **Positive:** Native packaging for Arch (`PKGBUILD`), Debian (`.deb`), and standalone `.AppImage`.
- **Negative:** Unsigned macOS binaries require `xattr -d com.apple.quarantine` on initial launch.
