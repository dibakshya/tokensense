# Tokensense

[![CI](https://github.fkinternal.com/dibakshya-c/tokensense/actions/workflows/ci.yml/badge.svg)](https://github.fkinternal.com/dibakshya-c/tokensense/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.fkinternal.com/dibakshya-c/tokensense)](https://goreportcard.com/report/github.fkinternal.com/dibakshya-c/tokensense)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Open-source AI token usage optimizer.** A local CLI tool that intercepts AI API calls, classifies each request by task type, and shows you where cheaper models could have been used — saving you money without losing quality.

## Features

- **Local HTTPS proxy** — transparently intercepts AI API calls (Anthropic, OpenAI, Google, Mistral, Cohere, Groq, xAI)
- **Task classification** — rule-based engine classifies each request (code generation, debugging, testing, etc.)
- **Daily reports** — terminal + HTML reports with cost breakdown and savings recommendations
- **Model advisor** — `tokensense ask "..."` recommends the optimal model for any task
- **Team reports** — export and merge usage data across team members
- **100% local** — no server, no account, no cloud dependency, no telemetry

## Quick Start (< 2 minutes)

### Install

```bash
# macOS / Linux
curl -fsSL https://github.fkinternal.com/dibakshya-c/tokensense/raw/main/scripts/install.sh | sh

# Homebrew
brew install tokensense/tap/tokensense

# Windows (PowerShell)
irm https://github.fkinternal.com/dibakshya-c/tokensense/raw/main/scripts/install.ps1 | iex

# From source
go install github.fkinternal.com/dibakshya-c/tokensense@latest
```

### Setup

```bash
tokensense setup
```

The interactive wizard will:
1. Ask your privacy preference (content classification or metadata-only)
2. Detect your AI tools (Cursor, Claude Desktop, VS Code, Windsurf)
3. Install a local CA certificate for HTTPS interception
4. Set your daily report time
5. Register and start the background proxy

### Use

```bash
# Work normally — AI calls route through the proxy automatically

# Check status
tokensense status

# View today's report
tokensense report

# Get model recommendations
tokensense ask "write unit tests for my auth module"

# View detected tools
tokensense tools status
```

## How It Works

```
Your AI Tool → Local HTTPS Proxy (127.0.0.1:7890) → AI API
                     ↓
              Task Classifier (in-memory, no persistence)
                     ↓
              SQLite Metadata Store (task type, model, cost — never prompt content)
                     ↓
              Daily Report + Model Advisor
```

1. **Proxy** — listens on `127.0.0.1:7890`, intercepts CONNECT requests to AI APIs
2. **Classifier** — reads request body in-memory to determine task type (code generation, debugging, etc.), then immediately discards the content
3. **Storage** — writes only metadata (provider, model, token count, cost, task type) to local SQLite
4. **Reports** — generates daily cost analysis with specific model swap recommendations
5. **Advisor** — classifies any task description and recommends the most cost-effective model

## Privacy

- **No prompt content is ever stored.** Classification happens in-memory; content is immediately discarded.
- **No data leaves your machine.** Everything runs locally.
- **No telemetry.** No analytics. No error reporting.
- **Metadata-only mode** available for maximum privacy (sees only provider, model, token count).
- **CA key is unique** per install with 0600 permissions.
- See [docs/privacy.md](docs/privacy.md) for details.

## Commands

| Command | Description |
|---------|-------------|
| `tokensense setup` | First-time setup wizard |
| `tokensense start` | Start the proxy daemon |
| `tokensense stop` | Stop the proxy daemon |
| `tokensense status` | Show proxy status and today's stats |
| `tokensense report` | View daily report (terminal) |
| `tokensense report --html --open` | Generate and open HTML report |
| `tokensense ask "..."` | Get model recommendations for a task |
| `tokensense tools status` | Show detected AI tools |
| `tokensense config set/get/list` | Manage configuration |
| `tokensense export` | Export usage data as JSON |
| `tokensense merge file1.json file2.json` | Merge exports into team report |
| `tokensense uninstall` | Remove everything |

## Configuration

Config stored in `~/.tokensense/config.yaml`:

```yaml
proxy_port: 7890
proxy_host: "127.0.0.1"
privacy_mode: "content"     # "content" or "metadata"
report_time: "18:00"
log_level: "info"
cloud_fallback: true
matrix_auto_update: true
confidence_threshold: 0.6
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:
- Adding or updating models in the model matrix
- Adding classifier test cases
- Development setup

## License

MIT
