package tui

import (
	"leitstand/internal/i18n"
)

// handleLeftRight handles left/right arrow key navigation across different settings tabs.
func (s *SettingsModal) handleLeftRight(delta int) bool {
	if s.activeTab == TabGeneral && s.focusField == FieldLanguage {
		s.selectedLangIndex = (s.selectedLangIndex + delta + len(i18n.SupportedLangs)) % len(i18n.SupportedLangs)
		i18n.SetLang(i18n.SupportedLangs[s.selectedLangIndex].Code)
		return true
	} else if s.focusField == FieldInterval {
		s.intervalIndex = (s.intervalIndex + delta + len(intervalOptions)) % len(intervalOptions)
		return true
	} else if s.activeTab == TabLogs && s.focusField == FieldLogPreset {
		s.selectedLogPreset = (s.selectedLogPreset + delta + len(s.logDirPresets)) % len(s.logDirPresets)
		return true
	} else if s.activeTab == TabAI && s.focusField == FieldAIProvider {
		prevProvider := s.aiProviders[s.aiProviderIndex]
		s.aiProviderIndex = (s.aiProviderIndex + delta + len(s.aiProviders)) % len(s.aiProviders)
		s.applyAIProviderPreset(prevProvider)
		return true
	}
	return false
}
