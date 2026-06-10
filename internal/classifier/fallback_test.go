package classifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGeminiFallback(t *testing.T) {
	fb := NewGeminiFallback("test-key")
	require.NotNil(t, fb)
	assert.Equal(t, "test-key", fb.apiKey)
	assert.NotNil(t, fb.httpClient)
	assert.Equal(t, fallbackTimeout, fb.httpClient.Timeout)
}

func TestGeminiFallbackEmptyKey(t *testing.T) {
	fb := NewGeminiFallback("")
	require.NotNil(t, fb)
	assert.Equal(t, "", fb.apiKey)
}

func TestGeminiFallbackClassifyWithoutNetwork(t *testing.T) {
	// Without a valid API key, this should fail gracefully
	fb := NewGeminiFallback("invalid-key")
	_, err := fb.Classify("write a function")
	// We expect an error since we can't actually reach the API in tests
	assert.Error(t, err)
}

func TestGeminiSystemPromptContainsAllTaskTypes(t *testing.T) {
	for _, tt := range AllTaskTypes() {
		assert.Contains(t, geminiSystemPrompt, tt,
			"system prompt should mention task type: %s", tt)
	}
}

func TestGeminiSystemPromptContainsComplexities(t *testing.T) {
	assert.Contains(t, geminiSystemPrompt, "low")
	assert.Contains(t, geminiSystemPrompt, "medium")
	assert.Contains(t, geminiSystemPrompt, "high")
}
