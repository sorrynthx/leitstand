package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"leitstand/internal/quickcmd"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the right-side slide drawer.
func (d *RunbookDrawer) View(drawerWidth, drawerHeight int) string {
	if d.filePicker != nil {
		return d.filePicker.View(drawerWidth, drawerHeight)
	}

	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("drawer_title"))
	b.WriteString(title + "\n")

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

	if d.isAdding || d.isEditing || d.isDeleting {
		b.WriteString(strings.Join(tabButtons, " ") + "\n\n")
		b.WriteString(d.renderCustomOverlay(drawerWidth, drawerHeight))
		boxStyle := lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(ColorPrimary).Padding(1, 2).Width(drawerWidth).Height(drawerHeight)
		return boxStyle.Render(b.String())
	}

	if d.detectedDistro != "" {
		detectedLine := lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("drawer_auto_os")) + " " +
			lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(d.detectedDistro)
		b.WriteString(detectedLine + "\n\n")
	} else {
		b.WriteString("\n")
	}

	b.WriteString(strings.Join(tabButtons, " ") + "\n\n")

	if d.bannerMessage != "" {
		b.WriteString(lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSuccess).
			Background(lipgloss.Color("#102A18")).
			Padding(0, 1).
			Width(drawerWidth - 6).
			Render("✨ "+d.bannerMessage) + "\n\n")
	}

	items := d.getItems()
	if len(items) == 0 && d.activeTab == quickcmd.OSTabCustom {
		emptyMsg := lipgloss.NewStyle().Foreground(ColorWarning).Padding(2, 1).Render(i18n.T("custom_cmd_empty"))
		b.WriteString(emptyMsg + "\n\n")
	} else {
		d.renderItemsList(&b, items, drawerWidth, drawerHeight)
	}

	var hintText string
	if d.activeTab == quickcmd.OSTabCustom {
		hintText = i18n.T("custom_cmd_hints")
	} else {
		hintText = i18n.T("drawer_hint")
	}
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(hintText))

	boxStyle := lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(ColorPrimary).Padding(1, 2).Width(drawerWidth).Height(drawerHeight)
	return boxStyle.Render(b.String())
}
