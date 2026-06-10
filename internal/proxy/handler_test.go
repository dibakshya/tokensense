package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractModelOpenAI(t *testing.T) {
	data := []byte("POST /v1/chat/completions HTTP/1.1\r\nContent-Type: application/json\r\n\r\n{\"model\":\"gpt-4o\",\"messages\":[{\"role\":\"user\",\"content\":\"hello\"}]}")
	model := extractModel(data)
	assert.Equal(t, "gpt-4o", model)
}

func TestExtractModelAnthropic(t *testing.T) {
	data := []byte("POST /v1/messages HTTP/1.1\r\nContent-Type: application/json\r\n\r\n{\"model\":\"claude-sonnet-4-5\",\"messages\":[{\"role\":\"user\",\"content\":\"hi\"}]}")
	model := extractModel(data)
	assert.Equal(t, "claude-sonnet-4-5", model)
}

func TestExtractModelMissingHeader(t *testing.T) {
	data := []byte("no headers here")
	model := extractModel(data)
	assert.Equal(t, "unknown", model)
}

func TestExtractModelNoModelField(t *testing.T) {
	data := []byte("POST /v1 HTTP/1.1\r\n\r\n{\"messages\":[]}")
	model := extractModel(data)
	assert.Equal(t, "unknown", model)
}

func TestExtractModelInvalidJSON(t *testing.T) {
	data := []byte("POST /v1 HTTP/1.1\r\n\r\n{invalid json}")
	model := extractModel(data)
	assert.Equal(t, "unknown", model)
}

func TestExtractModelEmptyData(t *testing.T) {
	model := extractModel([]byte{})
	assert.Equal(t, "unknown", model)
}

func TestExtractPromptTextOpenAI(t *testing.T) {
	data := []byte("POST /v1/chat HTTP/1.1\r\n\r\n{\"messages\":[{\"role\":\"user\",\"content\":\"write a function\"},{\"role\":\"system\",\"content\":\"you are helpful\"}]}")
	text := extractPromptText(data)
	assert.Contains(t, text, "write a function")
	assert.Contains(t, text, "you are helpful")
}

func TestExtractPromptTextGemini(t *testing.T) {
	data := []byte("POST /v1 HTTP/1.1\r\n\r\n{\"contents\":[{\"parts\":[{\"text\":\"debug this code\"}]}]}")
	text := extractPromptText(data)
	assert.Contains(t, text, "debug this code")
}

func TestExtractPromptTextEmptyData(t *testing.T) {
	assert.Equal(t, "", extractPromptText([]byte{}))
	assert.Equal(t, "", extractPromptText(nil))
}

func TestExtractPromptTextNoMessages(t *testing.T) {
	data := []byte("POST / HTTP/1.1\r\n\r\n{\"model\":\"gpt-4o\"}")
	assert.Equal(t, "", extractPromptText(data))
}

func TestExtractTokenCountsOpenAI(t *testing.T) {
	data := []byte("HTTP/1.1 200 OK\r\n\r\n{\"usage\":{\"prompt_tokens\":100,\"completion_tokens\":50}}")
	tokensIn, tokensOut := extractTokenCounts(data)
	assert.Equal(t, 100, tokensIn)
	assert.Equal(t, 50, tokensOut)
}

func TestExtractTokenCountsAnthropic(t *testing.T) {
	data := []byte("HTTP/1.1 200 OK\r\n\r\n{\"usage\":{\"input_tokens\":200,\"output_tokens\":100}}")
	tokensIn, tokensOut := extractTokenCounts(data)
	assert.Equal(t, 200, tokensIn)
	assert.Equal(t, 100, tokensOut)
}

func TestExtractTokenCountsEmpty(t *testing.T) {
	tokensIn, tokensOut := extractTokenCounts([]byte{})
	assert.Equal(t, 0, tokensIn)
	assert.Equal(t, 0, tokensOut)
}

func TestExtractTokenCountsInvalidJSON(t *testing.T) {
	data := []byte("HTTP/1.1 200 OK\r\n\r\n{bad json}")
	tokensIn, tokensOut := extractTokenCounts(data)
	assert.Equal(t, 0, tokensIn)
	assert.Equal(t, 0, tokensOut)
}

func TestExtractTokenCountsNoUsage(t *testing.T) {
	data := []byte("HTTP/1.1 200 OK\r\n\r\n{\"choices\":[]}")
	tokensIn, tokensOut := extractTokenCounts(data)
	assert.Equal(t, 0, tokensIn)
	assert.Equal(t, 0, tokensOut)
}

func TestCertCacheConcurrency(t *testing.T) {
	// Reset cache
	hostCertCache = &certCache{certs: make(map[string]*cachedCert)}

	// This test mainly ensures no data race under -race flag
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			hostCertCache.mu.RLock()
			_ = hostCertCache.certs["test.com"]
			hostCertCache.mu.RUnlock()
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestExtractPromptTextMultipleMessages(t *testing.T) {
	data := []byte("POST / HTTP/1.1\r\n\r\n{\"messages\":[{\"role\":\"system\",\"content\":\"you are a coder\"},{\"role\":\"user\",\"content\":\"write code\"},{\"role\":\"assistant\",\"content\":\"sure\"},{\"role\":\"user\",\"content\":\"add tests\"}]}")
	text := extractPromptText(data)
	assert.Contains(t, text, "you are a coder")
	assert.Contains(t, text, "write code")
	assert.Contains(t, text, "add tests")
}

func TestExtractPromptTextGeminiMultipleParts(t *testing.T) {
	data := []byte("POST / HTTP/1.1\r\n\r\n{\"contents\":[{\"parts\":[{\"text\":\"part1\"},{\"text\":\"part2\"}]},{\"parts\":[{\"text\":\"part3\"}]}]}")
	text := extractPromptText(data)
	assert.Contains(t, text, "part1")
	assert.Contains(t, text, "part2")
	assert.Contains(t, text, "part3")
}
