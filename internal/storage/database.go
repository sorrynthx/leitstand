package storage

import (
	"database/sql"
	"embed"
	"fmt"
	"leitstand/internal/config"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Storage wraps the SQLite database connection and operations.
type Storage struct {
	db *sql.DB
}

// Open initializes or connects to SQLite at the given path with WAL optimizations.
func Open(dbPath string) (*Storage, error) {
	if err := config.EnsureParentDirExists(dbPath); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	// Connect to SQLite
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database at %q: %w", dbPath, err)
	}

	// Apply recommended PRAGMA optimizations from architecture plan
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA cache_size = -64000;", // 64MB cache
		"PRAGMA temp_store = MEMORY;",
		"PRAGMA foreign_keys = ON;",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to execute %q: %w", pragma, err)
		}
	}

	s := &Storage{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	return s, nil
}

// Close closes the underlying database connection.
func (s *Storage) Close() error {
	return s.db.Close()
}

// DB returns the raw *sql.DB for advanced operations.
func (s *Storage) DB() *sql.DB {
	return s.db
}

func (s *Storage) migrate() error {
	// Read embedded migration file
	migrationSQL, err := migrationFS.ReadFile("migrations/001_initial.sql")
	if err != nil {
		return fmt.Errorf("failed to read embedded migration: %w", err)
	}

	if _, err := s.db.Exec(string(migrationSQL)); err != nil {
		return fmt.Errorf("failed to execute initial migration: %w", err)
	}

	return nil
}
