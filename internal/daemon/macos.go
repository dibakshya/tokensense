package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

const plistLabel = "dev.tokensense.proxy"

var plistTemplate = template.Must(template.New("plist").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{.Label}}</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.BinaryPath}}</string>
        <string>start</string>
        <string>--foreground</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>{{.LogPath}}</string>
    <key>StandardErrorPath</key>
    <string>{{.LogPath}}</string>
</dict>
</plist>
`))

type macOSService struct{}

func (s *macOSService) plistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", plistLabel+".plist")
}

func (s *macOSService) Install(binaryPath string) error {
	home, _ := os.UserHomeDir()
	logPath := filepath.Join(home, ".tokensense", "tokensense.log")

	f, err := os.Create(s.plistPath())
	if err != nil {
		return fmt.Errorf("cannot create plist: %w", err)
	}
	defer f.Close()

	data := struct {
		Label      string
		BinaryPath string
		LogPath    string
	}{
		Label:      plistLabel,
		BinaryPath: binaryPath,
		LogPath:    logPath,
	}

	if err := plistTemplate.Execute(f, data); err != nil {
		return fmt.Errorf("cannot write plist: %w", err)
	}

	return nil
}

func (s *macOSService) Uninstall() error {
	s.Stop()
	if err := os.Remove(s.plistPath()); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot remove plist: %w", err)
	}
	return nil
}

func (s *macOSService) Start() error {
	cmd := exec.Command("launchctl", "load", s.plistPath())
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("launchctl load failed: %s: %w", string(output), err)
	}
	return nil
}

func (s *macOSService) Stop() error {
	cmd := exec.Command("launchctl", "unload", s.plistPath())
	cmd.Run() // Ignore errors if already unloaded
	return nil
}

func (s *macOSService) Status() (string, error) {
	cmd := exec.Command("launchctl", "list", plistLabel)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "stopped", nil
	}
	return fmt.Sprintf("running\n%s", string(output)), nil
}
