package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"leitstand/internal/storage"
	"leitstand/internal/vault"

	tea "github.com/charmbracelet/bubbletea"
)

// updateHostListNavigation handles shortcuts when navigating outside the console.
// Returns (updatedModel, teaCmd, wasHandled).
func (m *Model) updateHostListNavigation(keyStr string) (tea.Model, tea.Cmd, bool) {
	if keyStr == "tab" {
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
		return m, nil, true
	}

	switch keyStr {
	case "q", "ctrl+c":
		m.cancel()
		return m, tea.Quit, true

	case "s", "c", "enter":
		m.activePane = PaneConsole
		m.consoleInput.Focus()
		m.statusMessage = "⌨️ Remote console focused. Type commands and press Enter."
		return m, nil, true

	case "a", "n":
		m.showAddModal = true
		m.addForm = NewHostForm()
		return m, nil, true

	case "e", "E":
		if !m.isDemo && len(m.hosts) > 0 && m.store != nil && m.vault != nil {
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

	case "up", "k":
		if len(m.hosts) > 0 {
			m.selectedIndex--
			if m.selectedIndex < 0 {
				m.selectedIndex = len(m.hosts) - 1
			}
			m.updateViewportContent()
			if m.cfg.Telemetry.PollingInterval > 0 {
				curHost := m.hosts[m.selectedIndex]
				return m, m.pollSingleHostCmd(curHost), true
			}
		}
		return m, nil, true

	case "down", "j":
		if len(m.hosts) > 0 {
			m.selectedIndex++
			if m.selectedIndex >= len(m.hosts) {
				m.selectedIndex = 0
			}
			m.updateViewportContent()
			if m.cfg.Telemetry.PollingInterval > 0 {
				curHost := m.hosts[m.selectedIndex]
				return m, m.pollSingleHostCmd(curHost), true
			}
		}
		return m, nil, true

	case "d", "x":
		if !m.isDemo && len(m.hosts) > 0 {
			m.hostToDelete = m.hosts[m.selectedIndex]
			m.showDeleteModal = true
			return m, nil, true
		}
		return m, nil, true

	case "p", ",":
		m.showSettingsModal = true
		m.settingsModal = NewSettingsModal(i18n.GetLang(), m.cfg.Telemetry.PollingInterval)
		return m, nil, true

	case "f", "F", "f6", "F6":
		if len(m.hosts) > 0 {
			curHost := m.hosts[m.selectedIndex]
			return m, m.openFileManagerCmd(curHost), true
		}
		return m, nil, true

	case "?", "ctrl+k":
		var distro string
		if len(m.hosts) > 0 && m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) {
			if si := m.sysInfos[m.hosts[m.selectedIndex].ID]; si != nil {
				distro = si.OSDistro
			}
		}
		m.showDrawer = true
		m.drawer = NewRunbookDrawer(distro)
		return m, nil, true

	case "t", "T":
		if len(m.hosts) > 0 {
			curHost := m.hosts[m.selectedIndex]
			m.statusMessage = fmt.Sprintf("💻 Launching full interactive terminal on %s...", curHost.Name)
			return m, m.launchInteractiveTerminalCmd(curHost), true
		}
		return m, nil, true

	case "r", "R":
		if len(m.hosts) > 0 {
			curHost := m.hosts[m.selectedIndex]
			m.hostStatus[curHost.ID] = HostStatusConnecting
			m.statusMessage = fmt.Sprintf("⏳ Connecting to '%s' (%s)...", curHost.Name, curHost.Address)
			return m, m.pollSingleHostCmd(curHost), true
		}
		return m, nil, true
	}

	return m, nil, false
}
