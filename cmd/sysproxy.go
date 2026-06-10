package cmd

// sysproxy.go — OS-level proxy management
//
// HTTPS_PROXY in .zshrc is only visible to apps launched from a terminal.
// GUI apps (Cursor, Claude Desktop, VS Code) launched from the Dock or
// Spotlight never see those env vars.
//
// The fix: configure the OS-level proxy via networksetup (macOS), which
// every app reads regardless of how it was launched.

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/dibakshya/tokensense/internal/config"
)

// EnableSystemProxy sets the macOS system-wide HTTP/HTTPS proxy so that
// GUI apps route through Tokensense without needing HTTPS_PROXY in the shell.
// On non-macOS platforms this is a no-op (env vars are sufficient).
func EnableSystemProxy() error {
	if runtime.GOOS != "darwin" {
		return nil
	}
	return setMacProxy(true)
}

// DisableSystemProxy removes the macOS system proxy configuration.
// Called on stop and uninstall. Best-effort — errors are ignored.
func DisableSystemProxy() {
	if runtime.GOOS == "darwin" {
		setMacProxy(false) //nolint:errcheck
	}
}

func setMacProxy(enable bool) error {
	host := config.GetString("proxy_host")
	if host == "" {
		host = "127.0.0.1"
	}
	port := fmt.Sprintf("%d", config.GetInt("proxy_port"))
	if port == "0" {
		port = "7890"
	}

	out, err := exec.Command("networksetup", "-listallnetworkservices").Output()
	if err != nil {
		return fmt.Errorf("networksetup unavailable: %w", err)
	}

	services := parseMacNetworkServices(string(out))
	if len(services) == 0 {
		return fmt.Errorf("no network services found")
	}

	state := "off"
	if enable {
		state = "on"
	}

	for _, svc := range services {
		if enable {
			// Set proxy address and bypass list
			exec.Command("networksetup", "-setwebproxy", svc, host, port).Run()           //nolint:errcheck
			exec.Command("networksetup", "-setsecurewebproxy", svc, host, port).Run()     //nolint:errcheck
			exec.Command("networksetup", "-setproxybypassdomains", svc,                   //nolint:errcheck
				"localhost", "127.0.0.1", "::1").Run()
		}
		// Toggle on/off
		exec.Command("networksetup", "-setwebproxystate", svc, state).Run()        //nolint:errcheck
		exec.Command("networksetup", "-setsecurewebproxystate", svc, state).Run()  //nolint:errcheck
	}

	return nil
}

func parseMacNetworkServices(output string) []string {
	var services []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "An asterisk") {
			continue
		}
		services = append(services, line)
	}
	return services
}
