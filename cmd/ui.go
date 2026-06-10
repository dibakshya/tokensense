package cmd

import (
	"fmt"
	"strings"
)

// ── ANSI helpers ────────────────────────────────────────────────────────────

const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiBlue   = "\033[34m"
	ansiCyan   = "\033[36m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiPurple = "\033[35m"
)

func bold(s string) string   { return ansiBold + s + ansiReset }
func dim(s string) string    { return ansiDim + s + ansiReset }
func blue(s string) string   { return ansiBlue + s + ansiReset }
func cyan(s string) string   { return ansiCyan + s + ansiReset }
func green(s string) string  { return ansiGreen + s + ansiReset }
func yellow(s string) string { return ansiYellow + s + ansiReset }
func purple(s string) string { return ansiPurple + s + ansiReset }

// ── Command catalogue ───────────────────────────────────────────────────────

type cmdEntry struct {
	cmd   string
	plain string // plain-English description for non-technical users
}

var trackCmds = []cmdEntry{
	{"tokensense status", "See if the proxy is on + how many AI calls happened today"},
	{"tokensense report", "View today's cost breakdown and where you could save money"},
	{"tokensense report --html --open", "Open a full visual report in your browser"},
}

var controlCmds = []cmdEntry{
	{"tokensense start", "Turn on tracking (also runs automatically at login)"},
	{"tokensense stop", "Pause tracking temporarily"},
	{"tokensense tools status", "See which AI tools (Cursor, Claude, Copilot…) are being tracked"},
}

var deepCmds = []cmdEntry{
	{`tokensense ask "I need to build an API"`, "Get model recommendations for any task you describe"},
	{"tokensense export", "Download all your usage data as JSON or CSV"},
	{"tokensense config list", "View and change settings"},
}

var devCmds = []cmdEntry{
	{"tokensense api", "Start a local JSON API on http://localhost:7891 for agents & tools"},
	{"tokensense report --json", "Get today's report as machine-readable JSON"},
	{"tokensense status --json", "Get proxy status as machine-readable JSON"},
}

// ── Welcome banner (shown once after setup) ─────────────────────────────────

func PrintWelcomeBanner() {
	w := "  "
	bar := "  " + strings.Repeat("─", 62)
	fmt.Println()
	fmt.Println(bold("  ╔══════════════════════════════════════════════════════════════╗"))
	fmt.Println(bold("  ║") + green("  🎉  Tokensense is set up and running!                      ") + bold("║"))
	fmt.Println(bold("  ╚══════════════════════════════════════════════════════════════╝"))
	fmt.Println()
	fmt.Println(w + dim("Your AI spending is now tracked silently in the background."))
	fmt.Println(w + dim("Here is everything you can do — no technical knowledge needed."))
	fmt.Println()

	printSection("📊  TRACK YOUR USAGE", trackCmds, bar)
	printSection("⚙️   CONTROL THE PROXY", controlCmds, bar)
	printSection("🔍  DIVE DEEPER", deepCmds, bar)
	printSection("🔌  FOR DEVELOPERS & AI AGENTS", devCmds, bar)

	fmt.Println(bar)
	fmt.Println(w + bold("Quick start:") + "  " + cyan("tokensense status") + "  →  " + cyan("tokensense report"))
	fmt.Println(w + dim("Docs & source: https://github.com/dibakshya/tokensense"))
	fmt.Println(bar)
	fmt.Println()
}

// ── Next-steps panel (shown after each command) ──────────────────────────────

// PrintNextSteps prints a compact "what can I do next?" panel.
// currentCmd is the command that just ran (e.g. "status", "report").
func PrintNextSteps(currentCmd string) {
	bar := "  " + strings.Repeat("─", 50)
	fmt.Println()
	fmt.Println(bar)
	fmt.Println(bold("  💡  What can you do next?"))
	fmt.Println(bar)

	switch currentCmd {
	case "start":
		printHint("📊 Check what's being tracked", "tokensense status")
		printHint("📈 View your usage report", "tokensense report")
	case "stop":
		printHint("▶️  Resume tracking", "tokensense start")
		printHint("📈 View your last report", "tokensense report")
	case "status":
		printHint("📈 See your cost breakdown", "tokensense report")
		printHint("🌐 Open visual report", "tokensense report --html --open")
		printHint("💬 Get model advice", `tokensense ask "describe your task"`)
	case "report":
		printHint("🌐 Open visual HTML report", "tokensense report --html --open")
		printHint("💬 Get AI model advice", `tokensense ask "describe your task"`)
		printHint("💾 Export raw data", "tokensense export")
	case "ask":
		printHint("📈 View full usage report", "tokensense report")
		printHint("🌐 Open visual report", "tokensense report --html --open")
	case "api":
		printHint("📖 API docs", "http://localhost:7891/v1/docs")
		printHint("📊 Live stats endpoint", "curl http://localhost:7891/v1/status")
	case "export":
		printHint("📈 View report", "tokensense report")
		printHint("⚙️  Change settings", "tokensense config list")
	default:
		printHint("📊 Check proxy status", "tokensense status")
		printHint("📈 View today's report", "tokensense report")
		printHint("💬 Get model advice", `tokensense ask "describe your task"`)
	}

	fmt.Println(bar)
	fmt.Println()
}

// ── Compact command reference (printed inside status/report) ─────────────────

func PrintCommandRef() {
	bar := "  " + strings.Repeat("─", 50)
	fmt.Println()
	fmt.Println(bar)
	fmt.Println(bold("  🗂  All Commands"))
	fmt.Println(bar)
	for _, e := range append(append(trackCmds, controlCmds...), deepCmds...) {
		fmt.Printf("  %-42s %s\n", cyan(e.cmd), dim(e.plain))
	}
	fmt.Println(bar)
	fmt.Printf("  %-42s %s\n", cyan("tokensense api"), dim("Start JSON API for developers & agents"))
	fmt.Println(bar)
	fmt.Println()
}

// ── Internal helpers ──────────────────────────────────────────────────────────

func printSection(title string, cmds []cmdEntry, bar string) {
	fmt.Println()
	fmt.Println("  " + bold(title))
	fmt.Println(bar)
	for _, e := range cmds {
		fmt.Printf("  %-46s %s\n", cyan(e.cmd), dim(e.plain))
	}
}

func printHint(label, cmd string) {
	fmt.Printf("  %-30s →  %s\n", label, cyan(cmd))
}
