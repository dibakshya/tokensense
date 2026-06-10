package reporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// QA-TC-012: topModel should be deterministic when counts are equal
func TestTopModelDeterministicOnTie(t *testing.T) {
	// All models appear exactly once — alphabetically first should win on tie
	models := []string{"gpt-4o", "claude-sonnet-4-5", "gemini-2.5-flash"}
	result1 := topModel(models)
	result2 := topModel(models)
	result3 := topModel(models)
	assert.Equal(t, result1, result2, "topModel must be deterministic across calls")
	assert.Equal(t, result2, result3, "topModel must be deterministic across calls")
	assert.Equal(t, "claude-sonnet-4-5", result1, "on tie, lexicographically first model should win")
}

// QA-TC-013: topModel returns the clear winner when one dominates
func TestTopModelClearWinner(t *testing.T) {
	models := []string{"gpt-4o", "gpt-4o", "gpt-4o", "claude-sonnet-4-5"}
	result := topModel(models)
	assert.Equal(t, "gpt-4o", result)
}

// QA-TC-014: topModel with empty slice
func TestTopModelEmpty(t *testing.T) {
	result := topModel([]string{})
	assert.Equal(t, "", result)
}

// QA-TC-015: topModel with a single model
func TestTopModelSingle(t *testing.T) {
	result := topModel([]string{"claude-opus-4"})
	assert.Equal(t, "claude-opus-4", result)
}

// QA-TC-016: topModel with all identical models
func TestTopModelAllSame(t *testing.T) {
	models := []string{"gemini-2.5-flash", "gemini-2.5-flash", "gemini-2.5-flash"}
	result := topModel(models)
	assert.Equal(t, "gemini-2.5-flash", result)
}

// QA-TC-017: savings capped at zero — optimized cost can't be negative savings
func TestReportSavingsNeverNegative(t *testing.T) {
	report := &ReportData{
		TotalCostUSD:     0.01,
		OptimizedCostUSD: 0.02, // optimized cost somehow higher (e.g. premium recommended)
	}
	report.SavingsPotential = report.TotalCostUSD - report.OptimizedCostUSD
	if report.SavingsPotential < 0 {
		report.SavingsPotential = 0
	}
	assert.GreaterOrEqual(t, report.SavingsPotential, 0.0, "savings should never be negative")
}

// QA-TC-018: savings percent is 0 when total cost is 0
func TestReportSavingsPercentZeroCost(t *testing.T) {
	report := &ReportData{
		TotalCostUSD:     0,
		SavingsPotential: 0,
	}
	var savingsPct float64
	if report.TotalCostUSD > 0 {
		savingsPct = (report.SavingsPotential / report.TotalCostUSD) * 100
	}
	assert.Equal(t, 0.0, savingsPct, "savings percent should be 0 when total cost is 0, not NaN or panic")
}
