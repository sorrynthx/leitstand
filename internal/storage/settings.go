package storage

import (
	"database/sql"
	"fmt"
)

// EnsureSettingsTable creates the settings table if it does not exist.
func (s *Storage) EnsureSettingsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS app_settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := s.db.Exec(query)
	return err
}

// GetSetting retrieves a setting value by key.
func (s *Storage) GetSetting(key string) (string, error) {
	_ = s.EnsureSettingsTable()
	var val string
	query := `SELECT value FROM app_settings WHERE key = ?;`
	err := s.db.QueryRow(query, key).Scan(&val)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get setting '%s': %w", key, err)
	}
	return val, nil
}

// SetSetting saves or updates a setting by key.
func (s *Storage) SetSetting(key, value string) error {
	_ = s.EnsureSettingsTable()
	query := `
	INSERT INTO app_settings (key, value, updated_at)
	VALUES (?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP;`
	_, err := s.db.Exec(query, key, value)
	if err != nil {
		return fmt.Errorf("failed to set setting '%s': %w", key, err)
	}
	return nil
}
