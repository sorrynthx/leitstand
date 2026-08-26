package tui

import (
	"errors"
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type VaultModalType int

const (
	VaultModalInit   VaultModalType = iota // First time setup: Select Lang + Create & Confirm
	VaultModalUnlock                       // Subsequent runs: Unlock
)

type VaultForm struct {
	modalType  VaultModalType
	inputs     []textinput.Model
	focusIndex int
	errMessage string
	selectedLangIndex int
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
		modalType:         modalType,
		inputs:            inputs,
		focusIndex:        0,
		selectedLangIndex: 0, // English default
	}
}

func (v *VaultForm) Update(msg tea.Msg) (bool, string, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Quick language switcher in Init modal: F1/F2/F3 or Ctrl+L
		if v.modalType == VaultModalInit {
			switch msg.String() {
			case "f1", "ctrl+1":
				v.selectedLangIndex = 0
				i18n.SetLang(i18n.LangEN)
				return false, "", nil
			case "f2", "ctrl+2":
				v.selectedLangIndex = 1
				i18n.SetLang(i18n.LangKO)
				return false, "", nil
			case "f3", "ctrl+3":
				v.selectedLangIndex = 2
				i18n.SetLang(i18n.LangDE)
				return false, "", nil
			}
		}

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

	var cmd tea.Cmd
	if v.focusIndex >= 0 && v.focusIndex < len(v.inputs) {
		v.inputs[v.focusIndex], cmd = v.inputs[v.focusIndex].Update(msg)
	}
	return false, "", cmd
}

func (v *VaultForm) SetError(err error) {
	if errors.Is(err, errors.New("vault password incorrect")) {
		v.errMessage = "❌ Incorrect password. Please try again."
	} else {
		v.errMessage = fmt.Sprintf("❌ %v", err)
	}
	if len(v.inputs) > 0 {
		v.inputs[0].SetValue("")
		if len(v.inputs) > 1 {
			v.inputs[1].SetValue("")
		}
		v.focusIndex = 0
		v.inputs[0].Focus()
	}
}

func (v *VaultForm) View(screenWidth, screenHeight int) string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		Render(i18n.T("modal_vault_title"))
	b.WriteString(title + "\n\n")

	if v.modalType == VaultModalInit {
		// Language selector badge bar
		var langBadges []string
		for i, opt := range i18n.SupportedLangs {
			if i == v.selectedLangIndex {
				langBadges = append(langBadges, lipgloss.NewStyle().Bold(true).Background(ColorPrimary).Foreground(lipgloss.Color("#000000")).Padding(0, 1).Render(fmt.Sprintf("[%d] %s", i+1, opt.Label)))
			} else {
				langBadges = append(langBadges, lipgloss.NewStyle().Foreground(ColorMuted).Padding(0, 1).Render(fmt.Sprintf("[%d] %s", i+1, opt.Label)))
			}
		}
		b.WriteString(strings.Join(langBadges, " ") + "\n\n")

		info := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ECEFF1")).
			Render(i18n.T("modal_vault_init") + "\n(Keys are encrypted with AES-256-GCM & Argon2id)\n")
		b.WriteString(info + "\n")
	} else {
		info := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ECEFF1")).
			Render(i18n.T("modal_vault_unlock") + "\n")
		b.WriteString(info + "\n")
	}

	if IsCapsLockOn() {
		capsBadge := lipgloss.NewStyle().
			Bold(true).
			Background(ColorWarning).
			Foreground(lipgloss.Color("#000000")).
			Padding(0, 1).
			Render(i18n.T("badge_caps_lock"))
		b.WriteString(capsBadge + "\n\n")
	}

	hasNonASCII := false
	for _, input := range v.inputs {
		for _, r := range input.Value() {
			if r > 127 {
				hasNonASCII = true
				break
			}
		}
	}

	if hasNonASCII {
		warnBox := lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true).
			Render(i18n.T("warn_non_ascii"))
		b.WriteString(warnBox + "\n\n")
	}

	if v.errMessage != "" {
		errBox := lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true).
			Render(v.errMessage)
		b.WriteString(errBox + "\n\n")
	}

	for _, input := range v.inputs {
		b.WriteString(input.View() + "\n\n")
	}

	hint := lipgloss.NewStyle().Foreground(ColorMuted).Render("[Enter] Submit    [Tab] Next Field    [Esc] Quit")
	if v.modalType == VaultModalInit {
		hint = lipgloss.NewStyle().Foreground(ColorMuted).Render("[F1/F2/F3] Switch Language    [Enter] Submit    [Tab] Next    [Esc] Quit")
	}
	b.WriteString(hint)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 3).
		Width(65)

	return lipgloss.Place(screenWidth, screenHeight, lipgloss.Center, lipgloss.Center, boxStyle.Render(b.String()))
}
