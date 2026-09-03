package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"leitstand/internal/quickcmd"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RunbookDrawer provides an OS-aware quick command runbook.
type RunbookDrawer struct {
	activeTab      quickcmd.OSTab
	selectedIndex  int
	detectedDistro string
}

// NewRunbookDrawer creates a new runbook drawer, auto-focusing on the detected host OS.
func NewRunbookDrawer(detectedDistro string) *RunbookDrawer {
	autoTab := quickcmd.DetectOSTab(detectedDistro)
	return &RunbookDrawer{
		activeTab:      autoTab,
		selectedIndex:  0,
		detectedDistro: detectedDistro,
	}
}

// Update handles key navigation inside the Runbook Drawer.
// Returns (done bool, chosenCommand string, cmd tea.Cmd)
func (d *RunbookDrawer) Update(msg tea.Msg) (bool, string, tea.Cmd) {
	items := quickcmd.Catalog[d.activeTab]

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "?", "ctrl+k":
			return true, "", nil // Close drawer

		case "1", "2", "3", "4", "5", "6":
			tabIdx := int(msg.String()[0] - '1')
			d.switchTab(quickcmd.OSTab(tabIdx))
			return false, "", nil

		case "tab", "right", "l":
			nextTab := (d.activeTab + 1) % quickcmd.OSTabCount
			d.switchTab(nextTab)
			return false, "", nil

		case "shift+tab", "left", "h":
			prevTab := d.activeTab - 1
			if prevTab < 0 {
				prevTab = quickcmd.OSTabCount - 1
			}
			d.switchTab(prevTab)
			return false, "", nil

		case "up", "k":
			if len(items) > 0 {
				d.selectedIndex--
				if d.selectedIndex < 0 {
					d.selectedIndex = len(items) - 1
				}
			}
			return false, "", nil

		case "down", "j":
			if len(items) > 0 {
				d.selectedIndex++
				if d.selectedIndex >= len(items) {
					d.selectedIndex = 0
				}
			}
			return false, "", nil

		case "pgup", "pageup", "ctrl+u", "ctrl+b":
			if len(items) > 0 {
				d.selectedIndex -= 5
				if d.selectedIndex < 0 {
					d.selectedIndex = 0
				}
			}
			return false, "", nil

		case "pgdown", "pagedown", "ctrl+d":
			if len(items) > 0 {
				d.selectedIndex += 5
				if d.selectedIndex >= len(items) {
					d.selectedIndex = len(items) - 1
				}
			}
			return false, "", nil

		case "enter":
			if d.activeTab == quickcmd.OSTabShortcuts {
				return true, "", nil
			}
			if len(items) > 0 && d.selectedIndex >= 0 && d.selectedIndex < len(items) {
				return true, items[d.selectedIndex].Command, nil
			}
			return true, "", nil
		}
	}

	return false, "", nil
}

func (d *RunbookDrawer) switchTab(tab quickcmd.OSTab) {
	d.activeTab = tab
	d.selectedIndex = 0
}

