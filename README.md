# Reqly

> [!CAUTION]
> Alpha, not stable. Expect breakage.
> Reqly is pre-1.0. It may crash, corrupt local data or break between commits. There are no stability guarantees, no migration path and no semver until v1.0.0. If you need a stable client today, use Postman, Insomnia, Bruno or HTTPie.

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="frontend/src/assets/logo-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="frontend/src/assets/logo-light.svg">
    <img alt="Reqly Logo" src="frontend/src/assets/logo-light.svg" width="180">
  </picture>
</p>

<p align="center">
  A local-first, Git-native API client. Alpha quality.
</p>

<p align="center">
  <a href="#overview">Overview</a> •
  <a href="#what-it-does">What it does</a> •
  <a href="#why-reqly">Why Reqly</a> •
  <a href="#architecture">Architecture</a> •
  <a href="#install">Install</a> •
  <a href="#build-from-source">Build from source</a> •
  <a href="#docs">Docs</a> •
  <a href="#license">License</a>
</p>

## Overview

Reqly keeps your API work on disk. Collections, environments, tests, scripts and request history live as plain text files in your repo. You commit them, review them in pull requests and diff them with Git. There is no cloud backend and no account. Payloads, headers, tokens and traffic never leave your machine.

We built it this way because cloud sync always felt like the wrong place for request data. It is a little more setup, and you lose one-click sharing, but you get version control that actually works. That trade feels worth it to us.

## What it does

- Request files in JSON or YAML that merge cleanly in Git.
- One Go core powers both interfaces, so the CLI and desktop read the same files the same way.
- Desktop app with Wails v3, React and Vite.
- CLI for scripts and CI. Commands include run, test, collection, mock, import, export, env, auth and history.
- Auth coverage for basic, bearer, API key, JWT signing, digest, OAuth 2.0 with client credentials, authorization code with PKCE and device flow, AWS SigV4 and Akamai EdgeGrid.
- WebSocket and SSE clients.
- Mock server generated from an OpenAPI spec, with delay and error injection.
- Retry policies, pagination runners, bulk execution, HAR import and export, plus cURL and OpenAPI import with Postman export.

P0 to P5 are done for core, CLI and desktop bindings. Thirteen deferred items remain and are tracked in code, like file download UI, NTLM, AWS and Azure helpers, XPath, gRPC bidi and SOAP rpc encoded.

## Why Reqly

|  | Cloud clients | Reqly |
| :--- | :--- | :--- |
| Storage | Proprietary cloud database | Plain text files on disk |
| Version control | App history, paid team sync | Git commits, branches, pull requests |
| Telemetry | Usage metrics, sync traffic | None |
| Interfaces | Desktop GUI | Desktop app and CLI |
| Account | Required | Not needed |

Cloud clients give you polished sync out of the box. Reqly gives you files you own. If you value Git history over one-click sharing, this fits. If you need hosted collaboration now, the table above is the honest tradeoff.

## Architecture

```
                    ┌─────────────────────────┐
                    │         Go Core         │
                    │       internal          │
                    └────────────┬────────────┘
         ┌───────────────────────┼───────────────────────┐
         ▼                       ▼                       ▼
   Desktop GUI                Cobra CLI              MCP Server
  Wails v3 plus React       apps cli               internal mcp
```

All execution logic lives in `internal`. Environment resolution, variable interpolation, auth, retries, cookie handling, masking and history run once, in Go. The desktop and CLI only parse input and render output. That is why behavior stays consistent without maintaining two copies.

## Install

Every push to `main` that merges a release PR creates a GitHub Release. The release attaches checksums and binaries for all platforms. Use the scripts below, or download manually from Releases.

### Linux and macOS

The `install.sh` script installs the CLI by default and the desktop app when you pass `--app`.

```bash
# CLI. Installs to /usr/local/bin or ~/.local/bin if not writable.
curl -fsSL https://raw.githubusercontent.com/Its-Satyajit/reqly/main/install.sh | sh

# Desktop. On Linux this tries AppImage, then .deb on apt systems, then .rpm on dnf/zypper.
# On macOS it installs Reqly.app to /Applications.
curl -fsSL https://raw.githubusercontent.com/Its-Satyajit/reqly/main/install.sh | bash -s -- --app

# Pin a version. Defaults to latest.
curl -fsSL https://raw.githubusercontent.com/Its-Satyajit/reqly/main/install.sh | bash -s -- --version v1.2.0
VERSION=v1.2.0 curl -fsSL https://raw.githubusercontent.com/Its-Satyajit/reqly/main/install.sh | sh

# Help
curl -fsSL https://raw.githubusercontent.com/Its-Satyajit/reqly/main/install.sh | bash -s -- --help
```

Linux details. The script detects `amd64` or `arm64` and your package manager. With `--app` it tries `reqly-linux-amd64.AppImage` then `Reqly.AppImage`, then `reqly-linux-amd64.deb` on apt and `reqly-linux-amd64.rpm` on dnf or zypper, with fallbacks to versioned names like `reqly_1.2.0_amd64.deb`. If none of those succeed it falls back to the desktop binary `reqly-desktop-linux-amd64`. Without `--app` it tries the same `.deb` then `reqly-linux-amd64` CLI binary. On pacman systems it first tries `yay` or `paru` for `reqly-bin`.

