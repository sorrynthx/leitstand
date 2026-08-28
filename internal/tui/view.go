package tui

import (
	"leitstand/internal/i18n"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// View renders the entire TUI viewport and orchestrates sub-panel layouts.
func (m *Model) View() string {
	if m.width > 0 && m.height > 0 {
		if m.width < m.cfg.TUI.MinCols || m.height < m.cfg.TUI.MinRows {
			return m.renderResolutionGuard()
		}
	}

	// 1. Overlay Modals
	if m.showVaultModal && m.vaultForm != nil {
		return m.vaultForm.View(m.width, m.height)
	}

	if m.showDeleteModal && m.hostToDelete != nil {
		return m.renderDeleteConfirmationModal()
	}

	if m.showAddModal && m.addForm != nil {
		return m.addForm.View(m.width, m.height)
	}

	if m.showEditModal && m.editForm != nil {
		return m.editForm.View(m.width, m.height)
	}

	if m.showEditorModal && m.editorModal != nil {
		return m.editorModal.View(m.width, m.height)
	}

	if m.showSudoModal && m.sudoModal != nil {
		return m.sudoModal.View(m.width, m.height)
	}

	if m.showSettingsModal && m.settingsModal != nil {
		return m.settingsModal.View(m.width, m.height)
	}

	if m.showFileManager && m.fileManager != nil {
		return m.fileManager.View(m.width, m.height)
	}

	// 2. Runbook Drawer Overlay
	if m.showDrawer && m.drawer != nil {
		drawerWidth := int(float64(m.width) * 0.58)
		if drawerWidth < 55 {
			drawerWidth = 55
		}
		if drawerWidth > m.width-4 {
			drawerWidth = m.width - 4
		}
		drawerHeight := m.height - 2
		if drawerHeight < 15 {
			drawerHeight = 15
		}
		return lipgloss.Place(m.width, m.height, lipgloss.Right, lipgloss.Center, m.drawer.View(drawerWidth, drawerHeight))
	}

	// 3. Master Cockpit Layout
	header := m.renderHeader()
	statusBar := m.renderStatusBar()

	headerHeight := 1
	statusBarHeight := 1
	availableHeight := m.height - headerHeight - statusBarHeight - 1
	if availableHeight < 14 {
		availableHeight = 14
	}

	// Full-screen Console Mode
	if m.fullScreenConsole {
		consolePane := m.renderRemoteConsole(m.width-2, availableHeight-2)
		mainView := lipgloss.JoinVertical(lipgloss.Left, header, consolePane, statusBar)
		return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, mainView)
	}

	leftWidth := int(float64(m.width) * 0.28)
	if leftWidth < 24 {
		leftWidth = 24
	}
	rightWidth := m.width - leftWidth - 2
	if rightWidth < 35 {
		rightWidth = 35
	}

	leftPane := m.renderHostListPane(leftWidth, availableHeight-2)

	var rightSide string
	if m.cfg.Telemetry.PollingInterval <= 0 {
		rightSide = m.renderRemoteConsole(rightWidth, availableHeight-2)
	} else {
		rightTopInnerHeight := 8
		rightTopPane := m.renderTelemetryDeck(rightWidth, rightTopInnerHeight)

		rightBottomInnerHeight := availableHeight - 10 - 2
		if rightBottomInnerHeight < 4 {
			rightBottomInnerHeight = 4
		}
		rightBottomPane := m.renderRemoteConsole(rightWidth, rightBottomInnerHeight)

		rightSide = lipgloss.JoinVertical(lipgloss.Left, rightTopPane, rightBottomPane)
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightSide)
	mainView := lipgloss.JoinVertical(lipgloss.Left, header, content, statusBar)

	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, mainView)
}

// renderHeader renders the top status header bar.
func (m *Model) renderHeader() string {
	title := TitleStyle.Render(i18n.T("app_title"))
	var modeBadge string
	if m.isDemo {
		modeBadge = BadgeStyle.Copy().Background(ColorWarning).Foreground(lipgloss.Color("#000000")).Render(i18n.T("demo_mode"))
	} else {
		modeBadge = BadgeStyle.Render(i18n.T("live_engine"))
	}

	timeStr := lipgloss.NewStyle().Foreground(ColorMuted).Render(time.Now().Format("2006-01-02 15:04:05 MST"))

	gapWidth := m.width - lipgloss.Width(title) - lipgloss.Width(modeBadge) - lipgloss.Width(timeStr) - 4
	if gapWidth < 1 {
		gapWidth = 1
	}

	gap := strings.Repeat(" ", gapWidth)
	return lipgloss.JoinHorizontal(lipgloss.Center, title, " ", modeBadge, gap, timeStr)
}

// renderStatusBar renders the bottom status bar with contextual shortcuts.
func (m *Model) renderStatusBar() string {
	var keys string
	if m.activePane == PaneConsole {
		keys = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_tab_new")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render(i18n.T("status_tab_switch")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_tab_close")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(i18n.T("status_complete")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(i18n.T("status_files")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(i18n.T("status_runbook")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_maximize")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_exit_console")) + "  "
	} else {
		keys = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_focus_console")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render(i18n.T("status_terminal")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(i18n.T("status_files")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(i18n.T("status_runbook")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_tab_new")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_reconnect")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_settings")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_add")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_edit")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_del")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_quit")) + "  "
	}

	status := m.statusMessage
	keysLen := runewidth.StringWidth(keys)
	statusLen := runewidth.StringWidth(status)
	availWidth := (m.width - 4) - keysLen - statusLen
	if availWidth < 1 {
		availWidth = 1
	}

	content := keys + strings.Repeat(" ", availWidth) + status
	return StatusBarStyle.Width(m.width - 2).Render(content)
}
