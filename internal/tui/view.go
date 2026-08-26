package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// View renders the entire TUI viewport.
func (m *Model) View() string {
	if m.width > 0 && m.height > 0 {
		if m.width < m.cfg.TUI.MinCols || m.height < m.cfg.TUI.MinRows {
			return m.renderResolutionGuard()
		}
	}

	if m.showVaultModal && m.vaultForm != nil {
		return m.vaultForm.View(m.width, m.height)
	}

	if m.showDeleteModal && m.hostToDelete != nil {
		return m.renderDeleteConfirmationModal()
	}

	if m.showAddModal && m.addForm != nil {
		return m.addForm.View(m.width, m.height)
	}

	header := m.renderHeader()
	statusBar := m.renderStatusBar()

	// Calculate exact available height for main content with safe terminal margin
	headerHeight := 1
	statusBarHeight := 1
	availableHeight := m.height - headerHeight - statusBarHeight - 2
	if availableHeight < 14 {
		availableHeight = 14
	}

	// Full-screen Console Mode
	if m.fullScreenConsole {
		consolePane := m.renderRemoteConsole(m.width-4, availableHeight-2)
		return lipgloss.JoinVertical(lipgloss.Left, header, consolePane, statusBar)
	}

	leftWidth := int(float64(m.width) * 0.28)
	if leftWidth < 25 {
		leftWidth = 25
	}
	rightWidth := m.width - leftWidth - 4
	if rightWidth < 45 {
		rightWidth = 50
	}

	// Left pane: inner height = availableHeight - 2 (outer box = availableHeight)
	leftPane := m.renderHostListPane(leftWidth, availableHeight-2)

	// Right Top pane: inner height 6 (outer box = 8)
	rightTopInnerHeight := 6
	rightTopPane := m.renderTelemetryDeck(rightWidth, rightTopInnerHeight)

	// Right Bottom pane: inner height = availableHeight - 8 - 2 (outer box = availableHeight - 8)
	rightBottomInnerHeight := availableHeight - 8 - 2
	if rightBottomInnerHeight < 4 {
		rightBottomInnerHeight = 4
	}
	rightBottomPane := m.renderRemoteConsole(rightWidth, rightBottomInnerHeight)

	rightSide := lipgloss.JoinVertical(lipgloss.Left, rightTopPane, rightBottomPane)
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightSide)

	return lipgloss.JoinVertical(lipgloss.Left, header, content, statusBar)
}

func (m *Model) renderDeleteConfirmationModal() string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render("⚠️  DELETE SERVER CONFIRMATION")
	b.WriteString(title + "\n\n")

	msg := fmt.Sprintf(
		"Are you sure you want to remove host '%s' (%s)?\n\n"+
			"All encrypted credentials and collected metrics for this host\n"+
			"will be permanently deleted.\n\n",
		lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(m.hostToDelete.Name),
		m.hostToDelete.Address,
	)
	b.WriteString(msg)

	actions := lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render("[y / Enter] Yes, Delete") +
		"    " +
		lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render("[n / Esc] Cancel")
	b.WriteString(actions)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorDanger).
		Padding(1, 3).
		Width(65)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, boxStyle.Render(b.String()))
}

func (m *Model) renderResolutionGuard() string {
	msg := fmt.Sprintf(
		"⚠️  Terminal Window Too Small\n\n"+
			"Current: %d cols × %d rows\n"+
			"Required: %d cols × %d rows\n\n"+
			"Please enlarge your terminal window.",
		m.width, m.height, m.cfg.TUI.MinCols, m.cfg.TUI.MinRows,
	)
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		WarningBoxStyle.Render(msg),
	)
}

func (m *Model) renderHeader() string {
	title := TitleStyle.Render("⚡ LEITSTAND COCKPIT")
	var modeBadge string
	if m.isDemo {
		modeBadge = BadgeStyle.Copy().Background(ColorWarning).Foreground(lipgloss.Color("#000000")).Render("DEMO MODE")
	} else {
		modeBadge = BadgeStyle.Render("LIVE ENGINE")
	}

	timeStr := lipgloss.NewStyle().Foreground(ColorMuted).Render(time.Now().Format("2006-01-02 15:04:05 MST"))

	gapWidth := m.width - lipgloss.Width(title) - lipgloss.Width(modeBadge) - lipgloss.Width(timeStr) - 4
	if gapWidth < 1 {
		gapWidth = 1
	}

	gap := strings.Repeat(" ", gapWidth)
	return lipgloss.JoinHorizontal(lipgloss.Center, title, " ", modeBadge, gap, timeStr)
}

