//go:build darwin

package cert

import (
	"fmt"
	"os/exec"

	"github.com/dibakshya/tokensense/internal/config"
)

// InjectCA adds the Tokensense CA cert to the macOS system keychain.
func InjectCA() error {
	certPath, err := config.CACertPath()
	if err != nil {
		return err
	}

	cmd := exec.Command("sudo", "security", "add-trusted-cert",
		"-d", "-r", "trustRoot",
		"-k", "/Library/Keychains/System.keychain",
		certPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("CA cert injection failed: %s. Run with sudo or as Administrator: %w", string(output), err)
	}
	return nil
}

// RemoveCA removes the Tokensense CA cert from the macOS system keychain.
func RemoveCA() error {
	cmd := exec.Command("sudo", "security", "delete-certificate",
		"-c", "Tokensense Local CA",
		"-t", "/Library/Keychains/System.keychain",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("CA cert removal failed: %s: %w", string(output), err)
	}
	return nil
}

// VerifyCA checks if the Tokensense CA is trusted.
func VerifyCA() bool {
	cmd := exec.Command("security", "find-certificate",
		"-c", "Tokensense Local CA",
		"/Library/Keychains/System.keychain",
	)
	return cmd.Run() == nil
}
