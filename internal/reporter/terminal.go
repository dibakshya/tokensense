package reporter

import (
	"fmt"
	"strings"
)

// RenderTerminal renders the report as a human-friendly terminal string.
func RenderTerminal(report *ReportData) string {
	var sb strings.Builder

	heavy := "  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	light := "  ──────────────────────────────────────────────────"

	sb.WriteString("\n")
	sb.WriteString(heavy + "\n")
	sb.WriteString(fmt.Sprintf("  📈  Tokensense  ·  Daily Report  ·  %s\n", report.Date))
	sb.WriteString(heavy + "\n\n")

	if report.TotalRequests == 0 {
		sb.WriteString("  No AI calls intercepted today.\n\n")
		sb.WriteString("  Possible reasons:\n")
		sb.WriteString("    • The proxy isn't running  →  tokensense start\n")
		sb.WriteString("    • HTTPS_PROXY isn't set in your terminal  →  restart your terminal\n")
		sb.WriteString("    • You haven't used an AI tool today yet\n\n")
		sb.WriteString(light + "\n")
		sb.WriteString("  💡 Try:  tokensense status   to check if the proxy is running\n")
		sb.WriteString(light + "\n\n")
		return sb.String()
	}

	// ── Summary ────────────────────────────────────────────────────────────
	sb.WriteString(fmt.Sprintf("  🔢  Total AI calls today:   %d\n", report.TotalRequests))
	sb.WriteString(fmt.Sprintf("  💰  Total spend:            $%.4f\n", report.TotalCostUSD))
	if report.SavingsPotential > 0.001 {
		sb.WriteString(fmt.Sprintf("  💡  You could save:         $%.4f today  (%.0f%%) by switching models\n",
			report.SavingsPotential, report.SavingsPercent))
	} else {
		sb.WriteString("  ✅  Model usage is already well-optimized!\n")
	}
	sb.WriteString("\n")

	// ── Task breakdown ─────────────────────────────────────────────────────
	sb.WriteString("  📂  Breakdown by task type\n")
	sb.WriteString(light + "\n")
	sb.WriteString(fmt.Sprintf("  %-20s  %6s  %-22s  %8s\n", "Task", "Calls", "Model used", "Cost"))
	sb.WriteString(light + "\n")
	for _, entry := range report.TaskBreakdown {
		sb.WriteString(fmt.Sprintf("  %-20s  %6d  %-22s  $%.4f\n",
			entry.TaskType, entry.Count, truncate(entry.TopModel, 22), entry.CostUSD))
		if entry.SavingUSD > 0.001 {
			sb.WriteString(fmt.Sprintf("    %s Switch to %-20s save $%.4f  %s\n",
				"→", truncate(entry.OptimizedModel, 20), entry.SavingUSD, entry.Indicator))
		}
	}
	sb.WriteString("\n")

	// ── Tool breakdown ─────────────────────────────────────────────────────
	if len(report.ToolBreakdown) > 0 {
		sb.WriteString("  🔧  Breakdown by AI tool\n")
		sb.WriteString(light + "\n")
		for _, entry := range report.ToolBreakdown {
			sb.WriteString(fmt.Sprintf("  %-22s  %6d calls   $%.4f\n",
				entry.Tool, entry.Count, entry.CostUSD))
		}
		sb.WriteString("\n")
	}

	// ── Top recommendation ─────────────────────────────────────────────────
	sb.WriteString(light + "\n")
	sb.WriteString(fmt.Sprintf("  ✦  Top tip: %s\n", report.TopRecommendation))
	sb.WriteString(light + "\n\n")

	// ── Command reference ─────────────────────────────────────────────────
	sb.WriteString(heavy + "\n")
	sb.WriteString("  🗂  What can you do with this data?\n")
	sb.WriteString(heavy + "\n")
	sb.WriteString(fmt.Sprintf("  %-48s  %s\n", "tokensense report --html --open", "→ Open a visual chart-based report in your browser"))
	sb.WriteString(fmt.Sprintf("  %-48s  %s\n", `tokensense ask "describe a task"`, "→ Get model recommendations for any task"))
	sb.WriteString(fmt.Sprintf("  %-48s  %s\n", "tokensense report --json", "→ Machine-readable JSON (for agents & tools)"))
	sb.WriteString(fmt.Sprintf("  %-48s  %s\n", "tokensense export", "→ Download raw data as CSV or JSON"))
	sb.WriteString(fmt.Sprintf("  %-48s  %s\n", "tokensense api", "→ Start local JSON API for developer integration"))
	sb.WriteString(heavy + "\n\n")

	return sb.String()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

// wrapWidth pads or truncates s to exactly width characters.
func wrapWidth(s string, width int) string {
	if len(s) >= width {
		return truncate(s, width)
	}
	return s + strings.Repeat(" ", width-len(s))
}
