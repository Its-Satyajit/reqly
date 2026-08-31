#!/usr/bin/env bash
# Reqly - Multi-distro installer for Linux and macOS (amd64/arm64)
# Copyright (C) 2026 It's Satyajit
# SPDX-License-Identifier: Apache-2.0
# Usage: curl -fsSL https://raw.githubusercontent.com/Its-Satyajit/reqly/main/install.sh | sh
#        curl -fsSL https://raw.githubusercontent.com/Its-Satyajit/reqly/main/install.sh | bash -s -- --app     # desktop app
#        curl -fsSL https://raw.githubusercontent.com/Its-Satyajit/reqly/main/install.sh | bash -s -- --version v1.2.0
set -e

REPO="Its-Satyajit/reqly"
BIN_NAME="reqly"
INSTALL_DIR="/usr/local/bin"
VERSION="${VERSION:-latest}"
WANT_APP=false
WANT_DESKTOP=false

for arg in "$@"; do
  case "$arg" in
    --app|--appimage|--desktop) WANT_APP=true; WANT_DESKTOP=true ;;
    --cli) WANT_APP=false; WANT_DESKTOP=false ;;
    --version=*|--version\ *) : ;; # handled via VERSION env
    -h|--help)
      echo "Usage: install.sh [--app|--desktop] [--cli] [--version vX.Y.Z]"
      echo "  --app, --desktop  Install desktop app (AppImage/.deb/.rpm on Linux, .app/.dmg on macOS)"
      echo "  --cli             Install CLI only (default)"
      echo "Env: VERSION=latest|v1.2.0  InstallDir override via INSTALL_DIR"
      exit 0
      ;;
  esac
done
# Also support VERSION as first arg like --version v1.2.0
for arg in "$@"; do
  if [ "$arg" = "--version" ]; then
    # next arg is version
    next_is_version=true
    continue
  fi
  if [ "${next_is_version:-}" = true ]; then
    VERSION="$arg"
    next_is_version=false
  fi
  case "$arg" in
    v*|latest) # bare version argument
      # only treat as version if previous arg was --version or if no other flag matched and it looks like a version
      if [ "$WANT_APP" = false ] && [ "$WANT_DESKTOP" = false ]; then
        # heuristic: if arg looks like v1.2.0, accept as version fallback
        case "$arg" in v*.*.*) VERSION="$arg" ;; esac
      fi
      ;;
  esac
done

detect_os() {
  case "$(uname -s)" in
    Linux*) echo "linux" ;;
    Darwin*) echo "darwin" ;;
    *) echo "unsupported" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) echo "amd64" ;;
  esac
}

detect_linux_pm() {
  if command -v pacman >/dev/null 2>&1; then echo "pacman"; return; fi
  if command -v apt-get >/dev/null 2>&1; then echo "apt"; return; fi
  if command -v dnf >/dev/null 2>&1; then echo "dnf"; return; fi
  if command -v zypper >/dev/null 2>&1; then echo "zypper"; return; fi
  echo "none"
}

download_url() {
  local os="$1" arch="$2" version="$3" ext="$4"
  if [ "$version" = "latest" ]; then
    echo "https://github.com/${REPO}/releases/latest/download/reqly-${os}-${arch}${ext}"
  else
    echo "https://github.com/${REPO}/releases/download/${version}/reqly-${os}-${arch}${ext}"
  fi
}

# Desktop-specific URL builder: supports generic names like Reqly.AppImage, Reqly.dmg, etc.
desktop_url() {
  local file="$1" version="$2"
  if [ "$version" = "latest" ]; then
    echo "https://github.com/${REPO}/releases/latest/download/${file}"
  else
    echo "https://github.com/${REPO}/releases/download/${version}/${file}"
  fi
}

