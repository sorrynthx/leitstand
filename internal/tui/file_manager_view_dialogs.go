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
			targetDesc := "선택된 항목"
			if len(paths) > 0 {
				baseName := path.Base(paths[0])
				if m.ActivePanel == PanelLocal {
					baseName = filepath.Base(paths[0])
				}
				targetDesc = fmt.Sprintf("'%s'", baseName)
				if len(paths) > 1 {
					targetDesc = fmt.Sprintf("'%s' 외 %d개 (총 %d개)", baseName, len(paths)-1, len(paths))
				}
			}
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(fmt.Sprintf("⚠️ [삭제 확인] %s 항목을 정말로 삭제하시겠습니까?  [Enter] 삭제 확정  [Esc] 취소", targetDesc)))
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
		rawHints := i18n.T("sftp_hints")
		truncHints := runewidth.Truncate(rawHints, width-2, "…")
		hints := lipgloss.NewStyle().Foreground(ColorMuted).Render(truncHints)
		b.WriteString(hints + "\n")
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
	b.WriteString(fmt.Sprintf("[%s] %.1f%%\n", RenderProgressBar(45, pct, ColorSuccess, ColorBorder), pct))
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("Speed: %.1f KB/s", m.BytesPerSec/1024.0)) + "\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorWarning).Render(i18n.T("sftp_transfer_wait")) + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("sftp_transfer_cancel_hint")))

	boxStyle := lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(ColorPrimary).Padding(1, 3).Width(55)
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

type SFTPRunbookItem struct {
	Title   string
	Command string
	Guide   string
}

var defaultSFTPRunbooks = []SFTPRunbookItem{
	{"📦 현재 폴더 전체 tar.gz 압축", "tar -czvf archive.tar.gz .", "현재 SFTP 디렉토리 전체를 archive.tar.gz로 압축 ( . = 현재 위치)"},
	{"📂 tar.gz 압축 파일 해제", "tar -xzvf archive.tar.gz", "현재 위치의 archive.tar.gz 압축 해제 (-x: 해제, -z: gzip)"},
	{"📦 현재 폴더 전체 zip 압축", "zip -r archive.zip .", "현재 위치 전체를 archive.zip으로 압축 (-r: 하위폴더 포함)"},
	{"📂 zip 압축 파일 해제", "unzip archive.zip", "현재 위치의 archive.zip 압축 해제"},
	{"🔒 전체 755 권한 복원", "chmod -R 755 .", "현재 위치 하위의 모든 디렉토리/파일 권한을 755로 복원"},
	{"📁 디렉토리만 755 권한 변경", "find . -type d -exec chmod 755 {} +", "파일 권한은 유지하고 디렉토리만 안전하게 755로 변경"},
	{"📄 파일만 644 권한 안전 변경", "find . -type f -exec chmod 644 {} +", "디렉토리 실행권한(x)은 유지하고 일반 파일만 644로 안전 변경"},
	{"👤 웹서버 소유자(www-data) 변경", "chown -R www-data:www-data .", "Nginx/Apache 등 웹서버 실행 계정으로 현재 위치 소유자 변경"},
	{"📊 폴더별 디스크 점유율 상위", "du -h --max-depth=1 | sort -hr", "현재 위치의 1단계 하위 폴더별 용량을 큰 순서로 정렬 출력"},
	{"🔍 100MB 이상 대용량 파일 검색", "find . -type f -size +100M", "현재 위치 및 하위에서 용량이 100MB를 초과하는 파일 검색"},
}

func (m *FileManagerModal) renderRunbookModal(width, height int) string {
	var b strings.Builder
	boxWidth := width - 6
	if boxWidth < 55 {
		boxWidth = 55
	}

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("📖 SFTP 매니저 단축키 가이드 & 데브옵스 런북") + "\n\n")

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("⌨️  [단축키 요약 Cheat Sheet]") + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorText).Render(" • [x/c/v] 잘라내기/복사/붙여넣기   • [Space] 선택   • [s] 정렬") + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorText).Render(" • [Tab] 패널전환   • [F5/F6] 전송/이동   • [F7/t] 새폴더/새파일") + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorText).Render(" • [F2/F8] 이름변경/삭제   • [.] 숨김파일   • [Esc] 닫기") + "\n\n")

	headerTitle := fmt.Sprintf("🚀 [데브옵스 추천 명령어 사적 (%d/%d)] (↑/↓/PgUp/PgDn 선택 후 Enter)", m.RunbookCursor+1, len(defaultSFTPRunbooks))
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
	if endIdx > len(defaultSFTPRunbooks) {
		endIdx = len(defaultSFTPRunbooks)
		if endIdx-visibleCount >= 0 {
			startIdx = endIdx - visibleCount
		}
	}

	for i := startIdx; i < endIdx; i++ {
		rb := defaultSFTPRunbooks[i]
		if i == m.RunbookCursor {
			b.WriteString(lipgloss.NewStyle().Bold(true).Background(ColorHighlight).Foreground(lipgloss.Color("#FFFFFF")).Render(fmt.Sprintf(" ▶ %s ", rb.Title)) + "\n")
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(fmt.Sprintf("    ➔ %s", rb.Command)) + "\n")
			b.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Render(fmt.Sprintf("    💡 %s", rb.Guide)) + "\n\n")
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(ColorText).Render(fmt.Sprintf("   %s", rb.Title)) + "\n")
			b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("    ➔ %s", rb.Command)) + "\n\n")
		}
	}

	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("[Enter] 명령어 주입   [PgUp/PgDn] 페이지 이동   [Esc] 닫기"))

	boxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorSecondary).Padding(1, 2).Width(boxWidth).Height(height - 2)
	content := boxStyle.Render(b.String())
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
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

func padStringToWidth(s string, targetWidth int) string {
	w := lipgloss.Width(s)
	if w >= targetWidth {
		return runewidth.Truncate(s, targetWidth, "")
	}
	return s + strings.Repeat(" ", targetWidth-w)
}
