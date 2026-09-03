package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMaintenance(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_maintenance.db")

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer s.Close()

	// 1. Test GetDBStats on fresh DB
	stats, err := s.GetDBStats()
	if err != nil {
		t.Fatalf("GetDBStats failed: %v", err)
	}
	if stats.HostCount != 0 {
		t.Errorf("Expected 0 hosts, got %d", stats.HostCount)
	}

	// 2. Insert test host and metrics
	h := &Host{
		Name:      "Test-Server",
		Address:   "192.168.1.100",
		Port:      22,
		Username:  "root",
		GroupName: "Default",
	}
	if _, err := s.CreateHost(h); err != nil {
		t.Fatalf("CreateHost failed: %v", err)
	}

	// Add old and recent metrics
	oldTime := time.Now().AddDate(0, 0, -20)
	recentTime := time.Now().AddDate(0, 0, -2)

	_ = s.SaveMetric(&MetricRecord{HostID: h.ID, Timestamp: oldTime, CPUPercent: 50.0})
	_ = s.SaveMetric(&MetricRecord{HostID: h.ID, Timestamp: recentTime, CPUPercent: 25.0})

	stats, _ = s.GetDBStats()
	if stats.MetricCount != 2 {
		t.Errorf("Expected 2 metrics, got %d", stats.MetricCount)
	}

	// 3. Test Vacuum
	if err := s.Vacuum(); err != nil {
		t.Fatalf("Vacuum failed: %v", err)
	}

	// 4. Test ExportMetricsCSV
	csvPath := filepath.Join(tmpDir, "metrics.csv")
	exportedCount, err := s.ExportMetricsCSV(csvPath, 30)
	if err != nil {
		t.Fatalf("ExportMetricsCSV failed: %v", err)
	}
	if exportedCount != 2 {
		t.Errorf("Expected 2 exported rows, got %d", exportedCount)
	}
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		t.Errorf("Expected CSV file to exist")
	}

	// 5. Test PruneAndVacuum (retain 7 days -> 1 old metric deleted)
	deleted, _, _, err := s.PruneAndVacuum(7)
	if err != nil {
		t.Fatalf("PruneAndVacuum failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("Expected 1 pruned row, got %d", deleted)
	}

	stats, _ = s.GetDBStats()
	if stats.MetricCount != 1 {
		t.Errorf("Expected 1 remaining metric, got %d", stats.MetricCount)
	}

	// 6. Test ExportHostsJSON & ImportHostsJSON
	jsonPath := filepath.Join(tmpDir, "hosts_backup.json")
	count, err := s.ExportHostsJSON(jsonPath)
	if err != nil {
		t.Fatalf("ExportHostsJSON failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 host exported, got %d", count)
	}

	// Re-importing into same DB should skip duplicate
	imported, skipped, err := s.ImportHostsJSON(jsonPath)
	if err != nil {
		t.Fatalf("ImportHostsJSON failed: %v", err)
	}
	if imported != 0 || skipped != 1 {
		t.Errorf("Expected 0 imported, 1 skipped, got imported=%d, skipped=%d", imported, skipped)
	}
}
