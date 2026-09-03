package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"leitstand/internal/storage"
	"leitstand/internal/vault"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
)

type SettingsTab int

const (
	TabGeneral SettingsTab = iota
	TabTelemetry
	TabLogs
	TabDatabase
	TabAbout
)

type SettingsField int

const (
	FieldLanguage SettingsField = iota
	FieldInterval
	FieldCPUThresh
	FieldRAMThresh
	FieldDiskThresh
	FieldLogPreset
	FieldLogCustomDir
	FieldSubmitBtn
	FieldCount
)

type LogDirPreset struct {
	Label string
	Path  string
}

type SettingsModal struct {
	activeTab         SettingsTab
	selectedLangIndex int
	intervalIndex     int
	logDirPresets     []LogDirPreset
	selectedLogPreset int
	inputs            []textinput.Model
	focusField        SettingsField
	errMessage        string
	successMessage    string
	filePicker        *FilePickerModal
	showConfirmSave   bool

	// Database & Maintenance State (TabDatabase)
	dbActionIndex     int
	retentionIndex    int
	dbStats           *storage.DBStats
	dbConfirmAction   int
	dbConfirmPrompt   string
	showRekeyModal    bool
	rekeyInputs       []textinput.Model
	rekeyFocus        int
	rekeyError        string
	pendingExportType int
	store             *storage.Storage
	vault             *vault.Vault
}

var intervalOptions = []struct {
	Label    string
	Duration time.Duration
}{
	{Label: "5 Seconds (Recommended)", Duration: 5 * time.Second},
	{Label: "10 Seconds (Relaxed)", Duration: 10 * time.Second},
	{Label: "30 Seconds (Minimal)", Duration: 30 * time.Second},
	{Label: "60 Seconds (Slow / Eco)", Duration: 60 * time.Second},
	{Label: "Off (Disabled / Full Console)", Duration: 0},
}

func NewSettingsModal(currLang i18n.Lang, currInterval time.Duration, cpuThresh, ramThresh, diskThresh float64, currLogDir string, store *storage.Storage, v *vault.Vault) *SettingsModal {
	langIdx := 0
	for i, opt := range i18n.SupportedLangs {
		if opt.Code == currLang {
			langIdx = i
			break
		}
	}

	intervalIdx := 0
	for i, opt := range intervalOptions {
		if opt.Duration == currInterval {
			intervalIdx = i
			break
		}
	}

	home, _ := os.UserHomeDir()
	docDir := filepath.Join(home, "Documents", "leitstand", "logs")
	localDir := filepath.Join(".", "logs")
	homeDir := filepath.Join(home, ".leitstand", "logs")

	presets := []LogDirPreset{
		{Label: i18n.T("settings_logs_preset_docs"), Path: docDir},
		{Label: i18n.T("settings_logs_preset_app"), Path: localDir},
		{Label: i18n.T("settings_logs_preset_home"), Path: homeDir},
		{Label: i18n.T("settings_logs_preset_custom"), Path: ""},
	}

	presetIdx := 0
	if currLogDir != "" {
		matched := false
		for i := 0; i < 3; i++ {
			if presets[i].Path == currLogDir {
				presetIdx = i
				matched = true
				break
			}
		}
		if !matched {
			presetIdx = 3
		}
	}

	inputs := make([]textinput.Model, 4)

	// Telemetry Threshold Inputs
	if cpuThresh <= 0 {
		cpuThresh = 85.0
	}
	inputs[0] = textinput.New()
	inputs[0].SetValue(fmt.Sprintf("%.0f", cpuThresh))
	inputs[0].Placeholder = "85"
	inputs[0].Width = 10

	if ramThresh <= 0 {
		ramThresh = 90.0
	}
	inputs[1] = textinput.New()
	inputs[1].SetValue(fmt.Sprintf("%.0f", ramThresh))
	inputs[1].Placeholder = "90"
	inputs[1].Width = 10

	if diskThresh <= 0 {
		diskThresh = 90.0
	}
	inputs[2] = textinput.New()
	inputs[2].SetValue(fmt.Sprintf("%.0f", diskThresh))
	inputs[2].Placeholder = "90"
	inputs[2].Width = 10

	// Custom Log Directory Input
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "e.g. D:\\MyLogs or /var/log/leitstand"
	if presetIdx == 3 {
		inputs[3].SetValue(currLogDir)
	}
	inputs[3].Width = 48

	sm := &SettingsModal{
		activeTab:         TabGeneral,
		selectedLangIndex: langIdx,
		intervalIndex:     intervalIdx,
		logDirPresets:     presets,
		selectedLogPreset: presetIdx,
		inputs:            inputs,
		focusField:        FieldLanguage,
		store:             store,
		vault:             v,
	}
	if store != nil {
		sm.dbStats, _ = store.GetDBStats()
	}
	return sm
}

func (s *SettingsModal) SetError(err error) {
	if err != nil {
		s.errMessage = err.Error()
	} else {
		s.errMessage = ""
	}
}

func (s *SettingsModal) inputIndexForField(f SettingsField) int {
	switch f {
	case FieldCPUThresh:
		return 0
	case FieldRAMThresh:
		return 1
	case FieldDiskThresh:
		return 2
	case FieldLogCustomDir:
		return 3
	default:
		return -1
	}
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
