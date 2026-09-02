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
	if m.showVaultModal {
		vf := m.vaultModal
		if vf == nil {
			vf = m.vaultForm
		}
		if vf != nil {
			return vf.View(m.width, m.height)
		}
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

	// 3. Master Cockpit Layout
	header := m.renderHeader()
	statusBar := m.renderStatusBar()

	headerHeight := 1
	statusBarHeight := 1
	availableHeight := m.height - headerHeight - statusBarHeight - 2
	if availableHeight < 14 {
		availableHeight = 14
	}

	var mainContent string
	if m.fullScreenConsole {
		consolePane := m.renderRemoteConsole(m.width-4, availableHeight-2)
		mainContent = lipgloss.JoinVertical(lipgloss.Left, header, consolePane, statusBar)
	} else {
		leftWidth := int(float64(m.width) * 0.28)
		if leftWidth < 25 {
			leftWidth = 25
		}
		rightWidth := m.width - leftWidth - 6
		if rightWidth < 45 {
			rightWidth = 45
		}

		leftPane := m.renderHostListPane(leftWidth, availableHeight-2)

		var rightSide string
		if !m.userHasNavigated {
			rightSide = m.renderWelcomePanel(rightWidth, availableHeight-2)
		} else {
			rightSide = m.renderRemoteConsole(rightWidth, availableHeight-2)
		}

		content := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightSide)
		mainContent = lipgloss.JoinVertical(lipgloss.Left, header, content, statusBar)
	}

	// 2. Runbook Drawer Overlay (100% Full Width Modal)
	if m.showDrawer && m.drawer != nil {
		drawerWidth := m.width - 4
		if drawerWidth < 55 {
			drawerWidth = 55
		}
		drawerHeight := m.height - 2
		if drawerHeight < 15 {
			drawerHeight = 15
		}
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.drawer.View(drawerWidth, drawerHeight))
	}



	return mainContent
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
	if len(m.hosts) == 0 {
		keys = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_add")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_settings")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_quit")) + "  "
	} else if m.activePane == PaneConsole {
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
	availWidth := (m.width - 1) - runewidth.StringWidth(keys) - runewidth.StringWidth(status) - 2
	if availWidth < 1 {
		availWidth = 1
	}

	content := keys + strings.Repeat(" ", availWidth) + status
	return StatusBarStyle.Width(m.width - 1).Render(content)
}

// renderWelcomePanel renders the clean text view without any panel borders or boxes before a server is selected.
func (m *Model) renderWelcomePanel(width, height int) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("welcome_title"))
	b.WriteString(title + "\n\n")

	subTitle := lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("welcome_subtitle"))
	b.WriteString(subTitle + "\n\n")

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render(i18n.T("welcome_shortcuts")) + "\n\n")
	b.WriteString("  • " + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render("↑ / ↓  (or j/k)") + "   : " + i18n.T("welcome_sc_1") + "\n")
	b.WriteString("  • " + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render("Enter / c / s") + "   : " + i18n.T("welcome_sc_2") + "\n")
	b.WriteString("  • " + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render("a / n") + "           : " + i18n.T("welcome_sc_3") + "\n")
	b.WriteString("  • " + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render("f / F6") + "          : " + i18n.T("welcome_sc_4") + "\n")
	b.WriteString("  • " + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render("t") + "             : " + i18n.T("welcome_sc_5") + "\n")
	b.WriteString("  • " + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render("?") + "             : " + i18n.T("welcome_sc_6") + "\n")
	b.WriteString("  • " + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render("p") + "             : " + i18n.T("welcome_sc_7") + "\n\n")

	if len(m.hosts) > 0 {
		statusStr := lipgloss.NewStyle().Foreground(ColorSuccess).Render(i18n.Tf("welcome_ready", len(m.hosts)))
		b.WriteString(statusStr)
	} else {
		statusStr := lipgloss.NewStyle().Foreground(ColorWarning).Render(i18n.T("welcome_no_hosts"))
		b.WriteString(statusStr)
	}

	return lipgloss.NewStyle().Padding(1, 2).Width(width).Height(height).Render(b.String())
}
