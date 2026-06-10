package classifier

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// QA-TC-025: Empty input defaults to chat with low confidence
func TestClassifyEmptyInput(t *testing.T) {
	c := NewRuleBasedClassifier()
	result, err := c.Classify("")
	require.NoError(t, err)
	assert.Equal(t, TaskChat, result.TaskType, "empty input should fall back to chat")
	assert.Less(t, result.Confidence, 0.4, "empty input should have very low confidence")
}

// QA-TC-026: Single-word input defaults to chat
func TestClassifySingleWord(t *testing.T) {
	c := NewRuleBasedClassifier()
	result, err := c.Classify("hello")
	require.NoError(t, err)
	assert.Equal(t, TaskChat, result.TaskType, "single word with no keywords should be chat")
}

// QA-TC-027: Debugging keywords produce correct task type
func TestClassifyDebuggingKeywords(t *testing.T) {
	c := NewRuleBasedClassifier()
	testCases := []string{
		"why does this crash",
		"null pointer exception in my code",
		"stack trace shows segfault",
		"deadlock in my goroutine",
		"returns 500 error",
	}
	for _, tc := range testCases {
		result, err := c.Classify(tc)
		require.NoError(t, err)
		assert.Equal(t, TaskDebugging, result.TaskType, "input %q should classify as debugging", tc)
	}
}

// QA-TC-028: Test generation should win over code generation for "write tests"
func TestClassifyWriteTestsIsTestGenNotCodeGen(t *testing.T) {
	c := NewRuleBasedClassifier()
	result, err := c.Classify("write tests for my authentication module")
	require.NoError(t, err)
	assert.Equal(t, TaskTestGeneration, result.TaskType,
		"'write tests' should classify as test_generation, not code_generation")
}

// QA-TC-029: Documentation should win over code generation for "write a docstring"
func TestClassifyWriteDocstringIsDocNotCode(t *testing.T) {
	c := NewRuleBasedClassifier()
	result, err := c.Classify("write a docstring for this function")
	require.NoError(t, err)
	assert.Equal(t, TaskDocumentation, result.TaskType,
		"'write a docstring' should classify as documentation, not code_generation")
}

// QA-TC-030: Complexity classification - high for security/architecture keywords
func TestClassifyComplexityHighKeywords(t *testing.T) {
	c := NewRuleBasedClassifier()
	result, err := c.Classify("design a distributed microservices architecture with security considerations")
	require.NoError(t, err)
	assert.Equal(t, ComplexityHigh, result.Complexity, "security + distributed + architecture = high complexity")
}

// QA-TC-031: Complexity classification - low for simple/basic keywords
func TestClassifyComplexityLow(t *testing.T) {
	c := NewRuleBasedClassifier()
	result, err := c.Classify("write a simple hello world example")
	require.NoError(t, err)
	assert.Equal(t, ComplexityLow, result.Complexity, "'simple hello world' should be low complexity")
}

// QA-TC-032: Confidence should be >= 0.4 for matched tasks
func TestClassifyConfidenceRange(t *testing.T) {
	c := NewRuleBasedClassifier()
	inputs := []string{
		"fix this bug in my code",
		"review this pull request",
		"implement a new API endpoint",
		"write unit tests for the auth module",
		"summarize this document",
	}
	for _, input := range inputs {
		result, err := c.Classify(input)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Confidence, 0.4,
			"confidence for %q should be at least 0.4", input)
		assert.LessOrEqual(t, result.Confidence, 0.95,
			"confidence for %q should never exceed 0.95", input)
	}
}

// QA-TC-033: Classifier source is always rule_based
func TestClassifySourceIsRuleBased(t *testing.T) {
	c := NewRuleBasedClassifier()
	result, err := c.Classify("implement a sorting algorithm")
	require.NoError(t, err)
	assert.Equal(t, SourceRuleBased, result.Source)
}

// QA-TC-034: Performance — classifier must complete in under 10ms
func TestClassifyPerformance(t *testing.T) {
	c := NewRuleBasedClassifier()
	input := "implement a distributed caching layer with Redis using consistent hashing for load balancing"
	start := time.Now()
	for i := 0; i < 100; i++ {
		_, _ = c.Classify(input)
	}
	elapsed := time.Since(start)
	assert.Less(t, elapsed.Milliseconds(), int64(100),
		"100 classifications should complete in under 100ms total")
}

// QA-TC-035: ProviderFromHost returns unknown for unrecognized hosts
func TestProviderFromHostUnknown(t *testing.T) {
	unknownHosts := []string{
		"example.com",
		"api.slack.com",
		"github.com",
		"localhost",
	}
	for _, h := range unknownHosts {
		p := ProviderFromHost(h)
		assert.Equal(t, ProviderUnknown, p, "host %q should map to unknown provider", h)
	}
}

// QA-TC-036: ProviderFromHost handles all known AI providers
func TestProviderFromHostAllKnown(t *testing.T) {
	cases := map[string]string{
		"api.anthropic.com":             ProviderAnthropic,
		"api.openai.com":                ProviderOpenAI,
		"generativelanguage.googleapis.com": ProviderGoogle,
		"api.mistral.ai":               ProviderMistral,
		"api.cohere.com":               ProviderCohere,
		"api.groq.com":                 ProviderGroq,
		"api.x.ai":                     ProviderXAI,
	}
	for host, expected := range cases {
		got := ProviderFromHost(host)
		assert.Equal(t, expected, got, "host %q should map to provider %q", host, expected)
	}
}

// QA-TC-037: Data extraction keywords
func TestClassifyDataExtraction(t *testing.T) {
	c := NewRuleBasedClassifier()
	cases := []string{
		"extract all emails from this text",
		"parse this json and pull out the user ids",
		"find all phone numbers in this document",
	}
	for _, tc := range cases {
		result, err := c.Classify(tc)
		require.NoError(t, err)
		assert.Equal(t, TaskDataExtraction, result.TaskType, "input %q should be data_extraction", tc)
	}
}

// QA-TC-038: Summarization keywords
func TestClassifySummarization(t *testing.T) {
	c := NewRuleBasedClassifier()
	cases := []string{
		"summarize this article",
		"give me a tldr of this document",
		"condense these meeting notes into key points",
	}
	for _, tc := range cases {
		result, err := c.Classify(tc)
		require.NoError(t, err)
		assert.Equal(t, TaskSummarization, result.TaskType, "input %q should be summarization", tc)
	}
}
