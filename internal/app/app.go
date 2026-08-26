package app

import (
	"fmt"
	"leitstand/internal/config"
	"leitstand/internal/ssh"
	"leitstand/internal/storage"
	"leitstand/internal/telemetry"
	"leitstand/internal/tui"
	"leitstand/internal/vault"

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
	store, err := storage.Open(cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
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
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui runtime error: %w", err)
	}

	return nil
}
