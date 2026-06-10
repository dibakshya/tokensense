package advisor

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/dibakshya/tokensense/internal/classifier"
)

func loadMatrix(t *testing.T) *classifier.ModelMatrix {
	t.Helper()
	data, err := os.ReadFile("../../data/model-matrix.yaml")
	require.NoError(t, err)
	matrix, err := classifier.ParseMatrix(data)
	require.NoError(t, err)
	return matrix
}

func TestRankModelsCodeGenLow(t *testing.T) {
	matrix := loadMatrix(t)
	recs := RankModels(matrix, classifier.TaskCodeGeneration, classifier.ComplexityLow)

	require.NotEmpty(t, recs)
	// First should be recommended and cheap
	assert.True(t, recs[0].Recommended)
	assert.Greater(t, recs[0].Quality, 0)
}

func TestRankModelsReasoningHigh(t *testing.T) {
	matrix := loadMatrix(t)
	recs := RankModels(matrix, classifier.TaskReasoning, classifier.ComplexityHigh)

	require.NotEmpty(t, recs)
	assert.True(t, recs[0].Recommended)
}

func TestRankModelsAllCombinations(t *testing.T) {
	matrix := loadMatrix(t)

	for _, taskType := range classifier.AllTaskTypes() {
		for _, complexity := range []string{classifier.ComplexityLow, classifier.ComplexityMedium, classifier.ComplexityHigh} {
			recs := RankModels(matrix, taskType, complexity)
			assert.NotEmpty(t, recs, "should have recs for %s/%s", taskType, complexity)

			for _, rec := range recs {
				assert.NotEmpty(t, rec.ModelID)
				assert.NotEmpty(t, rec.DisplayName)
				assert.Greater(t, rec.Quality, 0)
				assert.GreaterOrEqual(t, rec.CostMin, 0.0)
				assert.GreaterOrEqual(t, rec.CostMax, rec.CostMin)
			}
		}
	}
}

func TestRankModelsNilMatrix(t *testing.T) {
	recs := RankModels(nil, classifier.TaskCodeGeneration, classifier.ComplexityLow)
	assert.Nil(t, recs)
}

func TestEstimateSpeed(t *testing.T) {
	assert.Equal(t, 1.0, estimateSpeed("fast"))
	assert.Equal(t, 2.5, estimateSpeed("balanced"))
	assert.Equal(t, 5.0, estimateSpeed("premium"))
	assert.Equal(t, 3.0, estimateSpeed("unknown"))
}
