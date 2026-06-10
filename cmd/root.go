package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	versionStr string
	commitStr  string
	dateStr    string
)

// SetVersionInfo sets build-time version info from main.go ldflags.
func SetVersionInfo(version, commit, date string) {
	versionStr = version
	commitStr = commit
	dateStr = date
}

var rootCmd = &cobra.Command{
	Use:   "tokensense",
	Short: "AI token usage optimizer — local proxy for model cost analysis",
	Long: `Tokensense is an open-source AI token usage optimizer.
It intercepts AI API calls via a local HTTPS proxy, classifies each request 
by task type, and generates reports showing where cheaper models could be used.

Everything runs locally. No server. No account. No cloud dependency.`,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("tokensense %s\n  commit: %s\n  built:  %s\n", versionStr, commitStr, dateStr)
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
