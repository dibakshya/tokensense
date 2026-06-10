package detector

// ToolSourceFromName maps a detected tool name to a tool_source identifier for storage.
func ToolSourceFromName(name string) string {
	switch name {
	case "Cursor":
		return "cursor"
	case "Claude Desktop":
		return "claude_desktop"
	case "VS Code":
		return "vscode"
	case "Windsurf":
		return "windsurf"
	case "Direct API":
		return "direct"
	default:
		return "unknown"
	}
}

// AllToolNames returns the names of all known AI tools.
func AllToolNames() []string {
	return []string{
		"Cursor",
		"Claude Desktop",
		"VS Code",
		"Windsurf",
		"Direct API",
	}
}
