#!/usr/bin/env bash
# Reqly - Multi-distro installer for Linux and macOS (amd64/arm64)
# Usage: curl -fsSL https://raw.githubusercontent.com/Its-Satyajit/reqly/main/install.sh | sh
#        curl -fsSL https://raw.githubusercontent.com/Its-Satyajit/reqly/main/install.sh | bash -s -- --version
set -e

REPO="Its-Satyajit/reqly"
BIN_NAME="reqly"
INSTALL_DIR="/usr/local/bin"
VERSION="${VERSION:-latest}"
WANT_APPIMAGE=false

for arg in "$@"; do
  case "$arg" in
    --app|--appimage) WANT_APPIMAGE=true ;;
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

install_appimage() {
  local arch="$1"
  local tmpdir
  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT
  local url
  url=$(download_url "linux" "$arch" "$VERSION" ".AppImage")
  echo "Downloading Reqly AppImage ($arch)..."
  local target_dir="$HOME/Applications"
  mkdir -p "$target_dir"
  local target_path="$target_dir/Reqly.AppImage"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$target_path" || {
      # Try generic name without arch
      if [ "$VERSION" = "latest" ]; then
        url="https://github.com/${REPO}/releases/latest/download/Reqly.AppImage"
      else
        url="https://github.com/${REPO}/releases/download/${VERSION}/Reqly.AppImage"
      fi
      curl -fsSL "$url" -o "$target_path" || { echo "AppImage download failed. Falling back to CLI binary..."; install_binary "linux" "$arch"; return; }
    }
  elif command -v wget >/dev/null 2>&1; then
    wget -q "$url" -O "$target_path" || { echo "AppImage download failed. Falling back to CLI binary..."; install_binary "linux" "$arch"; return; }
  fi
  chmod +x "$target_path"
  echo "Reqly AppImage installed to: $target_path"
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
  if [ "$WANT_APPIMAGE" = true ]; then
    install_appimage "$arch"
    return
  fi

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
      local tmpdir
      tmpdir=$(mktemp -d)
      trap 'rm -rf "$tmpdir"' EXIT
      local url
      url=$(download_url "linux" "$arch" "$VERSION" ".deb")
      if command -v curl >/dev/null 2>&1 && curl -fsSL "$url" -o "$tmpdir/reqly.deb" 2>/dev/null; then
        sudo dpkg -i "$tmpdir/reqly.deb" || sudo apt-get install -f -y
      else
        install_binary "linux" "$arch"
      fi
      ;;
    *)
      install_binary "linux" "$arch"
      ;;
  esac
}

install_darwin() {
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
