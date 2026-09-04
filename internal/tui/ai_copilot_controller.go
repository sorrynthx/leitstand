package tui

import (
	"fmt"
	"leitstand/internal/i18n"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// ToggleAICopilot toggles the AI Copilot sidecar modal on or off.
func (m *Model) ToggleAICopilot() (tea.Model, tea.Cmd) {
	if len(m.hosts) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(m.hosts) {
		m.statusMessage = i18n.T("welcome_no_hosts")
		return m, nil
	}

	curHost := m.hosts[m.selectedIndex]
	if m.showAICopilot {
		m.showAICopilot = false
		m.activePane = PaneConsole
		m.consoleInput.Focus()
		m.updateViewportContent()
		m.statusMessage = i18n.T("ai_copilot_closed")
		return m, textinput.Blink
	}

	if m.hostStatus[curHost.ID] != HostStatusOnline {
		m.statusMessage = i18n.T("ai_err_host_offline")
		return m, nil
	}

	hts := m.GetOrCreateHostTabs(curHost.ID, curHost.Name)
	activeTab := hts.ActiveTab()
	distro := ""
	if sysInfo, ok := m.sysInfos[curHost.ID]; ok && sysInfo != nil {
		distro = sysInfo.OSDistro
	}

	if m.aiCopilot == nil || m.aiCopilot.HostID != curHost.ID {
		m.aiCopilot = NewAICopilotModal(curHost.ID, curHost.Name, distro, m.cfg, m.store, m.vault)
	} else {
		m.aiCopilot.initClient()
	}
	m.aiCopilot.UpdateHostContext(activeTab, distro)
	m.aiCopilot.ResetForNewQuery()
	m.activePane = PaneConsole
	m.showAICopilot = true
	m.statusMessage = fmt.Sprintf("🤖 %s (%s)", i18n.T("ai_inline_title"), curHost.Name)
	return m, nil
}


// UpdateAICopilot delegates events to the active AI Copilot modal.
func (m *Model) UpdateAICopilot(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	if !m.showAICopilot || m.aiCopilot == nil {
		return m, nil, false
	}

	closeModal, injectedCmd, runNow, cmd := m.aiCopilot.Update(msg)
	if closeModal {
		m.showAICopilot = false
		if injectedCmd != "" && m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) {
			curHost := m.hosts[m.selectedIndex]
			hts := m.GetOrCreateHostTabs(curHost.ID, curHost.Name)
			tab := hts.ActiveTab()
			if tab != nil {
				tab.InputText = injectedCmd
			}

			if runNow {
				m.activePane = PaneConsole
				m.consoleInput.Focus()
				m.updateViewportContent()
				m.statusMessage = fmt.Sprintf("🚀 AI 명령어 실행: %s", injectedCmd)
				execCmd := m.execRemoteCmd(curHost, injectedCmd)
				return m, tea.Batch(execCmd, textinput.Blink), true
			}

			// Tab was pressed: copy to input for editing
			m.consoleInput.SetValue(injectedCmd)
			m.consoleInput.Focus()
			m.statusMessage = fmt.Sprintf("✏️ 명령어 복사 완료: %s", injectedCmd)
			return m, nil, true
		}

		m.activePane = PaneConsole
		m.consoleInput.Focus()
		m.updateViewportContent()
		m.statusMessage = i18n.T("ai_copilot_closed")
		return m, textinput.Blink, true
	}

	return m, cmd, true
}
