package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	AppDir     = ".tokensense"
	ConfigFile = "config.yaml"
)

// DefaultConfig returns the default configuration values.
func DefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"proxy_port":         7890,
		"proxy_host":         "127.0.0.1",
		"privacy_mode":       "content",
		"report_time":        "18:00",
		"log_level":          "info",
		"cloud_fallback":     true,
		"matrix_auto_update": true,
		"confidence_threshold": 0.6,
	}
}

// Dir returns the Tokensense config directory path (~/.tokensense).
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, AppDir), nil
}

// EnsureDir creates the config directory and subdirectories if they don't exist.
func EnsureDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}

	dirs := []string{
		dir,
		filepath.Join(dir, "reports"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return "", fmt.Errorf("cannot create directory %s: %w", d, err)
		}
	}
	return dir, nil
}

// ConfigPath returns the full path to config.yaml.
func ConfigPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFile), nil
}

// LoadConfig initializes viper with defaults and loads config from disk.
func LoadConfig() error {
	dir, err := EnsureDir()
	if err != nil {
		return err
	}

	for k, v := range DefaultConfig() {
		viper.SetDefault(k, v)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(dir)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return SaveConfig()
		}
		return fmt.Errorf("cannot read config: %w", err)
	}
	return nil
}

// SaveConfig writes the current viper configuration to disk.
func SaveConfig() error {
	cfgPath, err := ConfigPath()
	if err != nil {
		return err
	}
	return viper.WriteConfigAs(cfgPath)
}

// Get returns a config value by key.
func Get(key string) interface{} {
	return viper.Get(key)
}

// GetString returns a config value as string.
func GetString(key string) string {
	return viper.GetString(key)
}

// GetInt returns a config value as int.
func GetInt(key string) int {
	return viper.GetInt(key)
}

// GetFloat64 returns a config value as float64.
func GetFloat64(key string) float64 {
	return viper.GetFloat64(key)
}

// GetBool returns a config value as bool.
func GetBool(key string) bool {
	return viper.GetBool(key)
}

// Set sets a config value and saves to disk.
func Set(key string, value interface{}) error {
	viper.Set(key, value)
	return SaveConfig()
}

// AllSettings returns all config as a map.
func AllSettings() map[string]interface{} {
	return viper.AllSettings()
}

// DBPath returns the path to the SQLite database.
func DBPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "data.db"), nil
}

// LogPath returns the path to the log file.
func LogPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "tokensense.log"), nil
}

// ReportsDir returns the path to the reports directory.
func ReportsDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "reports"), nil
}

// MatrixPath returns the path to the cached model matrix.
func MatrixPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "model-matrix.yaml"), nil
}

// CAKeyPath returns the path to the CA private key.
func CAKeyPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "ca.key"), nil
}

// CACertPath returns the path to the CA certificate.
func CACertPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "ca.crt"), nil
}
