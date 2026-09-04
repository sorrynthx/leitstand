package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ViewInline renders the compact, high-efficiency inline AI Copilot bar docked at the bottom of the console tab.
func (m *AICopilotModal) ViewInline(width int) string {
	if width < 30 {
		width = 30
	}
	innerWidth := width - 4
	if innerWidth < 20 {
		innerWidth = 20
	}

	var b strings.Builder

	// 1. Title bar
	title := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("🤖 " + i18n.T("ai_inline_title"))
	modelBadge := lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("[%s]", m.cfg.AI.Model))
	var ctxInfo string
	if m.OSDistro != "" {
		ctxInfo = lipgloss.NewStyle().Foreground(lipgloss.Color("#90CAF9")).Render(" • " + m.OSDistro)
	}
	escHint := lipgloss.NewStyle().Foreground(ColorMuted).Render("[Esc: " + i18n.T("btn_close") + "]")

	leftHeader := lipgloss.JoinHorizontal(lipgloss.Center, title, " ", modelBadge, ctxInfo)
	gapW := innerWidth - lipgloss.Width(leftHeader) - lipgloss.Width(escHint) - 2
	if gapW < 1 {
		gapW = 1
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, leftHeader, strings.Repeat(" ", gapW), escHint) + "\n")

	// 2. Input Box
	m.Input.Prompt = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("🤖 ❯ ")
	m.Input.Placeholder = i18n.T("ai_inline_ph")
	m.Input.Width = innerWidth - 6
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Width(innerWidth - 2).
		Padding(0, 1).
		Render(m.Input.View())
	b.WriteString(inputBox + "\n")

	// 3. Dynamic Response or Actions
	if m.IsStreaming {
		streamBadge := lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(i18n.T("ai_streaming_status"))
		streamSnippet := m.StreamingContent
		if len(streamSnippet) > 100 {
			streamSnippet = "..." + streamSnippet[len(streamSnippet)-100:]
		}
		b.WriteString(streamBadge + " " + lipgloss.NewStyle().Foreground(ColorMuted).Render(streamSnippet) + "\n")
	} else if m.ExtractedCommand != "" {
		if m.Explanation != "" {
			exp := lipgloss.NewStyle().Foreground(lipgloss.Color("#90CAF9")).MaxWidth(innerWidth).Render("💡 " + m.Explanation)
			b.WriteString(exp + "\n")
		}

		if m.IsDangerous {
			warnBadge := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(ColorDanger).Padding(0, 1).Render(i18n.T("ai_warn_dangerous_cmd"))
			cmdCode := lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(" " + m.ExtractedCommand)
			b.WriteString(warnBadge + "\n" + cmdCode + "\n")
			btnEdit := lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(i18n.T("ai_btn_edit_cmd") + " (수동 확인)")
			btnEsc := lipgloss.NewStyle().Foreground(ColorMuted).Render("   [Esc: 취소]")
			b.WriteString(btnEdit + btnEsc + "\n")
		} else {
			cmdBadge := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#121212")).Background(ColorSuccess).Padding(0, 1).Render(i18n.T("ai_cmd_prefix"))
			cmdCode := lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(" " + m.ExtractedCommand)
			b.WriteString(cmdBadge + cmdCode + "\n")

			btnRun := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#121212")).Background(ColorSuccess).Padding(0, 1).Render(i18n.T("ai_btn_run_now"))
			btnEdit := lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render("   " + i18n.T("ai_btn_edit_cmd"))
			b.WriteString(btnRun + btnEdit + "\n")
		}
	} else if m.Explanation != "" {
		exp := lipgloss.NewStyle().Foreground(lipgloss.Color("#90CAF9")).MaxWidth(innerWidth).Render("💡 " + m.Explanation)
		btnEsc := lipgloss.NewStyle().Foreground(ColorMuted).Render("  [Esc: 닫기]")
		b.WriteString(exp + btnEsc + "\n")
	} else if m.StatusMessage != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorWarning).Render(m.StatusMessage) + "\n")
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSecondary).
		Width(width).
		Padding(0, 1).
		Render(b.String())
}
