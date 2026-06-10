#!/bin/sh
set -e
REPO="dibakshya/tokensense"
VERSION=$(curl -sfL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed 's/.*"v\([^"]*\)".*/\1/')
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
[ "$ARCH" = "x86_64" ] && ARCH="amd64"
[ "$ARCH" = "aarch64" ] && ARCH="arm64"
URL="https://github.com/${REPO}/releases/download/v${VERSION}/tokensense_${VERSION}_${OS}_${ARCH}.tar.gz"
echo "Downloading Tokensense $VERSION for $OS/$ARCH..."
curl -sfL "$URL" | tar -xz tokensense
chmod +x tokensense
sudo mv tokensense /usr/local/bin/tokensense
echo "Tokensense $VERSION installed. Run: tokensense setup"
