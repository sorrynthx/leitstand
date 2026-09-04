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
		s.errMessage = i18n.T("settings_err_cpu_thresh")
		return false
	}
	if err2 != nil || ramT < 1 || ramT > 100 {
		s.errMessage = i18n.T("settings_err_ram_thresh")
		return false
	}
	if err3 != nil || diskT < 1 || diskT > 100 {
		s.errMessage = i18n.T("settings_err_disk_thresh")
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

	aiProv := s.aiProviders[s.aiProviderIndex]
	aiEp := strings.TrimSpace(s.inputs[4].Value())
	aiKey := strings.TrimSpace(s.inputs[5].Value())
	aiModel := strings.TrimSpace(s.inputs[6].Value())
	aiRet, _ := strconv.Atoi(strings.TrimSpace(s.inputs[7].Value()))
	if aiRet <= 0 {
		aiRet = 3
	}
	aiMaxH, _ := strconv.Atoi(strings.TrimSpace(s.inputs[8].Value()))
	if aiMaxH <= 0 {
		aiMaxH = 20
	}

	return SettingsResult{
		Done:         true,
		SaveReq:      true,
		Lang:         selectedLang,
		Interval:     selectedInterval,
		CPUThresh:    cpuT,
		RAMThresh:    ramT,
		DiskThresh:   diskT,
		LogDir:       finalLogDir,
		AIProvider:   aiProv,
		AIEndpoint:   aiEp,
		AIKey:        aiKey,
		AIModel:      aiModel,
		AIRetention:  aiRet,
		AIMaxHistory: aiMaxH,
	}
}

