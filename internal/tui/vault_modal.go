package tui

import (
	"errors"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type VaultModalType int

const (
	VaultModalInit   VaultModalType = iota // First time setup: Create & Confirm
	VaultModalUnlock                       // Subsequent runs: Unlock
)

type VaultForm struct {
	modalType  VaultModalType
	inputs     []textinput.Model
	focusIndex int
	errMessage string
}

func NewVaultForm(modalType VaultModalType) *VaultForm {
	var inputs []textinput.Model

	if modalType == VaultModalInit {
		inputs = make([]textinput.Model, 2)
		inputs[0] = textinput.New()
		inputs[0].Placeholder = "Enter new master password"
		inputs[0].EchoMode = textinput.EchoPassword
		inputs[0].EchoCharacter = '•'
		inputs[0].Prompt = "Create Password:  "
		inputs[0].Focus()
		inputs[0].Width = 25

		inputs[1] = textinput.New()
		inputs[1].Placeholder = "Confirm master password"
		inputs[1].EchoMode = textinput.EchoPassword
		inputs[1].EchoCharacter = '•'
		inputs[1].Prompt = "Confirm Password: "
		inputs[1].Width = 25
	} else {
		inputs = make([]textinput.Model, 1)
		inputs[0] = textinput.New()
		inputs[0].Placeholder = "Enter master password"
		inputs[0].EchoMode = textinput.EchoPassword
		inputs[0].EchoCharacter = '•'
		inputs[0].Prompt = "Master Password: "
		inputs[0].Focus()
		inputs[0].Width = 25
	}

	return &VaultForm{
		modalType:  modalType,
		inputs:     inputs,
		focusIndex: 0,
	}
}

func (v *VaultForm) Update(msg tea.Msg) (bool, string, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return true, "", nil // Exit

		case "tab", "down":
			if len(v.inputs) > 1 {
				v.inputs[v.focusIndex].Blur()
				v.focusIndex = (v.focusIndex + 1) % len(v.inputs)
				v.inputs[v.focusIndex].Focus()
				return false, "", textinput.Blink
			}

		case "shift+tab", "up":
			if len(v.inputs) > 1 {
				v.inputs[v.focusIndex].Blur()
				v.focusIndex--
				if v.focusIndex < 0 {
					v.focusIndex = len(v.inputs) - 1
				}
				v.inputs[v.focusIndex].Focus()
				return false, "", textinput.Blink
			}

		case "enter":
			if v.modalType == VaultModalInit {
				if v.focusIndex == 0 {
					v.inputs[0].Blur()
					v.focusIndex = 1
					v.inputs[1].Focus()
					return false, "", textinput.Blink
				}
				p1 := v.inputs[0].Value()
				p2 := v.inputs[1].Value()
				if len(p1) < 4 {
					v.errMessage = "Password must be at least 4 characters"
					return false, "", nil
				}
				if p1 != p2 {
					v.errMessage = "Passwords do not match"
					return false, "", nil
				}
				return true, p1, nil
			} else {
				p := v.inputs[0].Value()
				if p == "" {
					v.errMessage = "Password cannot be empty"
					return false, "", nil
				}
				return true, p, nil
			}
		}
	}

	cmds := make([]tea.Cmd, len(v.inputs))
	for i := range v.inputs {
		v.inputs[i], cmds[i] = v.inputs[i].Update(msg)
	}

	return false, "", tea.Batch(cmds...)
}

func (v *VaultForm) SetError(err error) {
	if err != nil {
		if errors.Is(err, errors.New("invalid master password")) || strings.Contains(err.Error(), "invalid") {
			v.errMessage = "Incorrect master password. Try again."
		} else {
			v.errMessage = err.Error()
		}
		v.inputs[0].SetValue("")
		if len(v.inputs) > 1 {
			v.inputs[1].SetValue("")
			v.focusIndex = 0
			v.inputs[0].Focus()
			v.inputs[1].Blur()
		}
	}
}

func (v *VaultForm) View(termWidth, termHeight int) string {
	var b strings.Builder

	if v.modalType == VaultModalInit {
		title := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("🔒 INITIALIZE SECURE LOCAL VAULT")
		b.WriteString(title + "\n")
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("Welcome! Create a master password to encrypt server credentials locally.") + "\n\n")
	} else {
		title := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("🔓 UNLOCK LEITSTAND VAULT")
		b.WriteString(title + "\n")
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("Enter your master password to unlock server credentials.") + "\n\n")
	}

	if v.errMessage != "" {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render("❌ "+v.errMessage) + "\n\n")
	}

	for i := range v.inputs {
		b.WriteString(v.inputs[i].View() + "\n")
	}

	b.WriteString("\n")
	if v.modalType == VaultModalInit {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("[Tab/↓] Next  [Enter] Confirm & Launch  [Esc] Exit"))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("[Enter] Unlock & Launch Cockpit  [Esc] Exit"))
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 2).
		Width(65)

	return lipgloss.Place(termWidth, termHeight, lipgloss.Center, lipgloss.Center, boxStyle.Render(b.String()))
}