# Try to download first successful URL from list; prints target path on success
try_download() {
  local target="$1"; shift
  local url
  for url in "$@"; do
    echo "  Trying $url..." >&2
    if command -v curl >/dev/null 2>&1; then
      if curl -fsSL "$url" -o "$target" 2>/dev/null; then
        echo "$url"
        return 0
      fi
    elif command -v wget >/dev/null 2>&1; then
      if wget -q "$url" -O "$target" 2>/dev/null; then
        echo "$url"
        return 0
      fi
    fi
  done
  return 1
}

install_appimage() {
  local arch="$1"
  local tmpdir
  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT
  echo "Downloading Reqly Desktop AppImage ($arch)..."
  local target_dir="$HOME/Applications"
  mkdir -p "$target_dir"
  local target_path="$target_dir/Reqly.AppImage"
  local f1 f2 f3 f4
  f1=$(desktop_url "reqly-linux-${arch}.AppImage" "$VERSION")
  f2=$(desktop_url "Reqly.AppImage" "$VERSION")
  f3=$(desktop_url "reqly.AppImage" "$VERSION")
  # versioned nfpm AppImage names (fallback if release uses original)
  # these are guessed; try_download will handle failures gracefully
  local ver="${VERSION#v}"
  f4=$(desktop_url "reqly_${ver}_${arch}.AppImage" "$VERSION")

  if try_download "$target_path" "$f1" "$f2" "$f3" "$f4"; then
    chmod +x "$target_path"
    echo "Reqly AppImage installed to: $target_path"
    echo "Run: $target_path  or  chmod +x and double-click in file manager"
    return 0
  else
    echo "AppImage download failed. Falling back..."
    return 1
  fi
}

