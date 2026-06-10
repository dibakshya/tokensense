package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/dibakshya/tokensense/internal/config"
	"github.com/dibakshya/tokensense/internal/daemon"
	"github.com/dibakshya/tokensense/internal/storage"
	"github.com/dibakshya/tokensense/internal/updater"
)

var statusVerbose bool
var statusJSON bool

func init() {
	statusCmd.Flags().BoolVar(&statusVerbose, "verbose", false, "Show detailed classification metrics")
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output as JSON (for agent/developer integration)")
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show proxy status and today's usage stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.LoadConfig(); err != nil {
			return err
		}

		svc := daemon.New()
		daemonStatus, daemonErr := svc.Status()
		host := config.GetString("proxy_host")
		port := config.GetInt("proxy_port")
		privacyMode := config.GetString("privacy_mode")

		var requestsToday, requestsTotal int
		var classifierBreakdown map[string]int

		dbPath, dbErr := config.DBPath()
		if dbErr == nil {
			db, err := storage.Open(dbPath)
			if err == nil {
				defer db.Close()
				today := time.Now().Format("2006-01-02")
				requestsToday, _ = db.CountByDate(today)
				requestsTotal, _ = db.TotalCount()
				if statusVerbose {
					requests, _ := db.QueryByDate(today)
					classifierBreakdown = map[string]int{}
					for _, r := range requests {
						if r.ClassifierSource != nil {
							classifierBreakdown[*r.ClassifierSource]++
						}
					}
				}
			}
		}

		matrix, _ := updater.LoadCachedMatrix(BundledMatrix)
		var matrixVersion, matrixUpdated, matrixWarning string
		if matrix != nil {
			matrixVersion = matrix.Version
			matrixUpdated = matrix.LastUpdated
			matrixWarning = updater.CheckStaleness(matrix)
		}

		proxyRunning := daemonErr == nil

		// ── JSON output (for agents / developers) ────────────────────────
		if statusJSON {
			out := map[string]interface{}{
				"proxy_running":  proxyRunning,
				"proxy_address":  fmt.Sprintf("%s:%d", host, port),
				"privacy_mode":   privacyMode,
				"requests_today": requestsToday,
				"requests_total": requestsTotal,
				"matrix_version": matrixVersion,
				"matrix_updated": matrixUpdated,
				"generated_at":   time.Now().Format(time.RFC3339),
			}
			if proxyRunning {
				out["daemon_status"] = daemonStatus
			}
			if classifierBreakdown != nil {
				out["classifier_breakdown"] = classifierBreakdown
			}
			b, _ := json.MarshalIndent(out, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		// ── Human-friendly output ─────────────────────────────────────────
		bar := "  " + repeatStr("─", 50)
		fmt.Println()
		fmt.Println(bold("  ╔════════════════════════════════════════════════════╗"))
		fmt.Println(bold("  ║") + "  📊  Tokensense Status                             " + bold("║"))
		fmt.Println(bold("  ╚════════════════════════════════════════════════════╝"))
		fmt.Println()

		if proxyRunning {
			fmt.Println("  " + green("✅  Proxy is ON") + "  " + dim("— your AI calls are being tracked"))
		} else {
			fmt.Println("  " + yellow("⚠️   Proxy is OFF") + "  " + dim("— run: tokensense start"))
		}
		fmt.Println()

		fmt.Println(bar)
		fmt.Println(bold("  📈  Today's Activity"))
		fmt.Println(bar)
		fmt.Printf("  %-30s  %s\n", "AI calls tracked today:", bold(fmt.Sprintf("%d", requestsToday)))
		fmt.Printf("  %-30s  %s\n", "Total calls (all time):", bold(fmt.Sprintf("%d", requestsTotal)))
		fmt.Println()

		fmt.Println(bar)
		fmt.Println(bold("  ⚙️   Configuration"))
		fmt.Println(bar)
		fmt.Printf("  %-30s  %s\n", "Proxy address:", cyan(fmt.Sprintf("%s:%d", host, port)))
		fmt.Printf("  %-30s  %s\n", "Privacy mode:", cyan(privacyMode))
		if matrixVersion != "" {
			fmt.Printf("  %-30s  %s\n", "Model matrix:", dim(fmt.Sprintf("v%s, updated %s", matrixVersion, matrixUpdated)))
		}
		fmt.Println()

		if matrixWarning != "" {
			fmt.Printf("  %s\n\n", yellow("⚠  "+matrixWarning))
		}

		if statusVerbose && len(classifierBreakdown) > 0 {
			fmt.Println(bar)
			fmt.Println(bold("  🔬  Classification Breakdown (today)"))
			fmt.Println(bar)
			for src, count := range classifierBreakdown {
				pct := 0.0
				if requestsToday > 0 {
					pct = float64(count) / float64(requestsToday) * 100
				}
				fmt.Printf("  %-22s %3d calls  (%.0f%%)\n", src+":", count, pct)
			}
			fmt.Println()
		}

		fmt.Println(bar)
		fmt.Println(bold("  🗂  Quick Commands"))
		fmt.Println(bar)
		fmt.Printf("  %-46s  %s\n", cyan("tokensense report"), dim("Cost breakdown + savings tips"))
		fmt.Printf("  %-46s  %s\n", cyan("tokensense report --html --open"), dim("Visual report in your browser"))
		fmt.Printf("  %-46s  %s\n", cyan(`tokensense ask "describe your task"`), dim("Model recommendations"))
		fmt.Printf("  %-46s  %s\n", cyan("tokensense api"), dim("JSON API for agents & developers"))
		fmt.Printf("  %-46s  %s\n", cyan("tokensense status --json"), dim("Machine-readable output"))
		fmt.Println(bar)
		fmt.Println()

		return nil
	},
}

func repeatStr(ch string, n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += ch
	}
	return s
}
