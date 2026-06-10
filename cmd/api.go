package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/dibakshya/tokensense/internal/classifier"
	"github.com/dibakshya/tokensense/internal/config"
	"github.com/dibakshya/tokensense/internal/daemon"
	"github.com/dibakshya/tokensense/internal/reporter"
	"github.com/dibakshya/tokensense/internal/storage"
	"github.com/dibakshya/tokensense/internal/updater"
)

var apiPort int

func init() {
	apiCmd.Flags().IntVar(&apiPort, "port", 7891, "Port to listen on (default 7891)")
	rootCmd.AddCommand(apiCmd)
}

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the local JSON API for developer and agent integration",
	Long: `Starts a lightweight HTTP server on localhost that exposes Tokensense
data as JSON — perfect for integrating into AI tools, agents, dashboards,
or any custom automation.

Endpoints:
  GET  /v1/status            Proxy status + today's request count
  GET  /v1/report?date=...   Daily cost & savings report (JSON)
  POST /v1/classify          Classify a prompt and get model recommendations
  GET  /v1/usage?limit=N     Raw usage records (newest first)
  GET  /v1/docs              This help page

Example (agent integration):
  curl http://localhost:7891/v1/status
  curl http://localhost:7891/v1/report
  curl -X POST http://localhost:7891/v1/classify \
       -H "Content-Type: application/json" \
       -d '{"prompt":"write unit tests for my auth module"}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.LoadConfig(); err != nil {
			return err
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/v1/status", withCORS(handleAPIStatus))
		mux.HandleFunc("/v1/report", withCORS(handleAPIReport))
		mux.HandleFunc("/v1/classify", withCORS(handleAPIClassify))
		mux.HandleFunc("/v1/usage", withCORS(handleAPIUsage))
		mux.HandleFunc("/v1/docs", withCORS(handleAPIDocs))
		mux.HandleFunc("/", withCORS(handleAPIRoot))

		addr := fmt.Sprintf("127.0.0.1:%d", apiPort)

		fmt.Println()
		fmt.Println(bold("  ╔════════════════════════════════════════════════╗"))
		fmt.Println(bold("  ║") + green("  🔌  Tokensense API running                   ") + bold("║"))
		fmt.Println(bold("  ╚════════════════════════════════════════════════╝"))
		fmt.Println()
		fmt.Printf("  Listening on:  %s\n", cyan("http://"+addr))
		fmt.Println()
		fmt.Println(bold("  Available endpoints:"))
		fmt.Printf("  %-42s %s\n", cyan("GET  /v1/status"), dim("Proxy status + today's stats"))
		fmt.Printf("  %-42s %s\n", cyan("GET  /v1/report"), dim("Daily cost & savings report"))
		fmt.Printf("  %-42s %s\n", cyan("POST /v1/classify"), dim("Classify a prompt → model recommendation"))
		fmt.Printf("  %-42s %s\n", cyan("GET  /v1/usage?limit=N"), dim("Raw usage records (newest first)"))
		fmt.Printf("  %-42s %s\n", cyan("GET  /v1/docs"), dim("Full API documentation"))
		fmt.Println()
		fmt.Println(dim("  Press Ctrl+C to stop."))
		fmt.Println()

		return http.ListenAndServe(addr, mux)
	},
}

// ── Handlers ─────────────────────────────────────────────────────────────────

func handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := map[string]interface{}{
		"proxy_host":    config.GetString("proxy_host"),
		"proxy_port":    config.GetInt("proxy_port"),
		"privacy_mode":  config.GetString("privacy_mode"),
		"generated_at":  time.Now().Format(time.RFC3339),
	}

	// Daemon status
	svc := daemon.New()
	status, err := svc.Status()
	if err != nil {
		resp["proxy_running"] = false
		resp["daemon_status"] = "not running"
	} else {
		resp["proxy_running"] = true
		resp["daemon_status"] = status
	}

	// DB stats
	dbPath, err := config.DBPath()
	if err == nil {
		db, err := storage.Open(dbPath)
		if err == nil {
			defer db.Close()
			today := time.Now().Format("2006-01-02")
			count, _ := db.CountByDate(today)
			total, _ := db.TotalCount()
			resp["requests_today"] = count
			resp["requests_total"] = total
		}
	}

	jsonOK(w, resp)
}

func handleAPIReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	dbPath, err := config.DBPath()
	if err != nil {
		jsonError(w, "cannot find database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	db, err := storage.Open(dbPath)
	if err != nil {
		jsonError(w, "cannot open database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	matrix, _ := updater.LoadCachedMatrix(BundledMatrix)
	report, err := reporter.GenerateReport(db, matrix, date)
	if err != nil {
		jsonError(w, "cannot generate report: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, report)
}

func handleAPIClassify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "use POST with JSON body: {\"prompt\": \"your text\"}", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Prompt) == "" {
		jsonError(w, "request body must be JSON with a non-empty \"prompt\" field", http.StatusBadRequest)
		return
	}

	c := classifier.NewRuleBasedClassifier()
	result, err := c.Classify(req.Prompt)
	if err != nil {
		jsonError(w, "classification failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Model recommendations
	matrix, _ := updater.LoadCachedMatrix(BundledMatrix)
	var recommendations []map[string]interface{}
	if matrix != nil {
		recs := matrix.RankModels(result.TaskType, result.Complexity)
		for i, rec := range recs {
			if i >= 5 {
				break
			}
			recommendations = append(recommendations, map[string]interface{}{
				"model":                  rec.Model.ID,
				"provider":               rec.Model.Provider,
				"display_name":           rec.Model.DisplayName,
				"quality_score":          rec.Quality,
				"cost_per_request_usd":   rec.CostPerRequest,
				"is_recommended":         rec.IsRecommended,
				"reason":                 rec.Reason,
				"input_per_1m_usd":       rec.Model.Pricing.InputPer1MUSD,
				"output_per_1m_usd":      rec.Model.Pricing.OutputPer1MUSD,
			})
		}
	}

	jsonOK(w, map[string]interface{}{
		"task_type":       result.TaskType,
		"complexity":      result.Complexity,
		"confidence":      result.Confidence,
		"source":          result.Source,
		"recommendations": recommendations,
	})
}

func handleAPIUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 && n <= 1000 {
			limit = n
		}
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	dbPath, err := config.DBPath()
	if err != nil {
		jsonError(w, "cannot find database", http.StatusInternalServerError)
		return
	}
	db, err := storage.Open(dbPath)
	if err != nil {
		jsonError(w, "cannot open database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	records, err := db.QueryByDate(date)
	if err != nil {
		jsonError(w, "cannot query records", http.StatusInternalServerError)
		return
	}

	// Cap to limit (records are newest-first from DB)
	if len(records) > limit {
		records = records[:limit]
	}

	jsonOK(w, map[string]interface{}{
		"date":    date,
		"limit":   limit,
		"count":   len(records),
		"records": records,
	})
}

func handleAPIDocs(w http.ResponseWriter, r *http.Request) {
	docs := map[string]interface{}{
		"name":    "Tokensense Local API",
		"version": "v1",
		"base_url": fmt.Sprintf("http://127.0.0.1:%d", apiPort),
		"description": "Local JSON API for integrating Tokensense into AI tools, agents, and dashboards.",
		"endpoints": []map[string]interface{}{
			{
				"method":      "GET",
				"path":        "/v1/status",
				"description": "Proxy status, today's request count, and configuration",
				"example":     "curl http://localhost:7891/v1/status",
			},
			{
				"method":      "GET",
				"path":        "/v1/report",
				"description": "Daily cost & savings report as JSON",
				"query_params": map[string]string{
					"date": "YYYY-MM-DD (default: today)",
				},
				"example": "curl http://localhost:7891/v1/report?date=2025-01-15",
			},
			{
				"method":      "POST",
				"path":        "/v1/classify",
				"description": "Classify a prompt and get ranked model recommendations",
				"body":        map[string]string{"prompt": "string — the text to classify"},
				"example":     `curl -X POST http://localhost:7891/v1/classify -H "Content-Type: application/json" -d '{"prompt":"write unit tests"}'`,
			},
			{
				"method":      "GET",
				"path":        "/v1/usage",
				"description": "Raw usage records for a given date",
				"query_params": map[string]string{
					"date":  "YYYY-MM-DD (default: today)",
					"limit": "max records to return (default: 100, max: 1000)",
				},
				"example": "curl http://localhost:7891/v1/usage?limit=50",
			},
		},
		"agent_integration_example": map[string]interface{}{
			"description": "How to use this API inside an AI agent",
			"python_snippet": strings.TrimSpace(`
import requests

# Get today's cost report
report = requests.get("http://localhost:7891/v1/report").json()
print(f"Spent today: ${report['total_cost_usd']:.4f}")
print(f"Potential savings: ${report['savings_potential_usd']:.4f}")

# Classify a task before sending to an expensive model
result = requests.post("http://localhost:7891/v1/classify",
    json={"prompt": "write unit tests for my auth module"}).json()
print(f"Task: {result['task_type']}, Complexity: {result['complexity']}")
best = result['recommendations'][0]
print(f"Recommended model: {best['display_name']}")`),
		},
	}
	jsonOK(w, docs)
}

func handleAPIRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "" {
		http.Redirect(w, r, "/v1/docs", http.StatusMovedPermanently)
		return
	}
	jsonError(w, fmt.Sprintf("endpoint not found: %s — see /v1/docs", r.URL.Path), http.StatusNotFound)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h(w, r)
	}
}

func jsonOK(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(v) //nolint:errcheck
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}
