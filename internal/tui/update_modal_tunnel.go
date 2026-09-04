package tui

import (
	"leitstand/internal/i18n"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updateTunnelModal(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	if !m.showTunnelModal || m.tunnelModal == nil {
		return m, nil, false
	}

	done, statusMsg, cmd := m.tunnelModal.Update(msg)
	if statusMsg != "" {
		m.statusMessage = statusMsg
	}
	if done {
		m.showTunnelModal = false
		m.tunnelModal = nil
		if statusMsg == "" {
			m.statusMessage = i18n.T("tunnel_closed_msg")
		}
		return m, nil, true
	}
	return m, cmd, true
}
