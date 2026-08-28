package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorDanger).
		Padding(1, 3).
		Width(65)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, boxStyle.Render(b.String()))
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
