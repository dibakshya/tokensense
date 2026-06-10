package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

var systemdTemplate = template.Must(template.New("unit").Parse(`[Unit]
Description=Tokensense AI Token Optimizer Proxy
After=network.target

[Service]
Type=simple
ExecStart={{.BinaryPath}} start --foreground
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
`))

type linuxService struct{}

func (s *linuxService) unitPath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "systemd", "user")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "tokensense.service")
}

func (s *linuxService) Install(binaryPath string) error {
	f, err := os.Create(s.unitPath())
	if err != nil {
		return fmt.Errorf("cannot create systemd unit: %w", err)
	}
	defer f.Close()

	data := struct {
		BinaryPath string
	}{
		BinaryPath: binaryPath,
	}

	if err := systemdTemplate.Execute(f, data); err != nil {
		return fmt.Errorf("cannot write systemd unit: %w", err)
	}

	exec.Command("systemctl", "--user", "daemon-reload").Run()
	exec.Command("systemctl", "--user", "enable", "tokensense").Run()
	return nil
}

func (s *linuxService) Uninstall() error {
	s.Stop()
	exec.Command("systemctl", "--user", "disable", "tokensense").Run()
	if err := os.Remove(s.unitPath()); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot remove systemd unit: %w", err)
	}
	exec.Command("systemctl", "--user", "daemon-reload").Run()
	return nil
}

func (s *linuxService) Start() error {
	cmd := exec.Command("systemctl", "--user", "start", "tokensense")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl start failed: %s: %w", string(output), err)
	}
	return nil
}

func (s *linuxService) Stop() error {
	exec.Command("systemctl", "--user", "stop", "tokensense").Run()
	return nil
}

func (s *linuxService) Status() (string, error) {
	cmd := exec.Command("systemctl", "--user", "is-active", "tokensense")
	output, err := cmd.Output()
	if err != nil {
		return "stopped", nil
	}
	return string(output), nil
}
