package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"leitstand/internal/quickcmd"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (d *RunbookDrawer) renderItemsList(b *strings.Builder, items []quickcmd.CommandItem, drawerWidth, drawerHeight int) {
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
		if item.CategoryKey != lastCat {
			lastCat = item.CategoryKey
			catHeader := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("── " + i18n.T(item.CategoryKey) + " ──")
			b.WriteString(catHeader + "\n")
		}

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
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Background(lipgloss.Color("#1B2A32")).Render(lineText) + "\n")
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#ECEFF1")).Render(lineText) + "\n")
		}
	}

	if endIdx < len(items) {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("  ▼ (%d more below)", len(items)-endIdx)) + "\n")
	}
	b.WriteString("\n")

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
			descBox := lipgloss.NewStyle().Foreground(ColorText).Width(previewWidth).Render("💡 " + desc + "  " + lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("drawer_hint_close")))
			b.WriteString(cmdBox + "\n" + descBox + "\n\n")
		} else {
			cmdBox := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Background(lipgloss.Color("#151820")).Padding(0, 1).Width(previewWidth).Render("❯ " + selected.Command)
			descBox := lipgloss.NewStyle().Foreground(lipgloss.Color("#B0BEC5")).Width(previewWidth).Render("💡 " + desc + "  " + lipgloss.NewStyle().Foreground(ColorSuccess).Render(i18n.T("drawer_hint_inject")))
			b.WriteString(cmdBox + "\n" + descBox + "\n\n")
		}
	}
}
