package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (s *SettingsModal) renderAITab() string {
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("🤖 "+i18n.T("settings_tab_ai")) + "\n\n")

	// 1. Provider selector
	pLabel := i18n.T("settings_ai_provider_label")
	if s.focusField == FieldAIProvider {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("❯ " + pLabel + " "))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("  " + pLabel + " "))
	}
	currProvider := s.aiProviders[s.aiProviderIndex]
	pBtn := lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(fmt.Sprintf("◄ %s ►", currProvider))
	b.WriteString(pBtn + "\n\n")

	// 2. Endpoint
	b.WriteString(s.renderAIInputField(FieldAIEndpoint, 4, i18n.T("settings_ai_endpoint_label")))

	// 3. API Key (Vault Encrypted)
	b.WriteString(s.renderAIInputField(FieldAIKey, 5, i18n.T("settings_ai_key_label")))

	// 4. Model Name
	b.WriteString(s.renderAIInputField(FieldAIModel, 6, i18n.T("settings_ai_model_label")))

	// 5. Retention Days
	b.WriteString(s.renderAIInputField(FieldAIRetention, 7, i18n.T("settings_ai_retention_label")))

	// 6. Max History
	b.WriteString(s.renderAIInputField(FieldAIMaxHistory, 8, i18n.T("settings_ai_max_history_label")))

	// Save Button
	if s.focusField == FieldSubmitBtn {
		btn := lipgloss.NewStyle().Bold(true).Foreground(ColorBg).Background(ColorSuccess).Padding(0, 2).Render(i18n.T("settings_btn_save"))
		b.WriteString("  " + btn + "   " + lipgloss.NewStyle().Foreground(ColorWarning).Render("(Press Enter to Save)") + "\n")
	} else {
		btn := lipgloss.NewStyle().Foreground(ColorMuted).Padding(0, 2).Render(i18n.T("settings_btn_save"))
		b.WriteString("  " + btn + "\n")
	}

	if s.errMessage != "" {
		b.WriteString("\n" + lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render("❌ "+s.errMessage) + "\n")
	} else {
		b.WriteString("\n")
	}

	hints := lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("settings_hints_footer"))
	b.WriteString(hints)

	return b.String()
}

func (s *SettingsModal) renderAIInputField(field SettingsField, inputIdx int, label string) string {
	inp := s.inputs[inputIdx]
	inp.Prompt = fmt.Sprintf("%-28s ", label)

	if s.focusField == field {
		inp.PromptStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
		inp.TextStyle = lipgloss.NewStyle().Foreground(ColorText)
		return "❯ " + inp.View() + "\n\n"
	}
	inp.PromptStyle = lipgloss.NewStyle().Foreground(ColorMuted)
	inp.TextStyle = lipgloss.NewStyle().Foreground(ColorMuted)
	return "  " + inp.View() + "\n\n"
}
