package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func renderBar(percent float64, width int) string {
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
	case percent > 85.0:
		color = ColorDanger
	case percent > 50.0:
		color = ColorWarning
	default:
		color = ColorSuccess
	}
	return lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(ColorMuted).Render(strings.Repeat("░", width-filled))
}

func (m *Model) renderTelemetryDrawer(width, height int) string {
	var b strings.Builder

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 2).
		Width(width).
		Height(height)

	if len(m.hosts) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(m.hosts) {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("select_host_hint")))
		return boxStyle.Render(b.String())
	}

	selectedHost := m.hosts[m.selectedIndex]
	metric, hasMetric := m.metrics[selectedHost.ID]
	err, hasErr := m.errors[selectedHost.ID]
	sysInfo := m.sysInfos[selectedHost.ID]

	header := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("telemetry_drawer_title")) + "\n" +
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render("🖥️ "+selectedHost.Name) + " (" + selectedHost.Address + ":" + fmt.Sprintf("%d", selectedHost.Port) + ")\n\n"
	b.WriteString(header)

	if sysInfo != nil {
		specLine := fmt.Sprintf("🐧 %s: %s  |  ⚡ %s: %d  |  ⏱ %s: %s\n\n",
			i18n.T("sys_distro"),
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(sysInfo.OSDistro),
			i18n.T("sys_cores"),
			sysInfo.CPUCores,
			i18n.T("sys_uptime"),
			lipgloss.NewStyle().Foreground(ColorWarning).Render(formatCompactUptime(sysInfo.Uptime)),
		)
		b.WriteString(specLine)
	}

	st := m.hostStatus[selectedHost.ID]
	if st == HostStatusConnecting {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorWarning).Render("⏳ Connecting to server telemetry stream...") + "\n\n")
	} else if hasErr && err != nil {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorDanger).Render("❌ Connection Error: "+err.Error()) + "\n\n")
	} else if hasMetric && metric != nil {
		barLen := width - 40
		if barLen < 15 {
			barLen = 15
		}

		cpuGauge := renderBar(metric.CPUPercent, barLen)
		b.WriteString(fmt.Sprintf("💻 %s: %s  %.1f%%\n\n",
			i18n.T("cpu_usage"),
			cpuGauge,
			metric.CPUPercent,
		))

		var memPercent float64
		if metric.MemoryTotal > 0 {
			memPercent = (float64(metric.MemoryUsed) / float64(metric.MemoryTotal)) * 100.0
		}
		memGauge := renderBar(memPercent, barLen)
		b.WriteString(fmt.Sprintf("🧠 %s: %s  %.1f%% (%s / %s)\n\n",
			i18n.T("mem_usage"),
			memGauge,
			memPercent,
			formatBytes(metric.MemoryUsed),
			formatBytes(metric.MemoryTotal),
		))

		var diskPercent float64
		if metric.DiskTotal > 0 {
			diskPercent = (float64(metric.DiskUsed) / float64(metric.DiskTotal)) * 100.0
		}
		diskGauge := renderBar(diskPercent, barLen)
		b.WriteString(fmt.Sprintf("💾 %s: %s  %.1f%% (%s / %s)\n\n",
			i18n.T("disk_usage"),
			diskGauge,
			diskPercent,
			formatBytes(metric.DiskUsed),
			formatBytes(metric.DiskTotal),
		))

		b.WriteString(fmt.Sprintf("🌐 %s:  📥 Rx %s  |  📤 Tx %s\n\n",
			i18n.T("net_traffic"),
			formatNetworkSpeed(metric.NetRxBytes),
			formatNetworkSpeed(metric.NetTxBytes),
		))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("📡 Polling telemetry metrics...") + "\n\n")
	}

	footer := lipgloss.NewStyle().Foreground(ColorWarning).Render(i18n.T("telemetry_drawer_hint"))
	b.WriteString(footer)

	return boxStyle.Render(b.String())
}
