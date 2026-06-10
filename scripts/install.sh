#!/bin/sh
set -e

REPO="dibakshya/tokensense"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
[ "$ARCH" = "x86_64" ]  && ARCH="amd64"
[ "$ARCH" = "aarch64" ] && ARCH="arm64"

# ── Check for a published release ────────────────────────────────────────────
VERSION=$(curl -sfL "https://api.github.com/repos/$REPO/releases/latest" \
  | grep '"tag_name"' \
  | sed 's/.*"v\([^"]*\)".*/\1/')

if [ -z "$VERSION" ]; then
  echo ""
  echo "  No binary release found for $REPO."
  echo ""
  echo "  ── Option 1: install with Go (recommended) ──────────────────"
  echo "     go install github.com/$REPO@latest"
  echo "     echo 'export PATH=\"\$PATH:\$HOME/go/bin\"' >> ~/.zshrc && source ~/.zshrc"
  echo "     tokensense setup"
  echo ""
  echo "  ── Option 2: build from source ──────────────────────────────"
  echo "     git clone https://github.com/$REPO.git"
  echo "     cd tokensense"
  echo "     go build -o tokensense ."
  echo "     sudo mv tokensense /usr/local/bin/"
  echo ""
  exit 1
fi

# ── Download and install binary ───────────────────────────────────────────────
URL="https://github.com/${REPO}/releases/download/v${VERSION}/tokensense_${VERSION}_${OS}_${ARCH}.tar.gz"
echo ""
echo "  Downloading Tokensense v${VERSION} for ${OS}/${ARCH}..."
TMPDIR=$(mktemp -d)
curl -sfL "$URL" -o "$TMPDIR/tokensense.tar.gz"
tar -xzf "$TMPDIR/tokensense.tar.gz" -C "$TMPDIR"
BINARY=$(find "$TMPDIR" -name "tokensense" -not -name "*.tar.gz" -type f | head -1)
if [ -z "$BINARY" ]; then
  echo "  ❌ Could not find tokensense binary in downloaded archive."
  rm -rf "$TMPDIR"
  exit 1
fi
chmod +x "$BINARY"
sudo mv "$BINARY" /usr/local/bin/tokensense
rm -rf "$TMPDIR"
echo "  ✅ Tokensense v${VERSION} installed."
echo ""
echo "  Run:  tokensense setup"
echo ""
