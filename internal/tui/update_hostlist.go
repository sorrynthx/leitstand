package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"leitstand/internal/storage"
	"leitstand/internal/vault"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updateHostListNavigation(keyStr string) (tea.Model, tea.Cmd, bool) {
	if keyStr == "tab" {
		if len(m.hosts) == 0 {
			m.activePane = PaneHostList
			m.consoleInput.Blur()
			m.statusMessage = "⚠️ " + strings.ReplaceAll(i18n.T("no_hosts"), "\n", " ")
			m.updateViewportContent()
			return m, nil, true
		}
		if m.activePane == PaneConsole && strings.TrimSpace(m.consoleInput.Value()) != "" {
			return m.handleConsoleAutoCompletion()
		}
		switch m.activePane {
		case PaneHostList:
			m.activePane = PaneConsole
			m.consoleInput.Focus()
		case PaneConsole:
			m.activePane = PaneHostList
			m.consoleInput.Blur()
		default:
			m.activePane = PaneHostList
			m.consoleInput.Blur()
		}
		m.updateViewportContent()
		return m, nil, true
	}

	syncHostInput := func(oldIdx int) {
		if oldIdx >= 0 && oldIdx < len(m.hosts) {
			oldHost := m.hosts[oldIdx]
			if hts, ok := m.hostTabs[oldHost.ID]; ok && hts.ActiveTab() != nil {
				hts.ActiveTab().InputText = m.consoleInput.Value()
			}
		}
		if m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) {
			newHost := m.hosts[m.selectedIndex]
			if hts, ok := m.hostTabs[newHost.ID]; ok && hts.ActiveTab() != nil {
				m.consoleInput.SetValue(hts.ActiveTab().InputText)
				m.consoleInput.SetCursor(len(hts.ActiveTab().InputText))
			} else {
				m.consoleInput.SetValue("")
			}
		}
	}

	onHostNav := func(oldIdx int) (tea.Model, tea.Cmd, bool) {
		syncHostInput(oldIdx)
		curHost := m.hosts[m.selectedIndex]
		m.updateViewportContent()
		if m.hostStatus[curHost.ID] == HostStatusOnline {
			m.statusMessage = fmt.Sprintf("🟢 [%s] %s (%s)", i18n.T("status_online"), curHost.Name, curHost.Address)
			return m, m.pollHostMetric(curHost), true
		}
		m.statusMessage = i18n.Tf("status_host_selected", curHost.Name, curHost.Address)
		return m, nil, true
	}

	switch keyStr {
	case "q", "ctrl+c":
		m.cancel()
		return m, tea.Quit, true

	case "esc":
		m.activePane = PaneHostList
		m.consoleInput.Blur()
		m.updateViewportContent()
		return m, nil, true

	case "up", "k":
		if len(m.hosts) > 0 {
			oldIdx := m.selectedIndex
			m.selectedIndex--
			if m.selectedIndex < 0 {
				m.selectedIndex = len(m.hosts) - 1
			}
			return onHostNav(oldIdx)
		}
		return m, nil, true

	case "down", "j":
		if len(m.hosts) > 0 {
			oldIdx := m.selectedIndex
			m.selectedIndex++
			if m.selectedIndex >= len(m.hosts) {
				m.selectedIndex = 0
			}
			return onHostNav(oldIdx)
		}
		return m, nil, true

	case "pgup", "pageup", "ctrl+u", "ctrl+b":
		if len(m.hosts) > 0 {
			oldIdx := m.selectedIndex
			m.selectedIndex -= 5
			if m.selectedIndex < 0 {
				m.selectedIndex = 0
			}
			return onHostNav(oldIdx)
		}
		return m, nil, true

	case "pgdown", "pagedown", "ctrl+d":
		if len(m.hosts) > 0 {
			oldIdx := m.selectedIndex
			m.selectedIndex += 5
			if m.selectedIndex >= len(m.hosts) {
				m.selectedIndex = len(m.hosts) - 1
			}
			return onHostNav(oldIdx)
		}
		return m, nil, true

	case "home", "g":
		if len(m.hosts) > 0 {
			oldIdx := m.selectedIndex
			m.selectedIndex = 0
			return onHostNav(oldIdx)
		}
		return m, nil, true

	case "end":
		if len(m.hosts) > 0 {
			oldIdx := m.selectedIndex
			m.selectedIndex = len(m.hosts) - 1
			return onHostNav(oldIdx)
		}
		return m, nil, true

	case "r", "R":
		if len(m.hosts) > 0 && m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) {
			curHost := m.hosts[m.selectedIndex]
			delete(m.errors, curHost.ID)
			m.hostStatus[curHost.ID] = HostStatusConnecting
			m.updateViewportContent()
			m.statusMessage = i18n.Tf("status_reconnecting_host", curHost.Name, curHost.Address)
			return m, m.pollHostMetric(curHost), true
		}
		return m, nil, true

	case "d", "x":
		if !m.isDemo && len(m.hosts) > 0 && m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) {
			m.hostToDelete = m.hosts[m.selectedIndex]
			m.showDeleteModal = true
			return m, nil, true
		}
		return m, nil, true

	case "p", ",":
		m.showSettingsModal = true
		m.settingsModal = NewSettingsModal(i18n.GetLang(), m.cfg.Telemetry.PollingInterval, m.cfg.Telemetry.CPUThreshold, m.cfg.Telemetry.RAMThreshold, m.cfg.Telemetry.DiskThreshold)
		return m, nil, true

	case "m", "M", "f5", "F5":
		if len(m.hosts) > 0 && m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) {
			m.showDrawer = false
			m.showTelemetryDrawer = !m.showTelemetryDrawer
			m.updateViewportContent()
			if m.showTelemetryDrawer {
				m.statusMessage = i18n.T("telemetry_deck_expanded")
			} else {
				m.statusMessage = i18n.T("telemetry_deck_collapsed")
			}
			return m, nil, true
		}
		return m, nil, true

	case "?", "b", "B":
		if len(m.hosts) > 0 && m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) {
			curHost := m.hosts[m.selectedIndex]
			var distro string
			if si := m.sysInfos[curHost.ID]; si != nil {
				distro = si.OSDistro
			}
			m.showTelemetryDrawer = false
			m.showDrawer = true
			m.drawer = NewRunbookDrawer(distro)
			return m, nil, true
		}
		return m, nil, true

	case "f", "F", "f6", "F6":
		if len(m.hosts) > 0 && m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) {
			curHost := m.hosts[m.selectedIndex]
			delete(m.errors, curHost.ID)
			m.hostStatus[curHost.ID] = HostStatusConnecting
			return m, m.openFileManagerCmd(curHost), true
		}
		return m, nil, true

	case "t", "T", "ctrl+t":
		if len(m.hosts) > 0 && m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) {
			curHost := m.hosts[m.selectedIndex]
			delete(m.errors, curHost.ID)
			m.hostStatus[curHost.ID] = HostStatusConnecting
			return m, m.launchInteractiveTerminalCmd(curHost), true
		}
		return m, nil, true

	case "s", "c", "enter":
		if len(m.hosts) == 0 {
			m.statusMessage = "⚠️ " + strings.ReplaceAll(i18n.T("no_hosts"), "\n", " ")
			return m, nil, true
		}
		m.activePane = PaneConsole
		m.consoleInput.Focus()
		if m.selectedIndex < 0 && len(m.hosts) > 0 {
			m.selectedIndex = 0
		}
		if len(m.hosts) > 0 && m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) {
			curHost := m.hosts[m.selectedIndex]
			delete(m.errors, curHost.ID)
			m.hostStatus[curHost.ID] = HostStatusConnecting
			m.statusMessage = i18n.Tf("status_connecting_host", curHost.Name, curHost.Address)
			return m, m.pollHostMetric(curHost), true
		}
		m.statusMessage = "⌨️ Remote console focused. Type commands and press Enter."
		return m, nil, true

	case "a", "n":
		m.showAddModal = true
		m.addForm = NewHostForm()
		return m, nil, true

	case "e", "E":
		if !m.isDemo && len(m.hosts) > 0 && m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) && m.store != nil && m.vault != nil {
			curHost := m.hosts[m.selectedIndex]
			secret, _ := m.store.GetHostSecret(curHost.ID)
			var payload *storage.SecretPayload
			if secret != nil {
				decrypted, err := m.vault.Decrypt(secret.Nonce, secret.Ciphertext)
				if err == nil {
					payload, _ = storage.ParseSecretPayload(decrypted, secret.AuthMethod)
					vault.ZeroBytes(decrypted)
				}
			}
			m.hostToEdit = curHost
			m.editForm = NewEditHostForm(curHost, secret, payload)
			m.showEditModal = true
			return m, nil, true
		}
		return m, nil, true
	}

	return m, nil, false
}
