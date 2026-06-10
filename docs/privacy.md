# Privacy & Security

Tokensense is designed with privacy as a core principle. Here's exactly what it does and doesn't do with your data.

## What Tokensense Stores

**Metadata only** — stored in `~/.tokensense/data.db` (local SQLite):

| Field | Example | Purpose |
|-------|---------|---------|
| Provider | `anthropic` | Cost tracking |
| Model | `claude-sonnet-4-5` | Model usage analysis |
| Task type | `code_generation` | Optimization recommendations |
| Complexity | `medium` | Model matching |
| Token count | `1200 in, 800 out` | Cost calculation |
| Cost | `$0.0156` | Spend tracking |
| Latency | `2100ms` | Performance monitoring |
| Tool source | `cursor` | Per-tool breakdown |

## What Tokensense NEVER Stores

- **Prompts** — your prompts are never written to disk, logged, or stored
- **Responses** — AI responses pass through but are never persisted
- **API keys** — your API keys are forwarded but never stored by Tokensense
- **Personal data** — no user accounts, no identifiers

## How Classification Works

In **Content Mode** (default, recommended):

1. Proxy intercepts the HTTPS request
2. Request body is read into memory
3. Task classifier extracts the task type (keyword matching)
4. Classification result (task type + confidence score) is stored
5. **Request body bytes become eligible for garbage collection immediately**
6. No channel, no write, no log carries the prompt content

In **Metadata Mode**:

1. Proxy tunnels the request transparently (no TLS termination)
2. Only metadata visible without reading content is stored (provider, estimated token count)
3. Task type is set to `null`

## Network Activity

Tokensense makes exactly **two** types of outbound requests:

1. **Forwarding your AI API calls** — this is its core function
2. **Model matrix update** (daily, optional) — fetches updated model pricing from GitHub

   ```
   GET https://raw.githubusercontent.com/dibakshya/tokensense/main/data/model-matrix.yaml
   User-Agent: tokensense/{VERSION} (install_id/{ANONYMOUS_UUID})
   ```

   Disable with: `tokensense config set matrix_auto_update false`

3. **Gemini Flash advisor** (optional, on-demand) — only when you run `tokensense ask` and the rule-based classifier has low confidence

   Sends **only your typed task description** (not intercepted content).

   Disable with: `tokensense config set cloud_fallback false`

## CA Certificate Security

- **Unique per install** — each `tokensense setup` generates a new key pair
- **Key permissions** — CA private key is created with `0600` (owner-read-only)
- **Never shared** — the CA key never leaves `~/.tokensense/ca.key`
- **Removed on uninstall** — `tokensense uninstall` removes the cert from the OS trust store

## Proxy Security

- **Binds to 127.0.0.1 only** — hardcoded, never configurable to 0.0.0.0
- **No remote access** — the proxy only accepts local connections
- **TLS verification on** — outbound connections to AI APIs use full TLS verification

## Team Exports

The `tokensense export` command produces a JSON file containing only the metadata fields listed above. It reads from the `requests` table which by design contains no prompt content.
