package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SudoModal struct {
	hostName      string
	userName      string
	command       string
	passwordInput textinput.Model
	rememberPass  bool
	showPassword  bool
	launchPTY     bool
	errMessage    string
}

func NewSudoModal(hostName, userName, command string) *SudoModal {
	ti := textinput.New()
	ti.Placeholder = i18n.T("form_ph_pass")
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Focus()
	ti.Width = 36
	ti.Prompt = "Password: "

	return &SudoModal{
		hostName:      hostName,
		userName:      userName,
		command:       command,
		passwordInput: ti,
		rememberPass:  true,
		showPassword:  false,
	}
}

func (s *SudoModal) Update(msg tea.Msg) (done bool, password string, remember bool, launchPTY bool, cmd tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return true, "", false, false, nil // Cancel

		case "tab":
			s.rememberPass = !s.rememberPass
			return false, "", false, false, nil

		case "f2", "ctrl+p":
			s.showPassword = !s.showPassword
			if s.showPassword {
				s.passwordInput.EchoMode = textinput.EchoNormal
			} else {
				s.passwordInput.EchoMode = textinput.EchoPassword
			}
			return false, "", false, false, nil

		case "f3", "ctrl+t":
			// Request direct interactive PTY terminal launch (MobaXterm style)
			return true, "", false, true, nil

		case "enter":
			pass := s.passwordInput.Value()
			return true, pass, s.rememberPass, false, nil
		}
	}

	var tiCmd tea.Cmd
	s.passwordInput, tiCmd = s.passwordInput.Update(msg)
	return false, "", false, false, tiCmd
}

func (s *SudoModal) View(termWidth, termHeight int) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render("🔑 " + i18n.T("sudo_title"))
	b.WriteString(title + "\n\n")

	info := lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("Target Host: %s  •  Current User: %s", s.hostName, s.userName))
	b.WriteString(info + "\n")

	userHint := i18n.Tf("sudo_prompt", s.hostName, s.userName)
	b.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Render(userHint) + "\n\n")

	cmdPreview := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		Background(lipgloss.Color("#151820")).
		Padding(0, 1).
		Render("Command: " + s.command)
	b.WriteString(cmdPreview + "\n\n")

	if s.errMessage != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render("⚠️ "+s.errMessage) + "\n\n")
	}

	s.passwordInput.Placeholder = i18n.T("form_ph_pass")
	b.WriteString(RenderSecurityBadges(s.passwordInput))

	b.WriteString(s.passwordInput.View() + "\n\n")

	rememberBox := "[ ]"
	if s.rememberPass {
		rememberBox = lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render("[✓]")
	}
	rememberText := fmt.Sprintf("%s %s", rememberBox, i18n.T("sudo_remember_check"))
	b.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Render(rememberText) + "\n\n")

	viewModeText := "[F2] " + i18n.T("btn_apply")
	if s.showPassword {
		viewModeText = "[F2] " + i18n.T("btn_close")
	}
	hints := lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("[Enter] %s  [Tab] %s  %s  [Esc] %s", i18n.T("btn_apply"), i18n.T("sudo_remember_check"), viewModeText, i18n.T("btn_cancel")))
	b.WriteString(hints)

	return RenderModalContainer(b.String(), 72, ColorWarning, termWidth, termHeight)
}