install_deb() {
  local arch="$1"
  local tmpdir
  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT
  echo "Downloading Reqly .deb ($arch)..."
  local urls=()
  urls+=("$(desktop_url "reqly-linux-${arch}.deb" "$VERSION")")
  urls+=("$(desktop_url "reqly_${VERSION#v}_${arch}.deb" "$VERSION")")
  urls+=("$(desktop_url "reqly_${VERSION#v}_linux_${arch}.deb" "$VERSION")")
  urls+=("$(desktop_url "Reqly.deb" "$VERSION")")

  local deb_path="$tmpdir/reqly.deb"
  if try_download "$deb_path" "${urls[@]}"; then
    echo "Installing .deb..."
    if sudo dpkg -i "$deb_path" 2>/dev/null; then
      echo "Installed via dpkg"
      return 0
    else
      echo "dpkg failed, trying apt fix..."
      sudo apt-get install -f -y && return 0 || true
      return 1
    fi
  fi
  return 1
}

install_rpm() {
  local arch="$1"
  # dnf uses x86_64 / aarch64, map amd64->x86_64, arm64->aarch64 for rpm fallback
  local rpm_arch="$arch"
  [ "$arch" = "amd64" ] && rpm_arch="x86_64"
  [ "$arch" = "arm64" ] && rpm_arch="aarch64"
  local tmpdir
  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT
  echo "Downloading Reqly .rpm ($arch)..."
  local urls=()
  urls+=("$(desktop_url "reqly-linux-${arch}.rpm" "$VERSION")")
  urls+=("$(desktop_url "reqly-${VERSION#v}-1.${rpm_arch}.rpm" "$VERSION")")
  urls+=("$(desktop_url "reqly_${VERSION#v}_${arch}.rpm" "$VERSION")")

  local rpm_path="$tmpdir/reqly.rpm"
  if try_download "$rpm_path" "${urls[@]}"; then
    echo "Installing .rpm..."
    if command -v dnf >/dev/null 2>&1; then
      sudo dnf install -y "$rpm_path" && return 0 || true
    fi
    if command -v yum >/dev/null 2>&1; then
      sudo yum localinstall -y "$rpm_path" && return 0 || true
    fi
    if command -v zypper >/dev/null 2>&1; then
      sudo zypper install -y "$rpm_path" && return 0 || true
    fi
    sudo rpm -Uvh "$rpm_path" && return 0 || true
  fi
  return 1
}

install_arch_pkg() {
  local arch="$1"
  local tmpdir
  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT
  echo "Downloading Reqly Arch package ($arch)..."
  local urls=()
  urls+=("$(desktop_url "reqly.pkg.tar.zst" "$VERSION")")
  urls+=("$(desktop_url "reqly-${VERSION#v}-1-${arch}.pkg.tar.zst" "$VERSION")")
  urls+=("$(desktop_url "reqly-${VERSION#v}-${arch}.pkg.tar.zst" "$VERSION")")
  local pkg_path="$tmpdir/reqly.pkg.tar.zst"
  if try_download "$pkg_path" "${urls[@]}"; then
    echo "Installing Arch package..."
    if sudo pacman -U --noconfirm "$pkg_path" 2>/dev/null; then
      echo "Installed via pacman -U"
      return 0
    else
      echo "pacman -U failed, trying to extract binary..."
      tar -I zstd -xf "$pkg_path" -C "$tmpdir" 2>/dev/null || tar -xf "$pkg_path" -C "$tmpdir" 2>/dev/null || true
      local bin
      bin=$(find "$tmpdir" -name "reqly" -type f | head -n 1)
      if [ -n "$bin" ]; then
        chmod +x "$bin"
        if [ -w "/usr/local/bin" ]; then
          install -m 755 "$bin" "/usr/local/bin/reqly"
        else
          mkdir -p "$HOME/.local/bin"
          install -m 755 "$bin" "$HOME/.local/bin/reqly"
        fi
        echo "Installed reqly binary from Arch package"
        return 0
      fi
    fi
  fi
  return 1
}

install_desktop_binary_linux() {
  local arch="$1"
  local tmpdir
  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT
  echo "Downloading Reqly Desktop binary (linux/$arch)..."
  local urls=()
  urls+=("$(desktop_url "reqly-desktop-linux-${arch}" "$VERSION")")
  urls+=("$(desktop_url "reqly-linux-${arch}" "$VERSION")")
  urls+=("$(desktop_url "reqly" "$VERSION")")

  local bin_path="$tmpdir/reqly"
  if try_download "$bin_path" "${urls[@]}"; then
    chmod +x "$bin_path"
    local dest="/usr/local/bin/reqly-desktop"
    if [ -w "/usr/local/bin" ]; then
      install -m 755 "$bin_path" "$dest"
      echo "Installed desktop binary to $dest"
    else
      mkdir -p "$HOME/.local/bin"
      install -m 755 "$bin_path" "$HOME/.local/bin/reqly-desktop"
      echo "Installed to $HOME/.local/bin/reqly-desktop (ensure in PATH)"
    fi
    return 0
  fi
  return 1
}

install_binary() {
  local os="$1" arch="$2"
  local tmpdir
  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT
  local url
  url=$(download_url "$os" "$arch" "$VERSION" "")
  echo "Downloading Reqly CLI ($os/$arch)..."
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$tmpdir/reqly" || {
      echo "Trying tar.gz archive fallback..."
      url=$(download_url "$os" "$arch" "$VERSION" ".tar.gz")
      curl -fsSL "$url" -o "$tmpdir/reqly.tar.gz" || { echo "Download failed for $os/$arch"; exit 1; }
      tar -xzf "$tmpdir/reqly.tar.gz" -C "$tmpdir"
      local bin
      bin=$(find "$tmpdir" -name "reqly" -type f | head -n 1)
      if [ -n "$bin" ] && [ "$bin" != "$tmpdir/reqly" ]; then
        cp "$bin" "$tmpdir/reqly"
      fi
    }
  elif command -v wget >/dev/null 2>&1; then
    wget -q "$url" -O "$tmpdir/reqly" || {
      url=$(download_url "$os" "$arch" "$VERSION" ".tar.gz")
      wget -q "$url" -O "$tmpdir/reqly.tar.gz" || { echo "Download failed"; exit 1; }
      tar -xzf "$tmpdir/reqly.tar.gz" -C "$tmpdir"
      local bin
      bin=$(find "$tmpdir" -name "reqly" -type f | head -n 1)
      if [ -n "$bin" ] && [ "$bin" != "$tmpdir/reqly" ]; then
        cp "$bin" "$tmpdir/reqly"
      fi
    }
  else
    echo "Error: curl or wget is required"; exit 1
  fi

  chmod +x "$tmpdir/reqly"
  if [ -w "$INSTALL_DIR" ]; then
    install -m 755 "$tmpdir/reqly" "$INSTALL_DIR/$BIN_NAME"
  else
    mkdir -p "$HOME/.local/bin"
    install -m 755 "$tmpdir/reqly" "$HOME/.local/bin/$BIN_NAME"
    echo "Installed to $HOME/.local/bin (ensure it is in your PATH)"
    INSTALL_DIR="$HOME/.local/bin"
  fi
  echo "Installed $BIN_NAME to $INSTALL_DIR/$BIN_NAME"
  "$INSTALL_DIR/$BIN_NAME" --version || true
}

install_linux() {
  local arch
  arch=$(detect_arch)

  if [ "$WANT_DESKTOP" = true ]; then
    echo "Installing Reqly Desktop for Linux ($arch)..."
    local pm
    pm=$(detect_linux_pm)
    echo "Detected package manager: $pm"
    # On Arch, AppImage built on Ubuntu fails due to WebKit path /usr/lib/x86_64-linux-gnu
    # Prefer native Arch package first
    if [ "$pm" = "pacman" ]; then
      if command -v yay >/dev/null 2>&1 && yay -S --noconfirm reqly-bin 2>/dev/null; then return 0; fi
      if command -v paru >/dev/null 2>&1 && paru -S --noconfirm reqly-bin 2>/dev/null; then return 0; fi
      if install_arch_pkg "$arch"; then return 0; fi
      if install_desktop_binary_linux "$arch"; then return 0; fi
      if install_appimage "$arch"; then
        echo "Note: Ubuntu-built AppImage may fail on Arch with 'WebKitNetworkProcess No such file'. Use pacman package if it does."
        return 0
      fi
    else
      # apt/dnf: AppImage is portable and preferred
      if install_appimage "$arch"; then return 0; fi
      case "$pm" in
        apt)
          if install_deb "$arch"; then return 0; fi
          ;;
        dnf|zypper)
          if install_rpm "$arch"; then return 0; fi
          ;;
      esac
    fi
    # Fallback: desktop binary
    if install_desktop_binary_linux "$arch"; then return 0; fi
    echo "Desktop install failed, falling back to CLI..."
    install_binary "linux" "$arch"
    return
  fi

  # CLI only (default)
  local pm
  pm=$(detect_linux_pm)
  echo "Detected Linux ($arch), package manager: $pm"

  case "$pm" in
    pacman)
      if command -v yay >/dev/null 2>&1 && yay -S --noconfirm reqly-bin 2>/dev/null; then return; fi
      if command -v paru >/dev/null 2>&1 && paru -S --noconfirm reqly-bin 2>/dev/null; then return; fi
      install_binary "linux" "$arch"
      ;;
    apt)
      # Prefer .deb for desktop? For CLI we still try deb but it may contain desktop files
      # Check if user actually wants CLI: skip deb if it would install desktop; fallback to binary
      # We try deb first, if it exists it will install reqly (which includes CLI)
      local tmpdir
      tmpdir=$(mktemp -d)
      trap 'rm -rf "$tmpdir"' EXIT
      local deb_url
      deb_url=$(desktop_url "reqly-linux-${arch}.deb" "$VERSION")
      if command -v curl >/dev/null 2>&1 && curl -fsSL "$deb_url" -o "$tmpdir/reqly.deb" 2>/dev/null; then
        # deb exists – install it (covers both CLI + desktop assets)
        if sudo dpkg -i "$tmpdir/reqly.deb" 2>/dev/null; then
          echo "Installed via .deb"
          return
        else
          sudo apt-get install -f -y && return || true
        fi
      else
        install_binary "linux" "$arch"
      fi
      ;;
    dnf|zypper)
      if install_rpm "$arch"; then return 0; fi
      install_binary "linux" "$arch"
      ;;
    *)
      install_binary "linux" "$arch"
      ;;
  esac
}

install_darwin_app() {
  local arch
  arch=$(detect_arch)
  echo "Detected macOS ($arch: $([ "$arch" = "arm64" ] && echo "Apple Silicon" || echo "Intel"))..."
  echo "Downloading Reqly Desktop for macOS..."
  local tmpdir
  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT

  # Try .app.zip universal, then dmg, then arch-specific
  local urls=()
  urls+=("$(desktop_url "Reqly-macos-universal.app.zip" "$VERSION")")
  urls+=("$(desktop_url "reqly-macos-universal.app.zip" "$VERSION")")
  urls+=("$(desktop_url "Reqly.dmg" "$VERSION")")
  urls+=("$(desktop_url "Reqly-macos-universal.dmg" "$VERSION")")
  urls+=("$(desktop_url "reqly-desktop-macos-universal" "$VERSION")")

  local target=""
  for url in "${urls[@]}"; do
    local fname
    fname=$(basename "$url")
    target="$tmpdir/$fname"
    echo "  Trying $url..."
    if command -v curl >/dev/null 2>&1; then
      if curl -fsSL "$url" -o "$target" 2>/dev/null; then
        echo "Downloaded $fname"
        break
      fi
    elif command -v wget >/dev/null 2>&1; then
      if wget -q "$url" -O "$target" 2>/dev/null; then
        echo "Downloaded $fname"
        break
      fi
    fi
    target=""
  done

  if [ -z "$target" ] || [ ! -f "$target" ]; then
    echo "Desktop download failed, falling back to CLI..."
    install_darwin_cli
    return
  fi

  case "$target" in
    *.zip)
      echo "Installing .app from zip..."
      local app_dest="/Applications/Reqly.app"
      # Remove quarantine and handle existing
      if [ -d "$app_dest" ]; then
        echo "Removing existing $app_dest (requires sudo)..."
        rm -rf "$app_dest" 2>/dev/null || sudo rm -rf "$app_dest" || true
      fi
      unzip -q "$target" -d "$tmpdir" 2>/dev/null || tar -xzf "$target" -C "$tmpdir" 2>/dev/null || true
      local app_src
      app_src=$(find "$tmpdir" -name "*.app" -type d | head -n 1)
      if [ -n "$app_src" ] && [ -d "$app_src" ]; then
        if [ -w "/Applications" ]; then
          mv "$app_src" "$app_dest"
        else
          sudo mv "$app_src" "$app_dest"
        fi
        xattr -dr com.apple.quarantine "$app_dest" 2>/dev/null || true
        echo "Installed Reqly.app to $app_dest"
        echo "Run: open $app_dest"
      else
        echo "Failed to find .app bundle in archive"
        install_darwin_cli
      fi
      ;;
    *.dmg)
      echo "Installing from .dmg..."
      local mnt
      mnt=$(mktemp -d)
      if hdiutil attach "$target" -mountpoint "$mnt" -nobrowse -quiet 2>/dev/null; then
        local app_src
        app_src=$(find "$mnt" -name "*.app" -type d | head -n 1)
        if [ -n "$app_src" ]; then
          local app_dest="/Applications/Reqly.app"
          if [ -d "$app_dest" ]; then rm -rf "$app_dest" 2>/dev/null || sudo rm -rf "$app_dest" || true; fi
          if [ -w "/Applications" ]; then
            cp -R "$app_src" "$app_dest"
          else
            sudo cp -R "$app_src" "$app_dest"
          fi
          xattr -dr com.apple.quarantine "$app_dest" 2>/dev/null || true
          echo "Installed Reqly.app to $app_dest"
        else
          echo "No .app found in dmg"
        fi
        hdiutil detach "$mnt" -quiet 2>/dev/null || true
        rmdir "$mnt" 2>/dev/null || true
      else
        echo "Failed to attach dmg"
        install_darwin_cli
      fi
      ;;
    *)
      # Raw binary fallback
      chmod +x "$target"
      if [ -w "/usr/local/bin" ]; then
        install -m 755 "$target" "/usr/local/bin/reqly-desktop"
        echo "Installed desktop binary to /usr/local/bin/reqly-desktop"
      else
        mkdir -p "$HOME/bin"
        install -m 755 "$target" "$HOME/bin/reqly-desktop"
        echo "Installed to $HOME/bin/reqly-desktop"
      fi
      ;;
  esac
}

install_darwin_cli() {
  local arch
  arch=$(detect_arch)
  echo "Detected macOS ($arch: $([ "$arch" = "arm64" ] && echo "Apple Silicon" || echo "Intel"))..."
  local tmpdir
  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT
  local url
  url=$(download_url "darwin" "$arch" "$VERSION" "")
  echo "Downloading Reqly CLI (darwin/$arch)..."
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$tmpdir/reqly" || {
      url=$(download_url "darwin" "$arch" "$VERSION" ".tar.gz")
      curl -fsSL "$url" -o "$tmpdir/reqly.tar.gz" || { echo "Download failed"; exit 1; }
      tar -xzf "$tmpdir/reqly.tar.gz" -C "$tmpdir"
      local bin
      bin=$(find "$tmpdir" -name "reqly" -type f | head -n 1)
      if [ -n "$bin" ] && [ "$bin" != "$tmpdir/reqly" ]; then
        cp "$bin" "$tmpdir/reqly"
      fi
    }
  elif command -v wget >/dev/null 2>&1; then
    wget -q "$url" -O "$tmpdir/reqly" || {
      url=$(download_url "darwin" "$arch" "$VERSION" ".tar.gz")
      wget -q "$url" -O "$tmpdir/reqly.tar.gz" || { echo "Download failed"; exit 1; }
      tar -xzf "$tmpdir/reqly.tar.gz" -C "$tmpdir"
      local bin
      bin=$(find "$tmpdir" -name "reqly" -type f | head -n 1)
      if [ -n "$bin" ] && [ "$bin" != "$tmpdir/reqly" ]; then
        cp "$bin" "$tmpdir/reqly"
      fi
    }
  fi

  chmod +x "$tmpdir/reqly"
  if [ -w "/usr/local/bin" ]; then
    install -m 755 "$tmpdir/reqly" "/usr/local/bin/$BIN_NAME"
    xattr -d com.apple.quarantine "/usr/local/bin/$BIN_NAME" 2>/dev/null || true
    echo "Installed to /usr/local/bin/$BIN_NAME"
  else
    mkdir -p "$HOME/bin"
    install -m 755 "$tmpdir/reqly" "$HOME/bin/$BIN_NAME"
    xattr -d com.apple.quarantine "$HOME/bin/$BIN_NAME" 2>/dev/null || true
    echo "Installed to $HOME/bin/$BIN_NAME (ensure it is in your PATH)"
  fi
  echo "Run: reqly --version"
}

install_darwin() {
  if [ "$WANT_DESKTOP" = true ]; then
    install_darwin_app
  else
    install_darwin_cli
  fi
}

main() {
  local os
  os=$(detect_os)
  case "$os" in
    linux) install_linux ;;
    darwin) install_darwin ;;
    *) echo "Unsupported OS: $(uname -s). Use Windows install.ps1 on Windows."; exit 1 ;;
  esac
}

# Allow sourcing for tests
if [ "${BASH_SOURCE[0]}" = "$0" ] || [ "$0" = "sh" ] || [ "$0" = "bash" ]; then
  main "$@"
fi
