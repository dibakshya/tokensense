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

var startCmds = []cmdEntry{
	{"tokensense", "Open the browser dashboard (default — no arguments needed)"},
	{"tokensense dashboard", "Open the browser control panel — start/stop, reports, settings"},
	{"tokensense setup", "First-time setup wizard — run this once after installing"},
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
	{`tokensense ask "describe a task"`, "Get model recommendations for any task you describe"},
	{"tokensense export", "Download your usage data as JSON or CSV"},
	{"tokensense merge file1 file2", "Combine teammates' usage exports into one team report"},
	{"tokensense config list", "View all settings"},
	{"tokensense config set key value", "Change a setting (e.g. privacy_mode, report_time)"},
	{"tokensense uninstall", "Remove everything — cert, service, data, and shell config"},
}

var devCmds = []cmdEntry{
	{"tokensense api", "Start a local JSON API on http://localhost:7891 for agents & tools"},
	{"tokensense report --json", "Get today's report as machine-readable JSON"},
	{"tokensense status --json", "Get proxy status as machine-readable JSON"},
}

// ── Setup complete (shown once after tokensense setup) ───────────────────────
//
// proxyStarted = true  → daemon launched successfully by the wizard
// proxyStarted = false → daemon failed; user must run tokensense start

func PrintSetupComplete(proxyStarted bool) {
	w := "  "
	box  := "  " + strings.Repeat("═", 62)
	warn := bold(yellow("  ⚠️   RESTART YOUR TERMINAL BEFORE USING YOUR AI TOOLS"))
	sub  := "      " + yellow("This activates HTTPS_PROXY — without it nothing is tracked.")

	fmt.Println()
	fmt.Println(bold("  ╔" + strings.Repeat("═", 62) + "╗"))
	fmt.Println(bold("  ║") + green("  ✅  Tokensense setup complete!") + strings.Repeat(" ", 32) + bold("║"))
	fmt.Println(bold("  ╚" + strings.Repeat("═", 62) + "╝"))
	fmt.Println()

	// ── Top reminder — seen immediately ──────────────────────────────────
	fmt.Println(warn)
	fmt.Println(sub)
	fmt.Println()

	fmt.Println(bold(w + "YOUR NEXT STEPS:"))
	fmt.Println(box)
	fmt.Println()

	if proxyStarted {
		fmt.Println(green(w + "  ✅  Step 1 — proxy started automatically"))
		fmt.Println()
		fmt.Println(bold(w + "  Step 2 →  ") + bold(yellow("Restart your terminal")))
		fmt.Println(w + "           Activates HTTPS_PROXY so your AI tools route through")
		fmt.Println(w + "           the proxy. Nothing will be tracked until you do this.")
		fmt.Println()
		fmt.Println(green(w + "  ✅  Step 3 — browser dashboard is opening now…"))
		fmt.Println(w + "           It will appear in your browser in a moment.")
	} else {
		fmt.Println(bold(w + "  Step 1 →  " + cyan("tokensense start")))
		fmt.Println(w + "           starts the tracking proxy")
		fmt.Println()
		fmt.Println(bold(w + "  Step 2 →  ") + bold(yellow("Restart your terminal")))
		fmt.Println(w + "           Activates HTTPS_PROXY so your AI tools route through")
		fmt.Println(w + "           the proxy. Nothing will be tracked until you do this.")
		fmt.Println()
		fmt.Println(green(w + "  ✅  Step 3 — browser dashboard is opening now…"))
		fmt.Println(w + "           It will appear in your browser in a moment.")
	}

	fmt.Println()
	fmt.Println(box)
	fmt.Println()

	// ── Bottom reminder — seen after scrolling / reading the steps ───────
	fmt.Println(warn)
	fmt.Println(sub)
	fmt.Println()
	fmt.Println(w + dim("Prefer the terminal? Run: tokensense --help"))
	fmt.Println()
}

// PrintWelcomeBanner is kept for any callers outside setup.
func PrintWelcomeBanner() { PrintSetupComplete(true) }

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
		printHint("🌐 Open browser dashboard", "tokensense dashboard")
		printHint("📊 Check what's being tracked", "tokensense status")
		printHint("📈 View your usage report", "tokensense report")
	case "stop":
		printHint("🌐 Open browser dashboard", "tokensense dashboard")
		printHint("▶️  Resume tracking", "tokensense start")
	case "status":
		printHint("🌐 Open browser dashboard", "tokensense dashboard")
		printHint("📈 See your cost breakdown", "tokensense report")
		printHint("💬 Get model advice", `tokensense ask "describe your task"`)
	case "report":
		printHint("🌐 Open browser dashboard", "tokensense dashboard")
		printHint("🌐 Open visual HTML report", "tokensense report --html --open")
		printHint("💬 Get AI model advice", `tokensense ask "describe your task"`)
	case "ask":
		printHint("🌐 Open browser dashboard", "tokensense dashboard")
		printHint("📈 View full usage report", "tokensense report")
	case "api":
		printHint("📖 API docs", "http://localhost:7891/v1/docs")
		printHint("📊 Live stats endpoint", "curl http://localhost:7891/v1/status")
	case "export":
		printHint("🌐 Open browser dashboard", "tokensense dashboard")
		printHint("⚙️  Change settings", "tokensense config list")
	default:
		printHint("🌐 Open browser dashboard", "tokensense dashboard")
		printHint("📊 Check proxy status", "tokensense status")
		printHint("📈 View today's report", "tokensense report")
	}

	fmt.Println(bar)
	fmt.Println()
}

// ── Compact command reference (printed inside status/report) ─────────────────

func PrintCommandRef() {
	bar := "  " + strings.Repeat("─", 54)
	fmt.Println()
	fmt.Println(bar)
	fmt.Println(bold("  🗂  All Commands"))
	fmt.Println(bar)
	all := append(append(append(append(startCmds, trackCmds...), controlCmds...), deepCmds...), devCmds...)
	for _, e := range all {
		fmt.Printf("  %-44s %s\n", cyan(e.cmd), dim(e.plain))
	}
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
