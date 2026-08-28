package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderHostListPane renders the left server list panel with status indicators and grouping.
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
