package tui

import (
	"fmt"
	"leitstand/internal/ai"

	"github.com/charmbracelet/bubbles/textinput"
)

func createSettingsInputs(cpuThresh, ramThresh, diskThresh float64, currLogDir string, presetIdx int) []textinput.Model {
	inputs := make([]textinput.Model, 9)

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

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "e.g. D:\\MyLogs or /var/log/leitstand"
	if presetIdx == 3 {
		inputs[3].SetValue(currLogDir)
	}
	inputs[3].Width = 48

	inputs[4] = textinput.New()
	inputs[4].Placeholder = "https://api.groq.com/openai/v1"
	inputs[4].Width = 42

	inputs[5] = textinput.New()
	inputs[5].Placeholder = "gsk_... or sk-..."
	inputs[5].EchoMode = textinput.EchoPassword
	inputs[5].EchoCharacter = '•'
	inputs[5].Width = 42

	inputs[6] = textinput.New()
	inputs[6].Placeholder = "openai/gpt-oss-20b"
	inputs[6].Width = 42

	inputs[7] = textinput.New()
	inputs[7].Placeholder = "3"
	inputs[7].SetValue("3")
	inputs[7].Width = 10

	inputs[8] = textinput.New()
	inputs[8].Placeholder = "20"
	inputs[8].SetValue("20")
	inputs[8].Width = 10

	return inputs
}

func (s *SettingsModal) initAISettings() {
	if s.store == nil {
		return
	}
	if p, err := s.store.GetSetting("ai_provider"); err == nil && p != "" {
		for i, prov := range s.aiProviders {
			if prov == p {
				s.aiProviderIndex = i
				break
			}
		}
	}

	currProvider := s.aiProviders[s.aiProviderIndex]

	// Try provider-specific keys first, fallback to legacy global keys
	ep, _ := s.store.GetSetting("ai_" + currProvider + "_endpoint")
	if ep == "" {
		ep, _ = s.store.GetSetting("ai_endpoint")
	}
	if ep == "" {
		ep = ai.GetDefaultEndpoint(currProvider)
	}
	s.inputs[4].SetValue(ep)

	key, _ := s.store.GetSetting("ai_" + currProvider + "_api_key")
	if key == "" {
		key, _ = s.store.GetSetting("ai_api_key")
	}
	s.inputs[5].SetValue(key)

	model, _ := s.store.GetSetting("ai_" + currProvider + "_model")
	if model == "" {
		model, _ = s.store.GetSetting("ai_model")
	}
	if model == "" {
		model = ai.GetDefaultModel(currProvider)
	}
	s.inputs[6].SetValue(model)

	if ret, err := s.store.GetSetting("ai_retention_days"); err == nil && ret != "" {
		s.inputs[7].SetValue(ret)
	}
	if maxH, err := s.store.GetSetting("ai_max_history"); err == nil && maxH != "" {
		s.inputs[8].SetValue(maxH)
	}

	// Also cache the initial active provider profile
	if s.providerProfiles == nil {
		s.providerProfiles = make(map[string]*AIProviderProfile)
	}
	s.providerProfiles[currProvider] = &AIProviderProfile{
		Endpoint: ep,
		APIKey:   key,
		Model:    model,
	}
}
