package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"leitstand/internal/ssh"
	"path"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func (m *FileManagerModal) renderTransferDeck(width int) string {
	var b strings.Builder

	if width < 30 {
		width = 30
	}

	if m.ActivePrompt != PromptNone {
		switch m.ActivePrompt {
		case PromptMkdir:
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render("📁 New Folder: ") + m.SubInput.View())
		case PromptTouch:
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("📄 New File: ") + m.SubInput.View())
		case PromptRename:
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render("✏️ Rename: ") + m.SubInput.View())
		case PromptDeleteConfirm:
			paths := m.GetSelectedPaths()
			targetDesc := i18n.T("sftp_delete_selected_items")
			if len(paths) > 0 {
				baseName := path.Base(paths[0])
				if m.ActivePanel == PanelLocal {
					baseName = filepath.Base(paths[0])
				}
				targetDesc = fmt.Sprintf("'%s'", baseName)
				if len(paths) > 1 {
					targetDesc = i18n.Tf("sftp_delete_multi_desc", baseName, len(paths)-1, len(paths))
				}
			}
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(i18n.Tf("sftp_delete_confirm_prompt", targetDesc)))
		case PromptExitConfirm:
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(i18n.T("sftp_prompt_exit")))
		case PromptQuickCmd:
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("⚡ Quick Command: ") + m.QuickCmdInput.View())
		}
		b.WriteString("\n")
	} else if len(m.ClipboardPaths) > 0 {
		clipMode := "Copy"
		if m.ClipboardIsCut {
			clipMode = "Cut"
		}
		b.WriteString(lipgloss.NewStyle().Foreground(ColorWarning).Render(i18n.Tf("sftp_clipboard_wait", len(m.ClipboardPaths), clipMode)) + "\n")
	} else if m.StatusMessage != "" {
		b.WriteString(m.StatusMessage + "\n")
	} else {
		truncHints := runewidth.Truncate(i18n.T("sftp_hints"), width-2, "…")
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(truncHints) + "\n")
	}
	return b.String()
}

func (m *FileManagerModal) renderTransferProgressModal(width, height int) string {
	var b strings.Builder
	title := i18n.T("sftp_upload_title")
	if !m.TransferIsUpload {
		title = i18n.T("sftp_download_title")
	}
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(title) + "\n\n")
	b.WriteString(fmt.Sprintf("File: %s (%d / %d)\n", m.CurrentFileName, m.FileIndex, m.FileTotal))

	pct := 0.0
	if m.CurrentTotal > 0 {
		pct = (float64(m.CurrentBytes) / float64(m.CurrentTotal)) * 100.0
	}
	speedStr := fmt.Sprintf("%.1f KB/s", m.BytesPerSec/1024.0)
	if m.BytesPerSec >= 1024*1024 {
		speedStr = fmt.Sprintf("%.1f MB/s", m.BytesPerSec/(1024.0*1024.0))
	}
	b.WriteString(fmt.Sprintf("[%s] %.1f%%\n", RenderProgressBar(40, pct, ColorSuccess, ColorBorder), pct))
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("Size: %s / %s  •  Speed: %s", formatTransferBytes(m.CurrentBytes), formatTransferBytes(m.CurrentTotal), speedStr)) + "\n\n")

	if m.ShowTransferCancelPrompt {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(i18n.T("sftp_transfer_cancel_title")) + "\n")
		b.WriteString(lipgloss.NewStyle().Foreground(ColorWarning).Render(i18n.T("sftp_transfer_cancel_confirm")))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorWarning).Render(i18n.T("sftp_transfer_wait")) + "\n")
		b.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Render(i18n.T("sftp_transfer_bg_hint")))
	}

	boxStyle := lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(ColorPrimary).Padding(1, 3).Width(58)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, boxStyle.Render(b.String()))
}

func (m *FileManagerModal) renderQuickCmdOutputModal(width, height int) string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("sftp_cmd_output_title")) + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(m.CmdOutputTitle) + "\n\n")

	lines := strings.Split(m.CmdOutputContent, "\n")
	maxVisible := height - 10
	if maxVisible < 5 {
		maxVisible = 5
	}
	if m.CmdOutputScroll < 0 {
		m.CmdOutputScroll = 0
	}
	if m.CmdOutputScroll > len(lines)-1 && len(lines) > 0 {
		m.CmdOutputScroll = len(lines) - 1
	}

	endIdx := m.CmdOutputScroll + maxVisible
	if endIdx > len(lines) {
		endIdx = len(lines)
	}
	for i := m.CmdOutputScroll; i < endIdx; i++ {
		b.WriteString(lines[i] + "\n")
	}

	b.WriteString("\n" + lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("sftp_cmd_return_hint")))
	boxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorSecondary).Padding(1, 3).Width(width - 10).Height(height - 4)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, boxStyle.Render(b.String()))
}

func renderFileItemLine(it *ssh.FileItem, isSelected, isChecked bool, inClipCut, inClipCopy bool, targetWidth int) string {
	checkMark := "  "
	if inClipCut {
		checkMark = lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render("✂️ ")
	} else if inClipCopy {
		checkMark = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("📋 ")
	} else if isChecked {
		if isSelected {
			checkMark = "✔ "
		} else {
			checkMark = lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render("✔ ")
		}
	}

	icon := "📄 "
	nameColor := ColorText
	if it.IsDir {
		icon = "📁 "
		nameColor = ColorSecondary
	}

	sizeStr := it.FormatSize()
	timeStr := it.ModTime.Format("01-02 15:04")

	nameWidth := targetWidth - 24
	if nameWidth < 10 {
		nameWidth = 10
	}
	plainName := runewidth.Truncate(it.Name, nameWidth, "…")

	plainLeft := checkMark + icon + plainName
	lineRight := fmt.Sprintf("%8s  %11s", sizeStr, timeStr)

	gap := targetWidth - lipgloss.Width(plainLeft) - lipgloss.Width(lineRight)
	if gap < 1 {
		gap = 1
	}

	nameStyled := lipgloss.NewStyle().Foreground(nameColor).Render(plainName)
	if it.IsDir {
		nameStyled = lipgloss.NewStyle().Bold(true).Foreground(nameColor).Render(plainName)
	}

	fullLine := checkMark + icon + nameStyled + strings.Repeat(" ", gap) + lineRight

	if isSelected {
		return lipgloss.NewStyle().Bold(true).Background(ColorHighlight).Foreground(lipgloss.Color("#FFFFFF")).Render(fullLine)
	}
	return fullLine
}
