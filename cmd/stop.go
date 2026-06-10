package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.fkinternal.com/dibakshya-c/tokensense/internal/daemon"
)

func init() {
	rootCmd.AddCommand(stopCmd)
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Tokensense proxy daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := daemon.New()
		if err := svc.Stop(); err != nil {
			return fmt.Errorf("cannot stop daemon: %w", err)
		}
		fmt.Println("Tokensense proxy stopped.")
		return nil
	},
}
