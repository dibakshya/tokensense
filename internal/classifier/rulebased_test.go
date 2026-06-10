package classifier

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	Input      string `json:"input"`
	TaskType   string `json:"task_type"`
	Complexity string `json:"complexity"`
}

func loadTestCases(t *testing.T) []testCase {
	t.Helper()
	data, err := os.ReadFile("../../testdata/classifier_test_cases.json")
	require.NoError(t, err, "cannot read test cases file")

	var cases []testCase
	require.NoError(t, json.Unmarshal(data, &cases))
	return cases
}

func TestClassifierAccuracy(t *testing.T) {
	cases := loadTestCases(t)
	require.GreaterOrEqual(t, len(cases), 50, "need at least 50 test cases")

	classifier := NewRuleBasedClassifier()

	correct := 0
	for _, tc := range cases {
		result, err := classifier.Classify(tc.Input)
		require.NoError(t, err)

		if result.TaskType == tc.TaskType {
			correct++
		} else {
			t.Logf("MISS: input=%q expected=%s got=%s (confidence=%.2f)",
				tc.Input, tc.TaskType, result.TaskType, result.Confidence)
		}
	}

	accuracy := float64(correct) / float64(len(cases)) * 100
	t.Logf("Classifier accuracy: %.1f%% (%d/%d)", accuracy, correct, len(cases))
	assert.GreaterOrEqual(t, accuracy, 80.0, "classifier accuracy must be >= 80%%")
}

func TestClassifyCodeGeneration(t *testing.T) {
	c := NewRuleBasedClassifier()

	result, err := c.Classify("write a function that parses JSON")
	require.NoError(t, err)
	assert.Equal(t, TaskCodeGeneration, result.TaskType)
	assert.Greater(t, result.Confidence, 0.5)
}

func TestClassifyDebugging(t *testing.T) {
	c := NewRuleBasedClassifier()

	result, err := c.Classify("fix this null pointer exception in my code")
	require.NoError(t, err)
	assert.Equal(t, TaskDebugging, result.TaskType)
}

func TestClassifyTestGeneration(t *testing.T) {
	c := NewRuleBasedClassifier()

	result, err := c.Classify("write unit tests for my UserService")
	require.NoError(t, err)
	assert.Equal(t, TaskTestGeneration, result.TaskType)
}

func TestClassifyDocumentation(t *testing.T) {
	c := NewRuleBasedClassifier()

	result, err := c.Classify("write a docstring for this Go function")
	require.NoError(t, err)
	assert.Equal(t, TaskDocumentation, result.TaskType)
	assert.Greater(t, result.Confidence, 0.8)
}

func TestClassifyReasoning(t *testing.T) {
	c := NewRuleBasedClassifier()

	result, err := c.Classify("analyze the tradeoffs between microservices and monolith")
	require.NoError(t, err)
	assert.Equal(t, TaskReasoning, result.TaskType)
}

func TestClassifyChat(t *testing.T) {
	c := NewRuleBasedClassifier()

	result, err := c.Classify("what is the capital of France")
	require.NoError(t, err)
	assert.Equal(t, TaskChat, result.TaskType)
}

func TestClassifyComplexity(t *testing.T) {
	c := NewRuleBasedClassifier()

	// Low complexity
	result, err := c.Classify("write a simple hello world function")
	require.NoError(t, err)
	assert.Equal(t, ComplexityLow, result.Complexity)

	// High complexity
	result, err = c.Classify("build a distributed system with concurrent processing and security considerations for production")
	require.NoError(t, err)
	assert.Equal(t, ComplexityHigh, result.Complexity)
}

func TestClassifySource(t *testing.T) {
	c := NewRuleBasedClassifier()
	result, err := c.Classify("write tests for auth")
	require.NoError(t, err)
	assert.Equal(t, SourceRuleBased, result.Source)
}

func TestRuleBasedResponseTime(t *testing.T) {
	c := NewRuleBasedClassifier()

	for i := 0; i < 100; i++ {
		_, err := c.Classify("write unit tests for my auth module with edge cases for expired tokens and rate limiting")
		require.NoError(t, err)
	}
	// This test mainly checks that the classifier doesn't hang; actual timing
	// is verified by benchmarks
}

func BenchmarkClassify(b *testing.B) {
	c := NewRuleBasedClassifier()
	for i := 0; i < b.N; i++ {
		c.Classify("write unit tests for my UserService with edge cases for expired tokens")
	}
}
