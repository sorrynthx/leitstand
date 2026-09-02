package tui

import (
	"leitstand/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
)

type TerminalExitedMsg struct{}

func (m *Model) Init() tea.Cmd {
	if m.isDemo {
		m.hosts = CreateDemoHosts()
		return m.tickCmd()
	}

	return tea.Batch(
		m.fetchHostsCmd(),
		m.tickCmd(),
	)
}

func (m *Model) fetchHostsCmd() tea.Cmd {
	return func() tea.Msg {
		if m.store == nil {
			return []*storage.Host{}
		}
		hosts, err := m.store.ListHosts()
		if err != nil {
			return []*storage.Host{}
		}
		return hosts
	}
}

func (m *Model) loadHostsCmd() tea.Cmd {
	return m.fetchHostsCmd()
}

func (m *Model) GetOrCreateHostTabs(hostID int64, hostName string) *HostTabs {
	if hts, ok := m.hostTabs[hostID]; ok {
		return hts
	}
	newHts := NewHostTabState(hostID, hostName, "")
	m.hostTabs[hostID] = newHts
	return newHts
}

func (m *Model) CurrentActiveTab() *ConsoleTab {
	if len(m.hosts) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(m.hosts) {
		return nil
	}
	curHost := m.hosts[m.selectedIndex]
	hts := m.GetOrCreateHostTabs(curHost.ID, curHost.Name)
	return hts.ActiveTab()
}
