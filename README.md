# Tokensense

[![CI](https://github.com/dibakshya/tokensense/actions/workflows/ci.yml/badge.svg)](https://github.com/dibakshya/tokensense/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dibakshya/tokensense)](https://goreportcard.com/report/github.com/dibakshya/tokensense)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Open-source AI token usage optimizer.** A local CLI tool that intercepts your AI API calls, classifies each request by task type, and shows you exactly where cheaper models could save you money — without touching your data.

## Features

- **Browser dashboard** — start/stop tracking, live cost breakdown, and settings — all in your browser, no terminal needed
- **Local HTTPS proxy** — transparently intercepts AI API calls (Anthropic, OpenAI, Google, Mistral, Cohere, Groq, xAI)
- **Task classification** — rule-based engine classifies each request (code generation, debugging, testing, docs, reasoning)
- **Daily reports** — terminal + HTML reports with per-task cost breakdown and specific model swap recommendations
- **Model advisor** — `tokensense ask "..."` recommends the best model for any task description
- **Team reports** — export and merge usage data across team members (privacy-preserving)
- **Developer API** — local JSON API on `localhost:7891` for integrating cost data into agents and tools
- **100% local** — no server, no account, no cloud dependency, no telemetry

---

## Browser Dashboard

The easiest way to use Tokensense. No terminal commands needed after setup.

```bash
tokensense dashboard   # opens http://localhost:7892 in your browser
tokensense             # same — dashboard is the default with no arguments
```

The dashboard lets you:
- **Start / stop** the proxy with one click (big green/red button — state unmistakable at a glance)
- **Live cost breakdown** — refreshed every 6 seconds: spend, potential savings, task breakdown
- **Spot savings** — per task type: "💰 switch to haiku-3-5 → save $1.90/day"
- **Change settings** — privacy mode, daily report time — no config files to edit

The CLI is still fully available for power users and scripting.

---

## Quick Start

### 1. Install

```bash
# macOS / Linux — binary install (recommended)
curl -fsSL https://raw.githubusercontent.com/dibakshya/tokensense/main/scripts/install.sh | sh

# Homebrew
brew install tokensense/tap/tokensense

# Windows (PowerShell as Administrator)
irm https://raw.githubusercontent.com/dibakshya/tokensense/main/scripts/install.ps1 | iex

# Go install
go install github.com/dibakshya/tokensense@latest
# First time only — add Go's bin dir to PATH:
echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.zshrc && source ~/.zshrc
# (replace ~/.zshrc with ~/.bashrc if you use bash)
```

### 2. Setup (one-time, ~2 minutes)

```bash
tokensense setup
```

The interactive wizard will:
1. Ask your privacy preference (content classification or metadata-only)
2. Detect your AI tools (Cursor, Claude Desktop, VS Code, Windsurf)
3. Install a local CA certificate for HTTPS interception
4. Set your daily report time
5. Register and start the background proxy

### 3. Open the dashboard

```bash
tokensense dashboard
# or just: tokensense
```

Or use the terminal:

```bash
tokensense start          # start tracking
tokensense status         # see if proxy is on + today's call count
tokensense report         # view today's cost breakdown
tokensense ask "..."      # get model recommendation for any task
```

---

## How It Works

```
Your AI Tool → Local HTTPS Proxy (127.0.0.1:7890) → AI API
                     ↓
              Task Classifier (in-memory — content never stored)
                     ↓
              SQLite Metadata Store (task type, model, cost only)
                     ↓
              Browser Dashboard + Daily Reports + Model Advisor
```

1. **Proxy** — listens on `127.0.0.1:7890`, intercepts CONNECT tunnels to AI API domains
2. **Classifier** — reads the request body in-memory to determine task type, then immediately discards it
3. **Storage** — writes only metadata (provider, model, token count, cost, task type) to local SQLite
4. **Reports** — daily cost analysis with specific model swap recommendations
5. **Dashboard** — browser UI at `localhost:7892` with live stats and one-click proxy control
6. **Advisor** — classifies any task description and recommends the most cost-effective model

---

## Privacy

- **No prompt content is ever stored.** Classification happens in-memory; content is discarded immediately after.
- **No data leaves your machine.** Everything runs locally — proxy, database, dashboard, and API.
- **No telemetry.** No analytics, no error reporting, no install pings.
- **Metadata-only mode** — skips content reading entirely. Records only provider, model, and token count.
- **CA key is unique** per install, stored with 0600 permissions, never transmitted.
- See [docs/privacy.md](docs/privacy.md) for full details.

---

## Commands

| Command | What it does |
|---------|--------------|
| `tokensense` | Open the browser dashboard (default — no arguments needed) |
| `tokensense dashboard` | Open the browser control panel at localhost:7892 |
| `tokensense setup` | First-time setup wizard — run once after installing |
| `tokensense start` | Start the tracking proxy (also starts automatically at login) |
| `tokensense stop` | Pause tracking temporarily |
| `tokensense status` | Check if the proxy is on + today's call count |
| `tokensense status --json` | Same, as machine-readable JSON (for scripts / agents) |
| `tokensense report` | View today's cost breakdown and savings tips in the terminal |
| `tokensense report --html --open` | Generate and open a visual HTML report in your browser |
| `tokensense report --json` | Full report as machine-readable JSON |
| `tokensense report --date YYYY-MM-DD` | Report for a specific past date |
| `tokensense ask "describe a task"` | Get ranked model recommendations for any task |
| `tokensense api` | Start a local JSON API on localhost:7891 (for developers & agents) |
| `tokensense api --port 8080` | Same, on a custom port |
| `tokensense tools status` | See which AI tools (Cursor, Claude, Copilot…) are being tracked |
| `tokensense export` | Download your usage data as JSON or CSV |
| `tokensense merge file1 file2` | Combine teammates' exports into one team report |
| `tokensense config list` | View all current settings |
| `tokensense config get key` | Get the value of one setting |
| `tokensense config set key value` | Change a setting |
| `tokensense version` | Show version, commit, and build date |
| `tokensense uninstall` | Remove everything — cert, service, data, and shell config |

---

## Configuration

Config file: `~/.tokensense/config.yaml`

Use `tokensense config set` to change settings — no need to edit the file directly.

```bash
tokensense config list                        # view all settings
tokensense config set privacy_mode metadata   # switch to metadata-only mode
tokensense config set report_time 09:00       # change daily report time
```

| Key | Default | Values |
|-----|---------|--------|
| `proxy_port` | `7890` | Any free port |
| `proxy_host` | `127.0.0.1` | Loopback address |
| `privacy_mode` | `content` | `content` or `metadata` |
| `report_time` | `18:00` | `HH:MM` (24-hour) |
| `log_level` | `info` | `debug`, `info`, `warn`, `error` |
| `cloud_fallback` | `true` | `true` / `false` |
| `matrix_auto_update` | `true` | `true` / `false` |
| `confidence_threshold` | `0.6` | 0.0 – 1.0 |

---

## Developer & Agent Integration

Tokensense exposes a local JSON API for integrating cost data into your own tools, agents, or dashboards.

### Start the API

```bash
tokensense api              # starts on http://localhost:7891
tokensense api --port 8080  # custom port
```

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/v1/status` | Proxy on/off + today's request count |
| `GET` | `/v1/report?date=YYYY-MM-DD` | Full cost & savings report as JSON |
| `POST` | `/v1/classify` | Classify a prompt → task type + ranked model recommendations |
| `GET` | `/v1/usage?limit=N&date=YYYY-MM-DD` | Raw usage records (newest first) |
| `GET` | `/v1/docs` | Full API reference with examples |

### Python — Agent Cost Guard

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
console.log(`Could save:  $${report.savings_potential_usd.toFixed(4)}`);
```

---

## Uninstalling

Two steps: run the command, then delete the binary.

```bash
# Step 1 — removes the service, CA certificate, proxy env vars, and ~/.tokensense/
tokensense uninstall

# Step 2 — delete the binary (the command can't remove itself while running)
sudo rm /usr/local/bin/tokensense   # if installed via curl / install.sh
rm ~/go/bin/tokensense              # if installed via go install

# Step 3 — restart your terminal
```

**What `tokensense uninstall` removes:**
- Background proxy service (launchd on macOS / systemd on Linux / Windows Service)
- CA certificate from your OS trust store
- `HTTPS_PROXY`, `HTTP_PROXY`, and `NO_PROXY` lines from `~/.zshrc`, `~/.bashrc`, `~/.profile`
- Everything in `~/.tokensense/` — database, config, logs, CA private key

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:
- Adding or updating models in the model matrix
- Adding classifier test cases
- Development setup

## License

MIT — [Dibakshya Chakraborty](https://github.com/dibakshya)
