package reporter

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.fkinternal.com/dibakshya-c/tokensense/internal/classifier"
	"github.fkinternal.com/dibakshya-c/tokensense/internal/storage"
)

func testDB(t *testing.T) *storage.DB {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := storage.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func loadMatrix(t *testing.T) *classifier.ModelMatrix {
	t.Helper()
	data, err := os.ReadFile("../../data/model-matrix.yaml")
	require.NoError(t, err)
	matrix, err := classifier.ParseMatrix(data)
	require.NoError(t, err)
	return matrix
}

func seedRequests(t *testing.T, db *storage.DB, count int, date string) {
	t.Helper()
	taskTypes := []string{"code_generation", "debugging", "test_generation", "reasoning"}
	models := []string{"claude-sonnet-4-5", "claude-opus-4", "gpt-4o", "gemini-2.5-flash"}
	providers := []string{"anthropic", "anthropic", "openai", "google"}

	for i := 0; i < count; i++ {
		idx := i % len(taskTypes)
		taskType := taskTypes[idx]
		complexity := "medium"
		tokensIn := 1000 + i*100
		tokensOut := 500 + i*50
		cost := float64(tokensIn)/1_000_000*3.0 + float64(tokensOut)/1_000_000*15.0
		latency := 1500 + i*100
		source := "rule_based"
		confidence := 0.85
		tool := "cursor"

		req := &storage.RequestMetadata{
			Timestamp:            time.Now().UnixMilli(),
			DayDate:              date,
			Provider:             providers[idx],
			Model:                models[idx],
			TaskType:             &taskType,
			Complexity:           &complexity,
			TokensIn:             &tokensIn,
			TokensOut:            &tokensOut,
			CostUSD:              &cost,
			LatencyMs:            &latency,
			ContentMode:          1,
			ClassifierSource:     &source,
			ClassifierConfidence: &confidence,
			ToolSource:           &tool,
			Intercepted:          1,
		}
		require.NoError(t, db.Insert(req))
	}
}

func TestGenerateReport(t *testing.T) {
	db := testDB(t)
	matrix := loadMatrix(t)

	seedRequests(t, db, 10, "2026-06-10")

	report, err := GenerateReport(db, matrix, "2026-06-10")
	require.NoError(t, err)
	assert.Equal(t, 10, report.TotalRequests)
	assert.Greater(t, report.TotalCostUSD, 0.0)
	assert.NotEmpty(t, report.TaskBreakdown)
	assert.NotEmpty(t, report.ToolBreakdown)
	assert.NotEmpty(t, report.TopRecommendation)
}

func TestGenerateReportEmpty(t *testing.T) {
	db := testDB(t)
	matrix := loadMatrix(t)

	report, err := GenerateReport(db, matrix, "2026-06-10")
	require.NoError(t, err)
	assert.Equal(t, 0, report.TotalRequests)
	assert.Contains(t, report.TopRecommendation, "No requests intercepted")
}

func TestRenderTerminal(t *testing.T) {
	report := &ReportData{
		Date:             "2026-06-10",
		TotalRequests:    10,
		TotalCostUSD:     0.50,
		OptimizedCostUSD: 0.25,
		SavingsPotential: 0.25,
		SavingsPercent:   50.0,
		TaskBreakdown: []TaskBreakdownEntry{
			{TaskType: "code_generation", Count: 5, TopModel: "claude-opus-4", CostUSD: 0.30, OptimizedModel: "gemini-2.5-flash", SavingUSD: 0.28, Indicator: "💰"},
		},
		ToolBreakdown: []ToolBreakdownEntry{
			{Tool: "cursor", Count: 10, CostUSD: 0.50},
		},
		TopRecommendation: "Switch code_generation to Gemini 2.5 Flash",
	}

	output := RenderTerminal(report)
	assert.Contains(t, output, "Tokensense")
	assert.Contains(t, output, "2026-06-10")
	assert.Contains(t, output, "code_generation")
	assert.Contains(t, output, "cursor")
}

func TestRenderTerminalEmpty(t *testing.T) {
	report := &ReportData{
		Date:          "2026-06-10",
		TotalRequests: 0,
	}

	output := RenderTerminal(report)
	assert.Contains(t, output, "No requests intercepted")
}

func TestReportPerformance(t *testing.T) {
	db := testDB(t)
	matrix := loadMatrix(t)

	// Seed 30 days of data
	for i := 0; i < 30; i++ {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		seedRequests(t, db, 20, date)
	}

	start := time.Now()
	_, err := GenerateReport(db, matrix, time.Now().Format("2006-01-02"))
	require.NoError(t, err)
	duration := time.Since(start)

	assert.Less(t, duration, time.Second, "report generation should be < 1s")
}
