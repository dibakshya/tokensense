package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/dibakshya/tokensense/internal/config"
	"github.com/dibakshya/tokensense/internal/storage"
)

var exportOutput string

func init() {
	exportCmd.Flags().StringVar(&exportOutput, "output", "", "Output file path (default: tokensense-export.json)")
	rootCmd.AddCommand(exportCmd)
}

// ExportEntry is the anonymized export format for a single request.
type ExportEntry struct {
	DayDate              string   `json:"day_date"`
	Provider             string   `json:"provider"`
	Model                string   `json:"model"`
	TaskType             *string  `json:"task_type"`
	Complexity           *string  `json:"complexity"`
	TokensIn             *int     `json:"tokens_in"`
	TokensOut            *int     `json:"tokens_out"`
	CostUSD              *float64 `json:"cost_usd"`
	LatencyMs            *int     `json:"latency_ms"`
	ClassifierSource     *string  `json:"classifier_source"`
	ClassifierConfidence *float64 `json:"classifier_confidence"`
	ToolSource           *string  `json:"tool_source"`
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export anonymized usage data as JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.LoadConfig(); err != nil {
			return err
		}

		if exportOutput == "" {
			exportOutput = "tokensense-export.json"
		}

		dbPath, err := config.DBPath()
		if err != nil {
			return err
		}
		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("cannot open database: %w", err)
		}
		defer db.Close()

		// Export last 30 days
		endDate := time.Now().Format("2006-01-02")
		startDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")

		requests, err := db.QueryByDateRange(startDate, endDate)
		if err != nil {
			return fmt.Errorf("cannot query requests: %w", err)
		}

		var entries []ExportEntry
		for _, req := range requests {
			entries = append(entries, ExportEntry{
				DayDate:              req.DayDate,
				Provider:             req.Provider,
				Model:                req.Model,
				TaskType:             req.TaskType,
				Complexity:           req.Complexity,
				TokensIn:             req.TokensIn,
				TokensOut:            req.TokensOut,
				CostUSD:              req.CostUSD,
				LatencyMs:            req.LatencyMs,
				ClassifierSource:     req.ClassifierSource,
				ClassifierConfidence: req.ClassifierConfidence,
				ToolSource:           req.ToolSource,
			})
		}

		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return fmt.Errorf("cannot marshal export: %w", err)
		}

		if err := os.WriteFile(exportOutput, data, 0644); err != nil {
			return fmt.Errorf("cannot write export file: %w", err)
		}

		fmt.Printf("Exported %d entries to %s\n", len(entries), exportOutput)
		return nil
	},
}
