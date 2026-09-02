package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (s *SettingsModal) View(width, height int) string {
	if width <= 0 {
		width = 100
	}
	if height <= 0 {
		height = 30
	}

	modalWidth := width - 4
	modalHeight := height - 2
	if modalWidth < 65 {
		modalWidth = 65
	}
	if modalHeight < 15 {
		modalHeight = 15
	}

	var b strings.Builder

	headerTitle := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("settings_title"))
	b.WriteString(headerTitle + "\n\n")

	var tabGenHeader, tabTelemHeader, tabAboutHeader string
	if s.activeTab == TabGeneral {
		tabGenHeader = lipgloss.NewStyle().Bold(true).Foreground(ColorBg).Background(ColorPrimary).Padding(0, 2).Render("[1] " + i18n.T("tab_general") + " ●")
		tabTelemHeader = lipgloss.NewStyle().Foreground(ColorMuted).Background(ColorHighlight).Padding(0, 2).Render("[2] 📊 Telemetry ○")
		tabAboutHeader = lipgloss.NewStyle().Foreground(ColorMuted).Background(ColorHighlight).Padding(0, 2).Render("[3] " + i18n.T("tab_about") + " ○")
	} else if s.activeTab == TabTelemetry {
		tabGenHeader = lipgloss.NewStyle().Foreground(ColorMuted).Background(ColorHighlight).Padding(0, 2).Render("[1] " + i18n.T("tab_general") + " ○")
		tabTelemHeader = lipgloss.NewStyle().Bold(true).Foreground(ColorBg).Background(ColorSecondary).Padding(0, 2).Render("[2] 📊 Telemetry ●")
		tabAboutHeader = lipgloss.NewStyle().Foreground(ColorMuted).Background(ColorHighlight).Padding(0, 2).Render("[3] " + i18n.T("tab_about") + " ○")
	} else {
		tabGenHeader = lipgloss.NewStyle().Foreground(ColorMuted).Background(ColorHighlight).Padding(0, 2).Render("[1] " + i18n.T("tab_general") + " ○")
		tabTelemHeader = lipgloss.NewStyle().Foreground(ColorMuted).Background(ColorHighlight).Padding(0, 2).Render("[2] 📊 Telemetry ○")
		tabAboutHeader = lipgloss.NewStyle().Bold(true).Foreground(ColorBg).Background(ColorSuccess).Padding(0, 2).Render("[3] " + i18n.T("tab_about") + " ●")
	}

	b.WriteString(tabGenHeader + "   " + tabTelemHeader + "   " + tabAboutHeader + "\n\n")

	if s.activeTab == TabGeneral {
		b.WriteString(s.renderGeneralTab())
	} else if s.activeTab == TabTelemetry {
		b.WriteString(s.renderTelemetryTab())
	} else {
		b.WriteString(s.renderAboutTab())
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 3).
		Width(modalWidth).
		Height(modalHeight)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, boxStyle.Render(b.String()))
}

