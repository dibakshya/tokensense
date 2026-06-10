# Contributing to Tokensense

Thanks for your interest in contributing!

## Development Setup

```bash
git clone https://github.com/dibakshya/tokensense
cd tokensense
go mod download
go build -o tokensense .
go test -race ./...
```

**Requirements:** Go 1.22+, `CGO_ENABLED=0`

## Adding or Updating Models

The model matrix lives in `data/model-matrix.yaml`. To add or update a model:

1. Add the model entry following the existing YAML schema
2. Include `task_recommendations` for all 9 task types
3. Set `last_verified` to today's date
4. Run tests: `go test ./internal/classifier/...`
5. Submit a PR with the provider's pricing page as reference

### Model Entry Schema

```yaml
- id: model-id-string
  provider: provider-name
  display_name: "Human-Readable Name"
  tier: fast|balanced|premium
  context_window: 128000
  pricing:
    input_per_1m_usd: 0.00
    output_per_1m_usd: 0.00
  task_recommendations:
    code_generation:  {quality: 0-100, recommended_for_complexity: [low, medium, high]}
    # ... all 9 task types
  last_verified: "YYYY-MM-DD"
```

## Adding Classifier Test Cases

Test cases live in `testdata/classifier_test_cases.json`. To add cases:

1. Add entries with `input`, `task_type`, and `complexity` fields
2. Run `go test ./internal/classifier/... -run TestClassifierAccuracy -v`
3. Accuracy must remain ≥ 80%

## Code Standards

- `gofmt` and `goimports` required
- No `panic()` in production code paths
- `CGO_ENABLED=0` for all builds
- All prompt/content handling code must have: `// content is in-process only; never persisted or logged`
- Error messages must be actionable

## Pull Request Checklist

- [ ] `go test -race ./...` passes
- [ ] `go vet ./...` clean
- [ ] `TestClassifierAccuracy` ≥ 80%
- [ ] No new `CGO` dependencies
- [ ] No prompt content in logs or storage
