package tui

import (
	"leitstand/internal/i18n"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *FileManagerModal) handleSFTPKeyActions(keyStr string) (bool, tea.Cmd) {
	switch keyStr {
	case "x", "X", "ctrl+x":
		paths := m.GetSelectedPaths()
		if len(paths) > 0 {
			m.ClipboardPaths = paths
			m.ClipboardIsCut = true
			m.ClipboardIsLocal = (m.ActivePanel == PanelLocal)
			m.ClearSelections()
			m.StatusMessage = i18n.Tf("sftp_cut_notify", len(paths))
		}
		return false, nil

	case "c", "C", "ctrl+c":
		paths := m.GetSelectedPaths()
		if len(paths) > 0 {
			m.ClipboardPaths = paths
			m.ClipboardIsCut = false
			m.ClipboardIsLocal = (m.ActivePanel == PanelLocal)
			m.ClearSelections()
			m.StatusMessage = i18n.Tf("sftp_copy_notify", len(paths))
		}
		return false, nil

	case "p", "P", "ctrl+p", "v", "V", "ctrl+v":
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
			actionName := i18n.T("sftp_action_copy")
			if isMove {
				actionName = i18n.T("sftp_action_move")
			}
			m.StatusMessage = i18n.Tf("sftp_paste_notify", len(paths), actionName)
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
		actionName := i18n.T("sftp_action_upload")
		if !isUpload {
			actionName = i18n.T("sftp_action_download")
		}
		m.StatusMessage = i18n.Tf("sftp_transfer_start_notify", len(paths), actionName)
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
		m.StatusMessage = i18n.Tf("sftp_move_start_notify", len(paths))
		return false, func() tea.Msg {
			return TransferActionMsg{HostID: m.HostID, IsUpload: isUpload, IsMove: true, SrcPaths: paths, DestDirPath: destDir}
		}

	case "f7", "F7", "n":
		m.ActivePrompt = PromptMkdir
		m.SubInput.Reset()
		m.SubInput.Placeholder = i18n.T("sftp_mkdir_ph")
		m.SubInput.Focus()
		return false, textinput.Blink

	case "t", "T", "N":
		m.ActivePrompt = PromptTouch
		m.SubInput.Reset()
		m.SubInput.Placeholder = i18n.T("sftp_touch_ph")
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
