package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"leitstand/internal/storage"
	"leitstand/internal/vault"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles incoming messages and updates model state.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.initOrResizeViewport()
		return m, nil

	case tea.KeyMsg:
		keyStr := msg.String()

		// 1. Delegate to active modal if open
		if updatedModel, cmd, handled := m.updateActiveModals(msg); handled {
			return updatedModel, cmd
		}

		// 2. Delegate to Runbook Drawer if open
		if updatedModel, cmd, handled := m.updateRunbookDrawer(msg); handled {
			return updatedModel, cmd
		}

		// 3. Console Pane Active Key Handling
		if m.activePane == PaneConsole {
			return m.updateConsoleKeys(msg, keyStr)
		}

		// 4. Host Explorer & Global Navigation
		if updatedModel, cmd, handled := m.updateHostListNavigation(keyStr); handled {
			return updatedModel, cmd
		}

	case []*storage.Host:
		m.hosts = msg
		if len(m.hosts) > 0 {
			if m.selectedIndex >= len(m.hosts) {
				m.selectedIndex = 0
			}
			m.updateViewportContent()
			return m, m.pollActiveHostsCmd()
		}
		m.updateViewportContent()
		return m, nil

	case TerminalExitedMsg:
		m.statusMessage = "💻 Terminal session ended."
		return m, nil

	case MetricResultMsg:
		hostName := fmt.Sprintf("Host #%d", msg.HostID)
		for _, h := range m.hosts {
			if h.ID == msg.HostID {
				hostName = h.Name
				break
			}
		}

		if msg.Err != nil {
			m.errors[msg.HostID] = msg.Err
			m.hostStatus[msg.HostID] = HostStatusOffline
			if len(m.hosts) > 0 && m.hosts[m.selectedIndex].ID == msg.HostID {
				m.statusMessage = fmt.Sprintf("🔴 Host '%s' offline. Turn on VPN and press [r] to connect.", hostName)
			}
		} else {
			delete(m.errors, msg.HostID)
			m.hostStatus[msg.HostID] = HostStatusOnline
			m.pausedHosts[msg.HostID] = false
			m.metrics[msg.HostID] = msg.Record
			if msg.SysInfo != nil {
				m.sysInfos[msg.HostID] = msg.SysInfo
			}
			if len(m.hosts) > 0 && m.hosts[m.selectedIndex].ID == msg.HostID {
				if strings.HasPrefix(m.statusMessage, "⏳ Connecting") {
					m.statusMessage = fmt.Sprintf("🟢 Connected to '%s' successfully!", hostName)
				}
			}
		}
		return m, nil

	case OpenFileMsg:
		if msg.Err != nil {
			m.statusMessage = fmt.Sprintf("⚠️ Failed to open file '%s': %v", msg.FilePath, msg.Err)
			return m, nil
		}
		m.showEditorModal = true
		m.editorModal = NewEditorModal(msg.HostID, msg.HostName, msg.FilePath, msg.Content, m.width, m.height)
		m.statusMessage = fmt.Sprintf("✏️ Editing '%s' on %s. [Ctrl+S] Save, [Esc] Cancel.", msg.FilePath, msg.HostName)
		return m, nil

	case FileSavedMsg:
		if msg.Err != nil {
			if m.editorModal != nil {
				m.editorModal.SetError(msg.Err)
			}
			m.statusMessage = fmt.Sprintf("⚠️ Failed to save '%s': %v", msg.FilePath, msg.Err)
		} else {
			m.showEditorModal = false
			m.editorModal = nil
			m.statusMessage = fmt.Sprintf("✨ File '%s' saved to remote server successfully!", msg.FilePath)
		}
		return m, nil

	case StreamChunkMsg:
		if len(m.hosts) > 0 {
			for _, h := range m.hosts {
				if h.ID == msg.HostID {
					hts := m.GetOrCreateHostTabs(h.ID, h.Name)
					for _, tab := range hts.Tabs {
						if tab.ID == msg.TabID {
							if tab.IsScreenApp {
								tab.SetFrame(msg.Chunk)
							} else {
								tab.AppendLog(msg.Chunk)
							}
							break
						}
					}
					break
				}
			}
		}
		if msg.msgChan != nil {
			return m, listenStreamCmd(msg.msgChan)
		}
		return m, nil

	case StreamFinishedMsg:
		if len(m.hosts) > 0 {
			for _, h := range m.hosts {
				if h.ID == msg.HostID {
					hts := m.GetOrCreateHostTabs(h.ID, h.Name)
					for _, tab := range hts.Tabs {
						if tab.ID == msg.TabID {
							tab.IsStreaming = false
							tab.StreamCancel = nil
							if msg.Err != nil {
								tab.AppendLog(fmt.Sprintf("⚠️ Stream ended with error: %v", msg.Err))
							}
							break
						}
					}
					break
				}
			}
		}
		return m, nil

	case TransferActionMsg:
		if m.fileManager != nil {
			m.fileManager.IsTransferring = true
			m.fileManager.TransferIsUpload = msg.IsUpload
			m.fileManager.TransferDoneMsg = ""
		}
		return m, m.startFileTransferCmd(msg)

	case FileTransferProgressMsg:
		if m.fileManager != nil {
			m.fileManager.IsTransferring = !msg.IsDone
			m.fileManager.CurrentFileName = msg.FileName
			m.fileManager.FileIndex = msg.FileIndex
			m.fileManager.FileTotal = msg.FileTotal
			m.fileManager.CurrentBytes = msg.CurrentBytes
			m.fileManager.CurrentTotal = msg.TotalBytes
			m.fileManager.BytesPerSec = msg.BytesPerSec

			if msg.Err != nil {
				m.fileManager.StatusMessage = fmt.Sprintf("❌ %v", msg.Err)
				m.fileManager.TransferDoneMsg = ""
				m.fileManager.IsTransferring = false
				return m, nil
			} else if msg.IsDone {
				m.fileManager.IsTransferring = false
				if msg.IsMove {
					m.fileManager.TransferDoneMsg = fmt.Sprintf(i18n.T("sftp_move_done"), msg.FileTotal)
				} else {
					m.fileManager.TransferDoneMsg = fmt.Sprintf(i18n.T("sftp_transfer_done"), msg.FileTotal)
				}
				m.fileManager.TransferDoneTime = time.Now()
				m.fileManager.LocalSelected = make(map[string]bool)
				m.fileManager.RemoteSelected = make(map[string]bool)
				var curHost *storage.Host
				if len(m.hosts) > 0 {
					curHost = m.hosts[m.selectedIndex]
				}
				return m, tea.Batch(
					m.fileManager.RefreshLocalCmd(),
					m.fetchRemoteFilesCmd(curHost, m.fileManager.RemotePath, m.fileManager.RemotePath, m.fileManager.ShowHidden),
				)
			}
		}
		if msg.msgChan != nil {
			return m, listenStreamCmd(msg.msgChan)
		}
		return m, nil

	case FileOpActionMsg:
		return m, m.executeFileOpCmd(msg)

	case FileOpResultMsg:
		if m.fileManager != nil {
			if msg.Err != nil {
				m.fileManager.StatusMessage = fmt.Sprintf("⚠️ %v", msg.Err)
			} else {
				m.fileManager.StatusMessage = msg.Msg
			}
			m.fileManager.LocalSelected = make(map[string]bool)
			m.fileManager.RemoteSelected = make(map[string]bool)
			var curHost *storage.Host
			if len(m.hosts) > 0 {
				curHost = m.hosts[m.selectedIndex]
			}
			if msg.IsLocal {
				return m, m.fileManager.RefreshLocalCmd()
			}
			return m, m.fetchRemoteFilesCmd(curHost, m.fileManager.RemotePath, m.fileManager.RemotePath, m.fileManager.ShowHidden)
		}
		return m, nil

	case LocalFileListMsg:
		if m.fileManager != nil && m.fileManager.HostID == msg.HostID {
			if msg.Err != nil {
				m.fileManager.LocalPath = msg.OldPath
				m.fileManager.StatusMessage = fmt.Sprintf(i18n.T("sftp_perm_denied"), filepath.Base(msg.Path))
			} else {
				m.fileManager.LocalPath = msg.Path
				m.fileManager.LocalItems = msg.Items
				if m.fileManager.LocalCursor >= len(msg.Items) {
					if len(msg.Items) > 0 {
						m.fileManager.LocalCursor = len(msg.Items) - 1
					} else {
						m.fileManager.LocalCursor = 0
					}
				}
				if m.fileManager.LocalCursor < 0 {
					m.fileManager.LocalCursor = 0
				}
			}
		}
		return m, nil

	case RemoteFileListMsg:
		if m.fileManager != nil && m.fileManager.HostID == msg.HostID {
			if msg.Err != nil {
				m.fileManager.RemotePath = msg.OldPath
				m.fileManager.StatusMessage = fmt.Sprintf(i18n.T("sftp_perm_denied"), filepath.Base(msg.Path))
			} else {
				m.fileManager.RemoteItems = msg.Items
				if msg.Path != "" {
					m.fileManager.RemotePath = msg.Path
				}
				if m.fileManager.RemoteCursor >= len(msg.Items) {
					if len(msg.Items) > 0 {
						m.fileManager.RemoteCursor = len(msg.Items) - 1
					} else {
						m.fileManager.RemoteCursor = 0
					}
				}
				if m.fileManager.RemoteCursor < 0 {
					m.fileManager.RemoteCursor = 0
				}
			}
		}
		return m, nil

	case NavigateRemoteMsg:
		if m.fileManager != nil && m.fileManager.HostID == msg.HostID {
			var curHost *storage.Host
			for _, h := range m.hosts {
				if h.ID == msg.HostID {
					curHost = h
					break
				}
			}
			if curHost != nil {
				return m, m.fetchRemoteFilesCmd(curHost, msg.NewPath, msg.OldPath, m.fileManager.ShowHidden)
			}
		}
		return m, nil

	case FileManagerQuickCmdMsg:
		return m, m.executeFileManagerQuickCmd(msg)

	case FileManagerQuickCmdResultMsg:
		if m.fileManager != nil && m.fileManager.HostID == msg.HostID {
			if msg.Err != nil {
				m.fileManager.StatusMessage = fmt.Sprintf("⚠️ %v", msg.Err)
			} else {
				m.fileManager.StatusMessage = fmt.Sprintf("✨ Executed: %s", msg.Command)
			}
			if msg.NewCWD != "" && msg.NewCWD != msg.OldCWD {
				if msg.IsLocal {
					m.fileManager.LocalPath = msg.NewCWD
				} else {
					m.fileManager.RemotePath = msg.NewCWD
				}
			}
			m.fileManager.CmdOutputTitle = fmt.Sprintf("%s (in %s)", msg.Command, msg.OldCWD)
			m.fileManager.CmdOutputContent = msg.Output
			m.fileManager.CmdOutputScroll = 0
			m.fileManager.ShowCmdOutput = true

			// Auto refresh file lists
			if msg.IsLocal {
				return m, m.fileManager.RefreshLocalCmd()
			}
			var curHost *storage.Host
			for _, h := range m.hosts {
				if h.ID == msg.HostID {
					curHost = h
					break
				}
			}
			if curHost != nil {
				return m, m.fetchRemoteFilesCmd(curHost, m.fileManager.RemotePath, m.fileManager.RemotePath, m.fileManager.ShowHidden)
			}
		}
		return m, nil

	case FileManagerRefreshMsg:
		if m.fileManager != nil && m.fileManager.HostID == msg.HostID {
			if msg.RefreshRemote && len(m.hosts) > 0 {
				curHost := m.hosts[m.selectedIndex]
				return m, m.fetchRemoteFilesCmd(curHost, m.fileManager.RemotePath, m.fileManager.RemotePath, m.fileManager.ShowHidden)
			}
		}
		return m, nil

	case CmdResultMsg:
		if len(m.hosts) > 0 {
			for _, h := range m.hosts {
				if h.ID == msg.HostID {
					hts := m.GetOrCreateHostTabs(h.ID, h.Name)
					var targetTab *ConsoleTab
					for _, t := range hts.Tabs {
						if t.ID == msg.TabID {
							targetTab = t
							break
						}
					}
					if targetTab == nil {
						targetTab = hts.ActiveTab()
					}

					if targetTab != nil {
						cwdDisplay := msg.CWD
						if cwdDisplay == "" {
							cwdDisplay = "~"
						}

						if msg.NewCWD != "" {
							targetTab.CWD = msg.NewCWD
						}

						if msg.Err != nil {
							targetTab.AppendLog(fmt.Sprintf("[%s] ❯ %s\n❌ Error: %v\n%s", cwdDisplay, msg.Command, msg.Err, msg.Stderr))
							m.statusMessage = fmt.Sprintf("⚠️ Error executing '%s'", msg.Command)
						} else {
							out := msg.Stdout
							if out != "" {
								targetTab.AppendLog(fmt.Sprintf("[%s] ❯ %s\n%s", cwdDisplay, msg.Command, out))
							} else {
								targetTab.AppendLog(fmt.Sprintf("[%s] ❯ %s\n(no output)", cwdDisplay, msg.Command))
							}
							m.statusMessage = fmt.Sprintf("✅ Executed '%s' successfully (%s)", msg.Command, time.Now().Format("15:04:05"))
						}
					}
					break
				}
			}
		}
		m.updateViewportContent()
		return m, nil

	case TabCompletionMsg:
		if msg.NewInput != "" && msg.NewInput != m.consoleInput.Value() {
			m.consoleInput.SetValue(msg.NewInput)
			m.consoleInput.SetCursor(len(msg.NewInput))
		}
		if len(msg.Candidates) > 1 {
			m.statusMessage = fmt.Sprintf("🔍 Matches: %s", strings.Join(msg.Candidates, "  "))
		} else if len(msg.Candidates) == 1 {
			m.statusMessage = "✨ Completed"
		} else {
			m.statusMessage = "(no completion matches)"
		}
		return m, nil

	case TickMsg:
		return m, tea.Batch(
			m.pollActiveHostsCmd(),
			m.tickCmd(),
		)
	}

	return m, nil
}

