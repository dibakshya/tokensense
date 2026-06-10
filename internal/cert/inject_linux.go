//go:build linux

package cert

import (
	"fmt"
	"os"
	"os/exec"

	"github.fkinternal.com/dibakshya-c/tokensense/internal/config"
)

// InjectCA adds the Tokensense CA cert to the Linux trust store.
func InjectCA() error {
	certPath, err := config.CACertPath()
	if err != nil {
		return err
	}

	// Try Debian/Ubuntu path first
	if _, err := os.Stat("/usr/local/share/ca-certificates"); err == nil {
		return injectDebian(certPath)
	}
	// Try RHEL/Fedora/CentOS path
	if _, err := os.Stat("/etc/pki/ca-trust/source/anchors"); err == nil {
		return injectRHEL(certPath)
	}

	return fmt.Errorf("unsupported Linux distribution for CA injection. Manually trust: %s", certPath)
}

func injectDebian(certPath string) error {
	dest := "/usr/local/share/ca-certificates/tokensense.crt"
	cmd := exec.Command("sudo", "cp", certPath, dest)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cannot copy cert: %s: %w", string(output), err)
	}
	cmd = exec.Command("sudo", "update-ca-certificates")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("update-ca-certificates failed: %s: %w", string(output), err)
	}
	return nil
}

func injectRHEL(certPath string) error {
	dest := "/etc/pki/ca-trust/source/anchors/tokensense.crt"
	cmd := exec.Command("sudo", "cp", certPath, dest)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cannot copy cert: %s: %w", string(output), err)
	}
	cmd = exec.Command("sudo", "update-ca-trust")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("update-ca-trust failed: %s: %w", string(output), err)
	}
	return nil
}

// RemoveCA removes the Tokensense CA cert from the Linux trust store.
func RemoveCA() error {
	debianPath := "/usr/local/share/ca-certificates/tokensense.crt"
	rhelPath := "/etc/pki/ca-trust/source/anchors/tokensense.crt"

	if _, err := os.Stat(debianPath); err == nil {
		exec.Command("sudo", "rm", debianPath).Run()
		exec.Command("sudo", "update-ca-certificates", "--fresh").Run()
		return nil
	}
	if _, err := os.Stat(rhelPath); err == nil {
		exec.Command("sudo", "rm", rhelPath).Run()
		exec.Command("sudo", "update-ca-trust").Run()
		return nil
	}
	return nil
}

// VerifyCA checks if the Tokensense CA is trusted.
func VerifyCA() bool {
	certPath, err := config.CACertPath()
	if err != nil {
		return false
	}
	cmd := exec.Command("openssl", "verify", "-CApath", "/etc/ssl/certs", certPath)
	return cmd.Run() == nil
}