macOS details. With `--app` the script downloads `Reqly-macos-universal.app.zip` or `Reqly.dmg`, unpacks or mounts it, moves `Reqly.app` to `/Applications` and clears quarantine with `xattr -dr com.apple.quarantine`. Without `--app` it installs `reqly-darwin-arm64` or `reqly-darwin-amd64` to `/usr/local/bin`.

Prerequisites for the Linux packages. You need `libgtk-4-1` and `libwebkitgtk-6.0-4` on Ubuntu 24.04 or Debian 13, `gtk4` and `webkitgtk6.0` on Fedora or RHEL, and `gtk4` and `webkitgtk-6.0` on Arch. The `.deb` and `.rpm` declare these in `nfpm.yaml`.

### Windows

```powershell
# CLI. Installs to %LOCALAPPDATA%\reqly and adds that dir to your user Path.
irm https://raw.githubusercontent.com/Its-Satyajit/reqly/main/install.ps1 | iex

# Desktop. Runs the NSIS installer if present, else copies the desktop exe and creates shortcuts.
irm https://raw.githubusercontent.com/Its-Satyajit/reqly/main/install.ps1 | iex; Install-Reqly -Desktop

# Options
powershell -ExecutionPolicy Bypass -File install.ps1 -Version v1.2.0
powershell -ExecutionPolicy Bypass -File install.ps1 -Desktop -Version v1.2.0 -InstallDir "$env:LOCALAPPDATA\Programs\Reqly"
```

The desktop path tries `reqly-windows-amd64-installer.exe`, then `Reqly-windows-amd64-installer.exe`, then the plain desktop binary `reqly-desktop-windows-amd64.exe` or `Reqly.exe`. It installs to `%LOCALAPPDATA%\Programs\Reqly\Reqly.exe`, creates Desktop and Start Menu shortcuts and adds the install dir to your user Path. The CLI path tries `reqly-windows-amd64.exe` and falls back to `reqly-windows-amd64.zip`.

### Manual download

Open Releases at `https://github.com/Its-Satyajit/reqly/releases`.

CLI:
- `reqly-linux-amd64`, `reqly-linux-arm64`, `reqly-darwin-amd64`, `reqly-darwin-arm64`, `reqly-windows-amd64.exe`, `reqly-windows-arm64.exe` plus `checksums.txt`

Desktop Linux:
- `reqly-desktop-linux-amd64`, `Reqly.AppImage` and `reqly-linux-amd64.AppImage`, `reqly-linux-amd64.deb` and `reqly_1.2.0_amd64.deb`, `reqly-linux-amd64.rpm` plus `checksums-linux.txt`

Desktop macOS:
- `Reqly-macos-universal.app.zip`, `Reqly-macos-universal.dmg` and `Reqly.dmg`, `reqly-desktop-macos-universal` plus `checksums-macos.txt`

Desktop Windows:
- `reqly-desktop-windows-amd64.exe` and `Reqly.exe`, `reqly-windows-amd64-installer.exe` and `Reqly-windows-amd64-installer.exe`, `reqly-windows-amd64.msix` plus `checksums-windows.txt`

Verify after download:

```bash
curl -fsSL https://github.com/Its-Satyajit/reqly/releases/latest/download/checksums.txt -o checksums.txt
curl -fsSL https://github.com/Its-Satyajit/reqly/releases/latest/download/reqly-linux-amd64 -o reqly
sha256sum -c checksums.txt --ignore-missing
```

For desktop use `checksums-linux.txt`, `checksums-macos.txt` or `checksums-windows.txt` the same way.

## Build from source

You need Go 1.27, Node 24 with `nub`, and Wails v3.

```bash
git clone https://github.com/Its-Satyajit/reqly.git
cd reqly
nub install

# Core checks
go test ./...
go test -race ./...
nub run typecheck

# Try the CLI against the companion mock API
go run ./apps/cli run https://reqly-test-api.vercel.app/api/users
go run ./apps/cli run -H "Authorization: Bearer admin-token" https://reqly-test-api.vercel.app/api/auth/me

# Mock a spec locally
go run ./apps/cli mock path/to/openapi.yaml --port 4010

# Desktop in development mode
cd apps/desktop/backend
wails3 generate bindings -d frontend/bindings -i -ts
wails3 dev
```

The `Makefile` also provides `make frontend`, `make go-test`, `make desktop` and `make install-desktop`. `wails3 dev` builds the frontend and runs the app with hot reload.

## Docs

- Architecture decisions in `docs/adr`, 45 ADRs.
- Full UI spec in `docs/Reqly Complete UI Architecture, Pages, Panels, and Navigation Specification.md`.
- Planning history for P0 to P5 and the shell rebuild is archived in Git history. See `ROADMAP.md` and `Milestones`.

## Stability

Reqly is alpha research software until v1.0.0.

- Crashes, hangs, data loss in `.reqly` directories and workspace YAML, and silent misbehavior are expected.
- CLI flags, the request file schema and IPC bindings can change without notice. Pin a commit if you depend on them.
- There is no SLA, no security audit and no stable import or export round trip guarantee.

Keep using Postman, Insomnia, Bruno, HTTPie or Restish for daily work. Check back closer to P1.

## License

Apache 2.0. See `LICENSE`.
