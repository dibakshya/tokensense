package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/dibakshya/tokensense/internal/config"
	"github.com/dibakshya/tokensense/internal/reporter"
	"github.com/dibakshya/tokensense/internal/storage"
	"github.com/dibakshya/tokensense/internal/updater"
)

var (
	reportDate string
	reportHTML bool
	reportOpen bool
)

func init() {
	reportCmd.Flags().StringVar(&reportDate, "date", "", "Report date (YYYY-MM-DD, default: today)")
	reportCmd.Flags().BoolVar(&reportHTML, "html", false, "Generate HTML report")
	reportCmd.Flags().BoolVar(&reportOpen, "open", false, "Open HTML report in browser")
	rootCmd.AddCommand(reportCmd)
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "View daily usage report",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.LoadConfig(); err != nil {
			return err
		}

		if reportDate == "" {
			reportDate = time.Now().Format("2006-01-02")
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

		matrix, _ := updater.LoadCachedMatrix(BundledMatrix)

		report, err := reporter.GenerateReport(db, matrix, reportDate)
		if err != nil {
			return fmt.Errorf("cannot generate report: %w", err)
		}

		// Terminal output
		fmt.Print(reporter.RenderTerminal(report))

		// HTML output
		if reportHTML || reportOpen {
			htmlPath, err := reporter.RenderHTML(report)
			if err != nil {
				return fmt.Errorf("cannot generate HTML report: %w", err)
			}
			fmt.Printf("\nHTML report saved to: %s\n", htmlPath)

			if reportOpen {
				openBrowser(htmlPath)
			}
		}

		return nil
	},
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}
