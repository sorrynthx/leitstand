package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *FileManagerModal) handleSFTPKeyActions(keyStr string) (bool, tea.Cmd) {
	switch keyStr {
	case "x", "X":
		paths := m.GetSelectedPaths()
		if len(paths) > 0 {
			m.ClipboardPaths = paths
			m.ClipboardIsCut = true
			m.ClipboardIsLocal = (m.ActivePanel == PanelLocal)
			m.ClearSelections()
			m.StatusMessage = fmt.Sprintf("✂️ %d개 항목 잘라내기됨 (이동할 폴더로 이동 후 [v]를 누르세요)", len(paths))
		}
		return false, nil

	case "c", "C":
		paths := m.GetSelectedPaths()
		if len(paths) > 0 {
			m.ClipboardPaths = paths
			m.ClipboardIsCut = false
			m.ClipboardIsLocal = (m.ActivePanel == PanelLocal)
			m.ClearSelections()
			m.StatusMessage = fmt.Sprintf("📋 %d개 항목 복사됨 (붙여넣을 폴더로 이동 후 [v]를 누르세요)", len(paths))
		}
		return false, nil

	case "v", "V", "ctrl+v":
		if len(m.ClipboardPaths) > 0 {
			destIsLocal := (m.ActivePanel == PanelLocal)
			destDir := m.RemotePath
			if destIsLocal {
				destDir = m.LocalPath
			}
			isSameHost := (m.ClipboardIsLocal == destIsLocal)
			isUpload := (m.ClipboardIsLocal && !destIsLocal)
			isMove := m.ClipboardIsCut
			paths := m.ClipboardPaths
			m.ClipboardPaths = nil
			m.ClipboardIsCut = false
			m.ClearSelections()
			actionName := "복사"
			if isMove {
				actionName = "이동"
			}
			m.StatusMessage = fmt.Sprintf("🚀 %d개 항목 붙여넣기(%s) 시작...", len(paths), actionName)
			return false, func() tea.Msg {
				return TransferActionMsg{
					HostID:      m.HostID,
					IsUpload:    isUpload,
					IsMove:      isMove,
					IsSameHost:  isSameHost,
					IsLocalOp:   destIsLocal,
					SrcPaths:    paths,
					DestDirPath: destDir,
				}
			}
		}
		return false, nil

	case "?", "b", "B":
		m.ShowRunbook = true
		m.RunbookCursor = 0
		return false, nil

	case "f5", "F5":
		paths := m.GetSelectedPaths()
		if len(paths) == 0 {
			return false, nil
		}
		isUpload := (m.ActivePanel == PanelLocal)
		destDir := m.RemotePath
		if !isUpload {
			destDir = m.LocalPath
		}
		m.ClearSelections()
		actionName := "업로드"
		if !isUpload {
			actionName = "다운로드"
		}
		m.StatusMessage = fmt.Sprintf("🚀 %d개 항목 %s 전송 시작...", len(paths), actionName)
		return false, func() tea.Msg {
			return TransferActionMsg{HostID: m.HostID, IsUpload: isUpload, IsMove: false, SrcPaths: paths, DestDirPath: destDir}
		}

	case "f6", "F6", "m", "M":
		paths := m.GetSelectedPaths()
		if len(paths) == 0 {
			return false, nil
		}
		isUpload := (m.ActivePanel == PanelLocal)
		destDir := m.RemotePath
		if !isUpload {
			destDir = m.LocalPath
		}
		m.ClearSelections()
		m.StatusMessage = fmt.Sprintf("🚀 %d개 항목 이동 시작...", len(paths))
		return false, func() tea.Msg {
			return TransferActionMsg{HostID: m.HostID, IsUpload: isUpload, IsMove: true, SrcPaths: paths, DestDirPath: destDir}
		}

	case "f7", "F7", "n", "N":
		m.ActivePrompt = PromptMkdir
		m.SubInput.Reset()
		m.SubInput.Placeholder = "새 폴더 이름..."
		m.SubInput.Focus()
		return false, textinput.Blink

	case "t", "T":
		m.ActivePrompt = PromptTouch
		m.SubInput.Reset()
		m.SubInput.Placeholder = "새 파일 이름..."
		m.SubInput.Focus()
		return false, textinput.Blink

	case "f2", "F2", "r", "R":
		items := m.GetActiveItems()
		cursor := m.GetActiveCursor()
		if cursor >= 0 && cursor < len(items) && items[cursor].Name != ".." {
			m.ActivePrompt = PromptRename
			m.SubInput.Reset()
			m.SubInput.SetValue(items[cursor].Name)
			m.SubInput.Focus()
			return false, textinput.Blink
		}

	case "f8", "F8", "d", "D", "delete":
		paths := m.GetSelectedPaths()
		if len(paths) > 0 {
			m.ActivePrompt = PromptDeleteConfirm
			return false, nil
		}

	case "!", ":":
		m.ActivePrompt = PromptQuickCmd
		m.QuickCmdInput.Reset()
		m.QuickCmdInput.Focus()
		return false, textinput.Blink

	case ".":
		m.ShowHidden = !m.ShowHidden
		return false, tea.Batch(
			m.RefreshLocalCmd(),
			func() tea.Msg { return FileManagerRefreshMsg{HostID: m.HostID, RefreshRemote: true} },
		)

	case "s", "S":
		if m.ActivePanel == PanelLocal {
			m.LocalSort = (m.LocalSort + 1) % 3
		} else {
			m.RemoteSort = (m.RemoteSort + 1) % 3
		}
		return false, nil
	}

	return false, nil
}
