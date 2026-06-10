//go:build windows

package cert

import (
	"fmt"
	"os/exec"

	"github.com/dibakshya/tokensense/internal/config"
)

// InjectCA adds the Tokensense CA cert to the Windows trust store.
func InjectCA() error {
	certPath, err := config.CACertPath()
	if err != nil {
		return err
	}

	cmd := exec.Command("certutil", "-addstore", "-f", "ROOT", certPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("CA cert injection failed: %s. Run as Administrator: %w", string(output), err)
	}
	return nil
}

// RemoveCA removes the Tokensense CA cert from the Windows trust store.
func RemoveCA() error {
	cmd := exec.Command("certutil", "-delstore", "ROOT", "Tokensense Local CA")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("CA cert removal failed: %s: %w", string(output), err)
	}
	return nil
}

// VerifyCA checks if the Tokensense CA is trusted.
func VerifyCA() bool {
	certPath, err := config.CACertPath()
	if err != nil {
		return false
	}
	cmd := exec.Command("certutil", "-verify", certPath)
	return cmd.Run() == nil
}
