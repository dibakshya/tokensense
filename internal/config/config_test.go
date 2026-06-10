package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	defaults := DefaultConfig()
	assert.Equal(t, 7890, defaults["proxy_port"])
	assert.Equal(t, "127.0.0.1", defaults["proxy_host"])
	assert.Equal(t, "content", defaults["privacy_mode"])
	assert.Equal(t, "18:00", defaults["report_time"])
	assert.Equal(t, "info", defaults["log_level"])
	assert.Equal(t, true, defaults["cloud_fallback"])
	assert.Equal(t, true, defaults["matrix_auto_update"])
	assert.Equal(t, 0.6, defaults["confidence_threshold"])
}

func TestDir(t *testing.T) {
	dir, err := Dir()
	require.NoError(t, err)
	assert.Contains(t, dir, ".tokensense")
	assert.True(t, filepath.IsAbs(dir))
}

func TestEnsureDir(t *testing.T) {
	dir, err := EnsureDir()
	require.NoError(t, err)
	assert.DirExists(t, dir)
	assert.DirExists(t, filepath.Join(dir, "reports"))
}

func TestPathHelpers(t *testing.T) {
	dbPath, err := DBPath()
	require.NoError(t, err)
	assert.Contains(t, dbPath, "data.db")

	logPath, err := LogPath()
	require.NoError(t, err)
	assert.Contains(t, logPath, "tokensense.log")

	reportsDir, err := ReportsDir()
	require.NoError(t, err)
	assert.Contains(t, reportsDir, "reports")

	matrixPath, err := MatrixPath()
	require.NoError(t, err)
	assert.Contains(t, matrixPath, "model-matrix.yaml")

	caKeyPath, err := CAKeyPath()
	require.NoError(t, err)
	assert.Contains(t, caKeyPath, "ca.key")

	caCertPath, err := CACertPath()
	require.NoError(t, err)
	assert.Contains(t, caCertPath, "ca.crt")
}

func TestPathsAreAbsolute(t *testing.T) {
	fns := []func() (string, error){DBPath, LogPath, ReportsDir, MatrixPath, CAKeyPath, CACertPath, ConfigPath}
	for _, fn := range fns {
		p, err := fn()
		require.NoError(t, err)
		assert.True(t, filepath.IsAbs(p), "path should be absolute: %s", p)
	}
}

func TestConfigPath(t *testing.T) {
	cfgPath, err := ConfigPath()
	require.NoError(t, err)
	assert.Contains(t, cfgPath, "config.yaml")
}

func TestLoadAndSaveConfig(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()

	err := LoadConfig()
	require.NoError(t, err)

	// Should have defaults loaded
	assert.Equal(t, 7890, viper.GetInt("proxy_port"))
	assert.Equal(t, "127.0.0.1", viper.GetString("proxy_host"))
}

func TestSetAndGet(t *testing.T) {
	viper.Reset()
	_ = LoadConfig()

	assert.Equal(t, 7890, GetInt("proxy_port"))
	assert.Equal(t, "127.0.0.1", GetString("proxy_host"))
	assert.Equal(t, true, GetBool("cloud_fallback"))
	assert.Equal(t, 0.6, GetFloat64("confidence_threshold"))

	// Get generic
	val := Get("proxy_port")
	assert.NotNil(t, val)
}

func TestSetPersists(t *testing.T) {
	viper.Reset()
	_ = LoadConfig()

	err := Set("proxy_port", 9999)
	require.NoError(t, err)
	assert.Equal(t, 9999, GetInt("proxy_port"))

	// Restore default
	_ = Set("proxy_port", 7890)
}

func TestAllSettings(t *testing.T) {
	viper.Reset()
	_ = LoadConfig()

	settings := AllSettings()
	assert.NotEmpty(t, settings)
	_, hasPort := settings["proxy_port"]
	assert.True(t, hasPort)
}

func TestCAKeyPermissions(t *testing.T) {
	// If ca.key exists, verify its permissions
	keyPath, err := CAKeyPath()
	require.NoError(t, err)

	if info, err := os.Stat(keyPath); err == nil {
		perm := info.Mode().Perm()
		assert.Equal(t, os.FileMode(0600), perm, "CA key should have 0600 permissions, got %o", perm)
	}
	// If file doesn't exist, that's OK — test passes
}
