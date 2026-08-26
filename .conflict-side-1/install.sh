#!/usr/bin/env bash
# Reqly - Multi-distro installer for Linux and macOS (amd64/arm64)
# Usage: curl -fsSL https://raw.githubusercontent.com/Its-Satyajit/reqly/main/install.sh | sh
#        curl -fsSL https://raw.githubusercontent.com/Its-Satyajit/reqly/main/install.sh | bash -s -- --version
set -e

REPO="Its-Satyajit/reqly"
BIN_NAME="reqly"
INSTALL_DIR="/usr/local/bin"
VERSION="${VERSION:-latest}"

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

install_binary() {
  local os="$1" arch="$2"
  local tmpdir
  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT
  local url
  url=$(download_url "$os" "$arch" "$VERSION" ".tar.gz")
  echo "Downloading $url..."
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$tmpdir/reqly.tar.gz" || {
      echo "Trying binary fallback..."
      url=$(download_url "$os" "$arch" "$VERSION" "")
      curl -fsSL "$url" -o "$tmpdir/reqly" || { echo "Download failed"; exit 1; }
      chmod +x "$tmpdir/reqly"
      install -m 755 "$tmpdir/reqly" "$INSTALL_DIR/$BIN_NAME" 2>/dev/null || install -m 755 "$tmpdir/reqly" "$HOME/.local/bin/$BIN_NAME"
      echo "Installed $BIN_NAME to $INSTALL_DIR"
      return
    }
  elif command -v wget >/dev/null 2>&1; then
    wget -q "$url" -O "$tmpdir/reqly.tar.gz" || { echo "Download failed"; exit 1; }
  else
    echo "Need curl or wget"; exit 1
  fi
  tar -xzf "$tmpdir/reqly.tar.gz" -C "$tmpdir"
  # find binary
  local bin
  bin=$(find "$tmpdir" -name "reqly" -type f | head -n 1)
  if [ -z "$bin" ]; then bin="$tmpdir/reqly"; fi
  chmod +x "$bin"
  if [ -w "$INSTALL_DIR" ]; then
    install -m 755 "$bin" "$INSTALL_DIR/$BIN_NAME"
  else
    mkdir -p "$HOME/.local/bin"
    install -m 755 "$bin" "$HOME/.local/bin/$BIN_NAME"
    echo "Installed to $HOME/.local/bin (add to PATH)"
    INSTALL_DIR="$HOME/.local/bin"
  fi
  echo "Installed $BIN_NAME to $INSTALL_DIR/$BIN_NAME"
  "$INSTALL_DIR/$BIN_NAME" --version || true
}

install_linux() {
  local pm
  pm=$(detect_linux_pm)
  local arch
  arch=$(detect_arch)
  echo "Detected Linux ($arch), package manager: $pm"
  # Try native package first for amd64 (most distro support)
  case "$pm" in
    pacman)
      echo "Arch Linux detected — trying AUR (reqly-bin)..."
      if command -v yay >/dev/null 2>&1; then yay -S --noconfirm reqly-bin || install_binary "linux" "$arch"; return; fi
      if command -v paru >/dev/null 2>&1; then paru -S --noconfirm reqly-bin || install_binary "linux" "$arch"; return; fi
      install_binary "linux" "$arch"
      ;;
    apt)
      echo "Debian/Ubuntu detected..."
      # try .deb if available, else tar.gz
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
    dnf)
      echo "Fedora/RHEL detected..."
      install_binary "linux" "$arch"
      ;;
    zypper)
      echo "openSUSE detected..."
      install_binary "linux" "$arch"
      ;;
    *)
      install_binary "linux" "$arch"
      ;;
  esac
}

install_darwin() {
  local arch
  arch=$(detect_arch)
  echo "Detected macOS ($arch)..."
  # For macOS, use tar.gz (dmg requires hdiutil + quarantine handling)
  local tmpdir
  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT
  local url
  # Map arch: x86_64 -> amd64, arm64 stays
  url=$(download_url "darwin" "$arch" "$VERSION" ".tar.gz")
  echo "Downloading $url..."
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$tmpdir/reqly.tar.gz" || { echo "Download failed, trying binary..."; url=$(download_url "darwin" "$arch" "$VERSION" ""); curl -fsSL "$url" -o "$tmpdir/reqly"; }
  else
    wget -q "$url" -O "$tmpdir/reqly.tar.gz"
  fi
  if [ -f "$tmpdir/reqly.tar.gz" ]; then
    tar -xzf "$tmpdir/reqly.tar.gz" -C "$tmpdir"
    local bin
    bin=$(find "$tmpdir" -name "reqly" -type f | head -n 1)
    if [ -z "$bin" ]; then bin="$tmpdir/reqly"; fi
    chmod +x "$bin"
    # Try /usr/local/bin, else ~/bin
    if [ -w "/usr/local/bin" ]; then
      install -m 755 "$bin" "/usr/local/bin/$BIN_NAME"
      echo "Installed to /usr/local/bin/$BIN_NAME"
      # Remove quarantine for ad-hoc signed binary
      xattr -d com.apple.quarantine "/usr/local/bin/$BIN_NAME" 2>/dev/null || true
    else
      mkdir -p "$HOME/bin"
      install -m 755 "$bin" "$HOME/bin/$BIN_NAME"
      echo "Installed to $HOME/bin/$BIN_NAME (add to PATH)"
      xattr -d com.apple.quarantine "$HOME/bin/$BIN_NAME" 2>/dev/null || true
    fi
  elif [ -f "$tmpdir/reqly" ]; then
    chmod +x "$tmpdir/reqly"
    if [ -w "/usr/local/bin" ]; then
      install -m 755 "$tmpdir/reqly" "/usr/local/bin/$BIN_NAME"
      xattr -d com.apple.quarantine "/usr/local/bin/$BIN_NAME" 2>/dev/null || true
    else
      mkdir -p "$HOME/bin"
      install -m 755 "$tmpdir/reqly" "$HOME/bin/$BIN_NAME"
      xattr -d com.apple.quarantine "$HOME/bin/$BIN_NAME" 2>/dev/null || true
    fi
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
