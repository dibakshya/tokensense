package reporter

import (
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.fkinternal.com/dibakshya-c/tokensense/internal/config"
)

//go:embed templates/report.html.tmpl
var reportHTMLTemplate string

// RenderHTML renders the report as an HTML file and returns the file path.
func RenderHTML(report *ReportData) (string, error) {
	tmpl, err := template.New("report").Parse(reportHTMLTemplate)
	if err != nil {
		return "", fmt.Errorf("cannot parse HTML template: %w", err)
	}

	reportsDir, err := config.ReportsDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return "", fmt.Errorf("cannot create reports dir: %w", err)
	}

	filePath := filepath.Join(reportsDir, report.Date+".html")
	f, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("cannot create report file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, report); err != nil {
		return "", fmt.Errorf("cannot render HTML report: %w", err)
	}

	return filePath, nil
}