// updateConsoleKeys handles keyboard input when the console pane is active.
func (m *Model) updateConsoleKeys(msg tea.KeyMsg, keyStr string) (tea.Model, tea.Cmd) {
	if len(m.hosts) == 0 {
		return m, nil
	}

	curHost := m.hosts[m.selectedIndex]
	hts := m.GetOrCreateHostTabs(curHost.ID, curHost.Name)

	// Tab Management
	if keyStr == "ctrl+t" || keyStr == "ctrl+n" {
		hts.AddNewTab(m.viewport.Width, m.viewport.Height)
		m.consoleInput.SetValue("")
		m.updateViewportContent()
		m.statusMessage = fmt.Sprintf(i18n.T("tab_new_msg"), len(hts.Tabs))
		return m, nil
	}

	if keyStr == "ctrl+w" {
		closed := hts.CloseActiveTab()
		if closed {
			m.statusMessage = i18n.T("tab_closed_msg")
			m.updateViewportContent()
		}
		return m, nil
	}

	// Tab Cycling: Alt+Left / Alt+Right or Alt+P / Alt+N or Ctrl+PgUp / Ctrl+PgDn
	if keyStr == "alt+left" || keyStr == "alt+p" || keyStr == "ctrl+pgup" {
		hts.PrevTab()
		m.updateViewportContent()
		m.statusMessage = fmt.Sprintf("📌 Tab #%d (%s)", hts.ActiveIndex+1, hts.ActiveTab().Title)
		return m, nil
	}

	if keyStr == "alt+right" || keyStr == "alt+n" || keyStr == "ctrl+pgdown" {
		hts.NextTab()
		m.updateViewportContent()
		m.statusMessage = fmt.Sprintf("📌 Tab #%d (%s)", hts.ActiveIndex+1, hts.ActiveTab().Title)
		return m, nil
	}

	// Tab Direct Jump: Alt+1 ~ Alt+8, and Alt+9 (Jump to Last Tab)
	if strings.HasPrefix(keyStr, "alt+") && len(keyStr) == 5 {
		digitRune := keyStr[4]
		if digitRune == '9' {
			lastIdx := len(hts.Tabs) - 1
			if hts.SwitchTab(lastIdx) {
				m.updateViewportContent()
				m.statusMessage = fmt.Sprintf("📌 Switched to Last Tab #%d", lastIdx+1)
				return m, nil
			}
		} else if digitRune >= '1' && digitRune <= '8' {
			idx := int(digitRune - '1')
			if hts.SwitchTab(idx) {
				m.updateViewportContent()
				m.statusMessage = fmt.Sprintf("📌 Switched to Tab #%d", idx+1)
				return m, nil
			}
		}
	}

	switch keyStr {
	case "esc":
		m.activePane = PaneHostList
		m.consoleInput.Blur()
		if m.fullScreenConsole {
			m.fullScreenConsole = false
			m.initOrResizeViewport()
		}
		m.statusMessage = "📋 Returned to Host Explorer. Press [c] to focus console."
		return m, nil

	case "alt+t":
		m.statusMessage = fmt.Sprintf("💻 Launching full interactive terminal on %s...", curHost.Name)
		return m, m.launchInteractiveTerminalCmd(curHost)

	case "alt+f", "f6":
		return m, m.openFileManagerCmd(curHost)

	case "ctrl+c":
		tab := m.CurrentActiveTab()
		if tab != nil && tab.IsStreaming {
			if tab.StreamCancel != nil {
				tab.StreamCancel()
				tab.StreamCancel = nil
			}
			tab.IsStreaming = false
			tab.AppendLog("⏹️ " + fmt.Sprintf(i18n.T("tab_streaming_cancel"), tab.StreamCmd))
			m.statusMessage = fmt.Sprintf(i18n.T("tab_streaming_cancel"), tab.StreamCmd)
			m.updateViewportContent()
			return m, nil
		}
		m.cancel()
		return m, tea.Quit

	case "tab":
		return m, m.completeInputCmd()

	case "up":
		tab := m.CurrentActiveTab()
		if tab != nil && len(tab.CmdHistory) > 0 {
			if tab.HistoryIndex < 0 {
				tab.HistoryIndex = len(tab.CmdHistory) - 1
			} else if tab.HistoryIndex > 0 {
				tab.HistoryIndex--
			}
			m.consoleInput.SetValue(tab.CmdHistory[tab.HistoryIndex])
			m.consoleInput.SetCursor(len(m.consoleInput.Value()))
			return m, nil
		}
		return m, nil

	case "down":
		tab := m.CurrentActiveTab()
		if tab != nil && len(tab.CmdHistory) > 0 && tab.HistoryIndex >= 0 {
			tab.HistoryIndex++
			if tab.HistoryIndex >= len(tab.CmdHistory) {
				tab.HistoryIndex = -1
				m.consoleInput.SetValue("")
			} else {
				m.consoleInput.SetValue(tab.CmdHistory[tab.HistoryIndex])
				m.consoleInput.SetCursor(len(m.consoleInput.Value()))
			}
			return m, nil
		}
		return m, nil

	case "pgup", "ctrl+u":
		tab := m.CurrentActiveTab()
		if tab != nil {
			tab.Viewport.LineUp(6)
		}
		return m, nil

	case "pgdown", "ctrl+d":
		tab := m.CurrentActiveTab()
		if tab != nil {
			tab.Viewport.LineDown(6)
		}
		return m, nil

	case "ctrl+l":
		tab := m.CurrentActiveTab()
		if tab != nil {
			tab.Logs = make([]string, 0)
			tab.UpdateViewportContent()
			m.statusMessage = i18n.T("console_cleared")
		}
		return m, nil

	case "ctrl+o":
		m.fullScreenConsole = !m.fullScreenConsole
		m.initOrResizeViewport()
		return m, nil

	case "ctrl+k", "f5":
		var distro string
		if si := m.sysInfos[curHost.ID]; si != nil {
			distro = si.OSDistro
		}
		m.showDrawer = true
		m.drawer = NewRunbookDrawer(distro)
		return m, nil

	case "?":
		if m.consoleInput.Value() == "" {
			var distro string
			if si := m.sysInfos[curHost.ID]; si != nil {
				distro = si.OSDistro
			}
			m.showDrawer = true
			m.drawer = NewRunbookDrawer(distro)
			return m, nil
		}
		var cmd tea.Cmd
		m.consoleInput, cmd = m.consoleInput.Update(msg)
		return m, cmd

	case "enter":
		cmdText := strings.TrimSpace(m.consoleInput.Value())
		if cmdText != "" {
			m.consoleInput.SetValue("")
			tab := hts.ActiveTab()
			if tab != nil {
				tab.CmdHistory = append(tab.CmdHistory, cmdText)
				tab.HistoryIndex = -1
			}

			// Exit / Logout Handling for Root Session
			if tab != nil && tab.IsRoot && (cmdText == "exit" || cmdText == "logout") {
				tab.IsRoot = false
				tab.SetAutoTitle(hts.ActiveIndex, "")
				tab.AppendLog("👋 Logged out from root session. Returned to standard user ($).")
				m.updateViewportContent()
				m.statusMessage = "Logged out from root session."
				return m, nil
			}

			// Built-in clear / cls command support
			if strings.EqualFold(cmdText, "clear") || strings.EqualFold(cmdText, "cls") {
				if tab != nil {
					tab.Logs = make([]string, 0)
					tab.UpdateViewportContent()
				}
				m.statusMessage = i18n.T("console_cleared")
				return m, nil
			}

			// Built-in sftp / files command support
			if strings.EqualFold(cmdText, "sftp") || strings.EqualFold(cmdText, "files") || strings.EqualFold(cmdText, "filemanager") || strings.EqualFold(cmdText, "mc") {
				return m, m.openFileManagerCmd(curHost)
			}

			// Direct in-app file editor trigger
			if strings.HasPrefix(cmdText, "edit ") || strings.HasPrefix(cmdText, "vi ") || strings.HasPrefix(cmdText, "vim ") || strings.HasPrefix(cmdText, "nano ") {
				targetFile := strings.TrimSpace(cmdText[strings.Index(cmdText, " "):])
				m.statusMessage = fmt.Sprintf("📖 Opening '%s' for editing on %s...", targetFile, curHost.Name)
				return m, m.openRemoteFileCmd(curHost, targetFile)
			}

			// Root privilege mode switch trigger (su / sudo -i)
			cmdLower := strings.ToLower(cmdText)
			if cmdLower == "su" || cmdLower == "su -" || cmdLower == "su root" || cmdLower == "sudo su" || cmdLower == "sudo -i" {
				if tab != nil {
					if m.isDemo {
						tab.IsRoot = true
						tab.SetAutoTitle(hts.ActiveIndex, "")
						tab.AppendLog(fmt.Sprintf("✨ [ROOT Session] Switched to root mode on %s.\nCommands in this tab run as root (root#). Type 'exit' to log out.", curHost.Name))
						m.updateViewportContent()
						m.statusMessage = "👑 Root session activated for this tab."
						return m, nil
					}
					if cachedPass, ok := m.sudoCache[curHost.ID]; ok && cachedPass != "" {
						tab.IsRoot = true
						tab.SetAutoTitle(hts.ActiveIndex, "")
						tab.AppendLog(fmt.Sprintf("✨ [ROOT Session] Root session activated on %s.\nCommands in this tab run with elevated privileges (root#). Type 'exit' to log out.", curHost.Name))
						m.updateViewportContent()
						m.statusMessage = "👑 Root session activated for this tab."
						return m, nil
					}
					m.pendingSudoCmd = "su"
					m.sudoModal = NewSudoModal(curHost.Name, "Elevate Tab to Root (su root)")
					m.showSudoModal = true
					return m, nil
				}
			}

			// Elevated Privilege Sudo Trigger for single command
			if strings.HasPrefix(cmdLower, "sudo ") {
				if cachedPass, ok := m.sudoCache[curHost.ID]; ok && cachedPass != "" {
					m.statusMessage = fmt.Sprintf("⏳ Running with elevated privilege: '%s'...", cmdText)
					return m, m.execSudoCmd(curHost, cmdText, cachedPass)
				}
				m.pendingSudoCmd = cmdText
				m.sudoModal = NewSudoModal(curHost.Name, cmdText)
				m.showSudoModal = true
				return m, nil
			}

			// If tab is in Root mode, run commands with root privileges automatically
			if tab != nil && tab.IsRoot {
				if cachedPass, ok := m.sudoCache[curHost.ID]; ok && cachedPass != "" {
					m.statusMessage = fmt.Sprintf("⏳ Running as root: '%s' on %s...", cmdText, curHost.Name)
					return m, m.execSudoCmd(curHost, cmdText, cachedPass)
				}
				if m.isDemo {
					m.statusMessage = fmt.Sprintf("⏳ Running as root: '%s' on %s...", cmdText, curHost.Name)
					return m, m.execSudoCmd(curHost, cmdText, "")
				}
			}

			m.statusMessage = fmt.Sprintf("⏳ Running '%s' on %s...", cmdText, curHost.Name)
			return m, m.execRemoteCmd(curHost, cmdText)
		}
		return m, nil

	default:
		var cmd tea.Cmd
		m.consoleInput, cmd = m.consoleInput.Update(msg)
		return m, cmd
	}
}

func (m *Model) changeVaultPassword(currPassword, newPassword string) error {
	if m.store == nil || m.vault == nil {
		return fmt.Errorf("vault or storage not available")
	}

	tempVerifyVault := vault.New()
	err := m.store.UnlockVault(tempVerifyVault, currPassword)
	if err != nil {
		return fmt.Errorf("current password verification failed: %w", err)
	}
	tempVerifyVault.Lock()

	newVault := vault.New()
	err = m.store.RekeyVault(m.vault, newVault, newPassword)
	if err != nil {
		return fmt.Errorf("failed to change master password: %w", err)
	}

	m.vault.Lock()
	m.vault = newVault
	return nil
}
