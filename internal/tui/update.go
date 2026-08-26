package tui

import (
	"fmt"
	"leitstand/internal/storage"
	"leitstand/internal/telemetry"
	"leitstand/internal/vault"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles incoming messages and updates model state.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.initOrResizeViewport()
		return m, nil

	case tea.KeyMsg:
		if m.showVaultModal && m.vaultForm != nil {
			done, pass, cmd := m.vaultForm.Update(msg)
			if done {
				if pass == "" {
					// User pressed Esc to quit
					m.cancel()
					return m, tea.Quit
				}

				// Attempt Init or Unlock
				isInit, _ := m.store.IsVaultInitialized()
				var err error
				if !isInit {
					err = m.store.InitVault(m.vault, pass)
				} else {
					err = m.store.UnlockVault(m.vault, pass)
				}

				if err != nil {
					m.vaultForm.SetError(err)
					return m, nil
				}

				// Success! Unlock cockpit
				m.showVaultModal = false
				m.statusMessage = "✨ Vault unlocked successfully!"
				return m, tea.Batch(m.loadHostsCmd(), m.pollActiveHostsCmd())
			}
			return m, cmd
		}

		if m.showDeleteModal {
			switch msg.String() {
			case "y", "Y", "enter":
				if m.hostToDelete != nil && m.store != nil {
					_ = m.store.DeleteHost(m.hostToDelete.ID)
					delete(m.metrics, m.hostToDelete.ID)
					delete(m.errors, m.hostToDelete.ID)
					delete(m.consoleLogs, m.hostToDelete.ID)
					m.statusMessage = fmt.Sprintf("🗑️ Host '%s' removed successfully.", m.hostToDelete.Name)
					m.showDeleteModal = false
					m.hostToDelete = nil
					return m, m.loadHostsCmd()
				}
				m.showDeleteModal = false
				return m, nil

			case "n", "N", "esc":
				m.showDeleteModal = false
				m.hostToDelete = nil
				m.statusMessage = "Delete cancelled."
				return m, nil
			}
			return m, nil
		}

		if m.showAddModal && m.addForm != nil {
			done, data, cmd := m.addForm.Update(msg)
			if done {
				m.showAddModal = false
				if data != nil {
					// Save new host
					return m, m.saveNewHostCmd(data)
				}
			}
			return m, cmd
		}

		// Tab to switch between Host Explorer -> Telemetry Deck -> Remote Console
		if msg.String() == "tab" {
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
			return m, nil
		}

		if m.activePane == PaneConsole {
			switch msg.String() {
			case "esc":
				m.activePane = PaneHostList
				m.consoleInput.Blur()
				return m, nil

			case "pgup", "ctrl+u":
				m.viewport.LineUp(6)
				return m, nil

			case "pgdown", "ctrl+d":
				m.viewport.LineDown(6)
				return m, nil

			case "ctrl+l":
				if len(m.hosts) > 0 {
					curHost := m.hosts[m.selectedIndex]
					delete(m.consoleLogs, curHost.ID)
					m.updateViewportContent()
					m.statusMessage = "Console cleared."
				}
				return m, nil

			case "ctrl+o":
				m.fullScreenConsole = !m.fullScreenConsole
				m.initOrResizeViewport()
				return m, nil

			case "enter":
				cmdText := strings.TrimSpace(m.consoleInput.Value())
				if cmdText != "" && len(m.hosts) > 0 {
					m.consoleInput.SetValue("")
					curHost := m.hosts[m.selectedIndex]

					// Built-in clear / cls command support
					if strings.EqualFold(cmdText, "clear") || strings.EqualFold(cmdText, "cls") {
						delete(m.consoleLogs, curHost.ID)
						m.updateViewportContent()
						m.statusMessage = "✨ Console cleared."
						return m, nil
					}

					m.statusMessage = fmt.Sprintf("Running '%s' on %s...", cmdText, curHost.Name)
					return m, m.execRemoteCmd(curHost, cmdText)
				}
				return m, nil

			default:
				var cmd tea.Cmd
				m.consoleInput, cmd = m.consoleInput.Update(msg)
				return m, cmd
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			m.cancel()
			return m, tea.Quit

		case "a", "n":
			m.showAddModal = true
			m.addForm = NewHostForm()
			return m, nil

		case "up", "k":
			if len(m.hosts) > 0 {
				m.selectedIndex--
				if m.selectedIndex < 0 {
					m.selectedIndex = len(m.hosts) - 1
				}
				m.updateViewportContent()
			}
			return m, nil

		case "down", "j":
			if len(m.hosts) > 0 {
				m.selectedIndex++
				if m.selectedIndex >= len(m.hosts) {
					m.selectedIndex = 0
				}
				m.updateViewportContent()
			}
			return m, nil

		case "d", "x":
			if !m.isDemo && len(m.hosts) > 0 {
				m.hostToDelete = m.hosts[m.selectedIndex]
				m.showDeleteModal = true
				return m, nil
			}
			return m, nil

		case "r":
			m.statusMessage = "⏳ Refreshing telemetry..."
			return m, tea.Batch(m.loadHostsCmd(), m.pollActiveHostsCmd())
		}

	case []*storage.Host:
		m.hosts = msg
		if len(m.hosts) > 0 && m.selectedIndex >= len(m.hosts) {
			m.selectedIndex = 0
		}
		// Trigger initial poll
		return m, m.pollActiveHostsCmd()

	case MetricResultMsg:
		if msg.Err != nil {
			m.errors[msg.HostID] = msg.Err
			m.statusMessage = fmt.Sprintf("⚠️ Connection error on host ID %d", msg.HostID)
		} else {
			delete(m.errors, msg.HostID)
			m.metrics[msg.HostID] = msg.Record
			if msg.SysInfo != nil {
				m.sysInfos[msg.HostID] = msg.SysInfo
			}
			m.statusMessage = fmt.Sprintf("✨ Telemetry updated (%s)", time.Now().Format("15:04:05"))
		}
		return m, nil

	case CmdResultMsg:
		if msg.Err != nil {
			m.appendConsoleLog(msg.HostID, fmt.Sprintf("❌ Error: %v\n%s", msg.Err, msg.Stderr))
			m.statusMessage = fmt.Sprintf("⚠️ Command '%s' failed", msg.Command)
		} else {
			m.appendConsoleLog(msg.HostID, fmt.Sprintf("❯ %s\n%s", msg.Command, msg.Stdout))
			m.statusMessage = fmt.Sprintf("✅ Executed '%s'", msg.Command)
		}
		m.updateViewportContent()
		m.viewport.GotoBottom()
		return m, nil

	case TickMsg:
		// Schedule next tick and poll
		return m, tea.Batch(
			m.pollActiveHostsCmd(),
			m.tickCmd(),
		)
	}

	return m, nil
}

