package tui

import (
	"context"
	"leitstand/internal/config"
	"leitstand/internal/storage"
	"leitstand/internal/telemetry"
	"leitstand/internal/vault"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// MetricResultMsg delivers telemetry polling results to the Bubbletea update loop.
type MetricResultMsg struct {
	HostID  int64
	Record  *storage.MetricRecord
	SysInfo *telemetry.SysInfo
	Err     error
}

// TickMsg triggers periodic UI refresh / polling.
type TickMsg time.Time

type PaneType int

const (
	PaneHostList PaneType = iota
	PaneTelemetryDeck
	PaneConsole
)

// CmdResultMsg carries remote command execution results back to TUI.
type CmdResultMsg struct {
	HostID  int64
	Command string
	Stdout  string
	Stderr  string
	Err     error
}

// Model is the main Bubbletea TUI model.
type Model struct {
	cfg               *config.AppConfig
	store             *storage.Storage
	vault             *vault.Vault
	collector         *telemetry.Collector
	hosts             []*storage.Host
	selectedIndex     int
	metrics           map[int64]*storage.MetricRecord
	sysInfos          map[int64]*telemetry.SysInfo
	errors            map[int64]error
	width             int
	height            int
	statusMessage     string
	isDemo            bool
	activePane        PaneType
	showAddModal      bool
	addForm           *HostForm
	showVaultModal    bool
	vaultForm         *VaultForm
	showDeleteModal   bool
	hostToDelete      *storage.Host
	consoleInput      textinput.Model
	consoleLogs       map[int64][]string
	viewport          viewport.Model
	viewportReady     bool
	fullScreenConsole bool
	ctx               context.Context
	cancel            context.CancelFunc
}

// NewModel creates an initialized Bubbletea Model.
func NewModel(cfg *config.AppConfig, store *storage.Storage, v *vault.Vault, col *telemetry.Collector, isDemo bool) *Model {
	ctx, cancel := context.WithCancel(context.Background())

	showVault := false
	var vForm *VaultForm
	if !isDemo && store != nil && v != nil && !v.IsUnlocked() {
		showVault = true
		isInit, _ := store.IsVaultInitialized()
		if !isInit {
			vForm = NewVaultForm(VaultModalInit)
		} else {
			vForm = NewVaultForm(VaultModalUnlock)
		}
	}

	ci := textinput.New()
	ci.Placeholder = "Type remote command (e.g. uname -a, docker ps, free -m, top -b -n 1) and press Enter..."
	ci.Prompt = "❯ "
	ci.Width = 60

	return &Model{
		cfg:            cfg,
		store:          store,
		vault:          v,
		collector:      col,
		hosts:          make([]*storage.Host, 0),
		selectedIndex:  0,
		metrics:        make(map[int64]*storage.MetricRecord),
		sysInfos:       make(map[int64]*telemetry.SysInfo),
		errors:         make(map[int64]error),
		statusMessage:  "Press [Tab] to switch panel, [a] Add host, [d] Delete, [r] Refresh",
		isDemo:         isDemo,
		activePane:     PaneHostList,
		showVaultModal: showVault,
		vaultForm:      vForm,
		consoleInput:   ci,
		consoleLogs:    make(map[int64][]string),
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Init starts the Bubbletea program lifecycle.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadHostsCmd(),
		m.tickCmd(),
	)
}

func (m *Model) tickCmd() tea.Cmd {
	interval := m.cfg.Telemetry.PollingInterval
	if interval <= 0 {
		interval = 3 * time.Second
	}
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m *Model) loadHostsCmd() tea.Cmd {
	return func() tea.Msg {
		if m.isDemo {
			return []*storage.Host{
				{ID: 101, Name: "prod-api-01", Address: "10.0.1.11", Port: 22, Username: "ubuntu", GroupName: "Production (EU)"},
				{ID: 102, Name: "prod-api-02", Address: "10.0.1.12", Port: 22, Username: "ubuntu", GroupName: "Production (EU)"},
				{ID: 103, Name: "prod-db-master", Address: "10.0.2.10", Port: 22, Username: "postgres", GroupName: "Databases"},
				{ID: 104, Name: "stage-k8s-node1", Address: "192.168.10.51", Port: 22, Username: "admin", GroupName: "Staging Cluster"},
				{ID: 105, Name: "stage-k8s-node2", Address: "192.168.10.52", Port: 22, Username: "admin", GroupName: "Staging Cluster"},
			}
		}

		if m.store == nil {
			return nil
		}
		hosts, err := m.store.ListHosts()
		if err != nil {
			return nil
		}
		return hosts
	}
}
