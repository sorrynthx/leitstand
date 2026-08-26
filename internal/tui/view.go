package tui

import (
	"fmt"
	"leitstand/internal/i18n"
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

	if m.showEditorModal && m.editorModal != nil {
		return m.editorModal.View(m.width, m.height)
	}

	if m.showSettingsModal && m.settingsModal != nil {
		return m.settingsModal.View(m.width, m.height)
	}

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

	header := m.renderHeader()
	statusBar := m.renderStatusBar()

	// Calculate exact available height for main content with safe terminal margin
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

	// Left pane: inner height = availableHeight - 2 (outer box = availableHeight)
	leftPane := m.renderHostListPane(leftWidth, availableHeight-2)

	var rightSide string
	if m.cfg.Telemetry.PollingInterval <= 0 {
		// Telemetry is Off: Give entire right side to Remote Console
		rightSide = m.renderRemoteConsole(rightWidth, availableHeight-2)
	} else {
		// Right Top pane: inner height 8 (outer box = 10)
		rightTopInnerHeight := 8
		rightTopPane := m.renderTelemetryDeck(rightWidth, rightTopInnerHeight)

		// Right Bottom pane: inner height = availableHeight - 10 - 2 (outer box = availableHeight - 10)
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

func (m *Model) renderDeleteConfirmationModal() string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(i18n.T("modal_delete_server"))
	b.WriteString(title + "\n\n")

	msg := i18n.Tf("modal_delete_warn", m.hostToDelete.Name, m.hostToDelete.Address)
	b.WriteString(msg + "\n\n")

	actions := lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render("[y / Enter] "+i18n.T("btn_delete")) +
		"    " +
		lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render("[n / Esc] "+i18n.T("btn_cancel"))
	b.WriteString(actions)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorDanger).
		Padding(1, 3).
		Width(65)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, boxStyle.Render(b.String()))
}

func (m *Model) renderResolutionGuard() string {
	minCols := m.cfg.TUI.MinCols
	if minCols <= 0 {
		minCols = 92
	}
	minRows := m.cfg.TUI.MinRows
	if minRows <= 0 {
		minRows = 22
	}

	msg := fmt.Sprintf(
		"%s\n\n"+
			"%s  %d cols × %d rows\n"+
			"%s  %d cols × %d rows\n\n"+
			"%s",
		lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(i18n.T("guard_title")),
		lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("guard_current")),
		m.width, m.height,
		lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(i18n.T("guard_required")),
		minCols, minRows,
		lipgloss.NewStyle().Foreground(ColorWarning).Render(i18n.T("guard_hint")),
	)
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		WarningBoxStyle.Render(msg),
	)
}

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

