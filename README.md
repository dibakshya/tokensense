# Tokensense

[![CI](https://github.com/dibakshya/tokensense/actions/workflows/ci.yml/badge.svg)](https://github.com/dibakshya/tokensense/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dibakshya/tokensense)](https://goreportcard.com/report/github.com/dibakshya/tokensense)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Open-source AI token usage optimizer.** A local CLI tool that intercepts AI API calls, classifies each request by task type, and shows you where cheaper models could have been used — saving you money without losing quality.

## Features

- **Local HTTPS proxy** — transparently intercepts AI API calls (Anthropic, OpenAI, Google, Mistral, Cohere, Groq, xAI)
- **Task classification** — rule-based engine classifies each request (code generation, debugging, testing, etc.)
- **Daily reports** — terminal + HTML reports with cost breakdown and savings recommendations
- **Model advisor** — `tokensense ask "..."` recommends the optimal model for any task
- **Team reports** — export and merge usage data across team members
- **100% local** — no server, no account, no cloud dependency, no telemetry

## Browser Dashboard

The easiest way to use Tokensense — no terminal commands needed after setup.

```bash
tokensense dashboard   # opens http://localhost:7892 in your browser
tokensense             # same thing — dashboard is the default
```

The dashboard lets you:
- **Start / stop** the proxy with one click (big green/red button)
- **See today's cost breakdown** — live, refreshed every 6 seconds
- **Spot savings** — highlighted recommendations per task type
- **Change settings** — privacy mode, report time — no config files

The CLI is still fully available for power users and scripting.

---

## Quick Start (< 2 minutes)

### Install

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/dibakshya/tokensense/main/scripts/install.sh | sh

# Homebrew
brew install tokensense/tap/tokensense

# Windows (PowerShell)
irm https://raw.githubusercontent.com/dibakshya/tokensense/main/scripts/install.ps1 | iex

# Go install
go install github.com/dibakshya/tokensense@latest
# Then add Go's bin dir to PATH so the command is findable (one-time):
echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.zshrc && source ~/.zshrc
# (replace ~/.zshrc with ~/.bashrc if you use bash)
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

| Command | What it does (plain English) |
|---------|------------------------------|
| `tokensense` | Opens the browser dashboard — default when run with no arguments |
| `tokensense dashboard` | Open the browser control panel (start/stop, reports, settings) |
| `tokensense setup` | Run this once after install — sets everything up with a wizard |
| `tokensense start` | Turn on tracking (also runs automatically at login) |
| `tokensense stop` | Pause tracking temporarily |
| `tokensense status` | See if the proxy is on and how many AI calls happened today |
| `tokensense status --json` | Same, but as machine-readable JSON (for agents/scripts) |
| `tokensense report` | View today's cost breakdown and savings tips |
| `tokensense report --html --open` | Open a visual chart report in your browser |
| `tokensense report --json` | Machine-readable JSON report (for agents/scripts) |
| `tokensense ask "..."` | Describe a task — get the best model for it |
| `tokensense api` | Start a local JSON API on port 7891 (for developers & agents) |
| `tokensense tools status` | See which AI tools (Cursor, Claude, Copilot…) are being tracked |
| `tokensense config set/get/list` | View and change settings |
| `tokensense export` | Download your usage data as JSON or CSV |
| `tokensense merge file1 file2` | Combine usage data from teammates into one report |
| `tokensense uninstall` | Remove everything cleanly |

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

## Developer & Agent Integration

Tokensense exposes a **local JSON API** so you can integrate it into your own AI tools, agents, or dashboards.

### Start the API

```bash
tokensense api              # starts on http://localhost:7891
tokensense api --port 8080  # custom port
```

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/v1/status` | Proxy status + today's request count |
| `GET` | `/v1/report?date=YYYY-MM-DD` | Full cost & savings report as JSON |
| `POST` | `/v1/classify` | Classify a prompt → task type + model recommendations |
| `GET` | `/v1/usage?limit=N&date=YYYY-MM-DD` | Raw usage records |
| `GET` | `/v1/docs` | Full API reference |

### Python Example (Agent Cost Guard)

```python
import requests

# Check today's spend before routing to an expensive model
report = requests.get("http://localhost:7891/v1/report").json()
if report["total_cost_usd"] > 0.50:
    print("Daily budget hit — routing to cheaper model")

# Classify a task before picking a model
result = requests.post("http://localhost:7891/v1/classify",
    json={"prompt": "write unit tests for my auth module"}).json()

print(result["task_type"])        # → "test_generation"
print(result["complexity"])       # → "medium"
best = result["recommendations"][0]
print(best["display_name"])       # → "Gemini 2.5 Flash"
print(best["cost_per_request_usd"])
```

### Shell / jq

```bash
# Get today's savings potential
tokensense report --json | jq '.savings_potential_usd'

# Check if proxy is running from a script
tokensense status --json | jq '.proxy_running'

# Classify a prompt via the API
curl -s -X POST http://localhost:7891/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"prompt":"summarize this document"}' | jq '.task_type'
```

### JavaScript / Node.js

```js
const res = await fetch('http://localhost:7891/v1/report');
const report = await res.json();
console.log(`Spent today: $${report.total_cost_usd.toFixed(4)}`);
console.log(`Could save: $${report.savings_potential_usd.toFixed(4)}`);
```

## Uninstalling

Two steps: run the command, then delete the binary.

```bash
# Step 1 — removes the service, CA certificate, proxy env vars, and all data in ~/.tokensense/
tokensense uninstall

# Step 2 — delete the binary (the command can't remove itself while running)
sudo rm /usr/local/bin/tokensense   # installed via curl / install.sh
rm ~/go/bin/tokensense              # installed via go install

# Step 3 — restart your terminal
```

**What `tokensense uninstall` removes:**
- Background proxy service (launchd on macOS / systemd on Linux / Windows Service)
- CA certificate from your OS trust store
- `HTTPS_PROXY` / `HTTP_PROXY` lines from `~/.zshrc`, `~/.bashrc`, `~/.profile`
- Everything in `~/.tokensense/` — database, config, logs, CA private key

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:
- Adding or updating models in the model matrix
- Adding classifier test cases
- Development setup

## License

MIT
