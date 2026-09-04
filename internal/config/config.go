package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// AppConfig represents the entire configuration for Leitstand.
type AppConfig struct {
	Database  DatabaseConfig  `mapstructure:"database"`
	Telemetry TelemetryConfig `mapstructure:"telemetry"`
	SSH       SSHConfig       `mapstructure:"ssh"`
	TUI       TUIConfig       `mapstructure:"tui"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	AI        AIConfig        `mapstructure:"ai"`
}

// AIConfig holds AI Copilot provider and retention settings.
type AIConfig struct {
	Provider      string `mapstructure:"provider"`
	Endpoint      string `mapstructure:"endpoint"`
	Model         string `mapstructure:"model"`
	RetentionDays int    `mapstructure:"retention_days"`
	MaxHistory    int    `mapstructure:"max_history"`
}

// LoggingConfig holds session audit logging directory settings.
type LoggingConfig struct {
	SessionLogDir string `mapstructure:"session_log_dir"`
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

// TelemetryConfig holds metric polling and retention settings.
type TelemetryConfig struct {
	PollingInterval  time.Duration `mapstructure:"polling_interval"`
	RawRetentionDays int           `mapstructure:"raw_retention_days"`
	CPUThreshold     float64       `mapstructure:"cpu_threshold"`
	RAMThreshold     float64       `mapstructure:"ram_threshold"`
	DiskThreshold    float64       `mapstructure:"disk_threshold"`
}

// SSHConfig holds global SSH defaults.
type SSHConfig struct {
	Timeout       time.Duration `mapstructure:"timeout"`
	KeepAliveSecs int           `mapstructure:"keep_alive_secs"`
}

// TUIConfig holds UI preferences and layout settings.
type TUIConfig struct {
	Language string `mapstructure:"language"`
	MinCols  int    `mapstructure:"min_cols"`
	MinRows  int    `mapstructure:"min_rows"`
}

// NewDefaultConfig returns an AppConfig populated with default values.
func NewDefaultConfig() *AppConfig {
	return &AppConfig{
		Database: DatabaseConfig{
			Path: DefaultDBPath(),
		},
		Telemetry: TelemetryConfig{
			PollingInterval:  DefaultPollingInterval,
			RawRetentionDays: 7,
			CPUThreshold:     85.0,
			RAMThreshold:     90.0,
			DiskThreshold:    90.0,
		},
		SSH: SSHConfig{
			Timeout:       DefaultSSHTimeout,
			KeepAliveSecs: 30,
		},
		TUI: TUIConfig{
			Language: DefaultLanguage,
			MinCols:  DefaultMinCols,
			MinRows:  DefaultMinRows,
		},
		Logging: LoggingConfig{
			SessionLogDir: DefaultSessionLogDir(),
		},
		AI: AIConfig{
			Provider:      "groq",
			Endpoint:      "https://api.groq.com/openai/v1",
			Model:         "llama-3.3-70b-versatile",
			RetentionDays: 3,
			MaxHistory:    20,
		},
	}
}

// Load loads configuration from custom path or standard search locations.
func Load(cfgFile string) (*AppConfig, error) {
	v := viper.New()

	defaults := NewDefaultConfig()
	v.SetDefault("database.path", defaults.Database.Path)
	v.SetDefault("telemetry.polling_interval", defaults.Telemetry.PollingInterval)
	v.SetDefault("telemetry.raw_retention_days", defaults.Telemetry.RawRetentionDays)
	v.SetDefault("ssh.timeout", defaults.SSH.Timeout)
	v.SetDefault("ssh.keep_alive_secs", defaults.SSH.KeepAliveSecs)
	v.SetDefault("tui.language", defaults.TUI.Language)
	v.SetDefault("tui.min_cols", defaults.TUI.MinCols)
	v.SetDefault("tui.min_rows", defaults.TUI.MinRows)
	v.SetDefault("logging.session_log_dir", defaults.Logging.SessionLogDir)
	v.SetDefault("ai.provider", defaults.AI.Provider)
	v.SetDefault("ai.endpoint", defaults.AI.Endpoint)
	v.SetDefault("ai.model", defaults.AI.Model)
	v.SetDefault("ai.retention_days", defaults.AI.RetentionDays)
	v.SetDefault("ai.max_history", defaults.AI.MaxHistory)

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName(DefaultConfigFileName)
		v.SetConfigType(DefaultConfigFileType)
		v.AddConfigPath(DefaultDataDir())
		v.AddConfigPath(".")
	}

	v.SetEnvPrefix("LEITSTAND")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Return error only if it's not a missing config file (missing is ok, we use defaults)
			if !os.IsNotExist(err) && cfgFile != "" {
				return nil, fmt.Errorf("failed to read config file %q: %w", cfgFile, err)
			}
		}
	}

	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return &cfg, nil
}

// EnsureDataDirExists creates the data directory if it does not exist.
func EnsureDataDirExists(dirPath string) error {
	if dirPath == "" {
		dirPath = DefaultDataDir()
	}
	return os.MkdirAll(dirPath, 0700)
}

// EnsureParentDirExists ensures the parent folder for a file exists.
func EnsureParentDirExists(filePath string) error {
	dir := filepath.Dir(filePath)
	return os.MkdirAll(dir, 0700)
}