func (m *Model) renderHostListPane(width, height int) string {
	var b strings.Builder

	titleColor := ColorMuted
	if m.activePane == PaneHostList {
		titleColor = ColorPrimary
	}
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(titleColor).Render("HOST EXPLORER") + "\n\n")

	if len(m.hosts) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("No hosts registered.\nPress 'a' to add a host."))
	} else {
		currentGroup := ""
		for i, h := range m.hosts {
			if h.GroupName != "" && h.GroupName != currentGroup {
				currentGroup = h.GroupName
				b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("📁 "+currentGroup) + "\n")
			}

			isSelected := i == m.selectedIndex
			dot := lipgloss.NewStyle().Foreground(ColorSuccess).Render("●")
			if _, hasErr := m.errors[h.ID]; hasErr {
				dot = lipgloss.NewStyle().Foreground(ColorDanger).Render("✖")
			}

			var hostLine string
			if isSelected {
				hostLine = SelectedHostStyle.Width(width - 6).Render(fmt.Sprintf("%s %s (%s)", dot, h.Name, h.Address))
			} else {
				hostLine = UnselectedHostStyle.Width(width - 6).Render(fmt.Sprintf("%s %s (%s)", dot, h.Name, h.Address))
			}
			b.WriteString(hostLine + "\n")
		}
	}

	paneBorder := PaneStyle
	if m.activePane == PaneHostList {
		paneBorder = ActivePaneStyle
	}

	return paneBorder.Width(width).Height(height).Render(b.String())
}

func (m *Model) renderTelemetryDeck(width, height int) string {
	var b strings.Builder

	if len(m.hosts) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("Select a host from the left panel to inspect live telemetry."))
		return PaneStyle.Width(width).Height(height).Render(b.String())
	}

	selectedHost := m.hosts[m.selectedIndex]
	metric, hasMetric := m.metrics[selectedHost.ID]
	err, hasErr := m.errors[selectedHost.ID]
	sysInfo := m.sysInfos[selectedHost.ID]

	deckTitleColor := ColorMuted
	if m.activePane == PaneTelemetryDeck {
		deckTitleColor = ColorPrimary
	}

	// Host Header
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(deckTitleColor).Render("TELEMETRY DECK") + " ── " +
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render(selectedHost.Name) + " (" + selectedHost.Address + ")\n")

	// System Specification Banner
	if sysInfo != nil {
		specLine := fmt.Sprintf("🐧 %s  |  ⚡ %d Cores  |  ⏱ %s\n",
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(sysInfo.OSDistro),
			sysInfo.CPUCores,
			lipgloss.NewStyle().Foreground(ColorWarning).Render(sysInfo.Uptime),
		)
		b.WriteString(specLine)
	} else {
		b.WriteString("\n")
	}

	if hasErr && err != nil {
		errBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDanger).
			Padding(0, 1).
			Width(width - 6)

		var errReason string
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "authenticate") || strings.Contains(errStr, "password") {
			errReason = "❌ Authentication Failed: Incorrect password or username rejected by server."
		} else if strings.Contains(errStr, "timeout") {
			errReason = "⏱️ Connection Timed Out: Server unreachable or firewall blocking port."
		} else if strings.Contains(errStr, "refused") {
			errReason = "🚫 Connection Refused: SSH daemon is not running on target port."
		} else {
			errReason = fmt.Sprintf("⚠️ SSH Error: %v", err)
		}

		details := fmt.Sprintf(
			"%s\n"+
				"Target: %s@%s:%d  |  Suggestions: Press [r] to retry, or [d] to delete host",
			lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(errReason),
			selectedHost.Username, selectedHost.Address, selectedHost.Port,
		)

		b.WriteString(errBox.Render(details) + "\n")
		paneBorder := PaneStyle
		if m.activePane == PaneTelemetryDeck {
			paneBorder = ActivePaneStyle
		}
		return paneBorder.Width(width).Height(height).Render(b.String())
	}

	if !hasMetric || metric == nil {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("⏳ Connecting and waiting for initial telemetry sample...\n"))
		paneBorder := PaneStyle
		if m.activePane == PaneTelemetryDeck {
			paneBorder = ActivePaneStyle
		}
		return paneBorder.Width(width).Height(height).Render(b.String())
	}

	// Safe gauge width to guarantee 0 line wrap on any screen width
	gaugeWidth := width - 52
	if gaugeWidth < 12 {
		gaugeWidth = 12
	} else if gaugeWidth > 24 {
		gaugeWidth = 24
	}

	// 1. CPU Section
	cpuColor := ColorSuccess
	if metric.CPUPercent > 85.0 {
		cpuColor = ColorDanger
	} else if metric.CPUPercent > 60.0 {
		cpuColor = ColorWarning
	}

	cpuBar := RenderProgressBar(gaugeWidth, metric.CPUPercent, cpuColor, ColorBorder)
	cpuVal := fmt.Sprintf("%5.1f%%", metric.CPUPercent)
	b.WriteString(fmt.Sprintf("%-10s [%s] %s\n", "CPU Usage", cpuBar, lipgloss.NewStyle().Bold(true).Foreground(cpuColor).Render(cpuVal)))

	// 2. Memory Section
	var memPercent float64
	if metric.MemoryTotal > 0 {
		memPercent = (float64(metric.MemoryUsed) / float64(metric.MemoryTotal)) * 100.0
	}
	memColor := ColorPrimary
	if memPercent > 85.0 {
		memColor = ColorDanger
	}
	memBar := RenderProgressBar(gaugeWidth, memPercent, memColor, ColorBorder)
	memVal := fmt.Sprintf("%5.1f%% (%s / %s)", memPercent, formatBytes(metric.MemoryUsed), formatBytes(metric.MemoryTotal))
	b.WriteString(fmt.Sprintf("%-10s [%s] %s\n", "Memory", memBar, lipgloss.NewStyle().Foreground(ColorPrimary).Render(memVal)))

	// 3. Disk Section
	var diskPercent float64
	if metric.DiskTotal > 0 {
		diskPercent = (float64(metric.DiskUsed) / float64(metric.DiskTotal)) * 100.0
	}
	diskColor := ColorSecondary
	if diskPercent > 90.0 {
		diskColor = ColorDanger
	}
	diskBar := RenderProgressBar(gaugeWidth, diskPercent, diskColor, ColorBorder)
	diskVal := fmt.Sprintf("%5.1f%% (%s / %s)", diskPercent, formatBytes(metric.DiskUsed), formatBytes(metric.DiskTotal))
	b.WriteString(fmt.Sprintf("%-10s [%s] %s\n", "Disk (/)", diskBar, lipgloss.NewStyle().Foreground(ColorSecondary).Render(diskVal)))

	// 4. Network Section
	rxStr := formatBytes(metric.NetRxBytes)
	txStr := formatBytes(metric.NetTxBytes)
	netLine := fmt.Sprintf("🌐 Net: ⬇ RX: %s   ⬆ TX: %s   ⏱ Tick: %s",
		lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(rxStr),
		lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(txStr),
		metric.Timestamp.Format("15:04:05"),
	)
	b.WriteString(netLine)

	paneBorder := PaneStyle
	if m.activePane == PaneTelemetryDeck {
		paneBorder = ActivePaneStyle
	}

	return paneBorder.Width(width).Height(height).Render(b.String())
}

