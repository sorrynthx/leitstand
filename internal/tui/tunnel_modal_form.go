package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"leitstand/internal/storage"
	"net"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (tm *TunnelModal) handleAddFormKeys(keyStr string) (bool, string, tea.Cmd) {
	switch keyStr {
	case "esc":
		tm.isAdding = false
		tm.errMessage = ""
		return false, "", nil

	case "tab", "down":
		tm.focusedInput = (tm.focusedInput + 1) % len(tm.inputs)
		for i := range tm.inputs {
			if i == tm.focusedInput {
				tm.inputs[i].Focus()
			} else {
				tm.inputs[i].Blur()
			}
		}
		return false, "", nil

	case "shift+tab", "up":
		tm.focusedInput--
		if tm.focusedInput < 0 {
			tm.focusedInput = len(tm.inputs) - 1
		}
		for i := range tm.inputs {
			if i == tm.focusedInput {
				tm.inputs[i].Focus()
			} else {
				tm.inputs[i].Blur()
			}
		}
		return false, "", nil

	case "enter":
		return tm.submitAddForm()
	}

	// Update text inputs
	var cmd tea.Cmd
	tm.inputs[tm.focusedInput], cmd = tm.inputs[tm.focusedInput].Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keyStr)})
	return false, "", cmd
}

func (tm *TunnelModal) submitAddForm() (bool, string, tea.Cmd) {
	name := strings.TrimSpace(tm.inputs[0].Value())
	localPortStr := strings.TrimSpace(tm.inputs[1].Value())
	remoteHost := strings.TrimSpace(tm.inputs[2].Value())
	remotePortStr := strings.TrimSpace(tm.inputs[3].Value())

	if name == "" {
		name = "Tunnel-" + localPortStr
	}
	if remoteHost == "" {
		remoteHost = "127.0.0.1"
	}

	localPort, err := strconv.Atoi(localPortStr)
	if err != nil || localPort < 1 || localPort > 65535 {
		tm.errMessage = i18n.T("tunnel_err_invalid_local_port")
		return false, "", nil
	}

	remotePort, err := strconv.Atoi(remotePortStr)
	if err != nil || remotePort < 1 || remotePort > 65535 {
		tm.errMessage = i18n.T("tunnel_err_invalid_remote_port")
		return false, "", nil
	}

	// Test if local port is available
	testListener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		tm.errMessage = fmt.Sprintf(i18n.T("tunnel_err_port_in_use"), localPort)
		return false, "", nil
	}
	_ = testListener.Close()

	tun := &storage.SSHTunnel{
		HostID:     tm.host.ID,
		Name:       name,
		LocalPort:  localPort,
		RemoteHost: remoteHost,
		RemotePort: remotePort,
	}

	if err := tm.store.SaveTunnel(tun); err != nil {
		tm.errMessage = fmt.Sprintf("⚠️ DB Error: %v", err)
		return false, "", nil
	}

	tm.tunnels, _ = tm.store.GetTunnelsByHost(tm.host.ID)
	tm.selectedIndex = len(tm.tunnels) - 1
	tm.isAdding = false
	tm.errMessage = ""

	return false, fmt.Sprintf(i18n.T("tunnel_added_success"), name), nil
}
