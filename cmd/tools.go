package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/dibakshya/tokensense/internal/detector"
)

func init() {
	toolsCmd.AddCommand(toolsStatusCmd)
	rootCmd.AddCommand(toolsCmd)
}

var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Manage detected AI tools",
}

var toolsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show detected AI tools and their intercept status",
	Run: func(cmd *cobra.Command, args []string) {
		tools := detector.Detect()

		fmt.Println("\nDetected AI Tools")
		fmt.Println("────────────────────────────────────────────")
		for _, t := range tools {
			fmt.Printf("  %s %-18s  %-8s  %s\n", t.StatusIcon, t.Name, t.InterceptMode, t.Notes)
		}
		fmt.Println()
	},
}