func (m *Model) pollActiveHostsCmd() tea.Cmd {
	if m.isDemo {
		return func() tea.Msg {
			m.generateDemoMetrics()
			return nil
		}
	}

	if m.collector == nil || len(m.hosts) == 0 {
		return nil
	}

	var cmds []tea.Cmd
	for _, host := range m.hosts {
		h := host
		cmds = append(cmds, func() tea.Msg {
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
		})
	}
	return tea.Batch(cmds...)
}

func (m *Model) generateDemoMetrics() {
	now := time.Now()
	distros := []string{"Ubuntu 24.04 LTS (Noble)", "Ubuntu 22.04.4 LTS (Jammy)", "Debian 12 (Bookworm)", "Rocky Linux 9.4 (Blue Onyx)", "Amazon Linux 2023"}
	kernels := []string{"Linux 6.8.0-40-generic x86_64", "Linux 6.5.0-35-generic x86_64", "Linux 6.1.0-21-amd64 x86_64", "Linux 5.14.0-427.el9 x86_64", "Linux 6.1.84-99.169.amzn2023.x86_64"}

	for i, h := range m.hosts {
		prev := m.metrics[h.ID]
		cpu := 25.0 + rand.Float64()*45.0
		if prev != nil {
			// Smooth transition
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
			NetRxBytes:  uint64(50000000 + rand.Intn(20000000)),
			NetTxBytes:  uint64(25000000 + rand.Intn(10000000)),
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

		// Encrypt password
		passBytes := []byte(data.Password)
		nonce, ciphertext, err := m.vault.Encrypt(passBytes)
		vault.ZeroBytes(passBytes)
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
			AuthMethod: "password",
			Nonce:      nonce,
			Ciphertext: ciphertext,
		})
		if err != nil {
			m.statusMessage = fmt.Sprintf("⚠️ Failed to save credentials: %v", err)
			return nil
		}

		m.statusMessage = fmt.Sprintf("✨ Host '%s' registered! Connecting...", data.Name)
		hosts, _ := m.store.ListHosts()
		return hosts
	}
}

