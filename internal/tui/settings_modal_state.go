package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
)

type SettingsTab int

const (
	TabGeneral SettingsTab = iota
	TabTelemetry
	TabAbout
)

type SettingsField int

const (
	FieldLanguage SettingsField = iota
	FieldCurrentPass
	FieldNewPass
	FieldConfirmPass
	FieldInterval
	FieldCPUThresh
	FieldRAMThresh
	FieldDiskThresh
	FieldSubmitBtn
	FieldCount
)

type SettingsModal struct {
	activeTab         SettingsTab
	selectedLangIndex int
	intervalIndex     int
	inputs            []textinput.Model
	focusField        SettingsField
	errMessage        string
	successMessage    string
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

func NewSettingsModal(currLang i18n.Lang, currInterval time.Duration, cpuThresh, ramThresh, diskThresh float64) *SettingsModal {
	langIdx := 0
	for i, opt := range i18n.SupportedLangs {
		if opt.Code == currLang {
			langIdx = i;
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

	inputs := make([]textinput.Model, 6)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Required only if changing password"
	inputs[0].EchoMode = textinput.EchoPassword
	inputs[0].EchoCharacter = '•'
	inputs[0].Width = 48

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Enter new master password"
	inputs[1].EchoMode = textinput.EchoPassword
	inputs[1].EchoCharacter = '•'
	inputs[1].Width = 48

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Confirm new master password"
	inputs[2].EchoMode = textinput.EchoPassword
	inputs[2].EchoCharacter = '•'
	inputs[2].Width = 48

	// Telemetry Threshold Inputs
	if cpuThresh <= 0 {
		cpuThresh = 85.0
	}
	inputs[3] = textinput.New()
	inputs[3].SetValue(fmt.Sprintf("%.0f", cpuThresh))
	inputs[3].Placeholder = "85"
	inputs[3].Width = 10

	if ramThresh <= 0 {
		ramThresh = 90.0
	}
	inputs[4] = textinput.New()
	inputs[4].SetValue(fmt.Sprintf("%.0f", ramThresh))
	inputs[4].Placeholder = "90"
	inputs[4].Width = 10

	if diskThresh <= 0 {
		diskThresh = 90.0
	}
	inputs[5] = textinput.New()
	inputs[5].SetValue(fmt.Sprintf("%.0f", diskThresh))
	inputs[5].Placeholder = "90"
	inputs[5].Width = 10

	return &SettingsModal{
		activeTab:         TabGeneral,
		selectedLangIndex: langIdx,
		intervalIndex:     intervalIdx,
		inputs:            inputs,
		focusField:        FieldLanguage,
	}
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
	case FieldCurrentPass:
		return 0
	case FieldNewPass:
		return 1
	case FieldConfirmPass:
		return 2
	case FieldCPUThresh:
		return 3
	case FieldRAMThresh:
		return 4
	case FieldDiskThresh:
		return 5
	default:
		return -1
	}
}
