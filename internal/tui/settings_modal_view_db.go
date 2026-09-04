package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var dbActions = []struct {
	TitleKey string
	DescKey  string
}{
	{"db_action_prune_title", "db_action_prune_desc"},
	{"db_action_export_metrics_title", "db_action_export_metrics_desc"},
	{"db_action_export_hosts_title", "db_action_export_hosts_desc"},
	{"db_action_import_hosts_title", "db_action_import_hosts_desc"},
	{"db_action_rekey_title", "db_action_rekey_desc"},
	{"db_action_reset_title", "db_action_reset_desc"},
}

var RetentionDays = []int{7, 14, 30}

func (s *SettingsModal) renderDatabaseTab() string {
	if s.showRekeyModal {
		return s.renderRekeyModal()
	}

	var b strings.Builder

	// 1. Top Database Summary Card
	dbPath := "data/leitstand.db"
	var dbSizeStr string = "0 KB"
	hostCount := 0
	var metricCount int64 = 0
	if s.dbStats != nil {
		if s.dbStats.Path != "" {
			dbPath = s.dbStats.Path
		}
		if s.dbStats.SizeBytes < 1024*1024 {
			dbSizeStr = fmt.Sprintf("%.1f KB", float64(s.dbStats.SizeBytes)/1024.0)
		} else {
			dbSizeStr = fmt.Sprintf("%.2f MB", float64(s.dbStats.SizeBytes)/(1024.0*1024.0))
		}
		hostCount = s.dbStats.HostCount
		metricCount = s.dbStats.MetricCount
	}

	var retentionBadges []string
	for i, d := range RetentionDays {
		label := i18n.Tf("db_days_format", d)
		if d == 7 {
			label = i18n.T("db_days_recommended")
		}
		if i == s.retentionIndex {
			retentionBadges = append(retentionBadges, lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorBg).
				Background(ColorSecondary).
				Padding(0, 1).
				Render("● "+label))
		} else {
			retentionBadges = append(retentionBadges, lipgloss.NewStyle().
				Foreground(ColorMuted).
				Background(ColorHighlight).
				Padding(0, 1).
				Render("○ "+label))
		}
	}
	retentionLine := "• " + i18n.T("db_retention_label") + ": " + strings.Join(retentionBadges, " ")

	summaryBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSecondary).
		Padding(0, 1).
		Width(68).
		Render(fmt.Sprintf(
			"%s\n• %s: %s   • %s: %s\n• %s: %d   • %s: %d\n%s",
			lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("🗄️ "+i18n.T("db_status_title")),
			i18n.T("db_path"), lipgloss.NewStyle().Foreground(ColorText).Render(dbPath),
			i18n.T("db_size"), lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(dbSizeStr),
			i18n.T("db_hosts_count"), hostCount,
			i18n.T("db_metrics_count"), metricCount,
			retentionLine,
		))
	b.WriteString(summaryBox + "\n\n")

	// 2. Execution Result Banner (Success or Error feedback)
	if s.successMessage != "" {
		b.WriteString(lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSuccess).
			Background(lipgloss.Color("#102A18")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSuccess).
			Padding(0, 1).
			Width(68).
			Render("✨ "+s.successMessage) + "\n\n")
	} else if s.errMessage != "" {
		b.WriteString(lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorDanger).
			Background(lipgloss.Color("#2A1010")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDanger).
			Padding(0, 1).
			Width(68).
			Render("❌ "+s.errMessage) + "\n\n")
	}

	// 3. Action List
	for i, act := range dbActions {
		cursor := "  "
		if i == s.dbActionIndex {
			cursor = "▶ "
		}

		title := i18n.T(act.TitleKey)
		desc := i18n.T(act.DescKey)

		if i == s.dbActionIndex {
			titleLine := lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Background(lipgloss.Color("#1B2A32")).Render(cursor + title)
			descLine := lipgloss.NewStyle().Foreground(lipgloss.Color("#B0BEC5")).Render("     " + desc)
			b.WriteString(titleLine + "\n" + descLine + "\n")
		} else {
			titleLine := lipgloss.NewStyle().Foreground(ColorText).Render(cursor + title)
			b.WriteString(titleLine + "\n")
		}
	}

	// 4. Confirmation Dialog Overlay or Navigation Tip
	if s.dbConfirmAction > 0 {
		b.WriteString("\n")
		confirmBox := lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorWarning).
			Background(lipgloss.Color("#241A10")).
			Padding(0, 1).
			Width(68).
			Render(fmt.Sprintf(
				"❓ %s\n   %s",
				lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(s.dbConfirmPrompt),
				lipgloss.NewStyle().Foreground(ColorText).Render(i18n.T("db_confirm_btns")),
			))
		b.WriteString(confirmBox + "\n")
	} else {
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("settings_db_tip")) + "\n")
	}

	return b.String()
}
