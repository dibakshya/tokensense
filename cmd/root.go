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

	// Show "what's next?" after every command that succeeds
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// These commands manage their own output / are long-running
		skip := map[string]bool{
			"setup":   true, // prints the full welcome banner instead
			"start":   true, // long-running foreground process
			"api":     true, // long-running server
			"version": true,
			"help":    true,
		}
		if skip[cmd.Name()] {
			return
		}
		PrintNextSteps(cmd.Name())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("\ntokensense %s\n  commit: %s\n  built:  %s\n\n", versionStr, commitStr, dateStr)
		PrintNextSteps("version")
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
