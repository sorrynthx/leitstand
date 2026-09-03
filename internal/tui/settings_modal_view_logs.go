package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (s *SettingsModal) renderLogsTab() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	mutedStyle := lipgloss.NewStyle().Foreground(ColorMuted)
	activeStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess)

	b.WriteString(titleStyle.Render(i18n.T("settings_logs_title")) + "\n")
	b.WriteString(mutedStyle.Render(i18n.T("settings_logs_desc")) + "\n\n")

	for i, preset := range s.logDirPresets {
		cursor := "  "
		radio := "○ "
		if i == s.selectedLogPreset {
			cursor = "❯ "
			radio = "● "
		}

		line := fmt.Sprintf("%s%s[%d] %s", cursor, radio, i+1, preset.Label)
		if i == s.selectedLogPreset {
			b.WriteString(activeStyle.Render(line) + "\n")
		} else {
			b.WriteString(mutedStyle.Render(line) + "\n")
		}

		if preset.Path != "" {
			indentPath := lipgloss.NewStyle().Foreground(ColorMuted).PaddingLeft(6).Render("Path: " + preset.Path)
			b.WriteString(indentPath + "\n")
		}
	}

	b.WriteString("\n")

	// Custom Directory input if preset 3 is selected
	if s.selectedLogPreset == 3 {
		inputHeader := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("Specify Custom Directory Path:")
		b.WriteString(inputHeader + "\n")
		b.WriteString(s.inputs[3].View() + "\n\n")
	}

	tipPrefix := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("💡 ")
	tipText := lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("settings_logs_tip"))
	b.WriteString(tipPrefix + tipText + "\n")

	return b.String()
}
