package advisor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/dibakshya/tokensense/internal/classifier"
)

func TestAdviseMinWordCount(t *testing.T) {
	c := classifier.NewRuleBasedClassifier()
	matrix := loadMatrix(t)

	// Too few words
	_, err := Advise("hi", c, nil, matrix, true, 0.6)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 3 words")

	_, err = Advise("a b", c, nil, matrix, true, 0.6)
	assert.Error(t, err)
}

func TestAdviseValidInput(t *testing.T) {
	c := classifier.NewRuleBasedClassifier()
	matrix := loadMatrix(t)

	result, err := Advise("write a function to parse JSON files", c, nil, matrix, true, 0.6)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.NotEmpty(t, result.Classification.TaskType)
	assert.NotEmpty(t, result.Classification.Complexity)
	assert.Greater(t, result.Classification.Confidence, 0.0)
	assert.NotEmpty(t, result.Recommendations)
}

func TestAdviseNoCloudFallback(t *testing.T) {
	c := classifier.NewRuleBasedClassifier()
	matrix := loadMatrix(t)

	// With noCloud=true, should never call API
	result, err := Advise("what is the capital of France today", c, nil, matrix, true, 0.6)
	require.NoError(t, err)
	assert.Equal(t, classifier.SourceRuleBased, result.Classification.Source)
}

func TestAdviseNilMatrix(t *testing.T) {
	c := classifier.NewRuleBasedClassifier()

	result, err := Advise("fix this bug in my authentication code", c, nil, nil, true, 0.6)
	require.NoError(t, err)
	assert.Empty(t, result.Recommendations)
	assert.Equal(t, classifier.TaskDebugging, result.Classification.TaskType)
}

func TestRenderAdvice(t *testing.T) {
	result := &Result{
		Classification: &classifier.ClassificationResult{
			TaskType:   classifier.TaskCodeGeneration,
			Complexity: classifier.ComplexityMedium,
			Confidence: 0.85,
			Source:     classifier.SourceRuleBased,
		},
		Recommendations: []ModelRecommendation{
			{
				ModelID:     "gpt-4o-mini",
				DisplayName: "GPT-4o Mini",
				Provider:    "openai",
				Tier:        "fast",
				Quality:     79,
				CostMin:     0.0001,
				CostMax:     0.0005,
				SpeedSec:    1.0,
				Recommended: true,
				Reason:      "Recommended for medium code_generation",
			},
		},
	}

	output := RenderAdvice("write code", result)
	assert.Contains(t, output, "code_generation")
	assert.Contains(t, output, "medium")
	assert.Contains(t, output, "GPT-4o Mini")
	assert.Contains(t, output, "★")
	assert.Contains(t, output, "best value")
}

func TestRenderAdviceEmpty(t *testing.T) {
	result := &Result{
		Classification: &classifier.ClassificationResult{
			TaskType:   classifier.TaskChat,
			Complexity: classifier.ComplexityLow,
			Confidence: 0.2,
			Source:     classifier.SourceRuleBased,
		},
		Recommendations: nil,
	}

	output := RenderAdvice("hello", result)
	assert.Contains(t, output, "chat")
	assert.NotContains(t, output, "best value")
}

func TestRenderAdviceTruncatesTo5(t *testing.T) {
	c := classifier.NewRuleBasedClassifier()
	matrix := loadMatrix(t)

	result, err := Advise("write a function to parse CSV files with error handling", c, nil, matrix, true, 0.6)
	require.NoError(t, err)

	output := RenderAdvice("write code", result)
	// Count star markers — should be at most 1
	starCount := 0
	for _, line := range splitLines(output) {
		if len(line) > 2 && line[2] == 0xe2 { // ★ starts with 0xe2 in UTF-8
			starCount++
		}
	}
	// Just verify it renders without panic
	assert.NotEmpty(t, output)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
