package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type SFTPRunbookItem struct {
	Title   string
	Command string
	Guide   string
}

func getSFTPRunbooks() []SFTPRunbookItem {
	return []SFTPRunbookItem{
		{i18n.T("sftp_rb_tar_pack_title"), "tar -czvf archive.tar.gz .", i18n.T("sftp_rb_tar_pack_guide")},
		{i18n.T("sftp_rb_tar_unpack_title"), "tar -xzvf archive.tar.gz", i18n.T("sftp_rb_tar_unpack_guide")},
		{i18n.T("sftp_rb_zip_pack_title"), "zip -r archive.zip .", i18n.T("sftp_rb_zip_pack_guide")},
		{i18n.T("sftp_rb_zip_unpack_title"), "unzip archive.zip", i18n.T("sftp_rb_zip_unpack_guide")},
		{i18n.T("sftp_rb_chmod_755_all_title"), "chmod -R 755 .", i18n.T("sftp_rb_chmod_755_all_guide")},
		{i18n.T("sftp_rb_chmod_755_dir_title"), "find . -type d -exec chmod 755 {} +", i18n.T("sftp_rb_chmod_755_dir_guide")},
		{i18n.T("sftp_rb_chmod_644_file_title"), "find . -type f -exec chmod 644 {} +", i18n.T("sftp_rb_chmod_644_file_guide")},
		{i18n.T("sftp_rb_chown_www_title"), "chown -R www-data:www-data .", i18n.T("sftp_rb_chown_www_guide")},
		{i18n.T("sftp_rb_du_usage_title"), "du -h --max-depth=1 | sort -hr", i18n.T("sftp_rb_du_usage_guide")},
		{i18n.T("sftp_rb_find_large_title"), "find . -type f -size +100M", i18n.T("sftp_rb_find_large_guide")},
	}
}

func (m *FileManagerModal) renderRunbookModal(width, height int) string {
	var b strings.Builder
	boxWidth := width - 6
	if boxWidth < 55 {
		boxWidth = 55
	}

	runbooks := getSFTPRunbooks()

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("sftp_runbook_title")) + "\n\n")

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render(i18n.T("sftp_cheatsheet_title")) + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorText).Render(i18n.T("sftp_cheatsheet_line1")) + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorText).Render(i18n.T("sftp_cheatsheet_line2")) + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorText).Render(i18n.T("sftp_cheatsheet_line3")) + "\n\n")

	headerTitle := fmt.Sprintf(i18n.T("sftp_runbook_header"), m.RunbookCursor+1, len(runbooks))
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render(headerTitle) + "\n")

	visibleCount := (height - 14) / 3
	if visibleCount < 2 {
		visibleCount = 2
	}

	startIdx := 0
	if m.RunbookCursor >= visibleCount {
		startIdx = m.RunbookCursor - visibleCount + 1
	}
	endIdx := startIdx + visibleCount
	if endIdx > len(runbooks) {
		endIdx = len(runbooks)
		if endIdx-visibleCount >= 0 {
			startIdx = endIdx - visibleCount
		}
	}

	for i := startIdx; i < endIdx; i++ {
		rb := runbooks[i]
		if i == m.RunbookCursor {
			b.WriteString(lipgloss.NewStyle().Bold(true).Background(ColorHighlight).Foreground(lipgloss.Color("#FFFFFF")).Render(fmt.Sprintf(" ▶ %s ", rb.Title)) + "\n")
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(fmt.Sprintf("    ➔ %s", rb.Command)) + "\n")
			b.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Render(fmt.Sprintf("    💡 %s", rb.Guide)) + "\n\n")
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(ColorText).Render(fmt.Sprintf("   %s", rb.Title)) + "\n")
			b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("    ➔ %s", rb.Command)) + "\n\n")
		}
	}

	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("sftp_runbook_nav")))

	boxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorSecondary).Padding(1, 2).Width(boxWidth).Height(height - 2)
	content := boxStyle.Render(b.String())
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
