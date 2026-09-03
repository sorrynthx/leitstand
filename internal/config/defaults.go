package config

import (
	"os"
	"path/filepath"
	"time"
)

const (
	DefaultConfigFileName = "config"
	DefaultConfigFileType = "yaml"
	DefaultAppDirName     = ".leitstand"
	DefaultDBFileName     = "leitstand.db"

	DefaultPollingInterval = 5 * time.Second
	DefaultSSHTimeout      = 10 * time.Second
	DefaultMinCols         = 98
	DefaultMinRows         = 24
	DefaultLanguage        = "en"
)

// DefaultDataDir returns the platform-appropriate default directory for leitstand data.
func DefaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", DefaultAppDirName)
	}
	return filepath.Join(home, DefaultAppDirName)
}

// DefaultDBPath returns the full path to the default SQLite database file.
func DefaultDBPath() string {
	return filepath.Join(DefaultDataDir(), DefaultDBFileName)
}

// DefaultSessionLogDir returns the platform-appropriate directory for saving session audit logs.
func DefaultSessionLogDir() string {
	home, err := os.UserHomeDir()
	if err == nil {
		return filepath.Join(home, "Documents", "leitstand", "logs")
	}
	return filepath.Join(".", "logs")
}
