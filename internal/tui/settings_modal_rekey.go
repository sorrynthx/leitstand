package tui

import (
	"leitstand/internal/i18n"
	"leitstand/internal/vault"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (s *SettingsModal) initRekeyInputs() {
	s.rekeyInputs = make([]textinput.Model, 3)
	for i := 0; i < 3; i++ {
		s.rekeyInputs[i] = textinput.New()
		s.rekeyInputs[i].EchoMode = textinput.EchoPassword
		s.rekeyInputs[i].EchoCharacter = '•'
		s.rekeyInputs[i].Width = 36
	}
	s.rekeyInputs[0].Placeholder = "Current Master Password"
	s.rekeyInputs[1].Placeholder = "New Master Password"
	s.rekeyInputs[2].Placeholder = "Confirm New Password"
	s.rekeyInputs[0].Focus()
	s.rekeyFocus = 0
	s.rekeyError = ""
}

func (s *SettingsModal) renderRekeyModal() string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(i18n.T("rekey_title"))
	b.WriteString(title + "\n\n")

	// Render Caps Lock and security badges
	badges := RenderSecurityBadges(s.rekeyInputs...)
	if badges != "" {
		b.WriteString(badges + "\n")
	}

	labels := []string{
		i18n.T("rekey_cur_pwd"),
		i18n.T("rekey_new_pwd"),
		i18n.T("rekey_confirm_pwd"),
	}

	for i := 0; i < 3; i++ {
		cursor := "  "
		if i == s.rekeyFocus {
			cursor = "▶ "
		}
		lbl := lipgloss.NewStyle().Bold(i == s.rekeyFocus).Foreground(ColorText).Width(24).Render(cursor + labels[i])
		b.WriteString(lbl + " " + s.rekeyInputs[i].View() + "\n\n")
	}

	if s.rekeyError != "" {
		errBox := lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render("⚠️ " + s.rekeyError)
		b.WriteString(errBox + "\n\n")
	}

	hint := lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("rekey_hints"))
	b.WriteString(hint + "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorWarning).
		Background(lipgloss.Color("#181B20")).
		Padding(1, 2).
		Width(66).
		Render(b.String())

	return box
}

func (s *SettingsModal) updateRekeyModal(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "esc":
		s.showRekeyModal = false
		return nil, true

	case "tab", "down":
		s.rekeyInputs[s.rekeyFocus].Blur()
		s.rekeyFocus = (s.rekeyFocus + 1) % 3
		s.rekeyInputs[s.rekeyFocus].Focus()
		return nil, true

	case "shift+tab", "up":
		s.rekeyInputs[s.rekeyFocus].Blur()
		s.rekeyFocus = (s.rekeyFocus + 2) % 3
		s.rekeyInputs[s.rekeyFocus].Focus()
		return nil, true

	case "enter":
		if s.rekeyFocus < 2 {
			s.rekeyInputs[s.rekeyFocus].Blur()
			s.rekeyFocus++
			s.rekeyInputs[s.rekeyFocus].Focus()
			return nil, true
		}
		fallthrough
	case "ctrl+s":
		curPwd := s.rekeyInputs[0].Value()
		newPwd := s.rekeyInputs[1].Value()
		confPwd := s.rekeyInputs[2].Value()

		if curPwd == "" || newPwd == "" {
			s.rekeyError = "Passwords cannot be empty"
			return nil, true
		}
		if newPwd != confPwd {
			s.rekeyError = i18n.T("rekey_err_mismatch")
			return nil, true
		}
		if s.store == nil || s.vault == nil {
			s.rekeyError = "Security vault unavailable"
			return nil, true
		}

		tempVault := vault.New()
		if err := s.store.UnlockVault(tempVault, curPwd); err != nil {
			s.rekeyError = i18n.T("rekey_err_cur_invalid")
			return nil, true
		}
		tempVault.Lock()

		newVault := vault.New()
		if err := s.store.RekeyVault(s.vault, newVault, newPwd); err != nil {
			s.rekeyError = err.Error()
			return nil, true
		}

		s.vault.Lock()
		*s.vault = *newVault
		s.showRekeyModal = false
		s.errMessage = ""
		s.successMessage = i18n.T("db_success_rekey")
		return nil, true
	}

	var cmd tea.Cmd
	s.rekeyInputs[s.rekeyFocus], cmd = s.rekeyInputs[s.rekeyFocus].Update(msg)
	return cmd, true
}
