package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
