package tui

import (
	"fmt"
	"leitstand/internal/storage"
	"leitstand/internal/telemetry"
	"leitstand/internal/vault"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) loadHostsCmd() tea.Cmd {
	return func() tea.Msg {
		if m.isDemo {
			return []*storage.Host{
				{ID: 101, Name: "prod-api-01", Address: "10.0.1.11", Port: 22, Username: "ubuntu", GroupName: "Production (EU)"},
				{ID: 102, Name: "prod-api-02", Address: "10.0.1.12", Port: 22, Username: "ubuntu", GroupName: "Production (EU)"},
				{ID: 103, Name: "prod-db-master", Address: "10.0.2.10", Port: 22, Username: "postgres", GroupName: "Databases"},
				{ID: 104, Name: "stage-k8s-node1", Address: "192.168.10.51", Port: 22, Username: "admin", GroupName: "Staging Cluster"},
				{ID: 105, Name: "stage-k8s-node2", Address: "192.168.10.52", Port: 22, Username: "admin", GroupName: "Staging Cluster"},
			}
		}

		if m.store == nil {
			return nil
		}
		hosts, err := m.store.ListHosts()
		if err != nil {
			return nil
		}
		return hosts
	}
}

func (m *Model) tickCmd() tea.Cmd {
	interval := m.cfg.Telemetry.PollingInterval
	if interval <= 0 {
		return nil // Disabled (Off mode)
	}
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m *Model) pollSingleHostCmd(h *storage.Host) tea.Cmd {
	return func() tea.Msg {
		if m.isDemo {
			return nil
		}
		if m.collector == nil || h == nil {
			return nil
		}
		rec, err := m.collector.CollectHost(h)
		var sysInfo *telemetry.SysInfo
		if state, exists := m.collector.GetHostState(h.ID); exists && state != nil {
			sysInfo = state.SysInfo
		}
		return MetricResultMsg{
			HostID:  h.ID,
			Record:  rec,
			SysInfo: sysInfo,
			Err:     err,
		}
	}
}

func (m *Model) pollActiveHostsCmd() tea.Cmd {
	if m.isDemo {
		return func() tea.Msg {
			m.generateDemoMetrics()
			return nil
		}
	}

	// If telemetry is disabled or no hosts available, do not poll
	if m.cfg.Telemetry.PollingInterval <= 0 || m.collector == nil || len(m.hosts) == 0 {
		return nil
	}

	// Poll only the currently selected host (Zero overhead for other hosts)
	if m.selectedIndex < 0 || m.selectedIndex >= len(m.hosts) {
		return nil
	}
	h := m.hosts[m.selectedIndex]
	if m.pausedHosts[h.ID] || m.hostStatus[h.ID] == HostStatusOffline {
		return nil
	}

	return func() tea.Msg {
		rec, err := m.collector.CollectHost(h)
		var sysInfo *telemetry.SysInfo
		if state, exists := m.collector.GetHostState(h.ID); exists && state != nil {
			sysInfo = state.SysInfo
		}
		return MetricResultMsg{
			HostID:  h.ID,
			Record:  rec,
			SysInfo: sysInfo,
			Err:     err,
		}
	}
}

func (m *Model) generateDemoMetrics() {
	now := time.Now()
	distros := []string{"Ubuntu 24.04 LTS (Noble)", "Ubuntu 22.04.4 LTS (Jammy)", "Debian 12 (Bookworm)", "Rocky Linux 9.4 (Blue Onyx)", "Amazon Linux 2023"}
	kernels := []string{"Linux 6.8.0-40-generic x86_64", "Linux 6.5.0-35-generic x86_64", "Linux 6.1.0-21-amd64 x86_64", "Linux 5.14.0-427.el9 x86_64", "Linux 6.1.84-99.169.amzn2023.x86_64"}

	for i, h := range m.hosts {
		prev := m.metrics[h.ID]
		cpu := 25.0 + rand.Float64()*45.0
		if prev != nil {
			cpu = prev.CPUPercent*0.7 + (20.0+rand.Float64()*50.0)*0.3
		}

		memTotal := uint64(16 * 1024 * 1024 * 1024)
		memUsed := uint64((float64(memTotal) * (0.45 + rand.Float64()*0.25)))
		diskTotal := uint64(256 * 1024 * 1024 * 1024)
		diskUsed := uint64(float64(diskTotal) * 0.58)

		m.metrics[h.ID] = &storage.MetricRecord{
			HostID:      h.ID,
			Timestamp:   now,
			CPUPercent:  cpu,
			MemoryTotal: memTotal,
			MemoryUsed:  memUsed,
			DiskUsed:    diskUsed,
			DiskTotal:   diskTotal,
			NetRxBytes:  uint64(50*1024 + rand.Intn(450*1024)), // 50 KB/s - 500 KB/s
			NetTxBytes:  uint64(10*1024 + rand.Intn(180*1024)), // 10 KB/s - 190 KB/s
		}

		m.sysInfos[h.ID] = &telemetry.SysInfo{
			OSDistro: distros[i%len(distros)],
			Kernel:   kernels[i%len(kernels)],
			Uptime:   fmt.Sprintf("up %d days, %d hrs", (i+1)*3+2, i*4+1),
			CPUCores: 4 + (i%3)*4,
		}
	}
}

