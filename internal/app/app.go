package app

import (
	"fmt"
	"leitstand/internal/config"
	"leitstand/internal/i18n"
	"leitstand/internal/logger"
	"leitstand/internal/ssh"
	"leitstand/internal/storage"
	"leitstand/internal/telemetry"
	"leitstand/internal/tui"
	"leitstand/internal/vault"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// App manages the entire Leitstand lifecycle.
type App struct {
	cfg       *config.AppConfig
	store     *storage.Storage
	vault     *vault.Vault
	pool      *ssh.Pool
	collector *telemetry.Collector
}

// New creates and wires all subsystems together.
func New(cfg *config.AppConfig) (*App, error) {
	_ = logger.Init("leitstand.log")
	logger.Infof("Initializing Leitstand application...")

	store, err := storage.Open(cfg.Database.Path)
	if err != nil {
		logger.Errorf("Failed to open database: %v", err)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if savedLang, err := store.GetSetting("language"); err == nil && savedLang != "" {
		i18n.SetLang(i18n.Lang(savedLang))
	}
	if savedInterval, err := store.GetSetting("polling_interval"); err == nil && savedInterval != "" {
		if dur, parseErr := time.ParseDuration(savedInterval); parseErr == nil {
			cfg.Telemetry.PollingInterval = dur
		}
	}
	if savedCPU, err := store.GetSetting("cpu_threshold"); err == nil && savedCPU != "" {
		if val, parseErr := strconv.ParseFloat(savedCPU, 64); parseErr == nil && val > 0 {
			cfg.Telemetry.CPUThreshold = val
		}
	}
	if savedRAM, err := store.GetSetting("ram_threshold"); err == nil && savedRAM != "" {
		if val, parseErr := strconv.ParseFloat(savedRAM, 64); parseErr == nil && val > 0 {
			cfg.Telemetry.RAMThreshold = val
		}
	}
	if savedDisk, err := store.GetSetting("disk_threshold"); err == nil && savedDisk != "" {
		if val, parseErr := strconv.ParseFloat(savedDisk, 64); parseErr == nil && val > 0 {
			cfg.Telemetry.DiskThreshold = val
		}
	}
	if savedLogDir, err := store.GetSetting("session_log_dir"); err == nil && savedLogDir != "" {
		cfg.Logging.SessionLogDir = savedLogDir
	}

	v := vault.New()
	pool := ssh.NewPool(cfg.SSH.Timeout)
	collector := telemetry.NewCollector(store, pool, v)

	return &App{
		cfg:       cfg,
		store:     store,
		vault:     v,
		pool:      pool,
		collector: collector,
	}, nil
}

// Close gracefully terminates open resources.
func (a *App) Close() {
	if a.pool != nil {
		a.pool.CloseAll()
	}
	if a.vault != nil {
		a.vault.Lock()
	}
	if a.store != nil {
		a.store.Close()
	}
}

// Store returns the underlying storage.
func (a *App) Store() *storage.Storage {
	return a.store
}

// Vault returns the app vault.
func (a *App) Vault() *vault.Vault {
	return a.vault
}

// RunCockpit starts the interactive Bubbletea TUI.
func (a *App) RunCockpit(isDemo bool) error {
	model := tui.NewModel(a.cfg, a.store, a.vault, a.collector, isDemo)
	defer model.Close()
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui runtime error: %w", err)
	}

	return nil
}
