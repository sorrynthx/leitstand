package ssh_test

import (
	"leitstand/internal/ssh"
	"leitstand/internal/storage"
	"testing"
	"time"
)

func TestTunnelManager_BasicLifecycle(t *testing.T) {
	pool := ssh.NewPool(5 * time.Second)
	tm := ssh.NewTunnelManager(pool)

	if tm.GetActiveCount() != 0 {
		t.Fatalf("expected 0 active tunnels, got %d", tm.GetActiveCount())
	}
	if tm.GetActiveSummary() != "" {
		t.Fatalf("expected empty summary, got %q", tm.GetActiveSummary())
	}

	tun := &storage.SSHTunnel{
		ID:         1,
		HostID:     10,
		Name:       "Test Web",
		LocalPort:  18080,
		RemoteHost: "127.0.0.1",
		RemotePort: 8080,
	}

	// Starting without a live client should fail safely
	_, err := tm.StartTunnel(tun, nil)
	if err == nil {
		t.Fatalf("expected error starting tunnel with nil client, got nil")
	}

	if tm.IsActive(tun.ID) {
		t.Fatalf("expected tunnel 1 to not be active")
	}

	// CloseAll should not panic
	tm.CloseAll()
}
