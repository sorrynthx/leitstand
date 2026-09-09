package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"leitstand/internal/profile"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderAboutTab renders the 50:50 split view with German beer ASCII art.
func (s *SettingsModal) renderAboutTab(totalWidth int) string {
	prof, links := profile.GetProfile(string(i18n.GetLang()))
	if totalWidth < 70 {
		totalWidth = 70
	}

	var b strings.Builder

	// 1. Top Header: German Beer Mug ASCII Art & App Tagline
	beerArt := s.renderBeerHeader(prof.Name)
	b.WriteString(beerArt + "\n")

	// 2. 50:50 Split Layout
	colWidth := (totalWidth - 8) / 2
	if colWidth < 30 {
		colWidth = 30
	}

	leftCol := s.renderAboutLeftCol(prof, colWidth)
	rightCol := s.renderAboutRightCol(prof, links, colWidth)

	divider := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#2A3342")).
		Padding(0, 1).
		Render("│\n│\n│\n│\n│\n│\n│\n│\n│\n│\n│\n│\n│\n│\n│\n│")

	splitView := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, divider, rightCol)
	b.WriteString(splitView + "\n\n")

	// 3. Footer
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("settings_footer_about")))
	return b.String()
}

func (s *SettingsModal) renderBeerHeader(devName string) string {
	foamStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	mugStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F1C40F"))
	handleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#D4AC0D"))
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	subStyle := lipgloss.NewStyle().Foreground(ColorMuted)
	tagStyle := lipgloss.NewStyle().Italic(true).Foreground(ColorSuccess)

	asciiMug := foamStyle.Render("       .---~~~~---.") + "\n" +
		foamStyle.Render("      (   ~ ~ ~ ~  )") + "\n" +
		foamStyle.Render("    (  ~  ") + mugStyle.Render("PROST!") + foamStyle.Render(" ~  )") + "    " + titleStyle.Render("⚡ LEITSTAND") + "\n" +
		foamStyle.Render("   ( ~ ~ ~ ~ ~ ~ ~ ~ )") + "   " + subStyle.Render("Modern, Agentless Terminal Server Cockpit") + "\n" +
		mugStyle.Render("  |-------------------|") + "\n" +
		mugStyle.Render("  |  .-------------.  |") + handleStyle.Render("===.") + "\n" +
		mugStyle.Render("  |  |  LEITSTAND  |  |") + handleStyle.Render("   |") + "  👤 " + lipgloss.NewStyle().Bold(true).Foreground(ColorText).Render(devName) + "\n" +
		mugStyle.Render("  |  | PURE GO TUI |  |") + handleStyle.Render("===·") + "  " + tagStyle.Render("\"Zero-Agent. Zero-Knowledge. 100% Pure Go.\"") + "\n" +
		mugStyle.Render("  |  '-------------'  |") + "\n" +
		mugStyle.Render("  '-------------------'")

	return asciiMug
}

func (s *SettingsModal) renderAboutLeftCol(prof profile.LangProfile, width int) string {
	boxStyle := lipgloss.NewStyle().Width(width).PaddingRight(1)

	var sb strings.Builder
	head := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("📖 " + prof.StoryTitle)
	sb.WriteString(head + "\n\n")

	bodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CFD8DC"))

	sb.WriteString(bodyStyle.Render(prof.StoryP1) + "\n\n")
	sb.WriteString(bodyStyle.Render(prof.StoryP2) + "\n\n")
	sb.WriteString(bodyStyle.Render(prof.StoryP3) + "\n\n")

	retryTag := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorSuccess).
		Render("🌱 Retry.Day — 실패해도 다시 일어서는 단단한 삶을 지향합니다.")
	if i18n.GetLang() != "ko" {
		retryTag = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSuccess).
			Render("🌱 Retry.Day — Resilience to stand up again and build forward.")
	}
	sb.WriteString(retryTag)

	return boxStyle.Render(sb.String())
}

func (s *SettingsModal) renderAboutRightCol(prof profile.LangProfile, links profile.Links, width int) string {
	boxStyle := lipgloss.NewStyle().Width(width).PaddingLeft(1)

	var sb strings.Builder

	// Career / Teammate Pitch
	careerHead := lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(prof.CareerTitle)
	careerDesc := lipgloss.NewStyle().Foreground(lipgloss.Color("#ECEFF1")).Render(prof.CareerDesc)
	sb.WriteString(careerHead + "\n")
	sb.WriteString(careerDesc + "\n\n")

	// Engineering Core
	techHead := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(prof.TechTitle)
	sb.WriteString(techHead + "\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(prof.TechP1) + "\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(prof.TechP2) + "\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(prof.TechP3) + "\n\n")

	// Links & Projects
	linkHead := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("🔗 Connect & Web Services")
	sb.WriteString(linkHead + "\n")
	sb.WriteString(fmt.Sprintf("  • Web    : %s\n", lipgloss.NewStyle().Foreground(ColorSuccess).Render(links.Web)))
	sb.WriteString(fmt.Sprintf("  • GitHub : %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#64B5F6")).Render(links.Git)))
	sb.WriteString(fmt.Sprintf("  • In     : %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#4FC3F7")).Render(links.LinkedIn)))

	return boxStyle.Render(sb.String())
}
