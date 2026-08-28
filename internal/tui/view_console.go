package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// renderTabBar renders the sliding window multi-tab bar with ◄/► scroll indicators and + [Ctrl+T] button.
func (m *Model) renderTabBar(hts *HostTabState, maxWidth int) string {
	if hts == nil || len(hts.Tabs) == 0 {
		return ""
	}

	btnText := "+ [Ctrl+T]"
	btnWidth := runewidth.StringWidth(btnText) + 2
	availForTabs := maxWidth - btnWidth - 2
	if availForTabs < 20 {
		availForTabs = 20
	}

	type renderedTab struct {
		rendered string
		width    int
		isActive bool
	}

	var renderedTabs []renderedTab
	for i, tab := range hts.Tabs {
		isActive := (i == hts.ActiveIndex)
		title := tab.Title
		if title == "" {
			title = fmt.Sprintf("%d:bash", i+1)
		}
		if tab.IsRoot {
			title = "👑" + title
		}
		if tab.IsStreaming {
			title += " 🔴"
		}

		var tabStyle lipgloss.Style
		if isActive {
			bgCol := ColorPrimary
			if tab.IsRoot {
				bgCol = ColorDanger
			}
			tabStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorBg).
				Background(bgCol).
				Padding(0, 1)
		} else {
			tabStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Background(ColorHighlight).
				Padding(0, 1)
		}

		rendered := tabStyle.Render(title)
		w := runewidth.StringWidth(title) + 2
		renderedTabs = append(renderedTabs, renderedTab{
			rendered: rendered,
			width:    w,
			isActive: isActive,
		})
	}

	// Calculate sliding window so activeIndex is always visible
	totalTabs := len(renderedTabs)
	startIdx := 0
	endIdx := totalTabs - 1

	// Check total width
	totalW := 0
	for _, rt := range renderedTabs {
		totalW += rt.width + 1
	}

	if totalW > availForTabs {
		active := hts.ActiveIndex
		if active < 0 {
			active = 0
		}
		if active >= totalTabs {
			active = totalTabs - 1
		}
		startIdx = active
		endIdx = active
		curW := renderedTabs[active].width + 4 // reserve space for ◄ and ►

		for {
			expanded := false
			if endIdx+1 < totalTabs && curW+renderedTabs[endIdx+1].width+1 <= availForTabs {
				endIdx++
				curW += renderedTabs[endIdx].width + 1
				expanded = true
			}
			if startIdx-1 >= 0 && curW+renderedTabs[startIdx-1].width+1 <= availForTabs {
				startIdx--
				curW += renderedTabs[startIdx].width + 1
				expanded = true
			}
			if !expanded {
				break
			}
		}
	}

	var visibleItems []string
	if startIdx > 0 {
		visibleItems = append(visibleItems, lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render("◄"))
	}
	for i := startIdx; i <= endIdx && i < len(renderedTabs); i++ {
		visibleItems = append(visibleItems, renderedTabs[i].rendered)
	}
	if endIdx < totalTabs-1 {
		visibleItems = append(visibleItems, lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render("►"))
	}

	newTabBtn := lipgloss.NewStyle().
		Foreground(ColorSecondary).
		Background(ColorHighlight).
		Padding(0, 1).
		Render(btnText)
	visibleItems = append(visibleItems, newTabBtn)

	return strings.Join(visibleItems, " ")
}

// renderRemoteConsole renders the interactive remote command console pane.
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
	hts := m.GetOrCreateHostTabs(curHost.ID, curHost.Name)
	activeTab := hts.ActiveTab()

	cwd := "~"
	if activeTab != nil && activeTab.CWD != "" {
		cwd = activeTab.CWD
	}

	// Render Tab Bar
	tabBar := m.renderTabBar(hts, width-4)
	b.WriteString(tabBar + "\n")

	st := m.hostStatus[curHost.ID]
	if st == HostStatusOffline {
		m.consoleInput.Placeholder = i18n.T("console_offline_ph")
		m.consoleInput.Prompt = lipgloss.NewStyle().Foreground(ColorDanger).Render("[Offline] ❯ ")
	} else if st == HostStatusConnecting {
		m.consoleInput.Placeholder = i18n.T("console_connecting_ph")
		m.consoleInput.Prompt = lipgloss.NewStyle().Foreground(ColorWarning).Render("[Connecting...] ❯ ")
	} else if activeTab != nil && activeTab.IsStreaming {
		m.consoleInput.Placeholder = "Streaming in progress... Press [Ctrl+C] to stop."
		m.consoleInput.Prompt = lipgloss.NewStyle().Foreground(ColorDanger).Render("[🔴 Streaming] ❯ ")
	} else if activeTab != nil && activeTab.IsRoot {
		m.consoleInput.Placeholder = "Type root command (e.g. systemctl restart, apt update, 'exit' to log out)..."
		m.consoleInput.Prompt = lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(fmt.Sprintf("[root@%s:%s]# ", curHost.Name, cwd))
	} else {
		m.consoleInput.Placeholder = i18n.T("console_placeholder")
		m.consoleInput.Prompt = lipgloss.NewStyle().Foreground(ColorPrimary).Render(fmt.Sprintf("[%s] ❯ ", cwd))
	}

	// Total inner lines = height. Title = 1 line, TabBar = 1 line, Input = 1 line, Viewport = height - 3
	vpHeight := height - 3
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

	// Command input prompt
	m.consoleInput.Width = width - 6
	b.WriteString(m.consoleInput.View())

	paneBorder := PaneStyle
	if m.activePane == PaneConsole {
		paneBorder = ActivePaneStyle
	}

	return paneBorder.Width(width).Height(height).Render(b.String())
}
