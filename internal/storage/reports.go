package storage

import (
	"fmt"
	"time"
)

// DailyReport represents a daily summary report row.
type DailyReport struct {
	Date               string
	GeneratedAt        int64
	TotalRequests      int
	TotalTokensIn      int
	TotalTokensOut     int
	TotalCostUSD       float64
	OptimizedCostUSD   float64
	SavingsPotentialUSD float64
	ReportJSON         *string
}

// UpsertReport inserts or updates a daily report.
func (db *DB) UpsertReport(report *DailyReport) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if report.GeneratedAt == 0 {
		report.GeneratedAt = time.Now().UnixMilli()
	}

	_, err := db.conn.Exec(`INSERT INTO daily_reports 
		(date, generated_at, total_requests, total_tokens_in, total_tokens_out, 
		 total_cost_usd, optimized_cost_usd, savings_potential_usd, report_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(date) DO UPDATE SET
			generated_at = excluded.generated_at,
			total_requests = excluded.total_requests,
			total_tokens_in = excluded.total_tokens_in,
			total_tokens_out = excluded.total_tokens_out,
			total_cost_usd = excluded.total_cost_usd,
			optimized_cost_usd = excluded.optimized_cost_usd,
			savings_potential_usd = excluded.savings_potential_usd,
			report_json = excluded.report_json`,
		report.Date, report.GeneratedAt, report.TotalRequests,
		report.TotalTokensIn, report.TotalTokensOut,
		report.TotalCostUSD, report.OptimizedCostUSD,
		report.SavingsPotentialUSD, report.ReportJSON,
	)
	if err != nil {
		return fmt.Errorf("cannot upsert report: %w", err)
	}
	return nil
}

// GetReport retrieves a daily report by date.
func (db *DB) GetReport(date string) (*DailyReport, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var report DailyReport
	err := db.conn.QueryRow(`SELECT 
		date, generated_at, total_requests, total_tokens_in, total_tokens_out,
		total_cost_usd, optimized_cost_usd, savings_potential_usd, report_json
		FROM daily_reports WHERE date = ?`, date).Scan(
		&report.Date, &report.GeneratedAt, &report.TotalRequests,
		&report.TotalTokensIn, &report.TotalTokensOut,
		&report.TotalCostUSD, &report.OptimizedCostUSD,
		&report.SavingsPotentialUSD, &report.ReportJSON,
	)
	if err != nil {
		return nil, err
	}
	return &report, nil
}
