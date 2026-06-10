package reporter

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.fkinternal.com/dibakshya-c/tokensense/internal/classifier"
	"github.fkinternal.com/dibakshya-c/tokensense/internal/storage"
)


// ReportData holds the computed report for a given date.
type ReportData struct {
	Date              string                `json:"date"`
	GeneratedAt       string                `json:"generated_at"`
	TotalRequests     int                   `json:"total_requests"`
	TotalTokensIn     int                   `json:"total_tokens_in"`
	TotalTokensOut    int                   `json:"total_tokens_out"`
	TotalCostUSD      float64               `json:"total_cost_usd"`
	OptimizedCostUSD  float64               `json:"optimized_cost_usd"`
	SavingsPotential  float64               `json:"savings_potential_usd"`
	SavingsPercent    float64               `json:"savings_percent"`
	TaskBreakdown     []TaskBreakdownEntry  `json:"task_breakdown"`
	ToolBreakdown     []ToolBreakdownEntry  `json:"tool_breakdown"`
	TopRecommendation string                `json:"top_recommendation"`
}

// TaskBreakdownEntry holds per-task-type statistics.
type TaskBreakdownEntry struct {
	TaskType       string  `json:"task_type"`
	Count          int     `json:"count"`
	TopModel       string  `json:"top_model"`
	CostUSD        float64 `json:"cost_usd"`
	OptimizedModel string  `json:"optimized_model"`
	SavingUSD      float64 `json:"saving_usd"`
	Indicator      string  `json:"indicator"`
}

// ToolBreakdownEntry holds per-tool statistics.
type ToolBreakdownEntry struct {
	Tool    string  `json:"tool"`
	Count   int     `json:"count"`
	CostUSD float64 `json:"cost_usd"`
}

