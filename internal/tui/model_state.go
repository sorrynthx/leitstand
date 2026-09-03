package tui

import (
	"context"
	"leitstand/internal/config"
	"leitstand/internal/i18n"
	"leitstand/internal/ssh"
	"leitstand/internal/storage"
	"leitstand/internal/telemetry"
	"leitstand/internal/vault"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

type Config = config.AppConfig
type Store = storage.Storage

type Pane int

const (
	PaneHostList Pane = iota
	PaneConsole
	PaneTelemetryDeck
)

type HostStatus int

const (
	HostStatusOffline HostStatus = iota
	HostStatusConnecting
	HostStatusOnline
)

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

type Model struct {
	ctx           context.Context
	cancel        context.CancelFunc
	cfg           *config.AppConfig
	store         *storage.Storage
	vault         *vault.Vault
	collector     *telemetry.Collector
	sshPool       *ssh.Pool
	hosts         []*storage.Host
	hostStatus    map[int64]HostStatus
	metrics       map[int64]*storage.MetricRecord
	sysInfos      map[int64]*telemetry.SysInfo
	errors        map[int64]error
	pausedHosts   map[int64]bool
	hostTabs      map[int64]*HostTabs
	sudoCache     map[int64]string
	sudoModeCache map[int64]int
	lastPollTime  time.Time

	activePane        Pane
	selectedIndex     int
	userHasNavigated  bool
	consoleInput      textinput.Model
	viewport          viewport.Model
	viewportReady     bool
	statusMessage     string
	fullScreenConsole bool
	width             int
	height            int

	showAddModal      bool
	addForm           *HostForm
	showEditModal     bool
	editForm          *HostForm
	hostToEdit        *storage.Host
	hostToDelete      *storage.Host
	showDeleteModal   bool
	showFormModal     bool
	formModal         *HostForm
	showVaultModal    bool
	vaultModal        *VaultModal
	vaultForm         *VaultModal
	showSudoModal     bool
	sudoModal         *SudoModal
	pendingSudoCmd    string
	showFileManager   bool
	fileManager       *FileManagerModal
	showEditorModal   bool
	editorModal       *EditorModal
	showSettingsModal bool
	settingsModal     *SettingsModal

	showDrawer          bool
	drawer              *RunbookDrawer
	showTelemetryDrawer bool
	isDemo              bool
}

func NewModel(c *config.AppConfig, s *storage.Storage, v *vault.Vault, collector *telemetry.Collector, isDemo bool) *Model {
	ctx, cancel := context.WithCancel(context.Background())

	if s != nil {
		if savedLang, err := s.GetSetting("language"); err == nil && savedLang != "" {
			i18n.SetLang(i18n.Lang(savedLang))
		}
		if savedInterval, err := s.GetSetting("polling_interval"); err == nil && savedInterval != "" {
			if dur, parseErr := time.ParseDuration(savedInterval); parseErr == nil && c != nil {
				c.Telemetry.PollingInterval = dur
			}
		}
		if savedLogDir, err := s.GetSetting("session_log_dir"); err == nil && savedLogDir != "" && c != nil {
			c.Logging.SessionLogDir = savedLogDir
		}
	}

	cInput := textinput.New()
	cInput.Prompt = "$ "
	cInput.Placeholder = "Type command and press Enter..."

	vp := viewport.New(80, 20)
	pool := ssh.NewPool(10 * time.Second)

	if collector == nil && s != nil && v != nil {
		collector = telemetry.NewCollector(s, pool, v)
	}

	var showVault bool
	var vModal *VaultModal
	if !isDemo && s != nil && v != nil && !v.IsUnlocked() {
		showVault = true
		hasVault, _ := s.IsVaultInitialized()
		if hasVault {
			vModal = NewVaultForm(VaultModalUnlock)
		} else {
			vModal = NewVaultForm(VaultModalInit)
		}
	}

	return &Model{
		ctx:               ctx,
		cancel:            cancel,
		cfg:               c,
		store:             s,
		vault:             v,
		collector:         collector,
		sshPool:           pool,
		hostStatus:        make(map[int64]HostStatus),
		metrics:           make(map[int64]*storage.MetricRecord),
		sysInfos:          make(map[int64]*telemetry.SysInfo),
		errors:            make(map[int64]error),
		pausedHosts:       make(map[int64]bool),
		hostTabs:          make(map[int64]*HostTabs),
		sudoCache:         make(map[int64]string),
		sudoModeCache:     make(map[int64]int),
		activePane:        PaneHostList,
		selectedIndex:     0,
		userHasNavigated:  false,
		consoleInput:      cInput,
		viewport:          vp,
		statusMessage:     "Press [Tab] to switch panes, [Enter] to select/connect server.",
		fullScreenConsole: false,
		showVaultModal:    showVault,
		vaultModal:        vModal,
		isDemo:            isDemo,
	}
}

func (m *Model) Close() {
	if m.cancel != nil {
		m.cancel()
	}
	if m.sshPool != nil {
		m.sshPool.CloseAll()
	}
}
