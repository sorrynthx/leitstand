package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func (m *Model) renderHostListPane(width, height int) string {
	if m.showTelemetryDrawer {
		return m.renderLeftTelemetryPane(width, height)
	}

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

			isSelected := (i == m.selectedIndex) && (m.selectedIndex >= 0)
			st := m.hostStatus[h.ID]
			var dot string
			var statusLabel string
			switch st {
			case HostStatusOnline:
				dot = lipgloss.NewStyle().Foreground(ColorSuccess).Render("●")
			case HostStatusOffline:
				dot = lipgloss.NewStyle().Foreground(ColorDanger).Render("✖")
				statusLabel = lipgloss.NewStyle().Foreground(ColorDanger).Render(" (" + i18n.T("host_offline") + ")")
			case HostStatusConnecting:
				if isSelected {
					dot = lipgloss.NewStyle().Foreground(ColorWarning).Render("◌")
					statusLabel = lipgloss.NewStyle().Foreground(ColorWarning).Render(" (" + i18n.T("host_connecting") + ")")
				} else {
					dot = lipgloss.NewStyle().Foreground(ColorMuted).Render("◌")
				}
			default:
				if _, hasErr := m.errors[h.ID]; hasErr {
					dot = lipgloss.NewStyle().Foreground(ColorDanger).Render("✖")
					statusLabel = lipgloss.NewStyle().Foreground(ColorDanger).Render(" (" + i18n.T("host_offline") + ")")
				} else {
					dot = lipgloss.NewStyle().Foreground(ColorMuted).Render("◌")
				}
			}

			cpuThresh := 85.0
			ramThresh := 90.0
			diskThresh := 90.0
			if m.cfg != nil {
				if m.cfg.Telemetry.CPUThreshold > 0 {
					cpuThresh = m.cfg.Telemetry.CPUThreshold
				}
				if m.cfg.Telemetry.RAMThreshold > 0 {
					ramThresh = m.cfg.Telemetry.RAMThreshold
				}
				if m.cfg.Telemetry.DiskThreshold > 0 {
					diskThresh = m.cfg.Telemetry.DiskThreshold
				}
			}

			// Resource Alert Threshold check (Configured Thresholds with 80% Smart Warning)
			var alertBadge string
			if metric, ok := m.metrics[h.ID]; ok && metric != nil {
				memPct := 0.0
				if metric.MemoryTotal > 0 {
					memPct = (float64(metric.MemoryUsed) / float64(metric.MemoryTotal)) * 100.0
				}
				diskPct := 0.0
				if metric.DiskTotal > 0 {
					diskPct = (float64(metric.DiskUsed) / float64(metric.DiskTotal)) * 100.0
				}
				if metric.CPUPercent >= cpuThresh || memPct >= ramThresh || diskPct >= diskThresh {
					alertBadge = lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Blink(true).Render("🔥 ")
				} else if metric.CPUPercent >= (cpuThresh*0.8) || memPct >= (ramThresh*0.8) || diskPct >= (diskThresh*0.8) {
					alertBadge = lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render("⚠️ ")
				}
			}

			var hostLine string
			hostText := fmt.Sprintf("%s %s%s (%s)%s", dot, alertBadge, h.Name, h.Address, statusLabel)
			if isSelected {
				hostLine = SelectedHostStyle.Width(width - 6).Render(hostText)
			} else {
				hostLine = lipgloss.NewStyle().Foreground(ColorMuted).Width(width - 6).Render(hostText)
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

func (m *Model) renderLeftTelemetryPane(width, height int) string {
	var b strings.Builder

	contentWidth := width - 4
	if contentWidth < 18 {
		contentWidth = 18
	}

	titleColor := ColorPrimary
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(titleColor).Render("📊 텔레메트리 콕핏") + "\n\n")

	if len(m.hosts) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(m.hosts) {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("select_host_hint")))
		return PaneStyle.Width(width).Height(height).Render(b.String())
	}

	selectedHost := m.hosts[m.selectedIndex]
	metric, hasMetric := m.metrics[selectedHost.ID]
	err, hasErr := m.errors[selectedHost.ID]
	sysInfo := m.sysInfos[selectedHost.ID]

	hostNameTrunc := runewidth.Truncate("🖥️ "+selectedHost.Name, contentWidth, "…")
	hostAddrTrunc := runewidth.Truncate("   "+selectedHost.Address, contentWidth, "…")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render(hostNameTrunc) + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(hostAddrTrunc) + "\n")

	if sysInfo != nil {
		distroTrunc := runewidth.Truncate("🐧 "+sysInfo.OSDistro, contentWidth, "…")
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(distroTrunc) + "\n")
		b.WriteString(fmt.Sprintf("⚡ Cores: %d  | ⏱ %s\n", sysInfo.CPUCores, formatUltraCompactUptime(sysInfo.Uptime)))
	}
	b.WriteString("\n")

	st := m.hostStatus[selectedHost.ID]
	if st == HostStatusConnecting {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorWarning).Render("⏳ 접속 확인 중...") + "\n\n")
	} else if hasErr && err != nil {
		errStr := runewidth.Truncate("❌ Err: "+err.Error(), contentWidth, "…")
		b.WriteString(lipgloss.NewStyle().Foreground(ColorDanger).Render(errStr) + "\n\n")
	} else if hasMetric && metric != nil {
		barWidth := contentWidth
		if barWidth < 10 {
			barWidth = 10
		}

		cpuThresh := 85.0
		ramThresh := 90.0
		diskThresh := 90.0
		if m.cfg != nil {
			if m.cfg.Telemetry.CPUThreshold > 0 {
				cpuThresh = m.cfg.Telemetry.CPUThreshold
			}
			if m.cfg.Telemetry.RAMThreshold > 0 {
				ramThresh = m.cfg.Telemetry.RAMThreshold
			}
			if m.cfg.Telemetry.DiskThreshold > 0 {
				diskThresh = m.cfg.Telemetry.DiskThreshold
			}
		}

		// 1. CPU
		cpuBadge := ""
		valStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
		if metric.CPUPercent >= cpuThresh {
			cpuBadge = lipgloss.NewStyle().Bold(true).Background(ColorDanger).Foreground(lipgloss.Color("#FFFFFF")).Render(" 🔥 OVERLOAD ") + " "
			valStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorDanger)
		} else if metric.CPUPercent >= (cpuThresh * 0.8) {
			cpuBadge = lipgloss.NewStyle().Bold(true).Background(ColorWarning).Foreground(lipgloss.Color("#000000")).Render(" ⚠️ WARN ") + " "
			valStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorWarning)
		}
		cpuLabel := fmt.Sprintf("%s💻 %s: %s", cpuBadge, i18n.T("cpu_usage"), valStyle.Render(fmt.Sprintf("%.1f%%", metric.CPUPercent)))
		b.WriteString(cpuLabel + "\n")
		b.WriteString(renderBarWithThreshold(metric.CPUPercent, barWidth, cpuThresh) + "\n")

		// 2. RAM
		var memPercent float64
		if metric.MemoryTotal > 0 {
			memPercent = (float64(metric.MemoryUsed) / float64(metric.MemoryTotal)) * 100.0
		}
		memBadge := ""
		memValStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
		if memPercent >= ramThresh {
			memBadge = lipgloss.NewStyle().Bold(true).Background(ColorDanger).Foreground(lipgloss.Color("#FFFFFF")).Render(" 🔥 DANGER ") + " "
			memValStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorDanger)
		} else if memPercent >= (ramThresh * 0.8) {
			memBadge = lipgloss.NewStyle().Bold(true).Background(ColorWarning).Foreground(lipgloss.Color("#000000")).Render(" ⚠️ WARN ") + " "
			memValStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorWarning)
		}
		memLabel := fmt.Sprintf("%s🧠 %s: %s (%s/%s)",
			memBadge,
			i18n.T("mem_usage"),
			memValStyle.Render(fmt.Sprintf("%.0f%%", memPercent)),
			formatUltraCompactBytes(metric.MemoryUsed),
			formatUltraCompactBytes(metric.MemoryTotal),
		)
		b.WriteString(memLabel + "\n")
		b.WriteString(renderBarWithThreshold(memPercent, barWidth, ramThresh) + "\n")

		// 3. DISK
		var diskPercent float64
		if metric.DiskTotal > 0 {
			diskPercent = (float64(metric.DiskUsed) / float64(metric.DiskTotal)) * 100.0
		}
		diskBadge := ""
		diskValStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
		if diskPercent >= diskThresh {
			diskBadge = lipgloss.NewStyle().Bold(true).Background(ColorDanger).Foreground(lipgloss.Color("#FFFFFF")).Render(" 🔥 FULL ") + " "
			diskValStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorDanger)
		} else if diskPercent >= (diskThresh * 0.8) {
			diskBadge = lipgloss.NewStyle().Bold(true).Background(ColorWarning).Foreground(lipgloss.Color("#000000")).Render(" ⚠️ WARN ") + " "
			diskValStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorWarning)
		}
		diskLabel := fmt.Sprintf("%s💾 %s: %s (%s/%s)",
			diskBadge,
			i18n.T("disk_usage"),
			diskValStyle.Render(fmt.Sprintf("%.0f%%", diskPercent)),
			formatUltraCompactBytes(metric.DiskUsed),
			formatUltraCompactBytes(metric.DiskTotal),
		)
		b.WriteString(diskLabel + "\n")
		b.WriteString(renderBarWithThreshold(diskPercent, barWidth, diskThresh) + "\n")

		// 4. NETWORK
		b.WriteString(fmt.Sprintf("🌐 %s:\n", i18n.T("net_traffic")))
		b.WriteString(fmt.Sprintf("   📥 Rx: %s | 📤 Tx: %s\n\n", formatNetworkSpeed(metric.NetRxBytes), formatNetworkSpeed(metric.NetTxBytes)))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("📡 텔레메트리 수집 중...") + "\n\n")
	}

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render("[F5] 서버목록으로 전환"))

	paneBorder := PaneStyle
	if m.activePane == PaneHostList {
		paneBorder = ActivePaneStyle
	}

	return paneBorder.Width(width).Height(height).Render(b.String())
}

