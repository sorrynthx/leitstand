package tui

import (
	"leitstand/internal/i18n"
	"os"
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
	LogDir     string
}

func (s *SettingsModal) Update(msg tea.Msg) (SettingsResult, tea.Cmd) {
	if s.showConfirmSave {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "enter", "y", "Y":
				return s.executeSave(), nil
			case "esc", "n", "N", "q":
				s.showConfirmSave = false
				return SettingsResult{}, nil
			}
		}
		return SettingsResult{}, nil
	}

	if s.filePicker != nil {
		done, pickedPath, cmd := s.filePicker.Update(msg)
		if done {
			if pickedPath != "" {
				if s.pendingExportType > 0 {
					s.handleDatabaseFilePicked(pickedPath)
					s.pendingExportType = 0
				} else {
					s.selectedLogPreset = 3
					s.inputs[3].SetValue(pickedPath)
				}
			}
			s.filePicker = nil
		}
		return SettingsResult{}, cmd
	}

	if s.activeTab == TabDatabase {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			cmd, handled := s.updateSettingsDatabaseTab(keyMsg)
			if handled {
				return SettingsResult{}, cmd
			}
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return SettingsResult{Done: true, SaveReq: false}, nil
		case "f1", "alt+1", "1":
			if s.inputIndexForField(s.focusField) < 0 || msg.String() != "1" {
				s.switchTab(TabGeneral)
				return SettingsResult{}, nil
			}
		case "f2", "alt+2", "2":
			if s.inputIndexForField(s.focusField) < 0 || msg.String() != "2" {
				s.switchTab(TabTelemetry)
				return SettingsResult{}, nil
			}
		case "f3", "alt+3", "3":
			if s.inputIndexForField(s.focusField) < 0 || msg.String() != "3" {
				s.switchTab(TabLogs)
				return SettingsResult{}, nil
			}
		case "f4", "alt+4", "4":
			if s.inputIndexForField(s.focusField) < 0 || msg.String() != "4" {
				s.switchTab(TabDatabase)
				if s.dbStats == nil && s.store != nil {
					s.dbStats, _ = s.store.GetDBStats()
				}
				return SettingsResult{}, nil
			}
		case "f5", "alt+5", "5":
			if s.inputIndexForField(s.focusField) < 0 || msg.String() != "5" {
				s.switchTab(TabAbout)
				return SettingsResult{}, nil
			}

		case "tab":
			if s.activeTab == TabAbout {
				s.switchTab(TabGeneral)
				return SettingsResult{}, nil
			}
			s.focusNext()
			return SettingsResult{}, textinput.Blink

		case "shift+tab":
			if s.activeTab == TabAbout {
				s.switchTab(TabGeneral)
				return SettingsResult{}, nil
			}
			s.focusPrev()
			return SettingsResult{}, textinput.Blink

		case "down", "j", "pgdn":
			if s.activeTab == TabLogs && s.focusField == FieldLogPreset {
				s.selectedLogPreset = (s.selectedLogPreset + 1) % len(s.logDirPresets)
				return SettingsResult{}, nil
			}
			if s.activeTab != TabAbout {
				s.focusNext()
				return SettingsResult{}, textinput.Blink
			}

		case "up", "k", "pgup":
			if s.activeTab == TabLogs && s.focusField == FieldLogPreset {
				s.selectedLogPreset--
				if s.selectedLogPreset < 0 {
					s.selectedLogPreset = len(s.logDirPresets) - 1
				}
				return SettingsResult{}, nil
			}
			if s.activeTab != TabAbout {
				s.focusPrev()
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
			} else if s.activeTab == TabLogs {
				s.selectedLogPreset--
				if s.selectedLogPreset < 0 {
					s.selectedLogPreset = len(s.logDirPresets) - 1
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
			} else if s.activeTab == TabLogs {
				s.selectedLogPreset = (s.selectedLogPreset + 1) % len(s.logDirPresets)
				return SettingsResult{}, nil
			}

		case " ", "space", "b", "B":
			if s.activeTab == TabLogs && s.selectedLogPreset == 3 {
				initDir := strings.TrimSpace(s.inputs[3].Value())
				if initDir == "" {
					initDir, _ = os.UserHomeDir()
				}
				s.filePicker = NewDirPickerModal(initDir, 80, 24)
				return SettingsResult{}, nil
			}
		case "ctrl+e":
			s.errMessage = "💡 세션 저장은 콘솔 화면에서 [Esc]로 설정을 닫은 후 [Ctrl+E]를 눌러주세요."
			return SettingsResult{}, nil
		case "enter":
			if s.activeTab == TabAbout {
				s.activeTab = TabGeneral
				s.focusField = FieldLanguage
				return SettingsResult{}, nil
			}
			if s.activeTab == TabLogs && s.selectedLogPreset == 3 && strings.TrimSpace(s.inputs[3].Value()) == "" {
				initDir, _ := os.UserHomeDir()
				s.filePicker = NewDirPickerModal(initDir, 80, 24)
				return SettingsResult{}, nil
			}
			if s.validateInputs() {
				s.showConfirmSave = true
			}
			return SettingsResult{}, nil
		}
	}

	var cmd tea.Cmd
	inputIdx := s.inputIndexForField(s.focusField)
	if inputIdx >= 0 && inputIdx < len(s.inputs) {
		s.inputs[inputIdx], cmd = s.inputs[inputIdx].Update(msg)
	}
	return SettingsResult{}, cmd
}
