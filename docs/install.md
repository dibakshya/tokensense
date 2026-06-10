# Installation Guide

## Quick Install

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/dibakshya/tokensense/main/scripts/install.sh | sh
```

### Homebrew (macOS / Linux)

```bash
brew install tokensense/tap/tokensense
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/dibakshya/tokensense/main/scripts/install.ps1 | iex
```

### From Source

```bash
go install github.com/dibakshya/tokensense@latest
```

## First-Time Setup

After installation, run the interactive setup wizard:

```bash
tokensense setup
```

This will:

1. **Privacy mode** — choose between content classification (recommended) or metadata-only
2. **Tool detection** — auto-detect Cursor, Claude Desktop, VS Code, Windsurf
3. **CA certificate** — install a local certificate for HTTPS interception
4. **Report time** — set your daily report generation time (default: 6 PM)
5. **Service registration** — register as an OS service for auto-start

## Platform-Specific Notes

### macOS

- The CA cert is added to the System Keychain (requires sudo/Touch ID)
- The daemon runs as a launchd LaunchAgent
- Service plist: `~/Library/LaunchAgents/dev.tokensense.proxy.plist`

### Linux

- The CA cert is added to `/usr/local/share/ca-certificates/` (Debian/Ubuntu) or `/etc/pki/ca-trust/source/anchors/` (RHEL/Fedora)
- The daemon runs as a systemd user service
- Unit file: `~/.config/systemd/user/tokensense.service`

### Windows

- The CA cert is added to the Windows certificate store (requires Administrator)
- The daemon runs as a Windows Service
- Service name: `TokensenseProxy`

## Verifying Installation

```bash
tokensense version
tokensense status
tokensense tools status
```

## Uninstalling

```bash
tokensense uninstall
```

This removes:
- CA certificate from OS trust store
- Service registration
- Shell profile proxy settings
- All data in `~/.tokensense/`