func formatUltraCompactUptime(u string) string {
	u = strings.TrimPrefix(u, "up ")
	u = strings.ReplaceAll(u, " weeks", "w")
	u = strings.ReplaceAll(u, " week", "w")
	u = strings.ReplaceAll(u, " days", "d")
	u = strings.ReplaceAll(u, " day", "d")
	u = strings.ReplaceAll(u, " hours", "h")
	u = strings.ReplaceAll(u, " hour", "h")
	parts := strings.Split(u, ",")
	if len(parts) > 2 {
		return strings.TrimSpace(parts[0]) + " " + strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(u)
}

func formatUltraCompactBytes(b uint64) string {
	gb := float64(b) / (1024 * 1024 * 1024)
	if gb >= 1.0 {
		return fmt.Sprintf("%.1fG", gb)
	}
	mb := float64(b) / (1024 * 1024)
	return fmt.Sprintf("%.0fM", mb)
}

func renderBarWithThreshold(percent float64, width int, thresh float64) string {
	if width < 4 {
		width = 4
	}
	filled := int((percent / 100.0) * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	var color lipgloss.Color
	switch {
	case percent >= thresh:
		color = ColorDanger
	case percent >= (thresh * 0.8):
		color = ColorWarning
	default:
		color = ColorSuccess
	}
	return lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(ColorMuted).Render(strings.Repeat("░", width-filled))
}
