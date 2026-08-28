package tui

import (
	"context"
	"fmt"
	"leitstand/internal/i18n"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
)

// ConsoleTab represents a single tab session within a host's console.
type ConsoleTab struct {
	ID           string
	Title        string
	CWD          string
	IsRoot       bool
	User         string
	Logs         []string
	CmdHistory   []string
	HistoryIndex int
	Viewport     viewport.Model
	InputText    string
	IsStreaming  bool
	IsScreenApp  bool
	StreamCmd    string
	StreamCancel context.CancelFunc
	CreatedAt    time.Time
}

// HostTabState holds all tabs and the active tab index for a specific host.
type HostTabState struct {
	HostID      int64
	Tabs        []*ConsoleTab
	ActiveIndex int
	mu          sync.RWMutex
}

// NewHostTabState creates an initial HostTabState with one default tab.
func NewHostTabState(hostID int64, hostName string, initialCWD string) *HostTabState {
	if initialCWD == "" {
		initialCWD = "~"
	}

	defaultTitle := fmt.Sprintf("1: %s", i18n.T("tab_default_title"))
	if defaultTitle == "1: tab_default_title" {
		defaultTitle = "1: bash"
	}

	tab := &ConsoleTab{
		ID:           fmt.Sprintf("tab_%d_%d", hostID, time.Now().UnixNano()),
		Title:        defaultTitle,
		CWD:          initialCWD,
		Logs:         make([]string, 0),
		CmdHistory:   make([]string, 0),
		HistoryIndex: -1,
		Viewport:     viewport.New(60, 10),
		CreatedAt:    time.Now(),
	}

	return &HostTabState{
		HostID:      hostID,
		Tabs:        []*ConsoleTab{tab},
		ActiveIndex: 0,
	}
}

// ActiveTab returns the currently focused tab.
func (hts *HostTabState) ActiveTab() *ConsoleTab {
	hts.mu.RLock()
	defer hts.mu.RUnlock()

	if len(hts.Tabs) == 0 {
		return nil
	}
	if hts.ActiveIndex < 0 || hts.ActiveIndex >= len(hts.Tabs) {
		hts.ActiveIndex = 0
	}
	return hts.Tabs[hts.ActiveIndex]
}

// AddNewTab creates and appends a new tab, setting it as active.
func (hts *HostTabState) AddNewTab(vpWidth, vpHeight int) *ConsoleTab {
	hts.mu.Lock()
	defer hts.mu.Unlock()

	tabNum := len(hts.Tabs) + 1
	tabTitle := fmt.Sprintf("%d: bash", tabNum)

	baseCWD := "~"
	if len(hts.Tabs) > 0 && hts.ActiveIndex >= 0 && hts.ActiveIndex < len(hts.Tabs) {
		baseCWD = hts.Tabs[hts.ActiveIndex].CWD
	}

	vp := viewport.New(vpWidth, vpHeight)

	tab := &ConsoleTab{
		ID:           fmt.Sprintf("tab_%d_%d", hts.HostID, time.Now().UnixNano()),
		Title:        tabTitle,
		CWD:          baseCWD,
		Logs:         make([]string, 0),
		CmdHistory:   make([]string, 0),
		HistoryIndex: -1,
		Viewport:     vp,
		CreatedAt:    time.Now(),
	}

	hts.Tabs = append(hts.Tabs, tab)
	hts.ActiveIndex = len(hts.Tabs) - 1
	return tab
}

