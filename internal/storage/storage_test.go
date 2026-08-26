package storage

import (
	"path/filepath"
	"testing"
	"time"

	"leitstand/internal/vault"
)

func TestStorageLifecycleAndMigrations(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_leitstand.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer store.Close()

	// 1. Test Host Creation & Retrieval
	host := &Host{
		Name:      "prod-app-01",
		Address:   "192.168.1.100",
		Port:      22,
		Username:  "ubuntu",
		GroupName: "Production",
	}

	id, err := store.CreateHost(host)
	if err != nil {
		t.Fatalf("failed to create host: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected valid host ID, got %d", id)
	}

	fetched, err := store.GetHost(id)
	if err != nil {
		t.Fatalf("failed to get host: %v", err)
	}
	if fetched.Name != "prod-app-01" || fetched.Address != "192.168.1.100" {
		t.Errorf("host data mismatch: %+v", fetched)
	}

	// 2. Test Host Secret Persistence
	secret := &HostSecret{
		HostID:     id,
		AuthMethod: "password",
		Nonce:      []byte("123456789012"), // 12-byte nonce
		Ciphertext: []byte("encrypted_secret_payload"),
	}
	if err := store.SaveHostSecret(secret); err != nil {
		t.Fatalf("failed to save host secret: %v", err)
	}

	fetchedSecret, err := store.GetHostSecret(id)
	if err != nil {
		t.Fatalf("failed to get host secret: %v", err)
	}
	if string(fetchedSecret.Ciphertext) != "encrypted_secret_payload" {
		t.Errorf("secret mismatch: got %s", string(fetchedSecret.Ciphertext))
	}

	// 3. Test Metrics Ingestion & Range Query
	now := time.Now().Truncate(time.Second)
	m1 := &MetricRecord{
		HostID:      id,
		Timestamp:   now.Add(-10 * time.Second),
		CPUPercent:  45.5,
		MemoryTotal: 16 * 1024 * 1024 * 1024,
		MemoryUsed:  8 * 1024 * 1024 * 1024,
		DiskUsed:    50 * 1024 * 1024 * 1024,
		DiskTotal:   100 * 1024 * 1024 * 1024,
		NetRxBytes:  10240,
		NetTxBytes:  20480,
	}
	m2 := &MetricRecord{
		HostID:      id,
		Timestamp:   now,
		CPUPercent:  78.2,
		MemoryTotal: 16 * 1024 * 1024 * 1024,
		MemoryUsed:  10 * 1024 * 1024 * 1024,
		DiskUsed:    50 * 1024 * 1024 * 1024,
		DiskTotal:   100 * 1024 * 1024 * 1024,
		NetRxBytes:  25600,
		NetTxBytes:  40960,
	}

	if err := store.SaveMetric(m1); err != nil {
		t.Fatalf("failed to save m1: %v", err)
	}
	if err := store.SaveMetric(m2); err != nil {
		t.Fatalf("failed to save m2: %v", err)
	}

	latest, err := store.GetLatestMetric(id)
	if err != nil {
		t.Fatalf("failed to get latest metric: %v", err)
	}
	if latest.CPUPercent != 78.2 {
		t.Errorf("expected latest CPU 78.2, got %f", latest.CPUPercent)
	}

	// Range query
	rangeResults, err := store.GetMetricsRange(id, now.Add(-30*time.Second), now.Add(5*time.Second))
	if err != nil {
		t.Fatalf("failed to query metrics range: %v", err)
	}
	if len(rangeResults) != 2 {
		t.Errorf("expected 2 metric records in range, got %d", len(rangeResults))
	}

	// 4. Test Cascade Deletion
	if err := store.DeleteHost(id); err != nil {
		t.Fatalf("failed to delete host: %v", err)
	}

	_, err = store.GetHost(id)
	if err != ErrHostNotFound {
		t.Errorf("expected ErrHostNotFound, got %v", err)
	}

	// Secret and metrics should be deleted via cascade
	_, err = store.GetHostSecret(id)
	if err == nil {
		t.Errorf("expected error when getting deleted host's secret")
	}
}

func TestStorageVaultIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "vault_test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer store.Close()

	// Initial check
	isInit, err := store.IsVaultInitialized()
	if err != nil {
		t.Fatalf("check init failed: %v", err)
	}
	if isInit {
		t.Fatal("expected uninitialized vault")
	}

	// Init vault
	v := vault.New()
	masterPass := "CorrectMasterPass456!"
	if err := store.InitVault(v, masterPass); err != nil {
		t.Fatalf("failed to init vault: %v", err)
	}

	isInit, err = store.IsVaultInitialized()
	if err != nil || !isInit {
		t.Fatal("vault should now be initialized")
	}

	// Try unlock with wrong password
	vWrong := vault.New()
	err = store.UnlockVault(vWrong, "WrongPassword!")
	if err != vault.ErrInvalidPassword {
		t.Errorf("expected ErrInvalidPassword, got: %v", err)
	}
	if vWrong.IsUnlocked() {
		t.Fatal("vault should remain locked on failure")
	}

	// Unlock with correct password
	vCorrect := vault.New()
	err = store.UnlockVault(vCorrect, masterPass)
	if err != nil {
		t.Fatalf("unlock with correct password failed: %v", err)
	}
	if !vCorrect.IsUnlocked() {
		t.Fatal("vault should be unlocked")
	}
}
