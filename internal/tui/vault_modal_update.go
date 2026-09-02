package tui

import (
	"errors"
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type VaultModalType int

const (
	VaultModalInit VaultModalType = iota
	VaultModalUnlock
)

type VaultForm struct {
	modalType         VaultModalType
	inputs            []textinput.Model
	focusIndex        int
	errMessage        string
	selectedLangIndex int
}

type VaultModal = VaultForm

func NewVaultForm(modalType VaultModalType) *VaultForm {
	var inputs []textinput.Model

	if modalType == VaultModalInit {
		inputs = make([]textinput.Model, 2)
		inputs[0] = textinput.New()
		inputs[0].Placeholder = i18n.T("vault_pass_create")
		inputs[0].EchoMode = textinput.EchoPassword
		inputs[0].EchoCharacter = '•'
		inputs[0].Focus()
		inputs[0].Width = 30

		inputs[1] = textinput.New()
		inputs[1].Placeholder = i18n.T("vault_pass_confirm")
		inputs[1].EchoMode = textinput.EchoPassword
		inputs[1].EchoCharacter = '•'
		inputs[1].Width = 30
	} else {
		inputs = make([]textinput.Model, 1)
		inputs[0] = textinput.New()
		inputs[0].Placeholder = i18n.T("vault_pass_enter")
		inputs[0].EchoMode = textinput.EchoPassword
		inputs[0].EchoCharacter = '•'
		inputs[0].Focus()
		inputs[0].Width = 30
	}

	langIdx := 0
	currLang := i18n.GetLang()
	for i, opt := range i18n.SupportedLangs {
		if opt.Code == currLang {
			langIdx = i
			break
		}
	}

	return &VaultForm{
		modalType:         modalType,
		inputs:            inputs,
		focusIndex:        0,
		selectedLangIndex: langIdx,
	}
}

func (vf *VaultForm) Update(msg tea.Msg) (bool, string, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return true, "", tea.Quit

		case "tab", "down":
			if vf.modalType == VaultModalInit {
				vf.inputs[vf.focusIndex].Blur()
				vf.focusIndex = (vf.focusIndex + 1) % len(vf.inputs)
				vf.inputs[vf.focusIndex].Focus()
				return false, "", textinput.Blink
			}

		case "shift+tab", "up":
			if vf.modalType == VaultModalInit {
				vf.inputs[vf.focusIndex].Blur()
				vf.focusIndex--
				if vf.focusIndex < 0 {
					vf.focusIndex = len(vf.inputs) - 1
				}
				vf.inputs[vf.focusIndex].Focus()
				return false, "", textinput.Blink
			}

		case "f1", "alt+1":
			vf.selectedLangIndex = 0
			if vf.selectedLangIndex < len(i18n.SupportedLangs) {
				i18n.SetLang(i18n.SupportedLangs[vf.selectedLangIndex].Code)
			}
			return false, "", nil

		case "f2", "alt+2":
			vf.selectedLangIndex = 1
			if vf.selectedLangIndex < len(i18n.SupportedLangs) {
				i18n.SetLang(i18n.SupportedLangs[vf.selectedLangIndex].Code)
			}
			return false, "", nil

		case "f3", "alt+3":
			vf.selectedLangIndex = 2
			if vf.selectedLangIndex < len(i18n.SupportedLangs) {
				i18n.SetLang(i18n.SupportedLangs[vf.selectedLangIndex].Code)
			}
			return false, "", nil

		case "left":
			vf.selectedLangIndex--
			if vf.selectedLangIndex < 0 {
				vf.selectedLangIndex = len(i18n.SupportedLangs) - 1
			}
			i18n.SetLang(i18n.SupportedLangs[vf.selectedLangIndex].Code)
			return false, "", nil

		case "right":
			vf.selectedLangIndex = (vf.selectedLangIndex + 1) % len(i18n.SupportedLangs)
			i18n.SetLang(i18n.SupportedLangs[vf.selectedLangIndex].Code)
			return false, "", nil

		case "enter":
			pass := strings.TrimSpace(vf.inputs[0].Value())
			if pass == "" {
				vf.errMessage = i18n.T("vault_err_empty")
				return false, "", nil
			}

			if vf.modalType == VaultModalInit {
				confirm := strings.TrimSpace(vf.inputs[1].Value())
				if len(pass) < 4 {
					vf.errMessage = i18n.T("vault_err_short")
					return false, "", nil
				}
				if pass != confirm {
					vf.errMessage = i18n.T("vault_err_mismatch")
					return false, "", nil
				}
			}
			return true, pass, nil
		}
	}

	var cmd tea.Cmd
	if vf.focusIndex < len(vf.inputs) {
		vf.inputs[vf.focusIndex], cmd = vf.inputs[vf.focusIndex].Update(msg)
	}
	return false, "", cmd
}

func (vf *VaultForm) SetError(err error) {
	if err != nil {
		if errors.Is(err, errors.New("cipher: message authentication failed")) || strings.Contains(err.Error(), "authentication failed") {
			vf.errMessage = i18n.T("vault_err_invalid")
		} else {
			vf.errMessage = fmt.Sprintf("Error: %v", err)
		}
	} else {
		vf.errMessage = ""
	}
}
