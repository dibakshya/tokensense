# Model Matrix

The model matrix is a versioned YAML file that contains pricing, quality scores, and task recommendations for AI models across providers.

## Location

- **Bundled**: `data/model-matrix.yaml` (embedded in the binary at build time)
- **Cached**: `~/.tokensense/model-matrix.yaml` (updated daily from GitHub)

## Auto-Update

By default, Tokensense fetches the latest matrix daily from:

```
https://raw.githubusercontent.com/dibakshya/tokensense/main/data/model-matrix.yaml
```

Disable with:
```bash
tokensense config set matrix_auto_update false
```

## Staleness

- **7 days**: informational message in `tokensense status`
- **60 days**: warning in `tokensense status --verbose`

## Schema

```yaml
version: "1"
last_updated: "YYYY-MM-DD"

task_types:
  - id: code_generation
    description: "Writing new code, functions, classes, scripts"

models:
  - id: model-identifier
    provider: anthropic|google|openai|mistral|cohere|groq|xai
    display_name: "Human-Readable Name"
    tier: fast|balanced|premium
    context_window: 128000
    pricing:
      input_per_1m_usd: 0.00
      output_per_1m_usd: 0.00
    task_recommendations:
      code_generation:
        quality: 0-100
        recommended_for_complexity: [low, medium, high]
    last_verified: "YYYY-MM-DD"
```

## Quality Scores

Quality scores are on a 0–100 scale based on community benchmarks:

- **90-100**: Best-in-class for this task type
- **80-89**: Very good, suitable for production use
- **70-79**: Good for routine tasks
- **60-69**: Acceptable for simple tasks
- **< 60**: Not recommended

## Contributing Models

See [CONTRIBUTING.md](../CONTRIBUTING.md) for how to add or update models.
