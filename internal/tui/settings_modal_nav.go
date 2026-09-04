package tui

func (s *SettingsModal) switchTab(newTab SettingsTab) {
	s.blurCurrent()
	s.activeTab = newTab
	s.errMessage = ""
	s.successMessage = ""
	switch newTab {
	case TabGeneral:
		s.focusField = FieldLanguage
	case TabTelemetry:
		s.focusField = FieldInterval
	case TabLogs:
		s.focusField = FieldLogPreset
	case TabAI:
		s.focusField = FieldAIProvider
	default:
		s.focusField = FieldLanguage
	}
	s.focusCurrent()
}

func (s *SettingsModal) focusNext() {
	s.blurCurrent()
	switch s.activeTab {
	case TabGeneral:
		if s.focusField == FieldLanguage {
			s.focusField = FieldSubmitBtn
		} else {
			s.focusField = FieldLanguage
		}
	case TabTelemetry:
		fields := []SettingsField{FieldInterval, FieldCPUThresh, FieldRAMThresh, FieldDiskThresh, FieldSubmitBtn}
		s.focusField = cycleField(s.focusField, fields, 1)
	case TabLogs:
		if s.selectedLogPreset == 3 {
			if s.focusField == FieldLogPreset {
				s.focusField = FieldLogCustomDir
			} else if s.focusField == FieldLogCustomDir {
				s.focusField = FieldSubmitBtn
			} else {
				s.focusField = FieldLogPreset
			}
		} else {
			if s.focusField == FieldLogPreset {
				s.focusField = FieldSubmitBtn
			} else {
				s.focusField = FieldLogPreset
			}
		}
	case TabAI:
		fields := []SettingsField{FieldAIProvider, FieldAIEndpoint, FieldAIKey, FieldAIModel, FieldAIRetention, FieldAIMaxHistory, FieldSubmitBtn}
		s.focusField = cycleField(s.focusField, fields, 1)
	}
	s.focusCurrent()
}

func (s *SettingsModal) focusPrev() {
	s.blurCurrent()
	switch s.activeTab {
	case TabGeneral:
		if s.focusField == FieldSubmitBtn {
			s.focusField = FieldLanguage
		} else {
			s.focusField = FieldSubmitBtn
		}
	case TabTelemetry:
		fields := []SettingsField{FieldInterval, FieldCPUThresh, FieldRAMThresh, FieldDiskThresh, FieldSubmitBtn}
		s.focusField = cycleField(s.focusField, fields, -1)
	case TabLogs:
		if s.selectedLogPreset == 3 {
			if s.focusField == FieldSubmitBtn {
				s.focusField = FieldLogCustomDir
			} else if s.focusField == FieldLogCustomDir {
				s.focusField = FieldLogPreset
			} else {
				s.focusField = FieldSubmitBtn
			}
		} else {
			if s.focusField == FieldSubmitBtn {
				s.focusField = FieldLogPreset
			} else {
				s.focusField = FieldSubmitBtn
			}
		}
	case TabAI:
		fields := []SettingsField{FieldAIProvider, FieldAIEndpoint, FieldAIKey, FieldAIModel, FieldAIRetention, FieldAIMaxHistory, FieldSubmitBtn}
		s.focusField = cycleField(s.focusField, fields, -1)
	}
	s.focusCurrent()
}


func cycleField(curr SettingsField, allowed []SettingsField, delta int) SettingsField {
	idx := -1
	for i, f := range allowed {
		if f == curr {
			idx = i
			break
		}
	}
	if idx == -1 {
		return allowed[0]
	}
	next := (idx + delta + len(allowed)) % len(allowed)
	return allowed[next]
}
