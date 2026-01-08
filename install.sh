#!/bin/bash
set -e

REPO="9roads/ccyolo"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64|amd64)
    ARCH="x86_64"
    ;;
  aarch64|arm64)
    ARCH="arm64"
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

case "$OS" in
  darwin|linux)
    ;;
  *)
    echo "Unsupported OS: $OS"
    echo "For Windows, use: scoop install ccyolo"
    exit 1
    ;;
esac

BINARY="ccyolo-${OS}-${ARCH}"

echo "Detecting latest version..."
LATEST=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "Failed to detect latest version"
  exit 1
fi

echo "Downloading ccyolo ${LATEST} for ${OS}/${ARCH}..."

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST}/${BINARY}"

TMP=$(mktemp)
curl -fsSL "$DOWNLOAD_URL" -o "$TMP"
chmod +x "$TMP"

echo "Installing to ${INSTALL_DIR}/ccyolo..."

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP" "${INSTALL_DIR}/ccyolo"
else
  echo "Need sudo to install to ${INSTALL_DIR}"
  sudo mv "$TMP" "${INSTALL_DIR}/ccyolo"
fi

echo ""
echo "ccyolo ${LATEST} installed successfully!"
echo ""
echo "Next steps:"
echo "  1. ccyolo install   # Register hook with Claude Code"
echo "  2. ccyolo setup     # Configure API key"
echo "  3. Restart Claude Code"