// View renders the right-side slide drawer.
func (d *RunbookDrawer) View(drawerWidth, drawerHeight int) string {
	var b strings.Builder

	// Header Title
	title := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("drawer_title"))
	b.WriteString(title + "\n")

	// Detected OS Badge
	if d.detectedDistro != "" {
		detectedLine := lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("drawer_auto_os")) + " " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(d.detectedDistro)
		b.WriteString(detectedLine + "\n\n")
	} else {
		b.WriteString("\n")
	}

	// Tab Bar
	var tabButtons []string
	for i, tInfo := range quickcmd.Tabs {
		var btnStyle lipgloss.Style
		if tInfo.Tab == d.activeTab {
			btnStyle = lipgloss.NewStyle().Bold(true).Background(ColorPrimary).Foreground(lipgloss.Color("#000000")).Padding(0, 1)
		} else {
			btnStyle = lipgloss.NewStyle().Bold(true).Background(ColorBorder).Foreground(lipgloss.Color("#B0BEC5")).Padding(0, 1)
		}
		tabButtons = append(tabButtons, btnStyle.Render(fmt.Sprintf("[%d] %s", i+1, tInfo.Badge)))
	}
	b.WriteString(strings.Join(tabButtons, " ") + "\n\n")

	// Items list with smooth sliding window scroll
	items := quickcmd.Catalog[d.activeTab]

	maxVisible := 7
	if drawerHeight > 28 {
		maxVisible = 10
	} else if drawerHeight < 20 {
		maxVisible = 5
	}

	startIdx := 0
	if d.selectedIndex >= maxVisible {
		startIdx = d.selectedIndex - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(items) {
		endIdx = len(items)
	}

	if startIdx > 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("  ▲ (%d more above)", startIdx)) + "\n")
	}

	var lastCat string
	for i := startIdx; i < endIdx; i++ {
		item := items[i]

		// Category Header
		if item.CategoryKey != lastCat {
			lastCat = item.CategoryKey
			catHeader := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("── " + i18n.T(item.CategoryKey) + " ──")
			b.WriteString(catHeader + "\n")
		}

		// Command Item
		cursor := "  "
		if i == d.selectedIndex {
			cursor = "▶ "
		}

		itemTitle := i18n.T(item.TitleKey)
		var lineText string
		if d.activeTab == quickcmd.OSTabShortcuts {
			keyBadge := lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Width(12).Render(item.Command)
			lineText = fmt.Sprintf("%s%s %s", cursor, keyBadge, itemTitle)
		} else {
			lineText = fmt.Sprintf("%s%s", cursor, itemTitle)
		}

		if i == d.selectedIndex {
			itemLine := lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Background(lipgloss.Color("#1B2A32")).Render(lineText)
			b.WriteString(itemLine + "\n")
		} else {
			itemLine := lipgloss.NewStyle().Foreground(lipgloss.Color("#ECEFF1")).Render(lineText)
			b.WriteString(itemLine + "\n")
		}
	}

	if endIdx < len(items) {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("  ▼ (%d more below)", len(items)-endIdx)) + "\n")
	}

	b.WriteString("\n")

	// Bottom Preview & Description Box of selected command
	if len(items) > 0 && d.selectedIndex < len(items) {
		selected := items[d.selectedIndex]
		desc := i18n.T(selected.DescKey)
		previewWidth := drawerWidth - 6
		if previewWidth < 30 {
			previewWidth = 30
		}

		if d.activeTab == quickcmd.OSTabShortcuts {
			headerText := fmt.Sprintf("⌨️ [%s] ── %s", selected.Command, i18n.T(selected.TitleKey))
			cmdBox := lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Background(lipgloss.Color("#151820")).Padding(0, 1).Width(previewWidth).Render(headerText)
			descBox := lipgloss.NewStyle().Foreground(ColorText).Width(previewWidth).Render("💡 " + desc + "  " + lipgloss.NewStyle().Foreground(ColorMuted).Render("(Enter/Esc: 닫기)"))
			b.WriteString(cmdBox + "\n" + descBox + "\n\n")
		} else {
			cmdBox := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Background(lipgloss.Color("#151820")).Padding(0, 1).Width(previewWidth).Render("❯ " + selected.Command)
			descBox := lipgloss.NewStyle().Foreground(lipgloss.Color("#B0BEC5")).Width(previewWidth).Render("💡 " + desc + "  " + lipgloss.NewStyle().Foreground(ColorSuccess).Render("(Enter: 콘솔 주입)"))
			b.WriteString(cmdBox + "\n" + descBox + "\n\n")
		}
	}

	// Footer Navigation Hint
	footerHint := lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("drawer_hint"))
	b.WriteString(footerHint)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 2).
		Width(drawerWidth).
		Height(drawerHeight)

	return boxStyle.Render(b.String())
}
