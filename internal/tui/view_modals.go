package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// RenderModalContainer renders a standardized double-bordered centered modal overlay.
func RenderModalContainer(body string, width int, borderColor lipgloss.Color, termWidth, termHeight int) string {
	if termWidth <= 0 {
		termWidth = 100
	}
	if termHeight <= 0 {
		termHeight = 30
	}
	if width <= 0 {
		width = 65
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(borderColor).
		Padding(1, 3).
		Width(width)

	content := boxStyle.Render(body)
	return lipgloss.Place(termWidth, termHeight, lipgloss.Center, lipgloss.Center, content)
}

// RenderSecurityBadges generates Caps Lock ON and Non-ASCII / Korean IME warning badges if applicable.
func RenderSecurityBadges(inputs ...textinput.Model) string {
	var b strings.Builder

	if IsCapsLockOn() {
		capsBadge := lipgloss.NewStyle().
			Bold(true).
			Background(ColorWarning).
			Foreground(ColorBg).
			Padding(0, 1).
			Render(i18n.T("badge_caps_lock"))
		b.WriteString(capsBadge + "\n\n")
	}

	hasNonASCII := false
	for _, input := range inputs {
		for _, r := range input.Value() {
			if r > 127 {
				hasNonASCII = true
				break
			}
		}
		if hasNonASCII {
			break
		}
	}

	if hasNonASCII {
		warnBox := lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true).
			Render(i18n.T("warn_non_ascii"))
		b.WriteString(warnBox + "\n\n")
	}

	return b.String()
}

// renderDeleteConfirmationModal renders the delete confirmation prompt dialog.
func (m *Model) renderDeleteConfirmationModal() string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(i18n.T("modal_delete_server"))
	b.WriteString(title + "\n\n")

	msg := i18n.Tf("modal_delete_warn", m.hostToDelete.Name, m.hostToDelete.Address)
	b.WriteString(msg + "\n\n")

	actions := lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render("[y / Enter] "+i18n.T("btn_delete")) +
		"    " +
		lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render("[n / Esc] "+i18n.T("btn_cancel"))
	b.WriteString(actions)

	return RenderModalContainer(b.String(), 65, ColorDanger, m.width, m.height)
}

// renderResolutionGuard renders a warning when the terminal dimensions are too small.
func (m *Model) renderResolutionGuard() string {
	minCols := m.cfg.TUI.MinCols
	if minCols <= 0 {
		minCols = 92
	}
	minRows := m.cfg.TUI.MinRows
	if minRows <= 0 {
		minRows = 22
	}

	msg := fmt.Sprintf(
		"%s\n\n"+
			"%s  %d cols × %d rows\n"+
			"%s  %d cols × %d rows\n\n"+
			"%s",
		lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(i18n.T("guard_title")),
		lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("guard_current")),
		m.width, m.height,
		lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(i18n.T("guard_required")),
		minCols, minRows,
		lipgloss.NewStyle().Foreground(ColorWarning).Render(i18n.T("guard_hint")),
	)
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		WarningBoxStyle.Render(msg),
	)
}
