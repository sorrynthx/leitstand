package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (vf *VaultForm) View(termWidth, termHeight int) string {
	if termWidth <= 0 {
		termWidth = 100
	}
	if termHeight <= 0 {
		termHeight = 30
	}

	var b strings.Builder

	titleText := i18n.T("vault_modal_title")
	if vf.modalType == VaultModalUnlock {
		titleText = i18n.T("vault_modal_unlock")
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		Render(titleText)
	b.WriteString(title + "\n\n")

	var langBadges []string
	for i, opt := range i18n.SupportedLangs {
		if i == vf.selectedLangIndex {
			langBadges = append(langBadges, lipgloss.NewStyle().
				Bold(true).
				Background(ColorPrimary).
				Foreground(ColorBg).
				Padding(0, 1).
				Render(fmt.Sprintf("[F%d] %s", i+1, opt.Label)))
		} else {
			langBadges = append(langBadges, lipgloss.NewStyle().
				Foreground(ColorMuted).
				Padding(0, 1).
				Render(fmt.Sprintf("[F%d] %s", i+1, opt.Label)))
		}
	}
	b.WriteString(strings.Join(langBadges, " ") + "\n\n")

	if vf.modalType == VaultModalInit {
		info := lipgloss.NewStyle().
			Foreground(ColorFg).
			Render(i18n.T("vault_help_setup") + "\n(Keys are encrypted with AES-256-GCM & Argon2id)\n")
		b.WriteString(info + "\n")
		if len(vf.inputs) >= 2 {
			vf.inputs[0].Placeholder = i18n.T("vault_ph")
			vf.inputs[1].Placeholder = i18n.T("vault_confirm_ph")
		}
	} else {
		info := lipgloss.NewStyle().
			Foreground(ColorFg).
			Render(i18n.T("vault_help_unlock") + "\n")
		b.WriteString(info + "\n")
		if len(vf.inputs) >= 1 {
			vf.inputs[0].Placeholder = i18n.T("vault_ph")
		}
	}

	b.WriteString(RenderSecurityBadges(vf.inputs...))

	if vf.errMessage != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render("❌ "+vf.errMessage) + "\n\n")
	}

	for _, input := range vf.inputs {
		b.WriteString("🔑 " + input.View() + "\n\n")
	}

	var hint string
	if vf.modalType == VaultModalInit {
		hint = lipgloss.NewStyle().Foreground(ColorMuted).Render("[F1/F2/F3 / ◄ ►] Switch Lang    [Tab] Next Field    [Enter] Submit    [Esc] Exit")
	} else {
		hint = lipgloss.NewStyle().Foreground(ColorMuted).Render("[F1/F2/F3 / ◄ ►] Switch Lang    [Enter] Unlock    [Esc] Exit")
	}
	b.WriteString(hint)

	return RenderModalContainer(b.String(), 78, ColorPrimary, termWidth, termHeight)
}
