package tui

import (
	"leitstand/internal/i18n"
	"leitstand/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
)

type TerminalExitedMsg struct{}

func (m *Model) Init() tea.Cmd {
	if m.isDemo {
		i18n.SetLang(i18n.LangEN)
		m.hosts = CreateDemoHosts()
		m.hostStatus[101] = HostStatusOnline
		m.userHasNavigated = true
		hts := m.GetOrCreateHostTabs(101, "prod-web-01")
		if tab := hts.ActiveTab(); tab != nil && len(tab.Logs) == 0 {
			tab.AppendLog("Connected to prod-web-01 (192.168.1.10) [Simulated SSH Session]")
			tab.AppendLog("Ubuntu 22.04 LTS (Jammy Jellyfish) - Linux 5.15.0-88-generic x86_64")
			tab.AppendLog("Type 'ls', 'uptime', 'free -h', or 'df -h' to test simulated execution.")
		}
		m.updateViewportContent()
		return tea.Batch(
			m.pollHostMetric(m.hosts[0]),
			m.tickCmd(),
		)
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