func (m *Model) execRemoteCmd(host *storage.Host, cmdText string) tea.Cmd {
	return func() tea.Msg {
		// Normalize commands that require TTY in non-interactive mode
		actualCmd := cmdText
		if strings.TrimSpace(cmdText) == "top" {
			actualCmd = "top -b -n 1 | head -n 35"
		}

		if m.isDemo {
			// Realistic demo simulation
			out := simulateDemoCmd(actualCmd, host.Name)
			return CmdResultMsg{
				HostID:  host.ID,
				Command: cmdText,
				Stdout:  out,
			}
		}

		if m.collector == nil || m.collector.Pool() == nil || m.vault == nil {
			return CmdResultMsg{
				HostID:  host.ID,
				Command: cmdText,
				Err:     fmt.Errorf("collector or vault not ready"),
			}
		}

		secret, err := m.store.GetHostSecret(host.ID)
		if err != nil {
			return CmdResultMsg{HostID: host.ID, Command: cmdText, Err: err}
		}

		decrypted, err := m.vault.Decrypt(secret.Nonce, secret.Ciphertext)
		if err != nil {
			return CmdResultMsg{HostID: host.ID, Command: cmdText, Err: err}
		}
		defer vault.ZeroBytes(decrypted)

		client, err := m.collector.Pool().GetOrCreate(host, secret, decrypted, nil)
		if err != nil {
			return CmdResultMsg{HostID: host.ID, Command: cmdText, Err: err}
		}

		stdout, stderr, err := client.Exec(actualCmd)
		return CmdResultMsg{
			HostID:  host.ID,
			Command: cmdText,
			Stdout:  string(stdout),
			Stderr:  string(stderr),
			Err:     err,
		}
	}
}

func (m *Model) appendConsoleLog(hostID int64, logEntry string) {
	logs := m.consoleLogs[hostID]
	logs = append(logs, logEntry)
	// Keep last 50 lines per host
	if len(logs) > 50 {
		logs = logs[len(logs)-50:]
	}
	m.consoleLogs[hostID] = logs
}

func simulateDemoCmd(cmd string, hostname string) string {
	cmdLower := strings.ToLower(cmd)
	if strings.HasPrefix(cmdLower, "uname") {
		return fmt.Sprintf("Linux %s 6.8.0-40-generic #40-Ubuntu SMP PREEMPT_DYNAMIC x86_64 GNU/Linux", hostname)
	}
	if strings.HasPrefix(cmdLower, "docker ps") {
		return "CONTAINER ID   IMAGE                 COMMAND                  CREATED         STATUS         PORTS                    NAMES\n" +
			"a1b2c3d4e5f6   nginx:alpine          \"/docker-entrypoint.…\"   2 days ago      Up 2 days      0.0.0.0:80->80/tcp       web-proxy\n" +
			"9f8e7d6c5b4a   postgres:16-alpine    \"docker-entrypoint.s…\"   3 weeks ago     Up 3 weeks     0.0.0.0:5432->5432/tcp   db-primary"
	}
	if strings.HasPrefix(cmdLower, "free") {
		return "               total        used        free      shared  buff/cache   available\n" +
			"Mem:        16384000     8192000     4096000      256000     4096000     8192000\n" +
			"Swap:        2097152           0     2097152"
	}
	return fmt.Sprintf("[%s]$ %s\nCommand executed successfully in demo simulation.", hostname, cmd)
}

func (m *Model) initOrResizeViewport() {
	vpWidth := m.width - int(float64(m.width)*0.30) - 8
	if vpWidth < 30 {
		vpWidth = 30
	}

	availHeight := m.height - 5
	if availHeight < 10 {
		availHeight = 10
	}

	vpHeight := availHeight - 8 - 4
	if m.fullScreenConsole {
		vpWidth = m.width - 6
		vpHeight = m.height - 7
	}
	if vpHeight < 4 {
		vpHeight = 4
	}

	if !m.viewportReady {
		m.viewport = viewport.New(vpWidth, vpHeight)
		m.viewportReady = true
	} else {
		m.viewport.Width = vpWidth
		m.viewport.Height = vpHeight
	}

	m.updateViewportContent()
}

func (m *Model) updateViewportContent() {
	if len(m.hosts) == 0 {
		m.viewport.SetContent("No host selected.")
		return
	}

	selectedHost := m.hosts[m.selectedIndex]
	logs := m.consoleLogs[selectedHost.ID]

	if len(logs) == 0 {
		welcomeMsg := fmt.Sprintf("Connected to %s (%s)\nType remote commands below and press Enter to execute.\nUse [PageUp/PageDown] or [Ctrl+U/Ctrl+D] to scroll output.", selectedHost.Name, selectedHost.Address)
		m.viewport.SetContent(welcomeMsg)
		return
	}

	content := strings.Join(logs, "\n\n")
	m.viewport.SetContent(content)
}
