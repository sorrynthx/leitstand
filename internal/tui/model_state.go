package tui

import (
	"context"
	"leitstand/internal/config"
	"leitstand/internal/i18n"
	"leitstand/internal/ssh"
	"leitstand/internal/storage"
	"leitstand/internal/telemetry"
	"leitstand/internal/vault"
	"strconv"
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

	tunnelMgr       *ssh.TunnelManager
	showTunnelModal bool
	tunnelModal     *TunnelModal

	showAICopilot bool
	aiCopilot     *AICopilotModal

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
		if p, err := s.GetSetting("ai_provider"); err == nil && p != "" && c != nil {
			c.AI.Provider = p
		}
		if ep, err := s.GetSetting("ai_endpoint"); err == nil && ep != "" && c != nil {
			c.AI.Endpoint = ep
		}
		if md, err := s.GetSetting("ai_model"); err == nil && md != "" && c != nil {
			c.AI.Model = md
		}
		if ret, err := s.GetSetting("ai_retention_days"); err == nil && ret != "" && c != nil {
			if n, err := strconv.Atoi(ret); err == nil && n > 0 {
				c.AI.RetentionDays = n
			}
		}
		if maxH, err := s.GetSetting("ai_max_history"); err == nil && maxH != "" && c != nil {
			if n, err := strconv.Atoi(maxH); err == nil && n > 0 {
				c.AI.MaxHistory = n
			}
		}
	}

	cInput := textinput.New()
	cInput.Prompt = "$ "
	cInput.Placeholder = "Type command and press Enter..."

	vp := viewport.New(80, 20)

	var pool *ssh.Pool
	if collector != nil && collector.Pool() != nil {
		pool = collector.Pool()
	} else {
		pool = ssh.NewPool(10 * time.Second)
	}
	tm := ssh.NewTunnelManager(pool)

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
		tunnelMgr:         tm,
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
	if m.tunnelMgr != nil {
		m.tunnelMgr.CloseAll()
	}
	if m.sshPool != nil {
		m.sshPool.CloseAll()
	}
}
