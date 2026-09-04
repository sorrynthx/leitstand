package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"leitstand/internal/storage"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderConsoleBody(width, height int, curHost *storage.Host, activeTab *ConsoleTab, st HostStatus) string {
	var b strings.Builder
	cwd := "~"
	if activeTab != nil && activeTab.CWD != "" {
		cwd = activeTab.CWD
	}

	if st == HostStatusOffline {
		m.consoleInput.Placeholder = i18n.T("console_offline_ph")
		m.consoleInput.Prompt = lipgloss.NewStyle().Foreground(ColorDanger).Render(fmt.Sprintf("[%s] ❯ ", i18n.T("status_offline")))
	} else if st == HostStatusConnecting {
		m.consoleInput.Placeholder = i18n.T("console_connecting_ph")
		m.consoleInput.Prompt = lipgloss.NewStyle().Foreground(ColorWarning).Render(fmt.Sprintf("[%s] ❯ ", i18n.T("status_connecting")))
	} else if activeTab != nil && activeTab.IsStreaming {
		m.consoleInput.Placeholder = i18n.T("console_streaming_ph")
		m.consoleInput.Prompt = lipgloss.NewStyle().Foreground(ColorDanger).Render("[🔴 Streaming] ❯ ")
	} else if activeTab != nil && activeTab.IsRoot {
		m.consoleInput.Placeholder = i18n.T("console_root_ph")
		m.consoleInput.Prompt = lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(fmt.Sprintf("[root@%s:%s]# ", curHost.Name, cwd))
	} else {
		m.consoleInput.Placeholder = i18n.T("console_placeholder")
		m.consoleInput.Prompt = lipgloss.NewStyle().Foreground(ColorPrimary).Render(fmt.Sprintf("[%s] ❯ ", cwd))
	}

	var statusBanner string
	if st == HostStatusConnecting {
		statusBanner = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Background(lipgloss.Color("#1E222B")).
			Padding(0, 1).
			Render(fmt.Sprintf("⏳ [%s] '%s' (%s:%d)", i18n.T("status_connecting"), curHost.Name, curHost.Address, curHost.Port))
	} else if st != HostStatusOnline {
		statusBanner = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF7B72")).
			Background(lipgloss.Color("#1E222B")).
			Padding(0, 1).
			Render(i18n.Tf("prompt_offline_host", i18n.T("status_offline"), curHost.Name, curHost.Address, curHost.Port))
	}

	if statusBanner != "" {
		b.WriteString(padToWidth(statusBanner, width-4) + "\n")
	}

	if m.showAICopilot && m.aiCopilot != nil {
		aiBar := m.aiCopilot.ViewInline(width - 4)
		aiBarH := lipgloss.Height(aiBar)
		vpHeight := height - aiBarH
		if statusBanner != "" {
			vpHeight -= lipgloss.Height(statusBanner)
		}
		if vpHeight < 2 {
			vpHeight = 2
		}

		if activeTab != nil {
			activeTab.Viewport.Width = width - 4
			activeTab.Viewport.Height = vpHeight
			b.WriteString(activeTab.Viewport.View() + "\n")
		} else {
			m.viewport.Width = width - 4
			m.viewport.Height = vpHeight
			b.WriteString(m.viewport.View() + "\n")
		}

		b.WriteString(aiBar)
		return b.String()
	}

	vpHeight := height - 2
	if statusBanner != "" {
		vpHeight--
	}
	if vpHeight < 2 {
		vpHeight = 2
	}

	if activeTab != nil {
		activeTab.Viewport.Width = width - 4
		activeTab.Viewport.Height = vpHeight
		b.WriteString(activeTab.Viewport.View() + "\n")
	} else {
		m.viewport.Width = width - 4
		m.viewport.Height = vpHeight
		b.WriteString(m.viewport.View() + "\n")
	}

	m.consoleInput.Width = width - 6
	b.WriteString(m.consoleInput.View())

	return b.String()
}
