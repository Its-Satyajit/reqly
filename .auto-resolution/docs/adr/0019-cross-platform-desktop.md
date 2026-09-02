# ADR 0019: Cross-Platform Desktop (M27)

## Status
Accepted

## Context
`ROADMAP.md:192` (`Linux build verified, macOS/Windows CI/signing`) is the last P0 build gap: `Taskfile.yml` has `linux:build`/`darwin:build`/`windows:build` (wails3) and `release.yml` has `Automated Tag Release Trigger` `git push origin v*.*.*` → `GoReleaser` + `Wails` matrix, but the OS matrix (`ubuntu-latest` `libgtk-4-dev` `libwebkitgtk-6.0-dev`, `macos-latest` `codesign -s -`, `windows-latest` `WebView2`) and `Linux Artifact Matrix` (`.deb`/`.AppImage`/`.tar.gz` + `AUR` + `install.sh` multi-distro) are not verified in CI, and `Unsigned Release Fallback` (ad-hoc `codesign -s -`, `xattr -d`, unsigned `.exe`) is the `0$` OSS fallback. The design questions are which OS matrix ships in M27, which artifacts, and which signing is deferred.

## Decision
1. **OS matrix `linux`/`darwin`/`windows` via `Taskfile.yml` + `release.yml` `strategy: matrix`.** `ubuntu-latest` (WebKitGTK 6.0) → `linux` `amd64` `AppImage`/`deb`/`tar.gz` via `Taskfile.yml` `linux:build` (`wails3 build` + `Taskfile.yml`), `macos-latest` → `darwin` `amd64`/`arm64` `dmg` (ad-hoc `codesign -s -`, `xattr -d`), `windows-latest` → `windows` `amd64` `.exe`/`.zip` (unsigned `WebView2`) via `Taskfile.yml` `windows:build`; `GoReleaser` CLI `goreleaser` job for `bin/reqly` (like `ci.yml` `frontend` `wails3 generate bindings`).
2. **Artifacts `linux` 3 + `darwin` 2 + `windows` 2 + `checksums.txt` + `bin/reqly`.** `dist/reqly_linux_amd64.tar.gz`/`deb`/`AppImage`, `dist/reqly_darwin_amd64.dmg`/`arm64.dmg`, `dist/reqly_windows_amd64.exe`/`zip`, `checksums.txt` (`sha256sum dist/*`), `bin/reqly` via `GoReleaser`; `install.sh` `Multi-Distro` (`pacman`/`apt`/`dnf`/`zypper` + `binary` fallback to `/usr/local/bin/reqly`) unchanged, no `AUR` automation for M27 (`reqly-bin` manual).
3. **`release.yml` `Automated Tag Release Trigger` `push-right` with `Unsigned Release Fallback` (`0$`).** `git push origin v*.*.*` → `GoReleaser` + `Wails` matrix + `checksums` + `Conventional Commits` `CHANGELOG.md` draft `gh release create --draft` + `checksums.txt`, `workflows/release.md` brief `Release v1.3.0 — CLI (3) + desktop (4) — checksums ok` (human `approve` before `gh release create` publish, `but land` not used). `Apple Developer`/`AzureSignTool` certs, `AUR` `makepkg`, `zypper` `openSUSE`, `homebrew` tap deferred to M27b.

## Considered Options
- **No `darwin`/`windows` matrix in M27 (linux only)** — rejected: `1.12` requires all three, even with `Unsigned Fallback`; `Wails` already supports `WebKit`/`WebView2`.
- **`internal/release` package** — rejected: `Taskfile.yml` + `release.yml` + `install.sh` are the seams (like `ci.yml`), no `internal/*` change, blast radius `3` files + `build/` assets.
- **Apple `Distribution` + `notarization` + `AzureSignTool` in M27** — rejected: needs `Apple Developer`/`Azure` secrets for `0$` OSS; `Unsigned Fallback` is `0$` per `CONTEXT.md:9`.

## Consequences
- **Positive:** `Taskfile.yml` + `release.yml` + `install.sh` close P0 `1.12` with one OS matrix, `Linux Artifact Matrix` + `Unsigned Fallback` + `Automated Tag Release Trigger` + `Multi-Distro Install Script` all verified via `wails3 task build` smoke + `release.yml` `sha256sum` + `install.sh` `shellcheck`/`bats`.
- **Trade-off:** `Apple`/`Azure` certs, `AUR` automation, `zypper` `openSUSE`, `homebrew` tap are M27b — `M27` is unsigned `0$`.

