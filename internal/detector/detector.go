package detector

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ToolStatus represents the detection status of an AI tool.
type ToolStatus struct {
	Name           string
	Detected       bool
	InterceptMode  string // "tls" | "tunnel" | "none"
	CertPinned     bool
	StatusIcon     string // ✅ | ⚠ | ❌
	Notes          string
}

// Detect scans for known AI tools on the system.
func Detect() []ToolStatus {
	tools := knownTools()
	var results []ToolStatus

	for _, tool := range tools {
		status := ToolStatus{
			Name:       tool.name,
			CertPinned: tool.certPinned,
		}

		status.Detected = tool.detectFn()

		if status.Detected {
			if tool.certPinned {
				status.InterceptMode = "tunnel"
				status.StatusIcon = "⚠"
				status.Notes = "Certificate-pinned; metadata-only mode (tunnel)"
			} else {
				status.InterceptMode = "tls"
				status.StatusIcon = "✅"
				status.Notes = "Full TLS interception available"
			}
		} else {
			status.InterceptMode = "none"
			status.StatusIcon = "❌"
			status.Notes = "Not detected"
		}

		results = append(results, status)
	}

	return results
}

type knownTool struct {
	name       string
	certPinned bool
	detectFn   func() bool
}

func knownTools() []knownTool {
	return []knownTool{
		{
			name:       "Cursor",
			certPinned: false,
			detectFn:   detectCursor,
		},
		{
			name:       "Claude Desktop",
			certPinned: true,
			detectFn:   detectClaudeDesktop,
		},
		{
			name:       "VS Code",
			certPinned: false,
			detectFn:   detectVSCode,
		},
		{
			name:       "Windsurf",
			certPinned: false,
			detectFn:   detectWindsurf,
		},
		{
			name:       "Direct API",
			certPinned: false,
			detectFn:   func() bool { return true }, // Always possible
		},
	}
}

func detectCursor() bool {
	switch runtime.GOOS {
	case "darwin":
		return pathExists("/Applications/Cursor.app") || isProcessRunning("Cursor")
	case "linux":
		return isProcessRunning("cursor") || pathExists(filepath.Join(homeDir(), ".config/Cursor"))
	case "windows":
		return pathExists(filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Cursor"))
	}
	return false
}

func detectClaudeDesktop() bool {
	switch runtime.GOOS {
	case "darwin":
		return pathExists("/Applications/Claude.app") || isProcessRunning("Claude")
	case "linux":
		return isProcessRunning("claude")
	case "windows":
		return pathExists(filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Claude"))
	}
	return false
}

func detectVSCode() bool {
	switch runtime.GOOS {
	case "darwin":
		return pathExists("/Applications/Visual Studio Code.app") || isProcessRunning("Code")
	case "linux":
		return isProcessRunning("code") || pathExists("/usr/bin/code")
	case "windows":
		return pathExists(filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Microsoft VS Code"))
	}
	return false
}

func detectWindsurf() bool {
	switch runtime.GOOS {
	case "darwin":
		return pathExists("/Applications/Windsurf.app") || isProcessRunning("Windsurf")
	case "linux":
		return isProcessRunning("windsurf")
	case "windows":
		return pathExists(filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Windsurf"))
	}
	return false
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isProcessRunning(name string) bool {
	switch runtime.GOOS {
	case "darwin", "linux":
		cmd := exec.Command("pgrep", "-i", name)
		return cmd.Run() == nil
	case "windows":
		cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq "+name+".exe")
		output, err := cmd.Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(output), name)
	}
	return false
}

func homeDir() string {
	h, _ := os.UserHomeDir()
	return h
}
