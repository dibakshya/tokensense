package storage

import "fmt"

// GetConfig reads a config key from the SQLite config table.
func (db *DB) GetConfig(key string) (string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var value string
	err := db.conn.QueryRow("SELECT value FROM config WHERE key = ?", key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

// SetConfig sets a config key-value pair in the SQLite config table.
func (db *DB) SetConfig(key, value string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.conn.Exec(`INSERT INTO config (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	if err != nil {
		return fmt.Errorf("cannot set config %s: %w", key, err)
	}
	return nil
}

// ListConfig returns all config key-value pairs.
func (db *DB) ListConfig() (map[string]string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	rows, err := db.conn.Query("SELECT key, value FROM config ORDER BY key")
	if err != nil {
		return nil, fmt.Errorf("cannot list config: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		result[k] = v
	}
	return result, nil
}
