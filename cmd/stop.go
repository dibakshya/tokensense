package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/dibakshya/tokensense/internal/config"
	"github.com/dibakshya/tokensense/internal/daemon"
)

func init() {
	rootCmd.AddCommand(stopCmd)
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Tokensense proxy daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		config.LoadConfig() //nolint:errcheck
		svc := daemon.New()
		if err := svc.Stop(); err != nil {
			return fmt.Errorf("cannot stop daemon: %w", err)
		}
		DisableSystemProxy()
		fmt.Println()
		fmt.Println(bold("  ⏹  Tokensense proxy stopped."))
		fmt.Println(dim("  System proxy cleared — AI tools will connect directly until you start again."))
		fmt.Println()
		return nil
	},
}