func (m *Model) saveNewHostCmd(data *HostFormData) tea.Cmd {
	return func() tea.Msg {
		if m.store == nil || m.vault == nil {
			return nil
		}

		payload := &storage.SecretPayload{
			Password:   data.Password,
			PrivateKey: data.KeyContent,
			Passphrase: data.Passphrase,
		}

		rawJSON, err := payload.Encode()
		if err != nil {
			m.statusMessage = fmt.Sprintf("⚠️ Failed to encode secret: %v", err)
			return nil
		}

		nonce, ciphertext, err := m.vault.Encrypt(rawJSON)
		vault.ZeroBytes(rawJSON)
		if err != nil {
			m.statusMessage = fmt.Sprintf("⚠️ Vault encryption failed: %v", err)
			return nil
		}

		h := &storage.Host{
			Name:      data.Name,
			Address:   data.Address,
			Port:      data.Port,
			Username:  data.Username,
			GroupName: data.Group,
		}

		id, err := m.store.CreateHost(h)
		if err != nil {
			m.statusMessage = fmt.Sprintf("⚠️ Failed to save host: %v", err)
			return nil
		}

		err = m.store.SaveHostSecret(&storage.HostSecret{
			HostID:     id,
			AuthMethod: data.AuthMethod,
			Nonce:      nonce,
			Ciphertext: ciphertext,
		})
		if err != nil {
			m.statusMessage = fmt.Sprintf("⚠️ Failed to save credentials: %v", err)
			return nil
		}

		m.statusMessage = fmt.Sprintf("✨ Host '%s' (%s) registered! Connecting...", data.Name, data.AuthMethod)
		hosts, _ := m.store.ListHosts()
		return hosts
	}
}

func (m *Model) updateExistingHostCmd(hostID int64, data *HostFormData) tea.Cmd {
	return func() tea.Msg {
		if m.store == nil || m.vault == nil {
			return nil
		}

		payload := &storage.SecretPayload{
			Password:   data.Password,
			PrivateKey: data.KeyContent,
			Passphrase: data.Passphrase,
		}

		rawJSON, err := payload.Encode()
		if err != nil {
			m.statusMessage = fmt.Sprintf("⚠️ Failed to encode secret: %v", err)
			return nil
		}

		nonce, ciphertext, err := m.vault.Encrypt(rawJSON)
		vault.ZeroBytes(rawJSON)
		if err != nil {
			m.statusMessage = fmt.Sprintf("⚠️ Vault encryption failed: %v", err)
			return nil
		}

		h := &storage.Host{
			ID:        hostID,
			Name:      data.Name,
			Address:   data.Address,
			Port:      data.Port,
			Username:  data.Username,
			GroupName: data.Group,
		}

		err = m.store.UpdateHost(h)
		if err != nil {
			m.statusMessage = fmt.Sprintf("⚠️ Failed to update host: %v", err)
			return nil
		}

		err = m.store.SaveHostSecret(&storage.HostSecret{
			HostID:     hostID,
			AuthMethod: data.AuthMethod,
			Nonce:      nonce,
			Ciphertext: ciphertext,
		})
		if err != nil {
			m.statusMessage = fmt.Sprintf("⚠️ Failed to update credentials: %v", err)
			return nil
		}

		// Close existing stale connection in pool
		if m.collector != nil && m.collector.Pool() != nil {
			m.collector.Pool().CloseHost(hostID)
		}

		m.pausedHosts[hostID] = false
		m.hostStatus[hostID] = HostStatusConnecting
		m.statusMessage = fmt.Sprintf("✨ Host '%s' updated! Reconnecting...", data.Name)

		hosts, _ := m.store.ListHosts()
		return hosts
	}
}
