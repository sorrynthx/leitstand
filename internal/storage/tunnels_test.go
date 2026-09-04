package storage_test

import (
	"leitstand/internal/storage"
	"path/filepath"
	"testing"
	"time"
)

func TestStorageTunnels(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_tunnels.db")

	store, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	host := &storage.Host{
		Name:      "TestHost",
		Address:   "192.168.14.119",
		Port:      22,
		Username:  "testuser",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if _, err := store.CreateHost(host); err != nil {
		t.Fatalf("failed to save host: %v", err)
	}

	// 1. Save tunnel
	tun := &storage.SSHTunnel{
		HostID:     host.ID,
		Name:       "n8n Web",
		LocalPort:  15678,
		RemoteHost: "127.0.0.1",
		RemotePort: 5678,
		AutoStart:  true,
	}
	if err := store.SaveTunnel(tun); err != nil {
		t.Fatalf("failed to save tunnel: %v", err)
	}
	if tun.ID == 0 {
		t.Fatalf("expected non-zero tunnel ID after insert")
	}

	// 2. Query tunnels by host
	tuns, err := store.GetTunnelsByHost(host.ID)
	if err != nil {
		t.Fatalf("failed to get tunnels by host: %v", err)
	}
	if len(tuns) != 1 {
		t.Fatalf("expected 1 tunnel, got %d", len(tuns))
	}
	if tuns[0].Name != "n8n Web" || tuns[0].LocalPort != 15678 || !tuns[0].AutoStart {
		t.Fatalf("unexpected tunnel data: %+v", tuns[0])
	}

	// 3. Update tunnel
	tun.Name = "n8n Workflow"
	tun.AutoStart = false
	if err := store.SaveTunnel(tun); err != nil {
		t.Fatalf("failed to update tunnel: %v", err)
	}

	allTuns, err := store.GetAllTunnels()
	if err != nil {
		t.Fatalf("failed to get all tunnels: %v", err)
	}
	if len(allTuns) != 1 || allTuns[0].Name != "n8n Workflow" || allTuns[0].AutoStart {
		t.Fatalf("unexpected allTuns data: %+v", allTuns[0])
	}

	// 4. Delete tunnel
	if err := store.DeleteTunnel(tun.ID); err != nil {
		t.Fatalf("failed to delete tunnel: %v", err)
	}
	tunsAfter, err := store.GetTunnelsByHost(host.ID)
	if err != nil {
		t.Fatalf("failed to get tunnels after delete: %v", err)
	}
	if len(tunsAfter) != 0 {
		t.Fatalf("expected 0 tunnels after delete, got %d", len(tunsAfter))
	}
}
