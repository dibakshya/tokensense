package detector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectReturnsSomething(t *testing.T) {
	results := Detect()
	assert.NotEmpty(t, results, "Detect should return at least one tool")

	// Direct API should always be detected
	found := false
	for _, r := range results {
		if r.Name == "Direct API" {
			found = true
			assert.True(t, r.Detected)
			assert.Equal(t, "tls", r.InterceptMode)
			assert.Equal(t, "✅", r.StatusIcon)
		}
	}
	assert.True(t, found, "Direct API tool should always be present")
}

func TestDetectToolStatuses(t *testing.T) {
	results := Detect()
	for _, r := range results {
		// Every result should have valid fields
		assert.NotEmpty(t, r.Name)
		assert.NotEmpty(t, r.InterceptMode)
		assert.NotEmpty(t, r.StatusIcon)
		assert.NotEmpty(t, r.Notes)

		// InterceptMode should be one of valid values
		validModes := map[string]bool{"tls": true, "tunnel": true, "none": true}
		assert.True(t, validModes[r.InterceptMode], "invalid mode: %s", r.InterceptMode)

		// StatusIcon should be one of valid values
		validIcons := map[string]bool{"✅": true, "⚠": true, "❌": true}
		assert.True(t, validIcons[r.StatusIcon], "invalid icon: %s", r.StatusIcon)

		// Claude Desktop should always be cert-pinned
		if r.Name == "Claude Desktop" {
			assert.True(t, r.CertPinned)
			if r.Detected {
				assert.Equal(t, "tunnel", r.InterceptMode)
			}
		}
	}
}

func TestToolSourceFromName(t *testing.T) {
	tests := map[string]string{
		"Cursor":        "cursor",
		"Claude Desktop": "claude_desktop",
		"VS Code":       "vscode",
		"Windsurf":      "windsurf",
		"Direct API":    "direct",
		"RandomTool":    "unknown",
		"":              "unknown",
	}
	for name, expected := range tests {
		assert.Equal(t, expected, ToolSourceFromName(name), "ToolSourceFromName(%q)", name)
	}
}

func TestAllToolNames(t *testing.T) {
	names := AllToolNames()
	assert.Len(t, names, 5)
	assert.Contains(t, names, "Cursor")
	assert.Contains(t, names, "Claude Desktop")
	assert.Contains(t, names, "VS Code")
	assert.Contains(t, names, "Windsurf")
	assert.Contains(t, names, "Direct API")
}

func TestPathExistsFunc(t *testing.T) {
	assert.True(t, pathExists("/"))
	assert.False(t, pathExists("/nonexistent/path/12345"))
}
