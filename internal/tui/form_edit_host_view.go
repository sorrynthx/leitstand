package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	ModalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorSecondary).
			Padding(0, 1)
)

func (f *HostForm) View(width, height int) string {
	f.width = width
	f.height = height

	if f.filePicker != nil {
		return f.filePicker.View(width, height)
	}

	var title string
	if f.isEditMode {
		title = i18n.T("form_edit_title")
	} else {
		title = i18n.T("form_add_title")
	}

	var b strings.Builder
	b.WriteString(HeaderStyle.Render(fmt.Sprintf("  %s  ", title)) + "\n\n")

	order := f.getFieldOrder()

	for idx, fieldIdx := range order {
		isFocused := (idx == f.focusIndex)

		if fieldIdx == -1 {
			labelStyle := lipgloss.NewStyle().Foreground(ColorMuted)
			if isFocused {
				labelStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
			}
			b.WriteString(labelStyle.Render(i18n.T("form_auth_method")))

			passBtnStr := i18n.T("form_auth_pass")
			keyBtnStr := i18n.T("form_auth_key")

			var passBtn, keyBtn string
			if f.authType == AuthTypePassword {
				passBtn = lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Background(lipgloss.Color("#1B2A1E")).Render(passBtnStr + " ●")
				keyBtn = lipgloss.NewStyle().Foreground(ColorMuted).Render(keyBtnStr + " ○")
			} else {
				passBtn = lipgloss.NewStyle().Foreground(ColorMuted).Render(passBtnStr + " ○")
				keyBtn = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Background(lipgloss.Color("#1A2634")).Render(keyBtnStr + " ●")
			}

			if isFocused {
				b.WriteString(passBtn + "  " + keyBtn + "  " + lipgloss.NewStyle().Foreground(ColorWarning).Render(i18n.T("form_switch_hint")) + "\n\n")
			} else {
				b.WriteString(passBtn + "  " + keyBtn + "\n\n")
			}
			continue
		}

		labelKey := ""
		phKey := ""
		switch fieldIdx {
		case 0:
			labelKey = "form_label_name"
			phKey = "form_ph_name"
		case 1:
			labelKey = "form_label_addr"
			phKey = "form_ph_addr"
		case 2:
			labelKey = "form_label_port"
			phKey = "form_ph_port"
		case 3:
			labelKey = "form_label_user"
			phKey = "form_ph_user"
		case 4:
			labelKey = "form_label_pass"
			phKey = "form_ph_pass"
		case 5:
			labelKey = "form_label_key"
			phKey = "form_ph_key"
		case 6:
			labelKey = "form_label_phrase"
			phKey = "form_ph_phrase"
		case 7:
			labelKey = "form_label_group"
			phKey = "form_ph_group"
		}

		if labelKey != "" {
			f.inputs[fieldIdx].Prompt = i18n.T(labelKey)
			f.inputs[fieldIdx].Placeholder = i18n.T(phKey)
			f.inputs[fieldIdx].Width = 42
		}

		inp := f.inputs[fieldIdx]

		if isFocused {
			inp.PromptStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
			inp.TextStyle = lipgloss.NewStyle().Foreground(ColorText)
		} else {
			inp.PromptStyle = lipgloss.NewStyle().Foreground(ColorMuted)
			inp.TextStyle = lipgloss.NewStyle().Foreground(ColorMuted)
		}

		b.WriteString(inp.View())
		if isFocused && fieldIdx == 5 && f.authType == AuthTypeKey {
			b.WriteString("  " + lipgloss.NewStyle().Foreground(ColorWarning).Bold(true).Render(i18n.T("form_browse_btn")))
		}
		b.WriteString("\n\n")
	}

	badges := RenderSecurityBadges(f.inputs...)
	if badges != "" {
		b.WriteString("\n" + badges)
	} else {
		b.WriteString("\n")
	}

	if f.errMessage != "" {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render("❌ "+f.errMessage) + "\n\n")
	} else {
		b.WriteString("\n")
	}

	helpText := lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("form_help_hint"))
	b.WriteString(helpText)

	if width <= 0 {
		width = 100
	}
	if height <= 0 {
		height = 30
	}

	modalWidth := width - 4
	modalHeight := height - 2
	if modalWidth < 65 {
		modalWidth = 65
	}
	if modalHeight < 15 {
		modalHeight = 15
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 3).
		Width(modalWidth).
		Height(modalHeight)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, boxStyle.Render(b.String()))
}