func (m *Model) renderRemoteConsole(width, height int) string {
	var b strings.Builder

	titleColor := ColorMuted
	if m.activePane == PaneConsole {
		titleColor = ColorPrimary
	}

	modeHint := "[Tab] Focus  [Ctrl+O] Fullscreen"
	if m.fullScreenConsole {
		modeHint = "[Ctrl+O] Exit Fullscreen"
	}

	titleLine := lipgloss.NewStyle().Bold(true).Foreground(titleColor).Render("REMOTE COMMAND CONSOLE") + "  " +
		lipgloss.NewStyle().Foreground(ColorMuted).Render(modeHint)
	b.WriteString(titleLine + "\n")

	if len(m.hosts) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("No host selected."))
		paneBorder := PaneStyle
		if m.activePane == PaneConsole {
			paneBorder = ActivePaneStyle
		}
		return paneBorder.Width(width).Height(height).Render(b.String())
	}

	// Total inner lines = height. Title = 1 line, Input = 1 line, Viewport = height - 2
	vpHeight := height - 2
	if vpHeight < 2 {
		vpHeight = 2
	}
	m.viewport.Width = width - 4
	m.viewport.Height = vpHeight

	b.WriteString(m.viewport.View() + "\n")

	// Command input prompt
	m.consoleInput.Width = width - 6
	b.WriteString(m.consoleInput.View())

	paneBorder := PaneStyle
	if m.activePane == PaneConsole {
		paneBorder = ActivePaneStyle
	}

	return paneBorder.Width(width).Height(height).Render(b.String())
}

func (m *Model) renderStatusBar() string {
	keys := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("[Tab]") + " Focus  " +
		lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("[PgUp/Dn]") + " Scroll  " +
		lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("[Ctrl+O]") + " Maximize  " +
		lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("[a]") + " Add  " +
		lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("[d]") + " Del  " +
		lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("[r]") + " Refresh  " +
		lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("[q]") + " Quit  "

	status := m.statusMessage
	availWidth := (m.width - 1) - runewidth.StringWidth(keys) - runewidth.StringWidth(status) - 2
	if availWidth < 1 {
		availWidth = 1
	}

	return StatusBarStyle.Width(m.width - 1).Render(keys + strings.Repeat(" ", availWidth) + status)
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
