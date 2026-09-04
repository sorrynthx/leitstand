package tui

import (
	"leitstand/internal/i18n"
	"leitstand/internal/logger"
	"path"
	"path/filepath"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *FileManagerModal) Update(msg tea.Msg) (bool, tea.Cmd) {
	if m.IsTransferring {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if m.ShowTransferCancelPrompt {
				switch keyMsg.String() {
				case "enter", "y", "Y":
					logger.Infof("[SFTP] User confirmed transfer cancellation -> calling TransferCancel()")
					m.IsTransferCanceled = true
					if m.TransferCancel != nil {
						m.TransferCancel()
					}
					m.IsTransferring = false
					m.IsTransferBackground = false
					m.ShowTransferCancelPrompt = false
					m.StatusMessage = i18n.T("sftp_transfer_cancelled_user")
					return false, nil
				case "esc", "n", "N", "q":
					m.ShowTransferCancelPrompt = false
					return false, nil
				}
				return false, nil
			}

			switch keyMsg.String() {
			case "esc", "q", "ctrl+c":
				m.ShowTransferCancelPrompt = true
				return false, nil
			case "b", "B":
				logger.Infof("[SFTP] User minimized transfer to background mode")
				m.IsTransferBackground = true
				return true, nil
			}
		}
		return false, nil
	}

	if m.ShowCmdOutput {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc", "enter", "q":
				m.ShowCmdOutput = false
				m.StatusMessage = ""
				return false, nil
			case "up", "k":
				if m.CmdOutputScroll > 0 {
					m.CmdOutputScroll--
				}
				return false, nil
			case "down", "j":
				m.CmdOutputScroll++
				return false, nil
			}
		}
		return false, nil
	}

	if m.ShowRunbook {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			runbooks := getSFTPRunbooks()
			switch keyMsg.String() {
			case "esc", "q":
				m.ShowRunbook = false
				return false, nil
			case "up", "k":
				if m.RunbookCursor > 0 {
					m.RunbookCursor--
				}
				return false, nil
			case "down", "j":
				if m.RunbookCursor < len(runbooks)-1 {
					m.RunbookCursor++
				}
				return false, nil
			case "pgup", "pageup":
				m.RunbookCursor -= 4
				if m.RunbookCursor < 0 {
					m.RunbookCursor = 0
				}
				return false, nil
			case "pgdown", "pagedown":
				m.RunbookCursor += 4
				if m.RunbookCursor >= len(runbooks) {
					m.RunbookCursor = len(runbooks) - 1
				}
				return false, nil
			case "enter":
				selectedCmd := runbooks[m.RunbookCursor].Command
				m.ShowRunbook = false
				m.ActivePrompt = PromptQuickCmd
				m.QuickCmdInput.Reset()
				m.QuickCmdInput.SetValue(selectedCmd)
				m.QuickCmdInput.Focus()
				return false, textinput.Blink
			}
		}
		return false, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok && m.ActivePrompt != PromptNone {
		done, cmd := m.UpdatePrompt(keyMsg)
		return done, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			if m.StatusMessage != "" {
				m.StatusMessage = ""
				return false, nil
			}
			if m.ActivePrompt == PromptExitConfirm {
				m.ActivePrompt = PromptNone
				return false, nil
			}
			m.ActivePrompt = PromptExitConfirm
			return false, nil

		case "tab":
			if m.ActivePanel == PanelLocal {
				m.ActivePanel = PanelRemote
				m.FocusLocal = false
			} else {
				m.ActivePanel = PanelLocal
				m.FocusLocal = true
			}
			m.StatusMessage = ""
			return false, nil

		case "up", "k":
			m.navigateCursor(-1)
			return false, nil

		case "down", "j":
			m.navigateCursor(1)
			return false, nil

		case "pgup", "pageup", "ctrl+u":
			m.navigateCursor(-10)
			return false, nil

		case "pgdown", "pagedown", "ctrl+d":
			m.navigateCursor(10)
			return false, nil

		case "home", "g":
			m.jumpCursor(true)
			return false, nil

		case "end", "G":
			m.jumpCursor(false)
			return false, nil

		case "space", " ":
			m.StatusMessage = ""
			m.ToggleSelection()
			return false, nil

		case "ctrl+a", "a":
			m.StatusMessage = ""
			m.SelectAll()
			return false, nil

		case "enter":
			m.StatusMessage = ""
			items := m.GetActiveItems()
			cursor := m.GetActiveCursor()
			if cursor >= 0 && cursor < len(items) {
				item := items[cursor]
				if item.IsDir {
					if m.ActivePanel == PanelLocal {
						m.LocalPath = item.Path
						m.LocalCursor = 0
						m.LocalSelected = make(map[string]bool)
						return false, m.RefreshLocalCmd()
					} else {
						oldP := m.RemotePath
						m.RemotePath = item.Path
						m.RemoteCursor = 0
						m.RemoteSelected = make(map[string]bool)
						return false, func() tea.Msg {
							return NavigateRemoteMsg{HostID: m.HostID, NewPath: item.Path, OldPath: oldP}
						}
					}
				}
			}
			return false, nil

		case "backspace":
			m.StatusMessage = ""
			if m.ActivePanel == PanelLocal {
				parent := filepath.Dir(m.LocalPath)
				if parent != m.LocalPath {
					m.LocalPath = parent
					m.LocalCursor = 0
					m.LocalSelected = make(map[string]bool)
					return false, m.RefreshLocalCmd()
				}
			} else {
				oldP := m.RemotePath
				parent := path.Dir(m.RemotePath)
				if parent != m.RemotePath {
					m.RemotePath = parent
					m.RemoteCursor = 0
					m.RemoteSelected = make(map[string]bool)
					return false, func() tea.Msg {
						return NavigateRemoteMsg{HostID: m.HostID, NewPath: parent, OldPath: oldP}
					}
				}
			}
			return false, nil

		default:
			return m.handleSFTPKeyActions(msg.String())
		}
	}

	return false, nil
}