func (s *SettingsModal) renderGeneralTab() string {
	var b strings.Builder

	// 1. Language selector
	langLabel := i18n.T("settings_lang_label")
	if s.focusField == FieldLanguage {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("❯ " + langLabel + " "))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("  " + langLabel + " "))
	}

	currLangOpt := i18n.SupportedLangs[s.selectedLangIndex]
	langBtn := lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(fmt.Sprintf("◄ %s ►", currLangOpt.Label))
	b.WriteString(langBtn + "\n\n")

	// 2. Vault Password Change Header
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("🔒 Master Vault Rekeying / Password Change") + "\n\n")

	prompts := []string{
		i18n.T("settings_curr_pass"),
		i18n.T("settings_new_pass"),
		i18n.T("settings_confirm_pass"),
	}

	placeholders := []string{
		i18n.T("settings_ph_curr_pass"),
		i18n.T("settings_ph_new_pass"),
		i18n.T("settings_ph_confirm_pass"),
	}

	for i := 0; i < 3; i++ {
		field := SettingsField(int(FieldCurrentPass) + i)
		inp := s.inputs[i]
		inp.Prompt = fmt.Sprintf("%-26s ", prompts[i])
		inp.Placeholder = placeholders[i]
		inp.Width = 55

		if s.focusField == field {
			inp.PromptStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
			inp.TextStyle = lipgloss.NewStyle().Foreground(ColorText)
			b.WriteString("❯ " + inp.View() + "\n\n")
		} else {
			inp.PromptStyle = lipgloss.NewStyle().Foreground(ColorMuted)
			inp.TextStyle = lipgloss.NewStyle().Foreground(ColorMuted)
			b.WriteString("  " + inp.View() + "\n\n")
		}
	}

	b.WriteString(RenderSecurityBadges(s.inputs...))

	// Submit Button / Hints
	if s.focusField == FieldSubmitBtn {
		btn := lipgloss.NewStyle().Bold(true).Foreground(ColorBg).Background(ColorSuccess).Padding(0, 2).Render(i18n.T("settings_btn_save"))
		b.WriteString("  " + btn + "   " + lipgloss.NewStyle().Foreground(ColorWarning).Render("(Press Enter to Save)") + "\n")
	} else {
		btn := lipgloss.NewStyle().Foreground(ColorMuted).Padding(0, 2).Render(i18n.T("settings_btn_save"))
		b.WriteString("  " + btn + "\n")
	}

	if s.errMessage != "" {
		b.WriteString("\n" + lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render("❌ "+s.errMessage) + "\n")
	} else if s.successMessage != "" {
		b.WriteString("\n" + lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render("✨ "+s.successMessage) + "\n")
	} else {
		b.WriteString("\n")
	}

	hints := lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("settings_footer_gen"))
	b.WriteString(hints)

	return b.String()
}

func (s *SettingsModal) renderAboutTab() string {
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("🚀 "+i18n.T("about_app_name")) + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("about_tagline")) + "\n\n")

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("👤 "+i18n.T("about_creator_title")) + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorText).Render("  • Author: "+i18n.T("about_creator_name")) + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorText).Render("  • Engine: Go 1.22 + Bubbletea TUI + SQLite Encrypted Storage") + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorText).Render("  • License: MIT Open Source License") + "\n\n")

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render("🎯 "+i18n.T("about_vision_title")) + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("  "+i18n.T("about_vision_desc")) + "\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("settings_footer_about")))
	return b.String()
}

func (s *SettingsModal) renderTelemetryTab() string {
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("📊 텔레메트리 폴링 및 자원 위험 임계치 설정") + "\n\n")

	// 1. Telemetry Polling Interval
	intervalLabel := i18n.T("settings_interval_label")
	if s.focusField == FieldInterval {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("❯ " + intervalLabel + " "))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("  " + intervalLabel + " "))
	}
	currIntervalOpt := intervalOptions[s.intervalIndex]
	intervalBtn := lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(fmt.Sprintf("◄ %s ►", currIntervalOpt.Label))
	b.WriteString(intervalBtn + "\n\n")

	// 2. CPU Threshold
	b.WriteString(s.renderThresholdField(FieldCPUThresh, 3, "💻 CPU 경고 임계치 (CPU Threshold %):", "85"))

	// 3. RAM Threshold
	b.WriteString(s.renderThresholdField(FieldRAMThresh, 4, "🧠 RAM 경고 임계치 (RAM Threshold %):", "90"))

	// 4. Disk Threshold
	b.WriteString(s.renderThresholdField(FieldDiskThresh, 5, "💾 Disk 경고 임계치 (Disk Threshold %):", "90"))

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

	hints := lipgloss.NewStyle().Foreground(ColorMuted).Render("[Tab/Shift+Tab] 이동  [←/→] 수치/옵션 조절  [1~3] 탭  [Esc] 닫기")
	b.WriteString(hints)

	return b.String()
}

func (s *SettingsModal) renderThresholdField(field SettingsField, inputIdx int, label, ph string) string {
	inp := s.inputs[inputIdx]
	inp.Prompt = fmt.Sprintf("%-38s ", label)
	inp.Placeholder = ph
	inp.Width = 10

	if s.focusField == field {
		inp.PromptStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
		inp.TextStyle = lipgloss.NewStyle().Foreground(ColorText)
		return "❯ " + inp.View() + " %\n\n"
	}
	inp.PromptStyle = lipgloss.NewStyle().Foreground(ColorMuted)
	inp.TextStyle = lipgloss.NewStyle().Foreground(ColorMuted)
	return "  " + inp.View() + " %\n\n"
}