// CloseActiveTab closes the current active tab. Returns false if only 1 tab remains.
func (hts *HostTabState) CloseActiveTab() bool {
	hts.mu.Lock()
	defer hts.mu.Unlock()

	if len(hts.Tabs) <= 1 {
		return false
	}

	closingTab := hts.Tabs[hts.ActiveIndex]
	if closingTab.StreamCancel != nil {
		closingTab.StreamCancel()
		closingTab.StreamCancel = nil
	}

	// Remove from slice
	hts.Tabs = append(hts.Tabs[:hts.ActiveIndex], hts.Tabs[hts.ActiveIndex+1:]...)

	// Adjust active index
	if hts.ActiveIndex >= len(hts.Tabs) {
		hts.ActiveIndex = len(hts.Tabs) - 1
	}

	// Re-number tab titles if they follow standard numbering
	for i, tab := range hts.Tabs {
		if strings.Contains(tab.Title, ":") {
			parts := strings.SplitN(tab.Title, ":", 2)
			tab.Title = fmt.Sprintf("%d:%s", i+1, parts[1])
		}
	}

	return true
}

// SwitchTab switches to a specific tab index (0-based).
func (hts *HostTabState) SwitchTab(index int) bool {
	hts.mu.Lock()
	defer hts.mu.Unlock()

	if index < 0 || index >= len(hts.Tabs) {
		return false
	}
	hts.ActiveIndex = index
	return true
}

// NextTab switches to the next tab in circular order.
func (hts *HostTabState) NextTab() {
	hts.mu.Lock()
	defer hts.mu.Unlock()

	if len(hts.Tabs) <= 1 {
		return
	}
	hts.ActiveIndex = (hts.ActiveIndex + 1) % len(hts.Tabs)
}

// PrevTab switches to the previous tab in circular order.
func (hts *HostTabState) PrevTab() {
	hts.mu.Lock()
	defer hts.mu.Unlock()

	if len(hts.Tabs) <= 1 {
		return
	}
	hts.ActiveIndex--
	if hts.ActiveIndex < 0 {
		hts.ActiveIndex = len(hts.Tabs) - 1
	}
}

// AppendLog adds a log line to a tab and updates its viewport.
func (tab *ConsoleTab) AppendLog(entry string) {
	tab.Logs = append(tab.Logs, entry)
	if len(tab.Logs) > 100 {
		tab.Logs = tab.Logs[len(tab.Logs)-100:]
	}
	tab.UpdateViewportContent()
}

// SetFrame replaces the tab content with a fresh live screen frame (top/htop/watch).
func (tab *ConsoleTab) SetFrame(frame string) {
	tab.Logs = []string{frame}
	tab.Viewport.SetContent(frame)
}

// UpdateViewportContent refreshes the tab's viewport content.
func (tab *ConsoleTab) UpdateViewportContent() {
	if len(tab.Logs) == 0 {
		welcomeMsg := fmt.Sprintf("Terminal Session Tab [%s]\nType remote commands below and press Enter to execute.\n[Ctrl+N] New Tab  [Ctrl+W] Close Tab  [Alt+1~9] Switch Tab", tab.Title)
		tab.Viewport.SetContent(welcomeMsg)
		return
	}
	tab.Viewport.SetContent(strings.Join(tab.Logs, "\n\n"))
	tab.Viewport.GotoBottom()
}

// SetAutoTitle automatically updates tab title based on running command or cwd.
func (tab *ConsoleTab) SetAutoTitle(tabIndex int, cmdText string) {
	trimmed := strings.TrimSpace(cmdText)
	if trimmed == "" {
		if tab.IsRoot {
			tab.Title = fmt.Sprintf("%d: root#", tabIndex+1)
		} else {
			tab.Title = fmt.Sprintf("%d: bash", tabIndex+1)
		}
		return
	}

	fields := strings.Fields(trimmed)
	shortCmd := fields[0]
	if len(fields) > 1 && (shortCmd == "sudo" || shortCmd == "tail" || shortCmd == "docker" || shortCmd == "journalctl" || shortCmd == "su") {
		shortCmd = fmt.Sprintf("%s %s", fields[0], fields[1])
	}

	if tab.IsRoot && !strings.HasPrefix(shortCmd, "root") {
		shortCmd = fmt.Sprintf("root: %s", shortCmd)
	}

	if len(shortCmd) > 18 {
		shortCmd = shortCmd[:18] + "…"
	}

	tab.Title = fmt.Sprintf("%d: %s", tabIndex+1, shortCmd)
}
