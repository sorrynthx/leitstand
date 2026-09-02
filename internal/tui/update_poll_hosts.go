package tui

import (
	"fmt"
	appstorage "leitstand/internal/storage"
	apptelemetry "leitstand/internal/telemetry"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type TickMsg time.Time

func (m *Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m *Model) pollActiveHostsCmd() tea.Cmd {
	if len(m.hosts) == 0 {
		return nil
	}

	// Skip polling if Vault is not unlocked yet and not in demo mode
	if !m.isDemo && (m.vault == nil || !m.vault.IsUnlocked()) {
		return nil
	}

	// Respect Telemetry Polling Interval setting (if <= 0, polling is Off)
	interval := m.cfg.Telemetry.PollingInterval
	if interval <= 0 {
		return nil
	}

	if !m.lastPollTime.IsZero() && time.Since(m.lastPollTime) < interval {
		return nil
	}

	if m.selectedIndex < 0 || m.selectedIndex >= len(m.hosts) {
		return nil
	}

	curHost := m.hosts[m.selectedIndex]
	if curHost == nil || m.pausedHosts[curHost.ID] {
		return nil
	}

	st := m.hostStatus[curHost.ID]
	if !m.isDemo && st != HostStatusOnline && st != HostStatusConnecting {
		return nil
	}

	m.lastPollTime = time.Now()
	return m.pollHostMetric(curHost)
}

func (m *Model) pollHostMetric(host *appstorage.Host) tea.Cmd {
	return func() tea.Msg {
		if m.isDemo {
			rec, sysInfo := CreateDemoMetrics(host.ID)
			return MetricResultMsg{
				HostID:  host.ID,
				Record:  rec,
				SysInfo: sysInfo,
				Err:     nil,
			}
		}

		if m.collector == nil {
			return MetricResultMsg{
				HostID: host.ID,
				Err:    fmt.Errorf("telemetry collector is not initialized"),
			}
		}

		client, err := m.getSSHClient(host)
		if err != nil {
			return MetricResultMsg{
				HostID: host.ID,
				Err:    err,
			}
		}

		ctx := m.ctx
		rec, sysInfo, err := m.collector.CollectFromClient(ctx, host, client)
		if err != nil {
			return MetricResultMsg{
				HostID: host.ID,
				Err:    err,
			}
		}

		return MetricResultMsg{
			HostID:  host.ID,
			Record:  rec,
			SysInfo: sysInfo,
			Err:     nil,
		}
	}
}

type MetricResultMsg struct {
	HostID  int64
	Record  *appstorage.MetricRecord
	SysInfo *apptelemetry.SysInfo
	Err     error
}
