package ssh

import (
	"fmt"
	"leitstand/internal/storage"
	"strings"
	"sync"
)

// TunnelManager coordinates live port forwarding tunnels across all hosts.
type TunnelManager struct {
	mu      sync.RWMutex
	tunnels map[int64]*ActiveTunnel
	pool    *Pool
}

// NewTunnelManager creates a new TunnelManager.
func NewTunnelManager(pool *Pool) *TunnelManager {
	return &TunnelManager{
		tunnels: make(map[int64]*ActiveTunnel),
		pool:    pool,
	}
}

// StartTunnel initiates an active port forward for the given SSHTunnel definition.
func (tm *TunnelManager) StartTunnel(tun *storage.SSHTunnel, client *Client) (*ActiveTunnel, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if existing, ok := tm.tunnels[tun.ID]; ok {
		_ = existing.Stop()
		delete(tm.tunnels, tun.ID)
	}

	at, err := NewActiveTunnel(tun, client)
	if err != nil {
		return nil, err
	}

	tm.tunnels[tun.ID] = at
	return at, nil
}

// StopTunnel terminates an active tunnel by its storage ID.
func (tm *TunnelManager) StopTunnel(tunnelID int64) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	at, ok := tm.tunnels[tunnelID]
	if !ok {
		return nil
	}

	err := at.Stop()
	delete(tm.tunnels, tunnelID)
	return err
}

// IsActive returns true if the tunnel ID is currently forwarding.
func (tm *TunnelManager) IsActive(tunnelID int64) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	_, ok := tm.tunnels[tunnelID]
	return ok
}

// GetActive retrieves the ActiveTunnel instance if running.
func (tm *TunnelManager) GetActive(tunnelID int64) (*ActiveTunnel, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	at, ok := tm.tunnels[tunnelID]
	return at, ok
}

// GetActiveCount returns the total number of running tunnels.
func (tm *TunnelManager) GetActiveCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.tunnels)
}

// GetActiveSummary returns a concise string representation of active tunnels (e.g., "15678➔5678").
func (tm *TunnelManager) GetActiveSummary() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if len(tm.tunnels) == 0 {
		return ""
	}

	if len(tm.tunnels) == 1 {
		for _, at := range tm.tunnels {
			return fmt.Sprintf("%d➔%d", at.Tunnel.LocalPort, at.Tunnel.RemotePort)
		}
	}

	var parts []string
	for _, at := range tm.tunnels {
		parts = append(parts, fmt.Sprintf("%d", at.Tunnel.LocalPort))
		if len(parts) >= 2 {
			break
		}
	}
	if len(tm.tunnels) > 2 {
		return fmt.Sprintf("%s,+%d", strings.Join(parts, ","), len(tm.tunnels)-2)
	}
	return strings.Join(parts, ",")
}

// CloseAll terminates all active tunnels cleanly.
func (tm *TunnelManager) CloseAll() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for id, at := range tm.tunnels {
		_ = at.Stop()
		delete(tm.tunnels, id)
	}
}
