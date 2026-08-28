package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SudoModal struct {
	hostName      string
	command       string
	passwordInput textinput.Model
	rememberPass  bool
	errMessage    string
}

func NewSudoModal(hostName, command string) *SudoModal {
	ti := textinput.New()
	ti.Placeholder = "Enter sudo/root password"
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Focus()
	ti.Width = 34
	ti.Prompt = "Password: "

	return &SudoModal{
		hostName:      hostName,
		command:       command,
		passwordInput: ti,
		rememberPass:  true,
	}
}

func (s *SudoModal) Update(msg tea.Msg) (done bool, password string, remember bool, cmd tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return true, "", false, nil // Cancel

		case "tab":
			s.rememberPass = !s.rememberPass
			return false, "", false, nil

		case "enter":
			pass := s.passwordInput.Value()
			return true, pass, s.rememberPass, nil
		}
	}

	var tiCmd tea.Cmd
	s.passwordInput, tiCmd = s.passwordInput.Update(msg)
	return false, "", false, tiCmd
}

func (s *SudoModal) View(termWidth, termHeight int) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render("🔑 ELEVATED PRIVILEGE REQUIRED (SUDO / ROOT)")
	b.WriteString(title + "\n\n")

	info := lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("Target Host: %s", s.hostName))
	b.WriteString(info + "\n")

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

	b.WriteString(s.passwordInput.View() + "\n\n")

	rememberBox := "[ ]"
	if s.rememberPass {
		rememberBox = lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render("[✓]")
	}
	rememberText := fmt.Sprintf("%s Remember password for this session (Tab to toggle)", rememberBox)
	b.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Render(rememberText) + "\n\n")

	hints := lipgloss.NewStyle().Foreground(ColorMuted).Render("[Enter] Run Command  [Tab] Toggle Remember  [Esc] Cancel")
	b.WriteString(hints)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorWarning).
		Padding(1, 3).
		Width(68)

	content := boxStyle.Render(b.String())

	return lipgloss.Place(termWidth, termHeight, lipgloss.Center, lipgloss.Center, content)
}
