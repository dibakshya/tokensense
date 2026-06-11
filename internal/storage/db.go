package storage

import (
	"database/sql"
	_ "embed"
	"fmt"
	"sync"

	_ "modernc.org/sqlite"
)

//go:embed migrations/001_initial.sql
var migrationSQL string

// DB wraps the SQLite database connection.
type DB struct {
	conn *sql.DB
	mu   sync.RWMutex
}

// Open opens or creates the SQLite database at the given path with WAL mode.
func Open(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open database: %w", err)
	}

	conn.SetMaxOpenConns(1)

	// Enable WAL mode and set busy timeout
	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("cannot enable WAL mode: %w", err)
	}
	// NORMAL synchronous is recommended for WAL mode: safe after OS crash, fast.
	if _, err := conn.Exec("PRAGMA synchronous=NORMAL"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("cannot set synchronous mode: %w", err)
	}
	if _, err := conn.Exec("PRAGMA busy_timeout=5000"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("cannot set busy timeout: %w", err)
	}
	// Checkpoint after every 100 WAL pages (default 1000) so data survives
	// an unexpected proxy kill with minimal loss.
	if _, err := conn.Exec("PRAGMA wal_autocheckpoint=100"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("cannot set wal_autocheckpoint: %w", err)
	}

	if err := migrate(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return &DB{conn: conn}, nil
}

func migrate(conn *sql.DB) error {
	_, err := conn.Exec(migrationSQL)
	if err != nil {
		return fmt.Errorf("cannot execute migration: %w", err)
	}
	return nil
}

// Checkpoint flushes the WAL to the main database file so data survives a
// process restart. Call this before closing the DB in long-running processes.
func (db *DB) Checkpoint() {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.conn.Exec("PRAGMA wal_checkpoint(TRUNCATE)") //nolint:errcheck
}

// Close closes the database connection.
func (db *DB) Close() error {
	db.Checkpoint()
	return db.conn.Close()
}

// Conn returns the underlying sql.DB for advanced queries.
func (db *DB) Conn() *sql.DB {
	return db.conn
}