// GenerateReport computes the daily report from DB rows and model matrix.
func GenerateReport(db *storage.DB, matrix *classifier.ModelMatrix, date string) (*ReportData, error) {
	requests, err := db.QueryByDate(date)
	if err != nil {
		return nil, fmt.Errorf("cannot query requests: %w", err)
	}

	report := &ReportData{
		Date:        date,
		GeneratedAt: time.Now().Format(time.RFC3339),
	}

	if len(requests) == 0 {
		report.TopRecommendation = "No requests intercepted today. Make sure HTTPS_PROXY is set."
		return report, nil
	}

	report.TotalRequests = len(requests)

	// Aggregate by task type
	taskMap := make(map[string]*taskAgg)
	toolMap := make(map[string]*toolAgg)

	for _, req := range requests {
		if req.TokensIn != nil {
			report.TotalTokensIn += *req.TokensIn
		}
		if req.TokensOut != nil {
			report.TotalTokensOut += *req.TokensOut
		}
		if req.CostUSD != nil {
			report.TotalCostUSD += *req.CostUSD
		}

		// Task aggregation
		taskType := "unknown"
		if req.TaskType != nil {
			taskType = *req.TaskType
		}
		if _, ok := taskMap[taskType]; !ok {
			taskMap[taskType] = &taskAgg{}
		}
		ta := taskMap[taskType]
		ta.count++
		if req.CostUSD != nil {
			ta.costUSD += *req.CostUSD
		}
		ta.models = append(ta.models, req.Model)
		if req.TokensIn != nil {
			ta.tokensIn += *req.TokensIn
		}
		if req.TokensOut != nil {
			ta.tokensOut += *req.TokensOut
		}

		// Tool aggregation
		tool := "unknown"
		if req.ToolSource != nil {
			tool = *req.ToolSource
		}
		if _, ok := toolMap[tool]; !ok {
			toolMap[tool] = &toolAgg{}
		}
		toolMap[tool].count++
		if req.CostUSD != nil {
			toolMap[tool].costUSD += *req.CostUSD
		}
	}

	// Calculate savings and build task breakdown
	maxSaving := 0.0
	maxSavingTask := ""
	maxSavingModel := ""

	for taskType, agg := range taskMap {
		entry := TaskBreakdownEntry{
			TaskType: taskType,
			Count:    agg.count,
			CostUSD:  agg.costUSD,
			TopModel: topModel(agg.models),
		}

		// Find cheapest recommended model for this task
		if matrix != nil && taskType != "unknown" {
			complexity := "medium" // Default assumption
			recs := matrix.RankModels(taskType, complexity)
			if len(recs) > 0 {
				cheapest := recs[0]
				optimizedCost := classifier.CostForRequest(cheapest.Model.Pricing, agg.tokensIn, agg.tokensOut)
				saving := agg.costUSD - optimizedCost
				if saving < 0 {
					saving = 0
				}
				entry.OptimizedModel = cheapest.Model.ID
				entry.SavingUSD = saving
				report.OptimizedCostUSD += optimizedCost

				if saving > 0.01 {
					entry.Indicator = "💰"
				} else {
					entry.Indicator = "✓"
				}

				if saving > maxSaving {
					maxSaving = saving
					maxSavingTask = taskType
					maxSavingModel = cheapest.Model.DisplayName
				}
			} else {
				report.OptimizedCostUSD += agg.costUSD
				entry.Indicator = "✓"
			}
		} else {
			report.OptimizedCostUSD += agg.costUSD
			entry.Indicator = "—"
		}

		report.TaskBreakdown = append(report.TaskBreakdown, entry)
	}

	// Sort task breakdown by cost descending
	sort.Slice(report.TaskBreakdown, func(i, j int) bool {
		return report.TaskBreakdown[i].CostUSD > report.TaskBreakdown[j].CostUSD
	})

	// Tool breakdown
	for tool, agg := range toolMap {
		report.ToolBreakdown = append(report.ToolBreakdown, ToolBreakdownEntry{
			Tool:    tool,
			Count:   agg.count,
			CostUSD: agg.costUSD,
		})
	}
	sort.Slice(report.ToolBreakdown, func(i, j int) bool {
		return report.ToolBreakdown[i].CostUSD > report.ToolBreakdown[j].CostUSD
	})

	// Savings
	report.SavingsPotential = report.TotalCostUSD - report.OptimizedCostUSD
	if report.SavingsPotential < 0 {
		report.SavingsPotential = 0
	}
	if report.TotalCostUSD > 0 {
		report.SavingsPercent = (report.SavingsPotential / report.TotalCostUSD) * 100
	}

	// Top recommendation
	if maxSaving > 0.01 {
		report.TopRecommendation = fmt.Sprintf("Switch %s tasks to %s — save $%.2f/day",
			maxSavingTask, maxSavingModel, maxSaving)
	} else {
		report.TopRecommendation = "Your model usage is already well-optimized!"
	}

	// Store the report
	reportJSON, _ := json.Marshal(report)
	reportJSONStr := string(reportJSON)
	dbReport := &storage.DailyReport{
		Date:                date,
		GeneratedAt:         time.Now().UnixMilli(),
		TotalRequests:       report.TotalRequests,
		TotalTokensIn:       report.TotalTokensIn,
		TotalTokensOut:      report.TotalTokensOut,
		TotalCostUSD:        report.TotalCostUSD,
		OptimizedCostUSD:    report.OptimizedCostUSD,
		SavingsPotentialUSD: report.SavingsPotential,
		ReportJSON:          &reportJSONStr,
	}
	if err := db.UpsertReport(dbReport); err != nil {
		return report, fmt.Errorf("cannot store report: %w", err)
	}

	return report, nil
}

type taskAgg struct {
	count    int
	costUSD  float64
	models   []string
	tokensIn int
	tokensOut int
}

type toolAgg struct {
	count   int
	costUSD float64
}

func topModel(models []string) string {
	counts := make(map[string]int)
	for _, m := range models {
		counts[m]++
	}

	// Collect and sort model IDs for deterministic tie-breaking
	ids := make([]string, 0, len(counts))
	for m := range counts {
		ids = append(ids, m)
	}
	sort.Strings(ids)

	best := ""
	bestCount := 0
	for _, m := range ids {
		if counts[m] > bestCount {
			best = m
			bestCount = counts[m]
		}
	}
	return best
}
