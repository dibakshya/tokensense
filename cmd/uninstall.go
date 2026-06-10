package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.fkinternal.com/dibakshya-c/tokensense/internal/cert"
	"github.fkinternal.com/dibakshya-c/tokensense/internal/config"
	"github.fkinternal.com/dibakshya-c/tokensense/internal/daemon"
)

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove all Tokensense artifacts from this machine",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("This will remove:")
		fmt.Println("  • CA certificate from OS trust store")
		fmt.Println("  • Tokensense service registration")
		fmt.Println("  • Shell profile proxy settings")
		fmt.Println("  • All data in ~/.tokensense/")
		fmt.Print("\nAre you sure? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Uninstall cancelled.")
			return nil
		}

		// Stop and remove service
		fmt.Print("Stopping service... ")
		svc := daemon.New()
		svc.Stop()
		svc.Uninstall()
		fmt.Println("done")

		// Remove CA from trust store
		fmt.Print("Removing CA certificate... ")
		if err := cert.RemoveCA(); err != nil {
			fmt.Printf("warning: %v\n", err)
		} else {
			fmt.Println("done")
		}

		// Remove shell profile entries
		fmt.Print("Cleaning shell profiles... ")
		cleanShellProfiles()
		fmt.Println("done")

		// Remove data directory
		fmt.Print("Removing ~/.tokensense/... ")
		dir, err := config.Dir()
		if err == nil {
			os.RemoveAll(dir)
		}
		fmt.Println("done")

		fmt.Println("\n✅ Tokensense uninstalled. Restart your terminal to clear proxy env vars.")
		return nil
	},
}

func cleanShellProfiles() {
	if runtime.GOOS == "windows" {
		cleanWindowsEnv()
		return
	}

	home, _ := os.UserHomeDir()
	profiles := []string{
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".profile"),
	}

	for _, profile := range profiles {
		data, err := os.ReadFile(profile)
		if err != nil {
			continue
		}

		content := string(data)
		startMarker := "# >>> tokensense proxy >>>"
		endMarker := "# <<< tokensense proxy <<<"

		startIdx := strings.Index(content, startMarker)
		endIdx := strings.Index(content, endMarker)
		if startIdx >= 0 && endIdx >= 0 {
			cleaned := content[:startIdx] + content[endIdx+len(endMarker):]
			cleaned = strings.TrimRight(cleaned, "\n") + "\n"
			os.WriteFile(profile, []byte(cleaned), 0644)
		}
	}
}

func cleanWindowsEnv() {
	for _, name := range []string{"HTTPS_PROXY", "HTTP_PROXY", "NO_PROXY"} {
		os.Unsetenv(name)
	}
}
