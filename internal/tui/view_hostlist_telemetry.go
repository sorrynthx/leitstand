package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func (m *Model) renderLeftTelemetryPane(width, height int) string {
	var b strings.Builder

	contentWidth := width - 4
	if contentWidth < 18 {
		contentWidth = 18
	}

	titleColor := ColorPrimary
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(titleColor).Render(i18n.T("telemetry_cockpit_title")) + "\n\n")

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
		b.WriteString(lipgloss.NewStyle().Foreground(ColorWarning).Render(i18n.T("host_checking_connection")) + "\n\n")
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
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("telemetry_collecting")) + "\n\n")
	}

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(i18n.T("status_switch_to_hostlist")))

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
