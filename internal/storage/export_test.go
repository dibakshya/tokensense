package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportContainsNoPromptContent(t *testing.T) {
	db := tempDB(t)

	// Simulate 100 classified requests
	prompts := []string{
		"write a Python function to parse CSV files",
		"debug this segfault in my C++ code",
		"review this authentication middleware",
		"create unit tests for the payment service",
		"summarize the quarterly engineering report",
	}

	for i := 0; i < 100; i++ {
		prompt := prompts[i%len(prompts)]
		_ = prompt // prompt is used only for classification, never stored

		taskType := "code_generation"
		complexity := "medium"
		source := "rule_based"
		confidence := 0.9
		tokens := 1000
		cost := 0.015
		tool := "cursor"

		req := &RequestMetadata{
			Timestamp:            time.Now().UnixMilli(),
			DayDate:              "2026-06-10",
			Provider:             "anthropic",
			Model:                "claude-sonnet-4-5",
			TaskType:             &taskType,
			Complexity:           &complexity,
			TokensIn:             &tokens,
			TokensOut:            &tokens,
			CostUSD:              &cost,
			ContentMode:          1,
			ClassifierSource:     &source,
			ClassifierConfidence: &confidence,
			ToolSource:           &tool,
			Intercepted:          1,
		}
		require.NoError(t, db.Insert(req))
	}

	// Verify exported data contains no prompt content
	results, err := db.QueryByDate("2026-06-10")
	require.NoError(t, err)
	assert.Len(t, results, 100)

	for _, r := range results {
		// Fields that could leak content
		assert.NotContains(t, r.Provider, "write")
		assert.NotContains(t, r.Provider, "debug")
		assert.NotContains(t, r.Provider, "review")
		assert.NotContains(t, r.Model, "parse")
		assert.NotContains(t, r.Model, "CSV")

		if r.TaskType != nil {
			assert.NotContains(t, *r.TaskType, "function")
			assert.NotContains(t, *r.TaskType, "Python")
		}

		if r.Complexity != nil {
			assert.NotContains(t, *r.Complexity, "write")
		}

		if r.ClassifierSource != nil {
			assert.NotContains(t, *r.ClassifierSource, "parse")
		}
	}
}

func TestDataRaceOnConcurrentInserts(t *testing.T) {
	db := tempDB(t)

	// Run with -race flag to detect data races
	done := make(chan struct{}, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- struct{}{} }()
			for j := 0; j < 10; j++ {
				taskType := "code_generation"
				req := &RequestMetadata{
					Timestamp: time.Now().UnixMilli(),
					DayDate:   "2026-06-10",
					Provider:  "test",
					Model:     "test",
					TaskType:  &taskType,
				}
				db.Insert(req)
			}
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	count, err := db.CountByDate("2026-06-10")
	require.NoError(t, err)
	assert.Equal(t, 100, count)
}
