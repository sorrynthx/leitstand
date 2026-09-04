package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"leitstand/internal/ssh"
	"leitstand/internal/storage"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// TunnelModal manages the interactive SSH Port Forwarding popup.
type TunnelModal struct {
	host             *storage.Host
	tunnels          []*storage.SSHTunnel
	selectedIndex    int
	tunnelMgr        *ssh.TunnelManager
	store            *storage.Storage
	sshPool          *ssh.Pool
	isAdding         bool
	isDeleting       bool
	focusedInput     int
	inputs           []textinput.Model
	errMessage       string
}

// NewTunnelModal creates and initializes the TunnelModal for the given host.
func NewTunnelModal(host *storage.Host, store *storage.Storage, sshPool *ssh.Pool, tm *ssh.TunnelManager) *TunnelModal {
	var tunnels []*storage.SSHTunnel
	if store != nil && host != nil {
		tunnels, _ = store.GetTunnelsByHost(host.ID)
	}

	inputs := make([]textinput.Model, 4)
	placeholders := []string{"n8n Web UI", "15678", "127.0.0.1", "5678"}

	for i := range inputs {
		t := textinput.New()
		t.Placeholder = placeholders[i]
		t.CharLimit = 64
		inputs[i] = t
	}
	inputs[2].SetValue("127.0.0.1")

	return &TunnelModal{
		host:          host,
		tunnels:       tunnels,
		selectedIndex: 0,
		tunnelMgr:     tm,
		store:         store,
		sshPool:       sshPool,
		inputs:        inputs,
	}
}

// Update handles key navigation inside the TunnelModal.
func (tm *TunnelModal) Update(msg tea.Msg) (bool, string, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return false, "", nil
	}
	keyStr := keyMsg.String()

	if tm.isAdding {
		return tm.handleAddFormKeys(keyStr)
	}

	if tm.isDeleting {
		switch keyStr {
		case "y", "Y", "enter":
			return tm.confirmDeleteSelectedTunnel()
		case "n", "N", "esc":
			tm.isDeleting = false
			return false, "", nil
		default:
			return false, "", nil
		}
	}

	switch keyStr {
	case "esc", "q":
		return true, "", nil

	case "up", "k":
		if len(tm.tunnels) > 0 {
			tm.selectedIndex--
			if tm.selectedIndex < 0 {
				tm.selectedIndex = len(tm.tunnels) - 1
			}
		}
		return false, "", nil

	case "down", "j":
		if len(tm.tunnels) > 0 {
			tm.selectedIndex++
			if tm.selectedIndex >= len(tm.tunnels) {
				tm.selectedIndex = 0
			}
		}
		return false, "", nil

	case " ", "enter":
		return tm.toggleSelectedTunnel()

	case "a", "n":
		tm.isAdding = true
		tm.isDeleting = false
		tm.errMessage = ""
		tm.focusedInput = 0
		for i := range tm.inputs {
			if i != 2 {
				tm.inputs[i].SetValue("")
			}
		}
		tm.inputs[0].Focus()
		return false, "", nil

	case "d", "x", "delete":
		if len(tm.tunnels) > 0 && tm.selectedIndex >= 0 && tm.selectedIndex < len(tm.tunnels) {
			tm.isDeleting = true
			tm.errMessage = ""
		}
		return false, "", nil
	}

	return false, "", nil
}

func (tm *TunnelModal) toggleSelectedTunnel() (bool, string, tea.Cmd) {
	if len(tm.tunnels) == 0 || tm.selectedIndex >= len(tm.tunnels) {
		return false, "", nil
	}

	tun := tm.tunnels[tm.selectedIndex]
	if tm.tunnelMgr.IsActive(tun.ID) {
		_ = tm.tunnelMgr.StopTunnel(tun.ID)
		msg := fmt.Sprintf(i18n.T("tunnel_stopped_msg"), tun.Name)
		return false, msg, nil
	}

	client, ok := tm.sshPool.Get(tm.host.ID)
	if !ok || client == nil || !client.IsAlive() {
		tm.errMessage = i18n.T("tunnel_err_not_connected")
		return false, "", nil
	}

	_, err := tm.tunnelMgr.StartTunnel(tun, client)
	if err != nil {
		tm.errMessage = fmt.Sprintf("⚠️ %v", err)
		return false, "", nil
	}

	tm.errMessage = ""
	msg := fmt.Sprintf(i18n.T("tunnel_started_msg"), tun.Name, tun.LocalPort, tun.RemoteHost, tun.RemotePort)
	return false, msg, nil
}

func (tm *TunnelModal) confirmDeleteSelectedTunnel() (bool, string, tea.Cmd) {
	tm.isDeleting = false
	if len(tm.tunnels) == 0 || tm.selectedIndex >= len(tm.tunnels) {
		return false, "", nil
	}

	tun := tm.tunnels[tm.selectedIndex]
	if tm.tunnelMgr.IsActive(tun.ID) {
		_ = tm.tunnelMgr.StopTunnel(tun.ID)
	}

	_ = tm.store.DeleteTunnel(tun.ID)
	tm.tunnels, _ = tm.store.GetTunnelsByHost(tm.host.ID)
	if tm.selectedIndex >= len(tm.tunnels) && tm.selectedIndex > 0 {
		tm.selectedIndex--
	}

	msg := fmt.Sprintf(i18n.T("tunnel_deleted_msg"), tun.Name)
	return false, msg, nil
}
