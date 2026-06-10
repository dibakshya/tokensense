package daemon

import (
	"fmt"
	"os/exec"
)

const serviceName = "TokensenseProxy"

type windowsService struct{}

func (s *windowsService) Install(binaryPath string) error {
	cmd := exec.Command("sc", "create", serviceName,
		"binpath=", fmt.Sprintf(`"%s" start --foreground`, binaryPath),
		"start=", "auto",
		"DisplayName=", "Tokensense Proxy",
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sc create failed: %s. Run as Administrator: %w", string(output), err)
	}
	return nil
}

func (s *windowsService) Uninstall() error {
	s.Stop()
	cmd := exec.Command("sc", "delete", serviceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sc delete failed: %s: %w", string(output), err)
	}
	return nil
}

func (s *windowsService) Start() error {
	cmd := exec.Command("sc", "start", serviceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sc start failed: %s: %w", string(output), err)
	}
	return nil
}

func (s *windowsService) Stop() error {
	exec.Command("sc", "stop", serviceName).Run()
	return nil
}

func (s *windowsService) Status() (string, error) {
	cmd := exec.Command("sc", "query", serviceName)
	output, err := cmd.Output()
	if err != nil {
		return "stopped", nil
	}
	return string(output), nil
}
