package proxy

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// QA-TC-001: extractPromptText should handle Anthropic content-array format (multimodal)
func TestExtractPromptTextAnthropicContentArray(t *testing.T) {
	data := []byte(`POST /v1/messages HTTP/1.1\r\nContent-Type: application/json\r\n\r\n` +
		`{"model":"claude-opus-4","messages":[{"role":"user","content":[{"type":"text","text":"describe this image"},{"type":"image","source":{"type":"base64","media_type":"image/jpeg","data":"abc123"}}]}]}`)

	// Inject proper \r\n\r\n separator
	raw := "POST /v1/messages HTTP/1.1\r\nContent-Type: application/json\r\n\r\n" +
		`{"model":"claude-opus-4","messages":[{"role":"user","content":[{"type":"text","text":"describe this image"},{"type":"image","source":{"type":"base64","media_type":"image/jpeg","data":"abc123"}}]}]}`
	_ = data
	text := extractPromptText([]byte(raw))
	assert.Contains(t, text, "describe this image", "should extract text from content-array blocks")
}

// QA-TC-002: content-array with multiple text blocks
func TestExtractPromptTextAnthropicMultiTextBlocks(t *testing.T) {
	raw := "POST /v1/messages HTTP/1.1\r\nContent-Type: application/json\r\n\r\n" +
		`{"messages":[{"role":"user","content":[{"type":"text","text":"first block"},{"type":"text","text":"second block"}]},{"role":"assistant","content":[{"type":"text","text":"assistant reply"}]}]}`
	text := extractPromptText([]byte(raw))
	assert.Contains(t, text, "first block")
	assert.Contains(t, text, "second block")
	assert.Contains(t, text, "assistant reply")
}

// QA-TC-003: image-only message (no text blocks) should return empty without panic
func TestExtractPromptTextImageOnly(t *testing.T) {
	raw := "POST /v1/messages HTTP/1.1\r\n\r\n" +
		`{"messages":[{"role":"user","content":[{"type":"image","source":{"data":"abc"}}]}]}`
	text := extractPromptText([]byte(raw))
	assert.Equal(t, "", text, "image-only message should return empty string, not panic")
}

// QA-TC-004: mixed string and array content in same conversation
func TestExtractPromptTextMixedContentFormats(t *testing.T) {
	raw := "POST /v1/messages HTTP/1.1\r\n\r\n" +
		`{"messages":[{"role":"system","content":"you are helpful"},{"role":"user","content":[{"type":"text","text":"array text"}]}]}`
	text := extractPromptText([]byte(raw))
	assert.Contains(t, text, "you are helpful", "string content should still be extracted")
	assert.Contains(t, text, "array text", "array text content should also be extracted")
}

// QA-TC-005: large request body — 256KB cap must not panic
func TestExtractModelLargeRequest(t *testing.T) {
	// Build a request body larger than the old 32KB cap
	bigValue := strings.Repeat("x", 300*1024)
	raw := "POST /v1/messages HTTP/1.1\r\n\r\n" +
		`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"` + bigValue + `"}]}`
	// extractModel only needs the model field near the top — should still work
	model := extractModel([]byte(raw))
	assert.Equal(t, "claude-sonnet-4-5", model)
}

// QA-TC-006: extractPromptText returns empty for nil input without panic
func TestExtractPromptTextNilSafety(t *testing.T) {
	require.NotPanics(t, func() {
		result := extractPromptText(nil)
		assert.Equal(t, "", result)
	})
}

// QA-TC-007: extractTokenCounts prefers input_tokens over prompt_tokens (Anthropic over OpenAI)
func TestExtractTokenCountsAnthropicPreferred(t *testing.T) {
	// Both Anthropic and OpenAI fields present — Anthropic should win if non-zero
	data := []byte("HTTP/1.1 200 OK\r\n\r\n{\"usage\":{\"input_tokens\":300,\"output_tokens\":150,\"prompt_tokens\":99,\"completion_tokens\":49}}")
	tokensIn, tokensOut := extractTokenCounts(data)
	assert.Equal(t, 300, tokensIn, "input_tokens should take priority when non-zero")
	assert.Equal(t, 150, tokensOut, "output_tokens should take priority when non-zero")
}

// QA-TC-008: extractTokenCounts fallback to OpenAI when Anthropic fields are zero
func TestExtractTokenCountsFallbackToOpenAI(t *testing.T) {
	data := []byte("HTTP/1.1 200 OK\r\n\r\n{\"usage\":{\"input_tokens\":0,\"output_tokens\":0,\"prompt_tokens\":77,\"completion_tokens\":33}}")
	tokensIn, tokensOut := extractTokenCounts(data)
	assert.Equal(t, 77, tokensIn, "should fall back to prompt_tokens when input_tokens is 0")
	assert.Equal(t, 33, tokensOut, "should fall back to completion_tokens when output_tokens is 0")
}

// QA-TC-009: extractModel handles Google Gemini-style body (no "model" in body; model in URL)
func TestExtractModelGeminiURL(t *testing.T) {
	// Gemini embeds model in URL, not body JSON — function should gracefully return "unknown"
	raw := "POST /v1beta/models/gemini-2.5-flash:generateContent HTTP/1.1\r\n\r\n{\"contents\":[{\"parts\":[{\"text\":\"hello\"}]}]}"
	model := extractModel([]byte(raw))
	// No "model" key in JSON body — expect "unknown" (graceful degradation)
	assert.Equal(t, "unknown", model)
}

// QA-TC-010: extractPromptText with empty messages array
func TestExtractPromptTextEmptyMessages(t *testing.T) {
	raw := "POST / HTTP/1.1\r\n\r\n{\"model\":\"gpt-4o\",\"messages\":[]}"
	text := extractPromptText([]byte(raw))
	assert.Equal(t, "", text)
}

// QA-TC-011: extractPromptText with tool_use content block (non-text type, no panic)
func TestExtractPromptTextToolUseBlock(t *testing.T) {
	raw := "POST /v1/messages HTTP/1.1\r\n\r\n" +
		`{"messages":[{"role":"assistant","content":[{"type":"tool_use","id":"toolu_01","name":"calculator","input":{"expr":"2+2"}}]},{"role":"user","content":"what is 2+2?"}]}`
	require.NotPanics(t, func() {
		text := extractPromptText([]byte(raw))
		assert.Contains(t, text, "what is 2+2?")
	})
}
