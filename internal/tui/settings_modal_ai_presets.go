package tui

import (
	"leitstand/internal/ai"
)

// applyAIProviderPreset saves current input values into the previous provider profile
// and restores or initializes the profile for the newly selected provider.
func (s *SettingsModal) applyAIProviderPreset(prevProvider string) {
	if s.aiProviderIndex < 0 || s.aiProviderIndex >= len(s.aiProviders) || len(s.inputs) < 9 {
		return
	}

	if s.providerProfiles == nil {
		s.providerProfiles = make(map[string]*AIProviderProfile)
	}

	// 1. Cache current field values into previous provider profile
	if prevProvider != "" {
		s.providerProfiles[prevProvider] = &AIProviderProfile{
			Endpoint: s.inputs[4].Value(),
			APIKey:   s.inputs[5].Value(),
			Model:    s.inputs[6].Value(),
		}
	}

	currProvider := s.aiProviders[s.aiProviderIndex]

	// 2. Restore cached profile if available, otherwise load defaults or DB
	if prof, ok := s.providerProfiles[currProvider]; ok && prof != nil {
		s.inputs[4].SetValue(prof.Endpoint)
		s.inputs[5].SetValue(prof.APIKey)
		s.inputs[6].SetValue(prof.Model)
	} else {
		// Try DB first
		ep := ""
		key := ""
		model := ""
		if s.store != nil {
			ep, _ = s.store.GetSetting("ai_" + currProvider + "_endpoint")
			key, _ = s.store.GetSetting("ai_" + currProvider + "_api_key")
			model, _ = s.store.GetSetting("ai_" + currProvider + "_model")
		}
		if ep == "" {
			ep = ai.GetDefaultEndpoint(currProvider)
		}
		if model == "" {
			model = ai.GetDefaultModel(currProvider)
		}
		s.inputs[4].SetValue(ep)
		s.inputs[5].SetValue(key)
		s.inputs[6].SetValue(model)
	}

	// 3. Update placeholder per provider
	switch currProvider {
	case ai.ProviderGroq:
		s.inputs[5].Placeholder = "gsk_... (Groq Cloud API Key)"
	case ai.ProviderOpenAI:
		s.inputs[5].Placeholder = "sk-... (OpenAI API Key)"
	case ai.ProviderOllama:
		s.inputs[5].Placeholder = "(Optional for local Ollama)"
	default:
		s.inputs[5].Placeholder = "API Key / Bearer Token"
	}
}
