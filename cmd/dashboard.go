package cmd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/dibakshya/tokensense/internal/config"
	"github.com/dibakshya/tokensense/internal/daemon"
	"github.com/dibakshya/tokensense/internal/reporter"
	"github.com/dibakshya/tokensense/internal/storage"
	"github.com/dibakshya/tokensense/internal/updater"
)

//go:embed dashboard.html
var dashboardHTML string

var dashPort int

func init() {
	dashboardCmd.Flags().IntVar(&dashPort, "port", 7892, "Port for the dashboard server (default 7892)")
	rootCmd.AddCommand(dashboardCmd)
}

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Open the browser control panel (easiest way to use Tokensense)",
	Long: `Opens a browser-based dashboard so you can control Tokensense
without typing any terminal commands.

The dashboard lets you:
  • Start and stop the proxy with one click
  • See today's AI cost breakdown and savings recommendations
  • Change privacy and report settings
  • View live stats as your AI tools run

It opens automatically in your browser. Press Ctrl+C to close it.`,
	RunE: runDashboard,
}

func runDashboard(cmd *cobra.Command, args []string) error {
	if err := config.LoadConfig(); err != nil {
		return err
	}

	addr := fmt.Sprintf("127.0.0.1:%d", dashPort)
	url := fmt.Sprintf("http://%s", addr)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleDashIndex)
	mux.HandleFunc("/api/status", withCORSDash(handleDashStatus))
	mux.HandleFunc("/api/report", withCORSDash(handleDashReport))
	mux.HandleFunc("/api/start", withCORSDash(handleDashStart))
	mux.HandleFunc("/api/stop", withCORSDash(handleDashStop))
	mux.HandleFunc("/api/config", withCORSDash(handleDashConfig))

	fmt.Println()
	fmt.Println(bold("  ╔══════════════════════════════════════════════════╗"))
	fmt.Println(bold("  ║") + green("  🌐  Tokensense Dashboard                     ") + bold("║"))
	fmt.Println(bold("  ╚══════════════════════════════════════════════════╝"))
	fmt.Println()
	fmt.Printf("  Opening:  %s\n", cyan(url))
	fmt.Println()
	fmt.Println(dim("  The dashboard stays open until you press Ctrl+C."))
	fmt.Println()

	openBrowser(url)
	return http.ListenAndServe(addr, mux)
}

// ── Page handler ──────────────────────────────────────────────────────────────

func handleDashIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, dashboardHTML)
}

// ── API: status ───────────────────────────────────────────────────────────────

func handleDashStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		dashJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := map[string]interface{}{
		"proxy_host":   config.GetString("proxy_host"),
		"proxy_port":   config.GetInt("proxy_port"),
		"privacy_mode": config.GetString("privacy_mode"),
		"generated_at": time.Now().Format(time.RFC3339),
	}

	svc := daemon.New()
	if _, err := svc.Status(); err != nil {
		resp["proxy_running"] = false
		resp["daemon_status"] = "not running"
	} else {
		resp["proxy_running"] = true
		resp["daemon_status"] = "running"
	}

	dbPath, err := config.DBPath()
	if err == nil {
		if db, err := storage.Open(dbPath); err == nil {
			defer db.Close()
			today := time.Now().Format("2006-01-02")
			count, _ := db.CountByDate(today)
			total, _ := db.TotalCount()
			resp["requests_today"] = count
			resp["requests_total"] = total
		}
	}

	dashJSONOK(w, resp)
}

// ── API: report ───────────────────────────────────────────────────────────────

func handleDashReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		dashJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	dbPath, err := config.DBPath()
	if err != nil {
		dashJSONError(w, "cannot find database", http.StatusInternalServerError)
		return
	}
	db, err := storage.Open(dbPath)
	if err != nil {
		dashJSONError(w, "cannot open database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	matrix, _ := updater.LoadCachedMatrix(BundledMatrix)
	report, err := reporter.GenerateReport(db, matrix, date)
	if err != nil {
		dashJSONError(w, "cannot generate report", http.StatusInternalServerError)
		return
	}

	dashJSONOK(w, report)
}

// ── API: start / stop proxy ───────────────────────────────────────────────────

func handleDashStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		dashJSONError(w, "use POST", http.StatusMethodNotAllowed)
		return
	}
	svc := daemon.New()
	if err := svc.Start(); err != nil {
		dashJSONError(w, "could not start proxy: "+err.Error(), http.StatusInternalServerError)
		return
	}
	dashJSONOK(w, map[string]string{"status": "started"})
}

func handleDashStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		dashJSONError(w, "use POST", http.StatusMethodNotAllowed)
		return
	}
	svc := daemon.New()
	if err := svc.Stop(); err != nil {
		dashJSONError(w, "could not stop proxy: "+err.Error(), http.StatusInternalServerError)
		return
	}
	dashJSONOK(w, map[string]string{"status": "stopped"})
}

// ── API: config ───────────────────────────────────────────────────────────────

func handleDashConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		dashJSONOK(w, map[string]interface{}{
			"proxy_port":   config.GetInt("proxy_port"),
			"proxy_host":   config.GetString("proxy_host"),
			"privacy_mode": config.GetString("privacy_mode"),
			"report_time":  config.GetString("report_time"),
		})

	case http.MethodPost:
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			dashJSONError(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		allowed := map[string]bool{"privacy_mode": true, "report_time": true}
		for k, v := range updates {
			if allowed[k] {
				config.Set(k, fmt.Sprintf("%v", v)) //nolint:errcheck
			}
		}
		if err := config.SaveConfig(); err != nil {
			dashJSONError(w, "could not save config: "+err.Error(), http.StatusInternalServerError)
			return
		}
		dashJSONOK(w, map[string]string{"status": "saved"})

	default:
		dashJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func withCORSDash(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h(w, r)
	}
}

func dashJSONOK(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(v) //nolint:errcheck
}

func dashJSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}
