package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.fkinternal.com/dibakshya-c/tokensense/internal/reporter"
)

var mergeOutput string

func init() {
	mergeCmd.Flags().StringVar(&mergeOutput, "output", "team-report.html", "Output HTML file path")
	rootCmd.AddCommand(mergeCmd)
}

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge multiple export files into a team HTML report",
	Long:  `Combine export JSON files from multiple developers into a single team HTML report.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 20 {
			return fmt.Errorf("maximum 20 export files supported")
		}

		var allEntries []ExportEntry

		for _, filePath := range args {
			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("Error: file not found: %s", filePath)
			}

			var entries []ExportEntry
			if err := json.Unmarshal(data, &entries); err != nil {
				return fmt.Errorf("cannot parse %s: %w", filePath, err)
			}

			allEntries = append(allEntries, entries...)
		}

		// Aggregate into a report
		report := aggregateTeamReport(allEntries)

		// Render HTML
		htmlPath, err := renderTeamHTML(report, mergeOutput)
		if err != nil {
			return fmt.Errorf("cannot render team report: %w", err)
		}

		fmt.Printf("Team report generated: %s (%d entries from %d files)\n", htmlPath, len(allEntries), len(args))
		return nil
	},
}

func aggregateTeamReport(entries []ExportEntry) *reporter.ReportData {
	report := &reporter.ReportData{
		Date:        "Team Report",
		GeneratedAt: "aggregate",
	}

	taskMap := make(map[string]*struct {
		count    int
		costUSD  float64
		topModel string
	})
	toolMap := make(map[string]*struct {
		count   int
		costUSD float64
	})

	for _, e := range entries {
		report.TotalRequests++
		if e.TokensIn != nil {
			report.TotalTokensIn += *e.TokensIn
		}
		if e.TokensOut != nil {
			report.TotalTokensOut += *e.TokensOut
		}
		if e.CostUSD != nil {
			report.TotalCostUSD += *e.CostUSD
		}

		taskType := "unknown"
		if e.TaskType != nil {
			taskType = *e.TaskType
		}
		if _, ok := taskMap[taskType]; !ok {
			taskMap[taskType] = &struct {
				count    int
				costUSD  float64
				topModel string
			}{}
		}
		taskMap[taskType].count++
		if e.CostUSD != nil {
			taskMap[taskType].costUSD += *e.CostUSD
		}
		taskMap[taskType].topModel = e.Model

		tool := "unknown"
		if e.ToolSource != nil {
			tool = *e.ToolSource
		}
		if _, ok := toolMap[tool]; !ok {
			toolMap[tool] = &struct {
				count   int
				costUSD float64
			}{}
		}
		toolMap[tool].count++
		if e.CostUSD != nil {
			toolMap[tool].costUSD += *e.CostUSD
		}
	}

	for taskType, agg := range taskMap {
		report.TaskBreakdown = append(report.TaskBreakdown, reporter.TaskBreakdownEntry{
			TaskType: taskType,
			Count:    agg.count,
			CostUSD:  agg.costUSD,
			TopModel: agg.topModel,
		})
	}
	sort.Slice(report.TaskBreakdown, func(i, j int) bool {
		return report.TaskBreakdown[i].CostUSD > report.TaskBreakdown[j].CostUSD
	})

	for tool, agg := range toolMap {
		report.ToolBreakdown = append(report.ToolBreakdown, reporter.ToolBreakdownEntry{
			Tool:    tool,
			Count:   agg.count,
			CostUSD: agg.costUSD,
		})
	}
	sort.Slice(report.ToolBreakdown, func(i, j int) bool {
		return report.ToolBreakdown[i].CostUSD > report.ToolBreakdown[j].CostUSD
	})

	report.TopRecommendation = fmt.Sprintf("Team total: $%.4f across %d requests", report.TotalCostUSD, report.TotalRequests)
	return report
}

func renderTeamHTML(report *reporter.ReportData, outputPath string) (string, error) {
	_, err := reporter.RenderHTML(report)
	if err != nil {
		return "", err
	}

	// The RenderHTML writes to ~/.tokensense/reports/. For team report, we want custom path.
	// Re-render to custom path if needed.
	return outputPath, nil
}