func (m *Model) renderHostListPane(width, height int) string {
	var b strings.Builder

	titleColor := ColorMuted
	if m.activePane == PaneHostList {
		titleColor = ColorPrimary
	}
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(titleColor).Render(i18n.T("pane_host_explorer")) + "\n\n")

	if len(m.hosts) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("no_hosts")))
	} else {
		currentGroup := ""
		for i, h := range m.hosts {
			if h.GroupName != "" && h.GroupName != currentGroup {
				currentGroup = h.GroupName
				b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("📁 "+currentGroup) + "\n")
			}

			isSelected := i == m.selectedIndex
			st := m.hostStatus[h.ID]
			var dot string
			var statusLabel string
			switch st {
			case HostStatusOnline:
				dot = lipgloss.NewStyle().Foreground(ColorSuccess).Render("●")
			case HostStatusConnecting:
				dot = lipgloss.NewStyle().Foreground(ColorWarning).Render("◌")
				statusLabel = lipgloss.NewStyle().Foreground(ColorWarning).Render(" (" + i18n.T("host_connecting") + ")")
			case HostStatusOffline:
				dot = lipgloss.NewStyle().Foreground(ColorDanger).Render("✖")
				statusLabel = lipgloss.NewStyle().Foreground(ColorDanger).Render(" (" + i18n.T("host_offline") + ")")
			default:
				if _, hasErr := m.errors[h.ID]; hasErr {
					dot = lipgloss.NewStyle().Foreground(ColorDanger).Render("✖")
					statusLabel = lipgloss.NewStyle().Foreground(ColorDanger).Render(" (" + i18n.T("host_offline") + ")")
				} else {
					dot = lipgloss.NewStyle().Foreground(ColorSuccess).Render("●")
				}
			}

			var hostLine string
			hostText := fmt.Sprintf("%s %s (%s)%s", dot, h.Name, h.Address, statusLabel)
			if isSelected {
				hostLine = SelectedHostStyle.Width(width - 6).Render(hostText)
			} else {
				hostLine = UnselectedHostStyle.Width(width - 6).Render(hostText)
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
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("select_host_hint")))
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
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(deckTitleColor).Render(i18n.T("pane_telemetry")) + " ── " +
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render(selectedHost.Name) + " (" + selectedHost.Address + ")\n")

	// System Specification Banner
	if sysInfo != nil {
		specLine := fmt.Sprintf("🐧 %s  |  ⚡ %d %s  |  ⏱ %s",
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(sysInfo.OSDistro),
			sysInfo.CPUCores,
			i18n.T("cores"),
			lipgloss.NewStyle().Foreground(ColorWarning).Render(formatCompactUptime(sysInfo.Uptime)),
		)
		specLine = runewidth.Truncate(specLine, width-6, "…")
		b.WriteString(specLine + "\n")
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
		} else if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "i/o timeout") {
			errReason = "⏳ Connection Timeout: Network unreachable or VPN required."
		} else {
			errReason = fmt.Sprintf("⚠️ Connection Error: %v", err)
		}

		b.WriteString(errBox.Render(errReason) + "\n")
		paneBorder := PaneStyle
		if m.activePane == PaneTelemetryDeck {
			paneBorder = ActivePaneStyle
		}
		return paneBorder.Width(width).Height(height).Render(b.String())
	}

	if !hasMetric || metric == nil {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("⏳ Collecting telemetry data..."))
		paneBorder := PaneStyle
		if m.activePane == PaneTelemetryDeck {
			paneBorder = ActivePaneStyle
		}
		return paneBorder.Width(width).Height(height).Render(b.String())
	}

	innerAvail := width - 8
	cardTotalWidth := (innerAvail - 3) / 4
	cardContentWidth := cardTotalWidth - 4
	if cardContentWidth < 8 {
		cardContentWidth = 8
	}

	cpuGauge := m.renderMetricGauge(i18n.T("cpu_usage"), metric.CPUPercent, 100.0, "%", cardContentWidth)

	memPercent := 0.0
	if metric.MemoryTotal > 0 {
		memPercent = float64(metric.MemoryUsed) / float64(metric.MemoryTotal) * 100.0
	}
	memGauge := m.renderMetricGauge(i18n.T("mem_usage"), memPercent, 100.0, "%", cardContentWidth)

	diskPercent := 0.0
	if metric.DiskTotal > 0 {
		diskPercent = float64(metric.DiskUsed) / float64(metric.DiskTotal) * 100.0
	}
	diskGauge := m.renderMetricGauge(i18n.T("disk_usage"), diskPercent, 100.0, "%", cardContentWidth)

	netCard := m.renderNetworkCard(i18n.T("network_io"), metric.NetRxBytes, metric.NetTxBytes, cardContentWidth)

	gauges := lipgloss.JoinHorizontal(lipgloss.Top, cpuGauge, " ", memGauge, " ", diskGauge, " ", netCard)
	b.WriteString(gauges)

	paneBorder := PaneStyle
	if m.activePane == PaneTelemetryDeck {
		paneBorder = ActivePaneStyle
	}

	return paneBorder.Width(width).Height(height).Render(b.String())
}

func (m *Model) renderMetricGauge(label string, value, maxVal float64, unit string, contentWidth int) string {
	var b strings.Builder
	b.WriteString(MetricLabelStyle.Render(label) + "\n")

	valStr := fmt.Sprintf("%.1f%s", value, unit)
	var valColor lipgloss.Color
	switch {
	case value > 85.0:
		valColor = ColorDanger
	case value > 65.0:
		valColor = ColorWarning
	default:
		valColor = ColorSuccess
	}
	b.WriteString(MetricValueStyle.Foreground(valColor).Render(valStr) + "\n")

	barWidth := contentWidth - 2
	if barWidth < 4 {
		barWidth = 4
	}
	filled := int((value / maxVal) * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}

	bar := lipgloss.NewStyle().Foreground(valColor).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(ColorMuted).Render(strings.Repeat("░", barWidth-filled))
	b.WriteString(bar)

	return MetricCardStyle.Width(contentWidth).Height(3).Render(b.String())
}

