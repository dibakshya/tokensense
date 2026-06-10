package reporter

import (
	"fmt"
	"strings"
)

// RenderTerminal renders the report to a terminal-friendly string using lipgloss-style formatting.
func RenderTerminal(report *ReportData) string {
	var sb strings.Builder

	border := "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	divider := "  ───────────────────────────────────────"

	sb.WriteString(fmt.Sprintf("\n%s\n", border))
	sb.WriteString(fmt.Sprintf("  Tokensense  ·  Daily Report  ·  %s\n", report.Date))
	sb.WriteString(fmt.Sprintf("%s\n\n", border))

	if report.TotalRequests == 0 {
		sb.WriteString("  No requests intercepted today.\n")
		sb.WriteString("  Make sure HTTPS_PROXY=http://127.0.0.1:7890 is set.\n\n")
		sb.WriteString(fmt.Sprintf("%s\n", border))
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("  Total spend:      $%.4f\n", report.TotalCostUSD))
	sb.WriteString(fmt.Sprintf("  Optimized could:  $%.4f  ↓%.0f%%  save $%.4f today\n\n",
		report.OptimizedCostUSD, report.SavingsPercent, report.SavingsPotential))

	// Task breakdown
	sb.WriteString("  Breakdown by task\n")
	sb.WriteString(fmt.Sprintf("%s\n", divider))
	for _, entry := range report.TaskBreakdown {
		sb.WriteString(fmt.Sprintf("  %-18s %3d calls  %-20s  $%.4f\n",
			entry.TaskType, entry.Count, entry.TopModel, entry.CostUSD))
		if entry.SavingUSD > 0.001 {
			sb.WriteString(fmt.Sprintf("    → Use %s  save $%.4f  [%s]\n",
				entry.OptimizedModel, entry.SavingUSD, entry.Indicator))
		}
	}

	// Tool breakdown
	sb.WriteString(fmt.Sprintf("\n  By tool\n%s\n", divider))
	for _, entry := range report.ToolBreakdown {
		sb.WriteString(fmt.Sprintf("  %-18s %3d calls   $%.4f\n",
			entry.Tool, entry.Count, entry.CostUSD))
	}

	// Top action
	sb.WriteString(fmt.Sprintf("\n  ✦ Top action: %s\n\n", report.TopRecommendation))

	sb.WriteString(fmt.Sprintf("%s\n", border))
	sb.WriteString("  tokensense report --open  ·  tokensense ask \"...\"\n")
	sb.WriteString(fmt.Sprintf("%s\n", border))

	return sb.String()
}
