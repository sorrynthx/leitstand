package tui

import (
	"leitstand/internal/i18n"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type SettingsResult struct {
	Done       bool
	SaveReq    bool
	Lang       i18n.Lang
	Interval   time.Duration
	CPUThresh  float64
	RAMThresh  float64
	DiskThresh float64
	CurrPass   string
	NewPass    string
}

func (s *SettingsModal) Update(msg tea.Msg) (SettingsResult, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return SettingsResult{Done: true, SaveReq: false}, nil

		case "f1", "alt+1", "1":
			if s.inputIndexForField(s.focusField) < 0 || msg.String() != "1" {
				s.activeTab = TabGeneral
				s.focusField = FieldLanguage
				return SettingsResult{}, nil
			}

		case "f2", "alt+2", "2":
			if s.inputIndexForField(s.focusField) < 0 || msg.String() != "2" {
				s.activeTab = TabTelemetry
				s.focusField = FieldInterval
				return SettingsResult{}, nil
			}

		case "f3", "alt+3", "3":
			if s.inputIndexForField(s.focusField) < 0 || msg.String() != "3" {
				s.activeTab = TabAbout
				return SettingsResult{}, nil
			}

		case "tab":
			if s.activeTab == TabAbout {
				s.activeTab = TabGeneral
				s.focusField = FieldLanguage
				return SettingsResult{}, nil
			}
			s.blurCurrent()
			s.focusField = (s.focusField + 1) % FieldCount
			s.focusCurrent()
			return SettingsResult{}, textinput.Blink

		case "shift+tab":
			if s.activeTab == TabAbout {
				s.activeTab = TabGeneral
				s.focusField = FieldLanguage
				return SettingsResult{}, nil
			}
			s.blurCurrent()
			s.focusField--
			if s.focusField < 0 {
				s.focusField = FieldCount - 1
			}
			s.focusCurrent()
			return SettingsResult{}, textinput.Blink

		case "down":
			if s.activeTab != TabAbout {
				s.blurCurrent()
				s.focusField = (s.focusField + 1) % FieldCount
				s.focusCurrent()
				return SettingsResult{}, textinput.Blink
			}

		case "up":
			if s.activeTab != TabAbout {
				s.blurCurrent()
				s.focusField--
				if s.focusField < 0 {
					s.focusField = FieldCount - 1
				}
				s.focusCurrent()
				return SettingsResult{}, textinput.Blink
			}

		case "left", "h":
			if s.activeTab == TabGeneral && s.focusField == FieldLanguage {
				s.selectedLangIndex--
				if s.selectedLangIndex < 0 {
					s.selectedLangIndex = len(i18n.SupportedLangs) - 1
				}
				i18n.SetLang(i18n.SupportedLangs[s.selectedLangIndex].Code)
				return SettingsResult{}, nil
			} else if s.focusField == FieldInterval {
				s.intervalIndex--
				if s.intervalIndex < 0 {
					s.intervalIndex = len(intervalOptions) - 1
				}
				return SettingsResult{}, nil
			}

		case "right", "l":
			if s.activeTab == TabGeneral && s.focusField == FieldLanguage {
				s.selectedLangIndex = (s.selectedLangIndex + 1) % len(i18n.SupportedLangs)
				i18n.SetLang(i18n.SupportedLangs[s.selectedLangIndex].Code)
				return SettingsResult{}, nil
			} else if s.focusField == FieldInterval {
				s.intervalIndex = (s.intervalIndex + 1) % len(intervalOptions)
				return SettingsResult{}, nil
			}

		case "enter":
			if s.activeTab == TabAbout {
				s.activeTab = TabGeneral
				s.focusField = FieldLanguage
				return SettingsResult{}, nil
			}

			selectedLang := i18n.SupportedLangs[s.selectedLangIndex].Code
			selectedInterval := intervalOptions[s.intervalIndex].Duration

			currPass := s.inputs[0].Value()
			newPass := s.inputs[1].Value()
			confirmPass := s.inputs[2].Value()

			if newPass != "" {
				if len(newPass) < 4 {
					s.errMessage = "New password must be at least 4 characters"
					return SettingsResult{}, nil
				}
				if newPass != confirmPass {
					s.errMessage = i18n.T("settings_pass_mismatch")
					return SettingsResult{}, nil
				}
				if currPass == "" {
					s.errMessage = "Current password is required to set new password"
					return SettingsResult{}, nil
				}
			}

			cpuT, err1 := strconv.ParseFloat(strings.TrimSpace(s.inputs[3].Value()), 64)
			ramT, err2 := strconv.ParseFloat(strings.TrimSpace(s.inputs[4].Value()), 64)
			diskT, err3 := strconv.ParseFloat(strings.TrimSpace(s.inputs[5].Value()), 64)

			if err1 != nil || cpuT < 1 || cpuT > 100 {
				s.errMessage = "⚠️ CPU 임계치는 1 ~ 100 사이의 숫자로 입력해 주세요."
				return SettingsResult{}, nil
			}
			if err2 != nil || ramT < 1 || ramT > 100 {
				s.errMessage = "⚠️ RAM 임계치는 1 ~ 100 사이의 숫자로 입력해 주세요."
				return SettingsResult{}, nil
			}
			if err3 != nil || diskT < 1 || diskT > 100 {
				s.errMessage = "⚠️ Disk 임계치는 1 ~ 100 사이의 숫자로 입력해 주세요."
				return SettingsResult{}, nil
			}

			return SettingsResult{
				Done:       true,
				SaveReq:    true,
				Lang:       selectedLang,
				Interval:   selectedInterval,
				CPUThresh:  cpuT,
				RAMThresh:  ramT,
				DiskThresh: diskT,
				CurrPass:   currPass,
				NewPass:    newPass,
			}, nil
		}
	}

	var cmd tea.Cmd
	inputIdx := s.inputIndexForField(s.focusField)
	if inputIdx >= 0 && inputIdx < len(s.inputs) {
		s.inputs[inputIdx], cmd = s.inputs[inputIdx].Update(msg)
	}
	return SettingsResult{}, cmd
}

func (s *SettingsModal) blurCurrent() {
	idx := s.inputIndexForField(s.focusField)
	if idx >= 0 && idx < len(s.inputs) {
		s.inputs[idx].Blur()
	}
}

func (s *SettingsModal) focusCurrent() {
	idx := s.inputIndexForField(s.focusField)
	if idx >= 0 && idx < len(s.inputs) {
		s.inputs[idx].Focus()
	}
}
