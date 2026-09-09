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
	Done             bool
	SaveReq          bool
	FactoryResetDone bool
	HostsImported    bool
	Lang             i18n.Lang
	Interval         time.Duration
	CPUThresh        float64
	RAMThresh        float64
	DiskThresh       float64
	LogDir           string
	AIProvider       string
	AIEndpoint       string
	AIKey            string
	AIModel          string
	AIRetention      int
	AIMaxHistory     int
	Profiles         map[string]*AIProviderProfile
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
			if s.factoryResetDone {
				return SettingsResult{
					Done:             true,
					SaveReq:          false,
					FactoryResetDone: true,
				}, nil
			}
			if handled {
				return SettingsResult{}, cmd
			}
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		isInput := s.inputIndexForField(s.focusField) >= 0
		switch msg.String() {
		case "esc":
			return SettingsResult{
				Done:             true,
				SaveReq:          false,
				FactoryResetDone: s.factoryResetDone,
				HostsImported:    s.hostsImported,
			}, nil
		case "f1", "alt+1", "1":
			if !isInput || msg.String() != "1" {
				s.switchTab(TabGeneral)
				return SettingsResult{}, nil
			}
		case "f2", "alt+2", "2":
			if !isInput || msg.String() != "2" {
				s.switchTab(TabTelemetry)
				return SettingsResult{}, nil
			}
		case "f3", "alt+3", "3":
			if !isInput || msg.String() != "3" {
				s.switchTab(TabLogs)
				return SettingsResult{}, nil
			}
		case "f4", "alt+4", "4":
			if !isInput || msg.String() != "4" {
				s.switchTab(TabDatabase)
				if s.dbStats == nil && s.store != nil {
					s.dbStats, _ = s.store.GetDBStats()
				}
				return SettingsResult{}, nil
			}
		case "f5", "alt+5", "5":
			if !isInput || msg.String() != "5" {
				s.switchTab(TabAI)
				return SettingsResult{}, nil
			}
		case "f6", "alt+6", "6":
			if !isInput || msg.String() != "6" {
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
			if isInput && msg.String() == "j" {
				break
			}
			if s.activeTab == TabLogs && s.focusField == FieldLogPreset {
				s.selectedLogPreset = (s.selectedLogPreset + 1) % len(s.logDirPresets)
				return SettingsResult{}, nil
			}
			if s.activeTab != TabAbout {
				s.focusNext()
				return SettingsResult{}, textinput.Blink
			}

		case "up", "k", "pgup":
			if isInput && msg.String() == "k" {
				break
			}
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
			if isInput {
				break
			}
			if s.handleLeftRight(-1) {
				return SettingsResult{}, nil
			}

		case "right", "l":
			if isInput {
				break
			}
			if s.handleLeftRight(1) {
				return SettingsResult{}, nil
			}

		case " ", "space", "b", "B":
			if isInput {
				break
			}
			if s.activeTab == TabLogs && s.selectedLogPreset == 3 {
				initDir := strings.TrimSpace(s.inputs[3].Value())
				if initDir == "" {
					initDir, _ = os.UserHomeDir()
				}
				s.filePicker = NewDirPickerModal(initDir, 80, 24)
				return SettingsResult{}, nil
			}
		case "ctrl+e":
			s.errMessage = i18n.T("settings_session_log_tip")
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
