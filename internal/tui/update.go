package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"leitstand/internal/storage"
	"leitstand/internal/vault"
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
		// 1. Settings / Preferences Modal
		if m.showSettingsModal && m.settingsModal != nil {
			done, saveReq, lang, interval, currPass, newPass, cmd := m.settingsModal.Update(msg)
			if done && !saveReq {
				m.showSettingsModal = false
				m.settingsModal = nil
				m.statusMessage = "Settings closed."
				return m, nil
			}
			if saveReq {
				i18n.SetLang(lang)
				if m.cfg != nil {
					m.cfg.Telemetry.PollingInterval = interval
				}
				if m.store != nil {
					_ = m.store.SetSetting("language", string(lang))
					_ = m.store.SetSetting("polling_interval", interval.String())
				}

				if newPass != "" {
					err := m.changeVaultPassword(currPass, newPass)
					if err != nil {
						m.settingsModal.SetError(err)
						return m, nil
					}
				}

				m.showSettingsModal = false
				m.settingsModal = nil
				m.statusMessage = i18n.T("settings_saved")
				return m, m.tickCmd()
			}
			return m, cmd
		}

		// 2. Quick Command Runbook Drawer
		if m.showDrawer && m.drawer != nil {
			done, chosenCmd, cmd := m.drawer.Update(msg)
			if done {
				m.showDrawer = false
				m.drawer = nil
				if chosenCmd != "" {
					m.consoleInput.SetValue(chosenCmd)
					m.consoleInput.SetCursor(len(chosenCmd))
					m.activePane = PaneConsole
					m.consoleInput.Focus()
					m.statusMessage = fmt.Sprintf("⌨️ Inserted: %s (Press Enter to run, or edit)", chosenCmd)
				}
				return m, nil
			}
			return m, cmd
		}

		// 2. In-app File Editor Modal
		if m.showEditorModal && m.editorModal != nil {
			done, saveReq, updatedContent, cmd := m.editorModal.Update(msg)
			if done {
				m.showEditorModal = false
				m.editorModal = nil
				m.statusMessage = "Editor closed."
				return m, nil
			}
			if saveReq {
				m.statusMessage = "⏳ Saving file to remote server..."
				return m, m.saveRemoteFileCmd(m.editorModal.HostID, m.editorModal.FilePath, updatedContent)
			}
			return m, cmd
		}

		// 3. Vault Unlock/Init Modal
		if m.showVaultModal && m.vaultForm != nil {
			done, pass, cmd := m.vaultForm.Update(msg)
			if done {
				if pass == "" {
					m.cancel()
					return m, tea.Quit
				}

				isInit, _ := m.store.IsVaultInitialized()
				var err error
				if !isInit {
					err = m.store.InitVault(m.vault, pass)
					if err == nil && m.store != nil {
						_ = m.store.SetSetting("language", string(i18n.GetLang()))
					}
				} else {
					err = m.store.UnlockVault(m.vault, pass)
				}

				if err != nil {
					m.vaultForm.SetError(err)
					return m, nil
				}

				m.showVaultModal = false
				m.statusMessage = "✨ Vault unlocked successfully!"
				return m, tea.Batch(m.loadHostsCmd(), m.pollActiveHostsCmd())
			}
			return m, cmd
		}

		// 3. Delete Host Confirmation Modal
		if m.showDeleteModal {
			switch msg.String() {
			case "y", "Y", "enter":
				if m.hostToDelete != nil && m.store != nil {
					_ = m.store.DeleteHost(m.hostToDelete.ID)
					delete(m.metrics, m.hostToDelete.ID)
					delete(m.errors, m.hostToDelete.ID)
					delete(m.consoleLogs, m.hostToDelete.ID)
					m.statusMessage = fmt.Sprintf("🗑️ Host '%s' removed successfully.", m.hostToDelete.Name)
					m.showDeleteModal = false
					m.hostToDelete = nil
					return m, m.loadHostsCmd()
				}
				m.showDeleteModal = false
				return m, nil

			case "n", "N", "esc":
				m.showDeleteModal = false
				m.hostToDelete = nil
				m.statusMessage = "Delete cancelled."
				return m, nil
			}
			return m, nil
		}

		// 4. Add Host Modal
		if m.showAddModal && m.addForm != nil {
			done, data, cmd := m.addForm.Update(msg)
			if done {
				m.showAddModal = false
				if data != nil {
					return m, m.saveNewHostCmd(data)
				}
			}
			return m, cmd
		}

		// 5. Console Pane Active Key Handling
		if m.activePane == PaneConsole {
			switch msg.String() {
			case "esc":
				m.activePane = PaneHostList
				m.consoleInput.Blur()
				if m.fullScreenConsole {
					m.fullScreenConsole = false
					m.initOrResizeViewport()
				}
				m.statusMessage = "📋 Returned to Host Explorer. Press [c] to focus console."
				return m, nil

			case "tab":
				return m, m.completeInputCmd()

			case "up":
				if len(m.cmdHistory) > 0 {
					if m.historyIndex < 0 {
						m.historyIndex = len(m.cmdHistory) - 1
					} else if m.historyIndex > 0 {
						m.historyIndex--
					}
					m.consoleInput.SetValue(m.cmdHistory[m.historyIndex])
					m.consoleInput.SetCursor(len(m.consoleInput.Value()))
					return m, nil
				}
				return m, nil

			case "down":
				if len(m.cmdHistory) > 0 && m.historyIndex >= 0 {
					m.historyIndex++
					if m.historyIndex >= len(m.cmdHistory) {
						m.historyIndex = -1
						m.consoleInput.SetValue("")
					} else {
						m.consoleInput.SetValue(m.cmdHistory[m.historyIndex])
						m.consoleInput.SetCursor(len(m.consoleInput.Value()))
					}
					return m, nil
				}
				return m, nil

			case "pgup", "ctrl+u":
				m.viewport.LineUp(6)
				return m, nil

			case "pgdown", "ctrl+d":
				m.viewport.LineDown(6)
				return m, nil

			case "ctrl+l":
				if len(m.hosts) > 0 {
					curHost := m.hosts[m.selectedIndex]
					delete(m.consoleLogs, curHost.ID)
					m.updateViewportContent()
					m.statusMessage = "Console cleared."
				}
				return m, nil

			case "ctrl+o":
				m.fullScreenConsole = !m.fullScreenConsole
				m.initOrResizeViewport()
				return m, nil

			case "ctrl+k", "f5":
				var distro string
				if len(m.hosts) > 0 && m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) {
					if si := m.sysInfos[m.hosts[m.selectedIndex].ID]; si != nil {
						distro = si.OSDistro
					}
				}
				m.showDrawer = true
				m.drawer = NewRunbookDrawer(distro)
				return m, nil

			case "?":
				if m.consoleInput.Value() == "" {
					var distro string
					if len(m.hosts) > 0 && m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) {
						if si := m.sysInfos[m.hosts[m.selectedIndex].ID]; si != nil {
							distro = si.OSDistro
						}
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
				if cmdText != "" && len(m.hosts) > 0 {
					m.consoleInput.SetValue("")
					m.cmdHistory = append(m.cmdHistory, cmdText)
					m.historyIndex = -1

					curHost := m.hosts[m.selectedIndex]

					// Built-in clear / cls command support
					if strings.EqualFold(cmdText, "clear") || strings.EqualFold(cmdText, "cls") {
						delete(m.consoleLogs, curHost.ID)
						m.updateViewportContent()
						m.statusMessage = "✨ Console cleared."
						return m, nil
					}

					// Direct in-app file editor trigger
					if strings.HasPrefix(cmdText, "edit ") || strings.HasPrefix(cmdText, "vi ") || strings.HasPrefix(cmdText, "vim ") || strings.HasPrefix(cmdText, "nano ") {
						targetFile := strings.TrimSpace(cmdText[strings.Index(cmdText, " "):])
						m.statusMessage = fmt.Sprintf("📖 Opening '%s' for editing on %s...", targetFile, curHost.Name)
						return m, m.openRemoteFileCmd(curHost, targetFile)
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

		// 6. Navigation when outside Console
		if msg.String() == "tab" {
			switch m.activePane {
			case PaneHostList:
				m.activePane = PaneTelemetryDeck
				m.consoleInput.Blur()
			case PaneTelemetryDeck:
				m.activePane = PaneConsole
				m.consoleInput.Focus()
			case PaneConsole:
				m.activePane = PaneHostList
				m.consoleInput.Blur()
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			m.cancel()
			return m, tea.Quit

		case "s", "c", "enter":
			m.activePane = PaneConsole
			m.consoleInput.Focus()
			m.statusMessage = "⌨️ Remote console focused. Type commands and press Enter."
			return m, nil

		case "a", "n":
			m.showAddModal = true
			m.addForm = NewHostForm()
			return m, nil

		case "up", "k":
			if len(m.hosts) > 0 {
				m.selectedIndex--
				if m.selectedIndex < 0 {
					m.selectedIndex = len(m.hosts) - 1
				}
				m.updateViewportContent()
				if m.cfg.Telemetry.PollingInterval > 0 {
					curHost := m.hosts[m.selectedIndex]
					return m, m.pollSingleHostCmd(curHost)
				}
			}
			return m, nil

		case "down", "j":
			if len(m.hosts) > 0 {
				m.selectedIndex++
				if m.selectedIndex >= len(m.hosts) {
					m.selectedIndex = 0
				}
				m.updateViewportContent()
				if m.cfg.Telemetry.PollingInterval > 0 {
					curHost := m.hosts[m.selectedIndex]
					return m, m.pollSingleHostCmd(curHost)
				}
			}
			return m, nil

		case "d", "x":
			if !m.isDemo && len(m.hosts) > 0 {
				m.hostToDelete = m.hosts[m.selectedIndex]
				m.showDeleteModal = true
				return m, nil
			}
			return m, nil

		case "p", ",":
			m.showSettingsModal = true
			m.settingsModal = NewSettingsModal(i18n.GetLang(), m.cfg.Telemetry.PollingInterval)
			return m, nil

		case "?", "ctrl+k":
			var distro string
			if len(m.hosts) > 0 && m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) {
				if si := m.sysInfos[m.hosts[m.selectedIndex].ID]; si != nil {
					distro = si.OSDistro
				}
			}
			m.showDrawer = true
			m.drawer = NewRunbookDrawer(distro)
			return m, nil

		case "r":
			if len(m.hosts) > 0 {
				curHost := m.hosts[m.selectedIndex]
				m.hostStatus[curHost.ID] = HostStatusConnecting
				m.pausedHosts[curHost.ID] = false
				m.statusMessage = fmt.Sprintf("⏳ Connecting to '%s' (%s)...", curHost.Name, curHost.Address)
				return m, m.pollSingleHostCmd(curHost)
			}
			return m, nil

		case "R":
			for _, h := range m.hosts {
				m.pausedHosts[h.ID] = false
				m.hostStatus[h.ID] = HostStatusConnecting
			}
			m.statusMessage = "⏳ Reconnecting all hosts..."
			return m, m.pollActiveHostsCmd()
		}

	case []*storage.Host:
		m.hosts = msg
		if len(m.hosts) > 0 && m.selectedIndex >= len(m.hosts) {
			m.selectedIndex = 0
		}
		if len(m.hosts) > 0 && m.cfg.Telemetry.PollingInterval > 0 {
			return m, m.pollSingleHostCmd(m.hosts[m.selectedIndex])
		}
		return m, nil

	case MetricResultMsg:
		var hostName string
		for _, h := range m.hosts {
			if h.ID == msg.HostID {
				hostName = h.Name
				break
			}
		}

		if msg.Err != nil {
			m.errors[msg.HostID] = msg.Err
			m.hostStatus[msg.HostID] = HostStatusOffline
			m.pausedHosts[msg.HostID] = true
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

	case CmdResultMsg:
		if msg.NewCWD != "" {
			m.hostCWD[msg.HostID] = msg.NewCWD
		}
		cwdDisplay := msg.CWD
		if cwdDisplay == "" {
			cwdDisplay = "~"
		}
		if msg.Err != nil {
			m.appendConsoleLog(msg.HostID, fmt.Sprintf("[%s] ❌ Error (%v):\n%s", cwdDisplay, msg.Err, msg.Stderr))
			m.statusMessage = fmt.Sprintf("⚠️ Command '%s' failed: %v", msg.Command, msg.Err)
		} else {
			out := msg.Stdout
			if out != "" {
				m.appendConsoleLog(msg.HostID, fmt.Sprintf("[%s] ❯ %s\n%s", cwdDisplay, msg.Command, out))
			} else {
				m.appendConsoleLog(msg.HostID, fmt.Sprintf("[%s] ❯ %s\n(no output)", cwdDisplay, msg.Command))
			}
			m.statusMessage = fmt.Sprintf("✅ Executed '%s' successfully (%s)", msg.Command, time.Now().Format("15:04:05"))
		}
		m.updateViewportContent()
		m.viewport.GotoBottom()
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

func (m *Model) changeVaultPassword(currPassword, newPassword string) error {
	if m.store == nil || m.vault == nil {
		return fmt.Errorf("vault or storage not available")
	}

	// Verify current password with a temporary verification vault
	tempVerifyVault := vault.New()
	err := m.store.UnlockVault(tempVerifyVault, currPassword)
	if err != nil {
		return fmt.Errorf("current password verification failed: %w", err)
	}
	tempVerifyVault.Lock()

	// Perform full rekeying
	newVault := vault.New()
	err = m.store.RekeyVault(m.vault, newVault, newPassword)
	if err != nil {
		return fmt.Errorf("failed to change master password: %w", err)
	}

	// Replace memory vault
	m.vault.Lock()
	m.vault = newVault
	return nil
}
