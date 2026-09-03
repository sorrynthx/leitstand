package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func (m *FileManagerModal) View(width, height int) string {
	m.Width = width
	m.Height = height

	if m.Width <= 0 || m.Height <= 0 {
		m.Width = 110
		m.Height = 32
	}

	if m.ShowRunbook {
		return m.renderRunbookModal(m.Width, m.Height)
	}

	if m.ShowCmdOutput {
		return m.renderQuickCmdOutputModal(m.Width, m.Height)
	}

	if m.IsTransferring {
		return m.renderTransferProgressModal(m.Width, m.Height)
	}

	halfWidth := (m.Width - 10) / 2
	if halfWidth < 25 {
		halfWidth = 25
	}

	panelInnerHeight := m.Height - 8
	if panelInnerHeight < 10 {
		panelInnerHeight = 10
	}

	localView := m.renderLocalPanel(halfWidth, panelInnerHeight)
	remoteView := m.renderRemotePanel(halfWidth, panelInnerHeight)

	mainDeck := lipgloss.JoinHorizontal(lipgloss.Top, localView, "  ", remoteView)

	var b strings.Builder
	titleHeader := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(i18n.T("sftp_title"))
	b.WriteString(titleHeader + "\n\n")
	b.WriteString(mainDeck + "\n")
	b.WriteString(m.renderTransferDeck(m.Width - 4))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(0, 1)

	content := boxStyle.Render(b.String())
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, content)
}

func (m *FileManagerModal) renderLocalPanel(width, height int) string {
	var b strings.Builder

	isFocused := (m.ActivePanel == PanelLocal)
	panelInnerWidth := width - 4

	sortTag := formatSortIndicator(m.LocalSort, m.LocalSortAsc)
	headerTitle := i18n.Tf("sftp_local_header", m.LocalPath) + " " + sortTag
	if isFocused {
		headerTitle = lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render("★ " + headerTitle)
	} else {
		headerTitle = lipgloss.NewStyle().Foreground(ColorMuted).Render("  " + headerTitle)
	}
	b.WriteString(padStringToWidth(headerTitle, panelInnerWidth) + "\n\n")

	items := m.GetSortedLocalItems()
	maxVisible := height - 4
	if maxVisible < 5 {
		maxVisible = 5
	}

	startIdx := 0
	if m.LocalCursor >= maxVisible {
		startIdx = m.LocalCursor - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(items) {
		endIdx = len(items)
	}

	if len(items) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("  "+i18n.T("sftp_empty_dir")+"\n"))
	} else {
		for i := startIdx; i < endIdx; i++ {
			it := items[i]
			isSelected := (i == m.LocalCursor && isFocused)
			isChecked := m.LocalSelected[it.Path]

			inClipCut := false
			inClipCopy := false
			for _, p := range m.ClipboardPaths {
				if p == it.Path {
					if m.ClipboardIsCut {
						inClipCut = true
					} else {
						inClipCopy = true
					}
					break
				}
			}

			itemLine := renderFileItemLine(it, isSelected, isChecked, inClipCut, inClipCopy, panelInnerWidth)
			b.WriteString(itemLine + "\n")
		}
	}

	paneBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Width(width).
		Height(height)

	if isFocused {
		paneBox = paneBox.BorderForeground(ColorSuccess)
	}

	return paneBox.Render(b.String())
}

func (m *FileManagerModal) renderRemotePanel(width, height int) string {
	var b strings.Builder

	isFocused := (m.ActivePanel == PanelRemote)
	panelInnerWidth := width - 4

	sortTag := formatSortIndicator(m.RemoteSort, m.RemoteSortAsc)
	headerTitle := fmt.Sprintf("🌐 %s: %s %s", m.HostName, m.RemotePath, sortTag)
	if isFocused {
		headerTitle = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("★ " + headerTitle)
	} else {
		headerTitle = lipgloss.NewStyle().Foreground(ColorMuted).Render("  " + headerTitle)
	}
	b.WriteString(padStringToWidth(headerTitle, panelInnerWidth) + "\n\n")

	items := m.GetSortedRemoteItems()
	maxVisible := height - 4
	if maxVisible < 5 {
		maxVisible = 5
	}

	startIdx := 0
	if m.RemoteCursor >= maxVisible {
		startIdx = m.RemoteCursor - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(items) {
		endIdx = len(items)
	}

	if len(items) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("  "+i18n.T("sftp_empty_dir")+"\n"))
	} else {
		for i := startIdx; i < endIdx; i++ {
			it := items[i]
			isSelected := (i == m.RemoteCursor && isFocused)
			isChecked := m.RemoteSelected[it.Path]

			inClipCut := false
			inClipCopy := false
			for _, p := range m.ClipboardPaths {
				if p == it.Path {
					if m.ClipboardIsCut {
						inClipCut = true
					} else {
						inClipCopy = true
					}
					break
				}
			}

			itemLine := renderFileItemLine(it, isSelected, isChecked, inClipCut, inClipCopy, panelInnerWidth)
			b.WriteString(itemLine + "\n")
		}
	}

	paneBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Width(width).
		Height(height)

	if isFocused {
		paneBox = paneBox.BorderForeground(ColorPrimary)
	}

	return paneBox.Render(b.String())
}

func formatSortIndicator(field SortField, asc bool) string {
	dir := "🔼"
	if !asc {
		dir = "🔽"
	}
	switch field {
	case SortByName:
		return "[이름순 " + dir + "]"
	case SortBySize:
		return "[크기순 " + dir + "]"
	case SortByModTime:
		return "[날짜순 " + dir + "]"
	default:
		return "[정렬 " + dir + "]"
	}
}

func formatTransferBytes(b int64) string {
	val := float64(b)
	switch {
	case val >= 1024*1024*1024:
		return fmt.Sprintf("%.2f GB", val/(1024*1024*1024))
	case val >= 1024*1024:
		return fmt.Sprintf("%.1f MB", val/(1024*1024))
	case val >= 1024:
		return fmt.Sprintf("%.1f KB", val/1024)
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func padStringToWidth(s string, targetWidth int) string {
	w := lipgloss.Width(s)
	if w >= targetWidth {
		return runewidth.Truncate(s, targetWidth, "")
	}
	return s + strings.Repeat(" ", targetWidth-w)
}
