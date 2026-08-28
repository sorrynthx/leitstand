package tui

import (
	"context"
	"leitstand/internal/config"
	"leitstand/internal/i18n"
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
	TabID   string
	Command string
	CWD     string
	NewCWD  string
	Stdout  string
	Stderr  string
	Err     error
}

// StreamChunkMsg delivers a real-time streaming output chunk to a specific tab.
type StreamChunkMsg struct {
	HostID  int64
	TabID   string
	Chunk   string
	msgChan <-chan tea.Msg
}

// StreamFinishedMsg signals that a background streaming task has finished or stopped.
type StreamFinishedMsg struct {
	HostID int64
	TabID  string
	Err    error
}

// TerminalExitedMsg indicates when an interactive SSH PTY session exits.
type TerminalExitedMsg struct {
	HostID int64
	Err    error
}

// TabCompletionMsg delivers remote tab completion results.
type TabCompletionMsg struct {
	HostID        int64
	OriginalInput string
	NewInput      string
	Candidates    []string
	Err           error
}

// OpenFileMsg opens the remote file in the in-app editor modal.
type OpenFileMsg struct {
	HostID   int64
	HostName string
	FilePath string
	Content  string
	Err      error
}

// FileSavedMsg indicates when remote file save finishes.
type FileSavedMsg struct {
	HostID   int64
	FilePath string
	Err      error
}

type HostConnStatus int

const (
	HostStatusUnknown HostConnStatus = iota
	HostStatusConnecting
	HostStatusOnline
	HostStatusOffline
)

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
	hostStatus        map[int64]HostConnStatus
	pausedHosts       map[int64]bool
	width             int
	height            int
	statusMessage     string
	isDemo            bool
	activePane        PaneType
	showAddModal      bool
	addForm           *HostForm
	showEditModal     bool
	editForm          *HostForm
	hostToEdit        *storage.Host
	showVaultModal    bool
	vaultForm         *VaultForm
	showDeleteModal   bool
	hostToDelete      *storage.Host
	showEditorModal   bool
	editorModal       *EditorModal
	showSudoModal     bool
	sudoModal         *SudoModal
	pendingSudoCmd    string
	sudoCache         map[int64]string
	showSettingsModal bool
	settingsModal     *SettingsModal
	showFileManager   bool
	fileManager       *FileManagerModal
	showDrawer        bool
	drawer            *RunbookDrawer
	consoleInput      textinput.Model
	hostTabs          map[int64]*HostTabState
	consoleLogs       map[int64][]string
	hostCWD           map[int64]string
	cmdHistory        []string
	historyIndex      int
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

		// Load saved language if available
		if savedLang, err := store.GetSetting("language"); err == nil && savedLang != "" {
			i18n.SetLang(i18n.Lang(savedLang))
		}

		// Load saved telemetry polling interval if available
		if savedInterval, err := store.GetSetting("polling_interval"); err == nil && savedInterval != "" {
			if savedInterval == "0" || savedInterval == "0s" {
				cfg.Telemetry.PollingInterval = 0
			} else if d, err := time.ParseDuration(savedInterval); err == nil {
				cfg.Telemetry.PollingInterval = d
			}
		}
	}

	ci := textinput.New()
	ci.Placeholder = "Type command (e.g. docker ps, ls -la, df -h)..."
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
		hostStatus:     make(map[int64]HostConnStatus),
		pausedHosts:    make(map[int64]bool),
		statusMessage:  "Press [r] Reconnect, [Tab/c] Console, [p] Settings, [a] Add Host",
		isDemo:         isDemo,
		activePane:     PaneHostList,
		showVaultModal: showVault,
		vaultForm:      vForm,
		consoleInput:   ci,
		hostTabs:       make(map[int64]*HostTabState),
		consoleLogs:    make(map[int64][]string),
		hostCWD:        make(map[int64]string),
		sudoCache:      make(map[int64]string),
		cmdHistory:     make([]string, 0),
		historyIndex:   -1,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// GetOrCreateHostTabs gets or initializes the multi-tab state for a host.
func (m *Model) GetOrCreateHostTabs(hostID int64, hostName string) *HostTabState {
	if m.hostTabs == nil {
		m.hostTabs = make(map[int64]*HostTabState)
	}
	hts, ok := m.hostTabs[hostID]
	if !ok {
		cwd := m.hostCWD[hostID]
		if cwd == "" {
			cwd = "~"
		}
		hts = NewHostTabState(hostID, hostName, cwd)
		m.hostTabs[hostID] = hts
	}
	return hts
}

// CurrentActiveTab returns the focused tab for the currently selected host.
func (m *Model) CurrentActiveTab() *ConsoleTab {
	if len(m.hosts) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(m.hosts) {
		return nil
	}
	curHost := m.hosts[m.selectedIndex]
	hts := m.GetOrCreateHostTabs(curHost.ID, curHost.Name)
	return hts.ActiveTab()
}

// Init starts the Bubbletea program lifecycle.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadHostsCmd(),
		m.tickCmd(),
	)
}
