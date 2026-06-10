package classifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderFromHost(t *testing.T) {
	tests := map[string]string{
		"api.anthropic.com":                              ProviderAnthropic,
		"generativelanguage.googleapis.com":               ProviderGoogle,
		"api.openai.com":                                  ProviderOpenAI,
		"api.mistral.ai":                                  ProviderMistral,
		"api.cohere.ai":                                   ProviderCohere,
		"api.groq.com":                                    ProviderGroq,
		"api.x.ai":                                        ProviderXAI,
		"random.example.com":                               ProviderUnknown,
		"":                                                ProviderUnknown,
		"localhost":                                        ProviderUnknown,
	}

	for host, expected := range tests {
		assert.Equal(t, expected, ProviderFromHost(host), "ProviderFromHost(%q)", host)
	}
}

func TestAllTaskTypes(t *testing.T) {
	types := AllTaskTypes()
	assert.Len(t, types, 9)

	expected := map[string]bool{
		TaskCodeGeneration: true,
		TaskCodeReview:     true,
		TaskDebugging:      true,
		TaskTestGeneration: true,
		TaskDocumentation:  true,
		TaskReasoning:      true,
		TaskDataExtraction: true,
		TaskSummarization:  true,
		TaskChat:           true,
	}

	for _, tt := range types {
		assert.True(t, expected[tt], "unexpected task type: %s", tt)
	}
}

func TestConstants(t *testing.T) {
	assert.Equal(t, 0.6, ConfidenceThreshold)
	assert.Equal(t, "rule_based", SourceRuleBased)
	assert.Equal(t, "api_fallback", SourceAPIFallback)
	assert.Equal(t, "api_fallback_timeout", SourceAPIFallbackTimeout)
	assert.Equal(t, "metadata_only", SourceMetadataOnly)
}

func TestClassificationResultFields(t *testing.T) {
	r := ClassificationResult{
		TaskType:   TaskCodeGeneration,
		Complexity: ComplexityHigh,
		Confidence: 0.95,
		Source:     SourceRuleBased,
	}
	assert.Equal(t, "code_generation", r.TaskType)
	assert.Equal(t, "high", r.Complexity)
	assert.Equal(t, 0.95, r.Confidence)
	assert.Equal(t, "rule_based", r.Source)
}

func TestContainsHelper(t *testing.T) {
	assert.True(t, contains("api.anthropic.com", "anthropic"))
	assert.False(t, contains("api.openai.com", "anthropic"))
	assert.False(t, contains("", "anthropic"))
	assert.True(t, contains("anthropic", "anthropic"))
	assert.False(t, contains("an", "anthropic"))
}
