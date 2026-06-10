package wizard

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Screen represents the current wizard screen.
type Screen int

const (
	ScreenWelcome Screen = iota
	ScreenPrivacy
	ScreenToolDetection
	ScreenCertInstall
	ScreenReportTime
	ScreenDone
)

// Config holds the wizard results.
type Config struct {
	PrivacyMode   string // "content" or "metadata"
	ReportTime    string // "HH:MM"
	InstallCert   bool
	CertInstalled bool
	CertError     string
	DetectedTools []string
}

// Model is the bubbletea model for the setup wizard.
type Model struct {
	screen   Screen
	config   Config
	cursor   int
	quitting bool
	done     bool
	err      error

	// Callbacks for actual operations
	onDetectTools func() []string
	onInstallCert func() error
}

// NewModel creates a new wizard model.
func NewModel(detectTools func() []string, installCert func() error) Model {
	return Model{
		screen: ScreenWelcome,
		config: Config{
			PrivacyMode: "content",
			ReportTime:  "18:00",
		},
		onDetectTools: detectTools,
		onInstallCert: installCert,
	}
}

// Result returns the wizard configuration after completion.
func (m Model) Result() Config {
	return m.config
}

// Done returns true if the wizard completed.
func (m Model) Done() bool {
	return m.done
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			return m.nextScreen()
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.screen == ScreenPrivacy && m.cursor < 1 {
				m.cursor++
			}
		case "backspace":
			if m.screen == ScreenReportTime && len(m.config.ReportTime) > 0 {
				m.config.ReportTime = m.config.ReportTime[:len(m.config.ReportTime)-1]
			}
		default:
			if m.screen == ScreenReportTime && len(msg.String()) == 1 {
				ch := msg.String()[0]
				if (ch >= '0' && ch <= '9') || ch == ':' {
					if len(m.config.ReportTime) < 5 {
						m.config.ReportTime += msg.String()
					}
				}
			}
		}
	}
	return m, nil
}

func (m Model) nextScreen() (tea.Model, tea.Cmd) {
	switch m.screen {
	case ScreenWelcome:
		m.screen = ScreenPrivacy
		m.cursor = 0
	case ScreenPrivacy:
		if m.cursor == 0 {
			m.config.PrivacyMode = "content"
		} else {
			m.config.PrivacyMode = "metadata"
		}
		m.screen = ScreenToolDetection
		if m.onDetectTools != nil {
			m.config.DetectedTools = m.onDetectTools()
		}
	case ScreenToolDetection:
		m.screen = ScreenCertInstall
	case ScreenCertInstall:
		m.config.InstallCert = true
		if m.onInstallCert != nil {
			if err := m.onInstallCert(); err != nil {
				m.config.CertError = err.Error()
			} else {
				m.config.CertInstalled = true
			}
		}
		m.screen = ScreenReportTime
	case ScreenReportTime:
		m.screen = ScreenDone
	case ScreenDone:
		m.done = true
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return "Setup cancelled.\n"
	}

	var sb strings.Builder

	switch m.screen {
	case ScreenWelcome:
		sb.WriteString(renderWelcome())
	case ScreenPrivacy:
		sb.WriteString(renderPrivacy(m.cursor))
	case ScreenToolDetection:
		sb.WriteString(renderToolDetection(m.config.DetectedTools))
	case ScreenCertInstall:
		sb.WriteString(renderCertInstall(m.config))
	case ScreenReportTime:
		sb.WriteString(renderReportTime(m.config.ReportTime))
	case ScreenDone:
		sb.WriteString(renderDone(m.config))
	}

	return sb.String()
}

func renderWelcome() string {
	return `
╔══════════════════════════════════════════════════════╗
║         Tokensense — AI Token Optimizer              ║
╚══════════════════════════════════════════════════════╝

  What Tokensense does:
  • Intercepts AI API calls via a local HTTPS proxy
  • Classifies each request by task type (never stores content)
  • Generates daily reports showing where cheaper models could be used
  • Provides an interactive advisor for optimal model selection

  What Tokensense does NOT do:
  • Send any data to remote servers (everything is local)
  • Store your prompts or AI responses
  • Modify or block your AI requests

  Press ENTER to continue, Q to quit
`
}

func renderPrivacy(cursor int) string {
	content := "  ● "
	metadata := "  ○ "
	if cursor == 1 {
		content = "  ○ "
		metadata = "  ● "
	}

	return fmt.Sprintf(`
  Privacy Mode
  ────────────────────────────────────────────

%sContent Mode (recommended)
    Reads prompts locally to classify tasks.
    Nothing written to disk. Nothing leaves your machine.
    3× better recommendations.

%sMetadata Only
    Sees only provider, model, token count, cost.
    No content ever seen.

  Change anytime: tokensense config set privacy_mode [content|metadata]

  Use ↑/↓ to select, ENTER to confirm
`, content, metadata)
}

func renderToolDetection(tools []string) string {
	var sb strings.Builder
	sb.WriteString("\n  Tool Detection\n  ────────────────────────────────────────────\n\n")
	if len(tools) == 0 {
		sb.WriteString("  No AI tools detected.\n")
	} else {
		for _, t := range tools {
			sb.WriteString(fmt.Sprintf("  ✅ %s\n", t))
		}
	}
	sb.WriteString("\n  Run 'tokensense tools status' anytime to re-check\n")
	sb.WriteString("\n  Press ENTER to continue\n")
	return sb.String()
}

func renderCertInstall(cfg Config) string {
	var sb strings.Builder
	sb.WriteString(`
  Certificate Install
  ────────────────────────────────────────────

  Tokensense needs to install a local CA certificate
  to intercept HTTPS traffic to AI APIs.

  Trust assurances:
  • Certificate is generated locally and unique to this install
  • Private key never leaves your machine
  • Removed automatically with 'tokensense uninstall'

`)
	if cfg.CertInstalled {
		sb.WriteString("  ✅ Certificate installed successfully!\n")
	} else if cfg.CertError != "" {
		sb.WriteString(fmt.Sprintf("  ❌ Error: %s\n", cfg.CertError))
		sb.WriteString("  You can retry with: tokensense setup --repair-cert\n")
	} else {
		sb.WriteString("  Press ENTER to install certificate\n")
	}
	return sb.String()
}

func renderReportTime(reportTime string) string {
	return fmt.Sprintf(`
  Report Time
  ────────────────────────────────────────────

  When should Tokensense generate your daily report?

  Time: %s

  Reports saved to: ~/.tokensense/reports/YYYY-MM-DD.html

  Type a time (HH:MM format), then press ENTER
`, reportTime)
}

func renderDone(cfg Config) string {
	certStatus := "❌ Not installed"
	if cfg.CertInstalled {
		certStatus = "✅ Installed"
	}
	return fmt.Sprintf(`
  Setup Complete!
  ────────────────────────────────────────────

  Summary:
  • Privacy mode: %s
  • Certificate:  %s
  • Report time:  %s

  Shell profile updated. Restart your terminal or run:
    source ~/.zshrc  (or ~/.bashrc)

  Key commands:
  • tokensense status    — check proxy and intercept status
  • tokensense report    — view today's report
  • tokensense ask "..." — get model recommendations

  Press ENTER to finish
`, cfg.PrivacyMode, certStatus, cfg.ReportTime)
}
