package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/dibakshya/tokensense/internal/config"
	"github.com/dibakshya/tokensense/internal/daemon"
	"github.com/dibakshya/tokensense/internal/storage"
	"github.com/dibakshya/tokensense/internal/updater"
)

var statusVerbose bool

func init() {
	statusCmd.Flags().BoolVar(&statusVerbose, "verbose", false, "Show detailed metrics")
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Tokensense daemon and proxy status",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.LoadConfig(); err != nil {
			return err
		}

		// Daemon status
		svc := daemon.New()
		status, err := svc.Status()
		if err != nil {
			fmt.Printf("⚠  Tokensense proxy is not running. Run: tokensense start\n")
		} else {
			fmt.Printf("Proxy status: %s\n", status)
		}

		host := config.GetString("proxy_host")
		port := config.GetInt("proxy_port")
		fmt.Printf("Proxy address: %s:%d\n", host, port)
		fmt.Printf("Privacy mode: %s\n", config.GetString("privacy_mode"))
		fmt.Printf("Report time: %s\n", config.GetString("report_time"))

		// DB stats
		dbPath, err := config.DBPath()
		if err == nil {
			db, err := storage.Open(dbPath)
			if err == nil {
				defer db.Close()
				today := time.Now().Format("2006-01-02")
				count, _ := db.CountByDate(today)
				total, _ := db.TotalCount()
				fmt.Printf("\nRequests today: %d\n", count)
				fmt.Printf("Total requests: %d\n", total)

				if statusVerbose {
					requests, _ := db.QueryByDate(today)
					ruleBased, apiFallback, metadataOnly := 0, 0, 0
					for _, r := range requests {
						if r.ClassifierSource != nil {
							switch *r.ClassifierSource {
							case "rule_based":
								ruleBased++
							case "api_fallback":
								apiFallback++
							case "metadata_only":
								metadataOnly++
							}
						}
					}
					if len(requests) > 0 {
						fmt.Printf("  rule_based:     %d (%.0f%%)\n", ruleBased, float64(ruleBased)/float64(len(requests))*100)
						fmt.Printf("  api_fallback:   %d (%.0f%%)\n", apiFallback, float64(apiFallback)/float64(len(requests))*100)
						fmt.Printf("  metadata_only:  %d (%.0f%%)\n", metadataOnly, float64(metadataOnly)/float64(len(requests))*100)
					}
				}
			}
		}

		// Matrix staleness
		matrix, err := updater.LoadCachedMatrix(BundledMatrix)
		if err == nil {
			warning := updater.CheckStaleness(matrix)
			if warning != "" {
				fmt.Printf("\n%s\n", warning)
			} else {
				fmt.Printf("\nMatrix: v%s, updated %s\n", matrix.Version, matrix.LastUpdated)
			}
		}

		return nil
	},
}
