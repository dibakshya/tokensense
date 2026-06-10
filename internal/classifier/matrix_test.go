package classifier

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadBundledMatrix(t *testing.T) *ModelMatrix {
	t.Helper()
	data, err := os.ReadFile("../../data/model-matrix.yaml")
	require.NoError(t, err)

	matrix, err := ParseMatrix(data)
	require.NoError(t, err)
	return matrix
}

func TestLoadMatrix(t *testing.T) {
	matrix := loadBundledMatrix(t)

	assert.Equal(t, "1", matrix.Version)
	assert.GreaterOrEqual(t, len(matrix.Models), 20, "matrix should have >= 20 models")
	assert.GreaterOrEqual(t, len(matrix.TaskTypes), 9, "matrix should have >= 9 task types")
}

func TestFindModel(t *testing.T) {
	matrix := loadBundledMatrix(t)

	model := matrix.FindModel("claude-opus-4")
	require.NotNil(t, model)
	assert.Equal(t, "anthropic", model.Provider)
	assert.Equal(t, "Claude Opus 4", model.DisplayName)
	assert.Equal(t, "premium", model.Tier)

	unknown := matrix.FindModel("nonexistent-model")
	assert.Nil(t, unknown)
}

func TestRankModels(t *testing.T) {
	matrix := loadBundledMatrix(t)

	recs := matrix.RankModels(TaskCodeGeneration, ComplexityLow)
	require.NotEmpty(t, recs)

	// First result should be recommended
	assert.True(t, recs[0].IsRecommended, "first ranked model should be recommended")

	// Among recommended, cheapest should be first
	for i := 1; i < len(recs); i++ {
		if recs[i-1].IsRecommended && recs[i].IsRecommended {
			assert.LessOrEqual(t, recs[i-1].CostPerRequest, recs[i].CostPerRequest,
				"recommended models should be sorted by cost ascending")
		}
	}
}

func TestRankModelsAllTaskTypes(t *testing.T) {
	matrix := loadBundledMatrix(t)

	for _, taskType := range AllTaskTypes() {
		for _, complexity := range []string{ComplexityLow, ComplexityMedium, ComplexityHigh} {
			recs := matrix.RankModels(taskType, complexity)
			assert.NotEmpty(t, recs, "should have recommendations for %s/%s", taskType, complexity)
		}
	}
}

func TestCostForRequest(t *testing.T) {
	pricing := Pricing{
		InputPer1MUSD:  15.00,
		OutputPer1MUSD: 75.00,
	}

	cost := CostForRequest(pricing, 1000, 500)
	expected := (1000.0/1_000_000)*15.00 + (500.0/1_000_000)*75.00
	assert.InDelta(t, expected, cost, 0.0001)
}

func TestStaleness(t *testing.T) {
	matrix := loadBundledMatrix(t)

	// Should not be stale (just created today)
	assert.False(t, matrix.IsStale(60))

	// Test with old matrix
	oldMatrix := &ModelMatrix{LastUpdated: "2025-01-01"}
	assert.True(t, oldMatrix.IsStale(60))
	assert.Greater(t, oldMatrix.IsStaleDays(), 365)
}

func TestParseBadMatrix(t *testing.T) {
	_, err := ParseMatrix([]byte("not valid yaml: [[["))
	assert.Error(t, err)
}
