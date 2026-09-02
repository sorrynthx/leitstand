package tui

import (
	"fmt"
	"leitstand/internal/storage"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func (m *Model) renderConsolePane(width, height int) string {
	if len(m.hosts) == 0 {
		return m.renderWelcomePanel(width, height)
	}

	if m.selectedIndex < 0 || m.selectedIndex >= len(m.hosts) {
		m.selectedIndex = 0
	}

	curHost := m.hosts[m.selectedIndex]
	st := m.hostStatus[curHost.ID]
	hts := m.GetOrCreateHostTabs(curHost.ID, curHost.Name)
	activeTab := hts.ActiveTab()

	var b strings.Builder

	b.WriteString(m.renderConsoleHeader(width, curHost, hts, activeTab) + "\n")
	b.WriteString(m.renderConsoleBody(width, height, curHost, activeTab, st))

	paneBorder := PaneStyle
	if m.activePane == PaneConsole {
		paneBorder = ActivePaneStyle
	}

	return paneBorder.Width(width).Height(height).Render(b.String())
}

func (m *Model) renderRemoteConsole(width, height int) string {
	return m.renderConsolePane(width, height)
}

func (m *Model) renderConsoleHeader(width int, curHost *storage.Host, hts *HostTabs, activeTab *ConsoleTab) string {
	titleStyle := TitleStyle
	if m.activePane == PaneConsole {
		titleStyle = TitleStyle.Copy().Bold(true).Foreground(ColorPrimary)
	}

	hostTitle := fmt.Sprintf("💻 %s (%s)", curHost.Name, curHost.Address)
	if activeTab != nil && activeTab.IsRoot {
		hostTitle += " 👑 [ROOT]"
	}

	headerText := titleStyle.Render(hostTitle)
	tabBar := m.renderTabBar(hts, width-lipgloss.Width(headerText)-4)

	gapWidth := width - lipgloss.Width(headerText) - lipgloss.Width(tabBar) - 4
	if gapWidth < 1 {
		gapWidth = 1
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, headerText, strings.Repeat(" ", gapWidth), tabBar)
}

func (m *Model) renderTabBar(hts *HostTabState, maxWidth int) string {
	if hts == nil || len(hts.Tabs) == 0 {
		return ""
	}

	btnText := "+ [Ctrl+N]"
	btnWidth := runewidth.StringWidth(btnText) + 2
	availForTabs := maxWidth - btnWidth - 2
	if availForTabs < 15 {
		availForTabs = 15
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
			title = fmt.Sprintf("%d: bash", i+1)
		}
		if tab.IsRoot && !strings.Contains(title, "👑") {
			title = "👑 " + title
		}
		if tab.IsStreaming && !strings.Contains(title, "🔴") {
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

	totalTabs := len(renderedTabs)
	startIdx := 0
	endIdx := totalTabs - 1

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
		curW := renderedTabs[active].width + 4

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
