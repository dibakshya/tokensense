package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/dibakshya/tokensense/internal/cert"
	"github.com/dibakshya/tokensense/internal/config"
	"github.com/dibakshya/tokensense/internal/daemon"
	"github.com/dibakshya/tokensense/internal/detector"
	"github.com/dibakshya/tokensense/internal/wizard"
)

const guideFilename = "token-optimization-guide.md"

var repairCert bool

func init() {
	setupCmd.Flags().BoolVar(&repairCert, "repair-cert", false, "Re-generate and re-inject CA certificate")
	rootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "First-time setup wizard",
	Long:  `Interactive setup wizard that configures the proxy, certificate, and OS service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ensure config directory exists
		if _, err := config.EnsureDir(); err != nil {
			return err
		}
		if err := config.LoadConfig(); err != nil {
			return err
		}

		if repairCert {
			return doRepairCert()
		}

		// Generate CA if not exists
		if !cert.CAExists() {
			fmt.Println("Generating CA certificate...")
			if err := cert.GenerateCA(); err != nil {
				return fmt.Errorf("cannot generate CA: %w", err)
			}
		}

		// Run wizard
		detectToolsFn := func() []string {
			tools := detector.Detect()
			var names []string
			for _, t := range tools {
				if t.Detected {
					names = append(names, t.Name)
				}
			}
			return names
		}

		installCertFn := func() error {
			return cert.InjectCA()
		}

		model := wizard.NewModel(detectToolsFn, installCertFn)
		p := tea.NewProgram(model)
		finalModel, err := p.Run()
		if err != nil {
			return fmt.Errorf("wizard error: %w", err)
		}

		wm := finalModel.(wizard.Model)
		if !wm.Done() {
			fmt.Println("Setup cancelled.")
			return nil
		}

		result := wm.Result()

		// Apply config
		config.Set("privacy_mode", result.PrivacyMode)
		config.Set("report_time", result.ReportTime)

		// Update shell profiles
		if err := updateShellProfiles(); err != nil {
			fmt.Printf("Warning: could not update shell profiles: %v\n", err)
			fmt.Println("Manually add to your shell profile:")
			fmt.Println("  export HTTPS_PROXY=http://127.0.0.1:7890")
			fmt.Println("  export HTTP_PROXY=http://127.0.0.1:7890")
			fmt.Println("  export NO_PROXY=localhost,127.0.0.1,::1")
		}

		// Register and start daemon
		binaryPath, err := os.Executable()
		if err != nil {
			binaryPath = "tokensense"
		}

		svc := daemon.New()
		if err := svc.Install(binaryPath); err != nil {
			fmt.Printf("Warning: cannot register service: %v\n", err)
		}
		proxyStarted := svc.Start() == nil

		// Set the OS-level system proxy so GUI apps (Cursor, Claude Desktop,
		// VS Code, etc.) route through Tokensense without needing a terminal restart.
		if err := EnableSystemProxy(); err != nil {
			fmt.Printf("  %s\n", yellow("⚠️  Could not set system proxy automatically: "+err.Error()))
			fmt.Println("  Set it manually: System Settings → Network → (your connection) → Proxies")
			fmt.Println("  Enable 'Web Proxy' and 'Secure Web Proxy', set server 127.0.0.1 port 7890")
		}

		// Save the token optimization guide to the desktop on first run
		if len(BundledGuide) > 0 {
			saveGuideToDesktop()
		}

		PrintSetupComplete(proxyStarted)

		// Automatically open the browser dashboard — no extra command needed.
		// New users land directly in the UI without having to know any commands.
		return runDashboard(cmd, args)
	},
}

// saveGuideToDesktop writes the bundled token guide to ~/Desktop on first setup.
func saveGuideToDesktop() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	var desktopPath string
	switch runtime.GOOS {
	case "windows":
		desktopPath = filepath.Join(home, "Desktop")
	default:
		desktopPath = filepath.Join(home, "Desktop")
	}

	if _, err := os.Stat(desktopPath); os.IsNotExist(err) {
		return // no Desktop directory (headless / Linux server)
	}

	guidePath := filepath.Join(desktopPath, guideFilename)

	// Don't overwrite if already present
	if _, err := os.Stat(guidePath); err == nil {
		return
	}

	if err := os.WriteFile(guidePath, BundledGuide, 0644); err != nil {
		return
	}

	fmt.Printf("\n📘 Token Optimization Guide saved to: %s\n", guidePath)
	fmt.Println("   Open it anytime for tips on reducing AI costs.")

	// Open the guide in the default viewer
	openBrowser(guidePath)
}

func doRepairCert() error {
	fmt.Println("Re-generating CA certificate...")
	if err := cert.GenerateCA(); err != nil {
		return fmt.Errorf("cannot generate CA: %w", err)
	}
	if err := cert.InjectCA(); err != nil {
		return fmt.Errorf("CA cert injection failed: permission denied. Run with sudo or as Administrator: %w", err)
	}
	fmt.Println("✅ CA certificate repaired and injected.")
	return nil
}

const shellBlock = `
# >>> tokensense proxy >>>
export HTTPS_PROXY=http://127.0.0.1:7890
export HTTP_PROXY=http://127.0.0.1:7890
export NO_PROXY=localhost,127.0.0.1,::1
# <<< tokensense proxy <<<
`

func updateShellProfiles() error {
	if runtime.GOOS == "windows" {
		return updateWindowsEnv()
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	profiles := []string{
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".profile"),
	}

	for _, profile := range profiles {
		if _, err := os.Stat(profile); err != nil {
			continue
		}

		data, err := os.ReadFile(profile)
		if err != nil {
			continue
		}

		// Skip if already present
		if strings.Contains(string(data), "tokensense proxy") {
			continue
		}

		// Create backup
		backupPath := profile + ".bak"
		os.WriteFile(backupPath, data, 0644)

		// Append block
		f, err := os.OpenFile(profile, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			continue
		}
		f.WriteString(shellBlock)
		f.Close()
	}

	return nil
}

func updateWindowsEnv() error {
	cmds := [][]string{
		{"setx", "HTTPS_PROXY", "http://127.0.0.1:7890"},
		{"setx", "HTTP_PROXY", "http://127.0.0.1:7890"},
		{"setx", "NO_PROXY", "localhost,127.0.0.1,::1"},
	}
	for _, c := range cmds {
		exec.Command(c[0], c[1:]...).Run()
	}
	return nil
}
