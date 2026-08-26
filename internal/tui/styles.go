package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Palette
	ColorPrimary   = lipgloss.Color("#00E5FF") // Cyan / Electric Blue
	ColorSecondary = lipgloss.Color("#7C4DFF") // Deep Purple
	ColorSuccess   = lipgloss.Color("#00E676") // Neon Green
	ColorWarning   = lipgloss.Color("#FFD600") // Amber
	ColorDanger    = lipgloss.Color("#FF1744") // Coral Red
	ColorMuted     = lipgloss.Color("#757575") // Slate Gray
	ColorBorder    = lipgloss.Color("#37474F") // Dark Slate Border
	ColorHighlight = lipgloss.Color("#263238") // Dark Teal Highlight
	ColorBg        = lipgloss.Color("#121212") // Deep Dark Background

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1)

	BadgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorSecondary).
			Padding(0, 1)

	PaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	ActivePaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1)

	SelectedHostStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(ColorHighlight).
				Padding(0, 1)

	UnselectedHostStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#B0BEC5")).
				Padding(0, 1)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ECEFF1")).
			Background(lipgloss.Color("#1E293B")).
			Padding(0, 1)

	WarningBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorWarning).
			Foreground(ColorWarning).
			Padding(1, 2).
			Align(lipgloss.Center)
)

// RenderProgressBar renders a sleek terminal progress bar.
func RenderProgressBar(width int, percent float64, filledColor, emptyColor lipgloss.Color) string {
	if width < 5 {
		width = 10
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filledLen := int((percent / 100.0) * float64(width))
	if filledLen > width {
		filledLen = width
	}
	emptyLen := width - filledLen

	filledStyle := lipgloss.NewStyle().Foreground(filledColor)
	emptyStyle := lipgloss.NewStyle().Foreground(emptyColor)

	var filledStr, emptyStr string
	for i := 0; i < filledLen; i++ {
		filledStr += "█"
	}
	for i := 0; i < emptyLen; i++ {
		emptyStr += "░"
	}

	return filledStyle.Render(filledStr) + emptyStyle.Render(emptyStr)
}
