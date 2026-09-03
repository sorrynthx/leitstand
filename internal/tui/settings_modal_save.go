package tui

import (
	"leitstand/internal/i18n"
	"strconv"
	"strings"
)

func (s *SettingsModal) validateInputs() bool {
	cpuT, err1 := strconv.ParseFloat(strings.TrimSpace(s.inputs[0].Value()), 64)
	ramT, err2 := strconv.ParseFloat(strings.TrimSpace(s.inputs[1].Value()), 64)
	diskT, err3 := strconv.ParseFloat(strings.TrimSpace(s.inputs[2].Value()), 64)

	if err1 != nil || cpuT < 1 || cpuT > 100 {
		s.errMessage = "⚠️ CPU 임계치는 1 ~ 100 사이의 숫자로 입력해 주세요."
		return false
	}
	if err2 != nil || ramT < 1 || ramT > 100 {
		s.errMessage = "⚠️ RAM 임계치는 1 ~ 100 사이의 숫자로 입력해 주세요."
		return false
	}
	if err3 != nil || diskT < 1 || diskT > 100 {
		s.errMessage = "⚠️ Disk 임계치는 1 ~ 100 사이의 숫자로 입력해 주세요."
		return false
	}

	s.errMessage = ""
	return true
}

func (s *SettingsModal) executeSave() SettingsResult {
	selectedLang := i18n.SupportedLangs[s.selectedLangIndex].Code
	selectedInterval := intervalOptions[s.intervalIndex].Duration

	cpuT, _ := strconv.ParseFloat(strings.TrimSpace(s.inputs[0].Value()), 64)
	ramT, _ := strconv.ParseFloat(strings.TrimSpace(s.inputs[1].Value()), 64)
	diskT, _ := strconv.ParseFloat(strings.TrimSpace(s.inputs[2].Value()), 64)

	finalLogDir := ""
	if s.selectedLogPreset >= 0 && s.selectedLogPreset < len(s.logDirPresets)-1 {
		finalLogDir = s.logDirPresets[s.selectedLogPreset].Path
	} else if len(s.inputs) > 3 {
		finalLogDir = strings.TrimSpace(s.inputs[3].Value())
	}

	return SettingsResult{
		Done:       true,
		SaveReq:    true,
		Lang:       selectedLang,
		Interval:   selectedInterval,
		CPUThresh:  cpuT,
		RAMThresh:  ramT,
		DiskThresh: diskT,
		LogDir:     finalLogDir,
	}
}
