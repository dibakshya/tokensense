package storage

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RequestMetadata represents a classified API request metadata row.
type RequestMetadata struct {
	ID                   string
	Timestamp            int64
	DayDate              string
	Provider             string
	Model                string
	TaskType             *string
	Complexity           *string
	TokensIn             *int
	TokensOut            *int
	CostUSD              *float64
	LatencyMs            *int
	ContentMode          int
	ClassifierSource     *string
	ClassifierConfidence *float64
	ToolSource           *string
	Intercepted          int
}

// Insert stores a new request metadata row.
func (db *DB) Insert(req *RequestMetadata) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if req.ID == "" {
		req.ID = uuid.New().String()
	}
	if req.Timestamp == 0 {
		req.Timestamp = time.Now().UnixMilli()
	}
	if req.DayDate == "" {
		req.DayDate = time.Now().Format("2006-01-02")
	}

	_, err := db.conn.Exec(`INSERT INTO requests 
		(id, timestamp, day_date, provider, model, task_type, complexity, 
		 tokens_in, tokens_out, cost_usd, latency_ms, content_mode, 
		 classifier_source, classifier_confidence, tool_source, intercepted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.ID, req.Timestamp, req.DayDate, req.Provider, req.Model,
		req.TaskType, req.Complexity, req.TokensIn, req.TokensOut,
		req.CostUSD, req.LatencyMs, req.ContentMode,
		req.ClassifierSource, req.ClassifierConfidence,
		req.ToolSource, req.Intercepted,
	)
	if err != nil {
		return fmt.Errorf("cannot insert request: %w", err)
	}
	return nil
}

// QueryByDate returns all request metadata for a given date (YYYY-MM-DD).
func (db *DB) QueryByDate(date string) ([]RequestMetadata, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	rows, err := db.conn.Query(`SELECT 
		id, timestamp, day_date, provider, model, task_type, complexity,
		tokens_in, tokens_out, cost_usd, latency_ms, content_mode,
		classifier_source, classifier_confidence, tool_source, intercepted
		FROM requests WHERE day_date = ? ORDER BY timestamp`, date)
	if err != nil {
		return nil, fmt.Errorf("cannot query requests: %w", err)
	}
	defer rows.Close()

	return scanRequests(rows)
}

// QueryByDateRange returns all request metadata between startDate and endDate (inclusive).
func (db *DB) QueryByDateRange(startDate, endDate string) ([]RequestMetadata, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	rows, err := db.conn.Query(`SELECT 
		id, timestamp, day_date, provider, model, task_type, complexity,
		tokens_in, tokens_out, cost_usd, latency_ms, content_mode,
		classifier_source, classifier_confidence, tool_source, intercepted
		FROM requests WHERE day_date >= ? AND day_date <= ? ORDER BY timestamp`,
		startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("cannot query requests: %w", err)
	}
	defer rows.Close()

	return scanRequests(rows)
}

// CountByDate returns the number of requests for a given date.
func (db *DB) CountByDate(date string) (int, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM requests WHERE day_date = ?", date).Scan(&count)
	return count, err
}

// TotalCount returns the total number of stored requests.
func (db *DB) TotalCount() (int, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM requests").Scan(&count)
	return count, err
}

func scanRequests(rows interface {
	Next() bool
	Scan(dest ...interface{}) error
}) ([]RequestMetadata, error) {
	var results []RequestMetadata
	for rows.Next() {
		var req RequestMetadata
		err := rows.Scan(
			&req.ID, &req.Timestamp, &req.DayDate, &req.Provider, &req.Model,
			&req.TaskType, &req.Complexity, &req.TokensIn, &req.TokensOut,
			&req.CostUSD, &req.LatencyMs, &req.ContentMode,
			&req.ClassifierSource, &req.ClassifierConfidence,
			&req.ToolSource, &req.Intercepted,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot scan request row: %w", err)
		}
		results = append(results, req)
	}
	return results, nil
}
