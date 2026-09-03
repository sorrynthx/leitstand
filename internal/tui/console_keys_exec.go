package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updateConsoleKeys(msg tea.KeyMsg, keyStr string) (tea.Model, tea.Cmd) {
	if len(m.hosts) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(m.hosts) {
		return m, nil
	}

	curHost := m.hosts[m.selectedIndex]
	st := m.hostStatus[curHost.ID]
	hts := m.GetOrCreateHostTabs(curHost.ID, curHost.Name)

	if model, cmd, handled := m.handleConsoleTabKeys(keyStr, hts); handled {
		return model, cmd
	}

	switch keyStr {
	case "esc":
		m.activePane = PaneHostList
		m.consoleInput.Blur()
		m.updateViewportContent()
		m.statusMessage = "📋 Switched focus to Host List."
		return m, nil

	case "tab":
		if strings.TrimSpace(m.consoleInput.Value()) == "" {
			m.activePane = PaneHostList
			m.consoleInput.Blur()
			m.updateViewportContent()
			m.statusMessage = "📋 Switched to Host List."
			return m, nil
		}
		newM, cmd, _ := m.handleConsoleAutoCompletion()
		return newM, cmd

	case "ctrl+t":
		return m, m.launchInteractiveTerminalCmd(curHost)
	case "f6", "ctrl+f":
		return m, m.openFileManagerCmd(curHost)
	case "ctrl+e":
		return m, m.exportCurrentSessionLogCmd(curHost, hts.ActiveTab())
	case "ctrl+p", "ctrl+P":
		m.showSettingsModal = true
		m.settingsModal = NewSettingsModal(i18n.GetLang(), m.cfg.Telemetry.PollingInterval, m.cfg.Telemetry.CPUThreshold, m.cfg.Telemetry.RAMThreshold, m.cfg.Telemetry.DiskThreshold, m.cfg.Logging.SessionLogDir, m.store, m.vault)
		return m, nil

	case "pgup":
		tab := hts.ActiveTab()
		if tab != nil {
			tab.Viewport.LineUp(6)
		}
		return m, nil

	case "pgdown":
		tab := hts.ActiveTab()
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
		m.updateViewportContent()
		return m, nil

	case "ctrl+k":
		var distro string
		if si := m.sysInfos[curHost.ID]; si != nil {
			distro = si.OSDistro
		}
		m.showTelemetryDrawer = false
		m.showDrawer = true
		m.drawer = NewRunbookDrawer(distro)
		return m, nil

	case "?":
		if m.consoleInput.Value() == "" {
			var distro string
			if si := m.sysInfos[curHost.ID]; si != nil {
				distro = si.OSDistro
			}
			m.showTelemetryDrawer = false
			m.showDrawer = true
			m.drawer = NewRunbookDrawer(distro)
			return m, nil
		}

	case "ctrl+c":
		tab := hts.ActiveTab()
		if tab != nil && tab.IsStreaming {
			tab.StopScreenApp(tab.StreamCmd)
			m.statusMessage = fmt.Sprintf(i18n.T("tab_streaming_cancel"), tab.StreamCmd)
			m.updateViewportContent()
			return m, nil
		}

	case "enter":
		if st == HostStatusConnecting {
			m.statusMessage = fmt.Sprintf("⏳ Server '%s' is connecting... Please wait.", curHost.Name)
			m.updateViewportContent()
			return m, nil
		}
		if st == HostStatusOffline {
			m.statusMessage = fmt.Sprintf("🔌 Server '%s' is offline. Press [r] to reconnect.", curHost.Name)
			m.updateViewportContent()
			return m, nil
		}

		cmdText := strings.TrimSpace(m.consoleInput.Value())
		if cmdText == "" {
			tab := hts.ActiveTab()
			if tab != nil {
				cwd := tab.CWD
				if cwd == "" {
					cwd = "~"
				}
				if tab.IsRoot {
					tab.AppendLog(fmt.Sprintf("[root@%s:%s]# ", curHost.Name, cwd))
				} else {
					tab.AppendLog(fmt.Sprintf("[%s] ❯ ", cwd))
				}
				m.updateViewportContent()
			}
			return m, nil
		}

		m.consoleInput.SetValue("")
		tab := hts.ActiveTab()
		if tab != nil {
			tab.CmdHistory = append(tab.CmdHistory, cmdText)
			tab.HistoryIndex = -1
		}

		if cmdText == "exit" || cmdText == "logout" {
			if tab != nil && tab.IsRoot {
				tab.IsRoot = false
				tab.SetAutoTitle(hts.ActiveIndex, "")
				tab.AppendLog("👋 Logged out from root session. Returned to standard user ($).")
				m.updateViewportContent()
				m.statusMessage = "Logged out from root session."
				return m, nil
			}
			if len(hts.Tabs) > 1 {
				closed := hts.CloseActiveTab()
				if closed {
					m.statusMessage = i18n.T("tab_closed_msg")
					m.updateViewportContent()
					return m, nil
				}
			} else if tab != nil {
				tab.AppendLog("👋 SSH session exited. Type 'clear' or press [Ctrl+T] for new tab.")
				m.updateViewportContent()
				m.statusMessage = "SSH session exited."
				return m, nil
			}
		}

		if strings.EqualFold(cmdText, "clear") || strings.EqualFold(cmdText, "cls") {
			if tab != nil {
				tab.Logs = make([]string, 0)
				tab.UpdateViewportContent()
			}
			m.viewport.SetContent("")
			m.statusMessage = i18n.T("console_cleared")
			return m, nil
		}

		if strings.EqualFold(cmdText, "sftp") || strings.EqualFold(cmdText, "files") || strings.EqualFold(cmdText, "filemanager") || strings.EqualFold(cmdText, "mc") {
			return m, m.openFileManagerCmd(curHost)
		}

		if strings.HasPrefix(cmdText, "edit ") || strings.HasPrefix(cmdText, "vi ") || strings.HasPrefix(cmdText, "vim ") || strings.HasPrefix(cmdText, "nano ") {
			targetFile := strings.TrimSpace(cmdText[strings.Index(cmdText, " "):])
			m.statusMessage = fmt.Sprintf("📖 Opening '%s' for editing on %s...", targetFile, curHost.Name)
			return m, m.openRemoteFileCmd(curHost, targetFile)
		}

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
					tab.AppendLog(fmt.Sprintf("⏳ [접속 및 권한 검증 중...] Authenticating root credentials on %s...", curHost.Name))
					m.updateViewportContent()
					m.statusMessage = fmt.Sprintf("⏳ [접속 중...] Authenticating root privilege on %s...", curHost.Name)
					return m, m.execSudoValidateAndElevateCmd(curHost, tab, cachedPass, true)
				}
				m.pendingSudoCmd = "su"
				m.sudoModal = NewSudoModal(curHost.Name, curHost.Username, "Elevate Tab to Root (su root)")
				m.showSudoModal = true
				return m, nil
			}
		}

		if strings.HasPrefix(cmdLower, "sudo ") {
			if cachedPass, ok := m.sudoCache[curHost.ID]; ok && cachedPass != "" {
				m.statusMessage = fmt.Sprintf("⏳ Running with elevated privilege: '%s'...", cmdText)
				return m, m.execSudoCmd(curHost, cmdText, cachedPass)
			}
			m.pendingSudoCmd = cmdText
			m.sudoModal = NewSudoModal(curHost.Name, curHost.Username, cmdText)
			m.showSudoModal = true
			return m, nil
		}

		if tab != nil && tab.IsRoot {
			pass := m.sudoCache[curHost.ID]
			if IsStreamingCommand(cmdText) {
				tab.AppendLog(fmt.Sprintf("❯ %s  [🔴 LIVE STREAMING - Press Ctrl+C to stop]", cmdText))
				return m, m.execStreamingCmdInTab(curHost, tab, cmdText)
			}
			m.statusMessage = fmt.Sprintf("⏳ Running as root: '%s' on %s...", cmdText, curHost.Name)
			return m, m.execSudoCmd(curHost, cmdText, pass)
		}

		m.statusMessage = fmt.Sprintf("⏳ Running '%s' on %s...", cmdText, curHost.Name)
		return m, m.execRemoteCmd(curHost, cmdText)
	}

	var cmd tea.Cmd
	m.consoleInput, cmd = m.consoleInput.Update(msg)
	return m, cmd
}