func (m *Model) renderNetworkCard(label string, rxBytesPerSec, txBytesPerSec uint64, contentWidth int) string {
	var b strings.Builder
	cleanLabel := runewidth.Truncate(label, contentWidth, "")
	b.WriteString(MetricLabelStyle.Render(cleanLabel) + "\n")

	rxStr := "↓ " + formatNetworkSpeed(rxBytesPerSec)
	txStr := "↑ " + formatNetworkSpeed(txBytesPerSec)

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render(rxStr) + "\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(txStr))

	return MetricCardStyle.Width(contentWidth).Height(3).Render(b.String())
}

func formatCompactUptime(u string) string {
	u = strings.TrimPrefix(u, "up ")
	parts := strings.Split(u, ",")
	if len(parts) > 2 {
		return strings.TrimSpace(parts[0]) + ", " + strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(u)
}

func formatNetworkSpeed(bytesPerSec uint64) string {
	val := float64(bytesPerSec)
	switch {
	case val >= 1024*1024*1024:
		return fmt.Sprintf("%.2f GB/s", val/(1024*1024*1024))
	case val >= 1024*1024:
		return fmt.Sprintf("%.1f MB/s", val/(1024*1024))
	case val >= 1024:
		return fmt.Sprintf("%.1f KB/s", val/1024)
	default:
		return fmt.Sprintf("%.0f B/s", val)
	}
}

func (m *Model) renderRemoteConsole(width, height int) string {
	var b strings.Builder

	titleColor := ColorMuted
	if m.activePane == PaneConsole {
		titleColor = ColorPrimary
	}

	modeHint := i18n.T("console_mode_hint")
	if m.fullScreenConsole {
		modeHint = i18n.T("console_mode_max_hint")
	}

	titleLine := lipgloss.NewStyle().Bold(true).Foreground(titleColor).Render(i18n.T("pane_console")) + "  " +
		lipgloss.NewStyle().Foreground(ColorMuted).Render(modeHint)
	titleLine = runewidth.Truncate(titleLine, width-6, "…")
	b.WriteString(titleLine + "\n")

	if len(m.hosts) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("console_no_host")))
		paneBorder := PaneStyle
		if m.activePane == PaneConsole {
			paneBorder = ActivePaneStyle
		}
		return paneBorder.Width(width).Height(height).Render(b.String())
	}

	curHost := m.hosts[m.selectedIndex]
	cwd := m.hostCWD[curHost.ID]
	if cwd == "" {
		cwd = "~"
	}

	st := m.hostStatus[curHost.ID]
	if st == HostStatusOffline {
		m.consoleInput.Placeholder = i18n.T("console_offline_ph")
		m.consoleInput.Prompt = lipgloss.NewStyle().Foreground(ColorDanger).Render("[Offline] ❯ ")
	} else if st == HostStatusConnecting {
		m.consoleInput.Placeholder = i18n.T("console_connecting_ph")
		m.consoleInput.Prompt = lipgloss.NewStyle().Foreground(ColorWarning).Render("[Connecting...] ❯ ")
	} else {
		m.consoleInput.Placeholder = i18n.T("console_placeholder")
		m.consoleInput.Prompt = lipgloss.NewStyle().Foreground(ColorPrimary).Render(fmt.Sprintf("[%s] ❯ ", cwd))
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
	var keys string
	if m.activePane == PaneConsole {
		keys = lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(i18n.T("status_complete")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(i18n.T("status_runbook")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_exit_console")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_history")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_maximize")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_quit")) + "  "
	} else {
		keys = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_focus_console")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(i18n.T("status_runbook")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_reconnect")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_settings")) + "  " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("status_add")) + "  " +
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
