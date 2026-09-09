package tui

import (
	"leitstand/internal/i18n"
	"leitstand/internal/storage"
	"leitstand/internal/telemetry"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleSettingsModalResult(res SettingsResult, cmd tea.Cmd) (tea.Model, tea.Cmd, bool) {
	if res.FactoryResetDone {
		m.hosts = nil
		m.selectedIndex = 0
		m.hostTabs = make(map[int64]*HostTabs)
		m.hostStatus = make(map[int64]HostStatus)
		m.metrics = make(map[int64]*storage.MetricRecord)
		m.sysInfos = make(map[int64]*telemetry.SysInfo)
		m.errors = make(map[int64]error)
		m.sudoCache = make(map[int64]string)
		m.sudoModeCache = make(map[int64]int)
		if m.tunnelMgr != nil {
			m.tunnelMgr.CloseAll()
		}
		if m.sshPool != nil {
			m.sshPool.CloseAll()
		}
		if m.vault != nil {
			m.vault.Lock()
		}
		m.showSettingsModal = false
		m.settingsModal = nil
		m.showVaultModal = true
		m.vaultModal = NewVaultForm(VaultModalInit)
		m.statusMessage = "✨ " + i18n.T("db_success_reset")
		m.updateViewportContent()
		return m, nil, true
	}

	if res.Done && !res.SaveReq {
		m.showSettingsModal = false
		m.settingsModal = nil
		m.statusMessage = "Settings closed."
		if res.HostsImported {
			return m, m.loadHostsCmd(), true
		}
		return m, nil, true
	}

	if res.SaveReq {
		m.applyAndPersistSettings(res)
		m.showSettingsModal = false
		m.settingsModal = nil
		m.statusMessage = "✨ " + i18n.T("settings_saved_msg")
		m.updateViewportContent()
		return m, nil, true
	}

	return m, cmd, true
}
