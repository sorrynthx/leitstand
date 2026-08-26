package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()
	if cfg.Telemetry.PollingInterval != 5*time.Second {
		t.Errorf("expected 5s polling interval, got %v", cfg.Telemetry.PollingInterval)
	}
	if cfg.TUI.MinCols != 98 || cfg.TUI.MinRows != 24 {
		t.Errorf("expected 98x24 resolution guard, got %dx%d", cfg.TUI.MinCols, cfg.TUI.MinRows)
	}
	if cfg.Database.Path == "" {
		t.Errorf("expected default database path not to be empty")
	}
}

func TestLoadCustomConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	customConfigPath := filepath.Join(tmpDir, "custom_config.yaml")

	yamlContent := `
database:
  path: "/tmp/custom.db"
telemetry:
  polling_interval: "2s"
  raw_retention_days: 14
ssh:
  timeout: "15s"
  keep_alive_secs: 60
tui:
  language: "ko"
  min_cols: 120
  min_rows: 35
`
	if err := os.WriteFile(customConfigPath, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("failed to write temporary config: %v", err)
	}

	cfg, err := Load(customConfigPath)
	if err != nil {
		t.Fatalf("failed to load custom config: %v", err)
	}

	if cfg.Database.Path != "/tmp/custom.db" {
		t.Errorf("expected /tmp/custom.db, got %s", cfg.Database.Path)
	}
	if cfg.Telemetry.PollingInterval != 2*time.Second {
		t.Errorf("expected 2s polling interval, got %v", cfg.Telemetry.PollingInterval)
	}
	if cfg.Telemetry.RawRetentionDays != 14 {
		t.Errorf("expected 14 days retention, got %d", cfg.Telemetry.RawRetentionDays)
	}
	if cfg.TUI.Language != "ko" {
		t.Errorf("expected 'ko' language, got %s", cfg.TUI.Language)
	}
	if cfg.TUI.MinCols != 120 || cfg.TUI.MinRows != 35 {
		t.Errorf("expected 120x35, got %dx%d", cfg.TUI.MinCols, cfg.TUI.MinRows)
	}
}

func TestLoadDefaultWhenFileMissing(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("expected no error when default config missing, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
}
