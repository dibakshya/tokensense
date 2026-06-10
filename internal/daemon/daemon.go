package daemon

import "runtime"

// ServiceManager provides platform-specific service management.
type ServiceManager interface {
	Install(binaryPath string) error
	Uninstall() error
	Start() error
	Stop() error
	Status() (string, error)
}

// New returns the platform-appropriate ServiceManager.
func New() ServiceManager {
	switch runtime.GOOS {
	case "darwin":
		return &macOSService{}
	case "linux":
		return &linuxService{}
	case "windows":
		return &windowsService{}
	default:
		return &macOSService{} // Fallback
	}
}
