package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// renderTelemetryDeck renders the top telemetry dashboard with CPU, RAM, Disk and Network cards.
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
