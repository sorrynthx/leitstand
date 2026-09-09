package tui

import (
	"leitstand/internal/i18n"
	"leitstand/internal/vault"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (s *SettingsModal) initResetInputs() {
	inp := textinput.New()
	inp.Placeholder = "Enter Master Password"
	inp.EchoMode = textinput.EchoPassword
	inp.EchoCharacter = '•'
	inp.Width = 36
	inp.Focus()
	s.resetInput = inp
	s.resetError = ""
}

func (s *SettingsModal) renderResetModal() string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(i18n.T("reset_modal_title"))
	b.WriteString(title + "\n\n")

	warnBox := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorDanger).
		Background(lipgloss.Color("#2A1010")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDanger).
		Padding(0, 1).
		Width(60).
		Render(i18n.T("reset_modal_warning"))
	b.WriteString(warnBox + "\n\n")

	// Security badge for caps lock
	badges := RenderSecurityBadges(s.resetInput)
	if badges != "" {
		b.WriteString(badges + "\n")
	}

	prompt := lipgloss.NewStyle().Foreground(ColorText).Render("🔒 " + i18n.T("reset_modal_prompt"))
	b.WriteString(prompt + "\n")
	b.WriteString("  " + s.resetInput.View() + "\n\n")

	if s.resetError != "" {
		errBox := lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render("⚠️ " + s.resetError)
		b.WriteString(errBox + "\n\n")
	}

	hints := lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("reset_modal_hints"))
	b.WriteString(hints + "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorDanger).
		Background(lipgloss.Color("#181B20")).
		Padding(1, 2).
		Width(66).
		Render(b.String())
}

func (s *SettingsModal) updateResetModal(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "esc":
		s.showResetModal = false
		s.resetError = ""
		return nil, true

	case "enter":
		pwd := strings.TrimSpace(s.resetInput.Value())
		if pwd == "" {
			s.resetError = i18n.T("vault_err_empty")
			return nil, true
		}

		if s.store == nil {
			s.resetError = "Database unavailable"
			return nil, true
		}

		// Verify master password against vault_meta
		tempVault := vault.New()
		if err := s.store.UnlockVault(tempVault, pwd); err != nil {
			s.resetError = i18n.T("reset_modal_err_pwd")
			return nil, true
		}
		tempVault.Lock()

		// Password is correct! Execute factory reset
		err := s.store.FactoryReset()
		if err != nil {
			s.resetError = err.Error()
			return nil, true
		}

		s.factoryResetDone = true
		s.showResetModal = false
		s.dbStats, _ = s.store.GetDBStats()
		s.errMessage = ""
		s.successMessage = "✨ " + i18n.T("db_success_reset")
		return nil, true
	}

	var cmd tea.Cmd
	s.resetInput, cmd = s.resetInput.Update(msg)
	return cmd, true
}
