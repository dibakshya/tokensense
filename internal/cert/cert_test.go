package cert

import (
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndLoadCA(t *testing.T) {
	// Set up temp directory as tokensense home
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	tokensenseDir := filepath.Join(dir, ".tokensense")
	require.NoError(t, os.MkdirAll(tokensenseDir, 0755))

	err := GenerateCA()
	require.NoError(t, err)

	// Verify files exist
	keyPath := filepath.Join(tokensenseDir, "ca.key")
	certPath := filepath.Join(tokensenseDir, "ca.crt")

	assert.FileExists(t, keyPath)
	assert.FileExists(t, certPath)

	// Verify key permissions (0600)
	info, err := os.Stat(keyPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm(), "CA key should have 0600 permissions")

	// Load and verify CA
	caCert, caKey, err := LoadCA()
	require.NoError(t, err)
	assert.True(t, caCert.IsCA)
	assert.Equal(t, "Tokensense Local CA", caCert.Subject.CommonName)
	assert.NotNil(t, caKey)
}

func TestSignForHost(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	tokensenseDir := filepath.Join(dir, ".tokensense")
	require.NoError(t, os.MkdirAll(tokensenseDir, 0755))

	require.NoError(t, GenerateCA())
	caCert, caKey, err := LoadCA()
	require.NoError(t, err)

	hostCert, hostKey, err := SignForHost("api.anthropic.com", caCert, caKey)
	require.NoError(t, err)
	assert.NotNil(t, hostCert)
	assert.NotNil(t, hostKey)

	assert.Equal(t, "api.anthropic.com", hostCert.Subject.CommonName)
	assert.Contains(t, hostCert.DNSNames, "api.anthropic.com")

	// Verify the cert is signed by our CA
	pool := x509.NewCertPool()
	pool.AddCert(caCert)
	_, err = hostCert.Verify(x509.VerifyOptions{
		Roots: pool,
	})
	assert.NoError(t, err, "host cert should be verifiable against our CA")
}

func TestUniqueCAPerGeneration(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	tokensenseDir := filepath.Join(dir, ".tokensense")
	require.NoError(t, os.MkdirAll(tokensenseDir, 0755))

	require.NoError(t, GenerateCA())
	cert1, _, err := LoadCA()
	require.NoError(t, err)

	// Regenerate
	require.NoError(t, GenerateCA())
	cert2, _, err := LoadCA()
	require.NoError(t, err)

	// Serial numbers should differ
	assert.NotEqual(t, cert1.SerialNumber, cert2.SerialNumber,
		"different CA generations should have different serial numbers")
}

func TestCAExists(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	assert.False(t, CAExists(), "CA should not exist yet")

	tokensenseDir := filepath.Join(dir, ".tokensense")
	require.NoError(t, os.MkdirAll(tokensenseDir, 0755))
	require.NoError(t, GenerateCA())

	assert.True(t, CAExists(), "CA should exist after generation")
}
