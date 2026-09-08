package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (d *RunbookDrawer) renderCustomOverlay(width, height int) string {
	var b strings.Builder

	if d.isDeleting {
		titleText := lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render("── " + i18n.T("custom_cmd_modal_del_title") + " ──")
		b.WriteString(titleText + "\n\n")

		if len(d.customCmds) > 0 && d.selectedIndex < len(d.customCmds) {
			cur := d.customCmds[d.selectedIndex]
			itemText := fmt.Sprintf("• %s (%s)", cur.Title, cur.Command)
			b.WriteString(lipgloss.NewStyle().Foreground(ColorText).Bold(true).Render(itemText) + "\n\n")
		}

		hints := lipgloss.NewStyle().Foreground(ColorWarning).Render(i18n.T("custom_cmd_del_hints"))
		b.WriteString(hints + "\n")
		return b.String()
	}

	titleText := i18n.T("custom_cmd_modal_add_title")
	if d.isEditing {
		titleText = i18n.T("custom_cmd_modal_edit_title")
	}
	subHeader := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("── " + titleText + " ──")
	b.WriteString(subHeader + "\n\n")

	labels := []string{
		i18n.T("custom_cmd_title_label"),
		i18n.T("custom_cmd_cmd_label"),
		i18n.T("custom_cmd_cat_label"),
		i18n.T("custom_cmd_desc_label"),
	}

	labelWidth := 18
	spacing := "\n\n"
	if height < 26 {
		spacing = "\n"
	}

	for i, input := range d.customInputs {
		lblStyle := lipgloss.NewStyle().Bold(true).Width(labelWidth)
		if i == d.customFocused {
			lblStyle = lblStyle.Foreground(ColorPrimary)
		} else {
			lblStyle = lblStyle.Foreground(lipgloss.Color("#B0BEC5"))
		}

		line := lipgloss.JoinHorizontal(
			lipgloss.Left,
			lblStyle.Render(labels[i]+": "),
			input.View(),
		)
		b.WriteString(line + spacing)
	}

	if d.errMessage != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render(d.errMessage) + "\n\n")
	}

	hints := lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("custom_cmd_form_hints"))
	b.WriteString(hints + "\n")

	return b.String()
}
