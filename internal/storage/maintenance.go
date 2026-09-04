package storage

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

// DBStats holds statistical metadata about the SQLite database.
type DBStats struct {
	Path        string
	SizeBytes   int64
	HostCount   int
	MetricCount int64
}

// GetDBStats returns high-level diagnostic statistics of the local database.
func (s *Storage) GetDBStats() (*DBStats, error) {
	stats := &DBStats{
		Path: s.dbPath,
	}

	if s.dbPath != "" {
		if fi, err := os.Stat(s.dbPath); err == nil {
			stats.SizeBytes = fi.Size()
		}
	}

	_ = s.db.QueryRow("SELECT COUNT(*) FROM hosts;").Scan(&stats.HostCount)
	_ = s.db.QueryRow("SELECT COUNT(*) FROM metrics_raw;").Scan(&stats.MetricCount)

	return stats, nil
}

// Vacuum executes SQLite VACUUM to defragment and reclaim unused space.
func (s *Storage) Vacuum() error {
	_, err := s.db.Exec("VACUUM;")
	if err != nil {
		return fmt.Errorf("failed to vacuum sqlite database: %w", err)
	}
	return nil
}

// PruneAndVacuum deletes metrics and AI chats older than retentionDays and then vacuums the database.
func (s *Storage) PruneAndVacuum(retentionDays int) (int64, int64, int64, error) {
	beforeStats, _ := s.GetDBStats()
	beforeSize := beforeStats.SizeBytes

	deletedMetrics, err := s.PruneMetricsOlderThan(retentionDays)
	if err != nil {
		return 0, beforeSize, beforeSize, err
	}

	deletedChats, _ := s.PruneAIChatsOlderThan(retentionDays)

	if err := s.Vacuum(); err != nil {
		return deletedMetrics + deletedChats, beforeSize, beforeSize, err
	}

	afterStats, _ := s.GetDBStats()
	return deletedMetrics + deletedChats, beforeSize, afterStats.SizeBytes, nil
}

// ExportMetricsCSV exports historical metrics of a given time window (in days) to a CSV file.
func (s *Storage) ExportMetricsCSV(targetPath string, days int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -days).Unix()
	query := `
		SELECT m.host_id, h.name, m.timestamp, m.cpu_percent, m.memory_total, m.memory_used,
		       m.disk_used, m.disk_total, m.net_rx_bytes, m.net_tx_bytes
		FROM metrics_raw m
		LEFT JOIN hosts h ON m.host_id = h.id
		WHERE m.timestamp >= ?
		ORDER BY m.timestamp ASC;
	`
	rows, err := s.db.Query(query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	f, err := os.Create(targetPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create metrics CSV file: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	header := []string{"HostID", "HostName", "Timestamp", "DateTime", "CPUPercent", "MemoryTotal", "MemoryUsed", "DiskUsed", "DiskTotal", "NetRxBytes", "NetTxBytes"}
	if err := w.Write(header); err != nil {
		return 0, err
	}

	var count int64
	for rows.Next() {
		var hostID int64
		var hostName sql.NullString
		var ts int64
		var cpu float64
		var memTot, memUsed, diskUsed, diskTot, netRx, netTx uint64

		if err := rows.Scan(&hostID, &hostName, &ts, &cpu, &memTot, &memUsed, &diskUsed, &diskTot, &netRx, &netTx); err != nil {
			return count, err
		}

		dt := time.Unix(ts, 0).Format("2006-01-02 15:04:05")
		name := hostName.String
		if name == "" {
			name = fmt.Sprintf("Host-%d", hostID)
		}

		record := []string{
			strconv.FormatInt(hostID, 10),
			name,
			strconv.FormatInt(ts, 10),
			dt,
			fmt.Sprintf("%.2f", cpu),
			strconv.FormatUint(memTot, 10),
			strconv.FormatUint(memUsed, 10),
			strconv.FormatUint(diskUsed, 10),
			strconv.FormatUint(diskTot, 10),
			strconv.FormatUint(netRx, 10),
			strconv.FormatUint(netTx, 10),
		}
		if err := w.Write(record); err != nil {
			return count, err
		}
		count++
	}

	return count, nil
}

// HostBackup represents a clean exported host definition.
type HostBackup struct {
	Name      string `json:"name"`
	Address   string `json:"address"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	GroupName string `json:"group_name,omitempty"`
}

// ExportHostsJSON exports all configured hosts into a structured JSON file.
func (s *Storage) ExportHostsJSON(targetPath string) (int, error) {
	hosts, err := s.ListHosts()
	if err != nil {
		return 0, fmt.Errorf("failed to list hosts: %w", err)
	}

	var backups []HostBackup
	for _, h := range hosts {
		backups = append(backups, HostBackup{
			Name:      h.Name,
			Address:   h.Address,
			Port:      h.Port,
			Username:  h.Username,
			GroupName: h.GroupName,
		})
	}

	data, err := json.MarshalIndent(backups, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("failed to encode hosts backup: %w", err)
	}

	if err := os.WriteFile(targetPath, data, 0600); err != nil {
		return 0, fmt.Errorf("failed to write backup file: %w", err)
	}

	return len(backups), nil
}

// ImportHostsJSON reads a JSON backup and imports non-duplicate hosts.
func (s *Storage) ImportHostsJSON(sourcePath string) (int, int, error) {
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read backup file: %w", err)
	}

	var backups []HostBackup
	if err := json.Unmarshal(data, &backups); err != nil {
		return 0, 0, fmt.Errorf("invalid backup JSON format: %w", err)
	}

	existing, err := s.ListHosts()
	if err != nil {
		return 0, 0, err
	}

	existingMap := make(map[string]bool)
	for _, h := range existing {
		key := fmt.Sprintf("%s@%s:%d", h.Username, h.Address, h.Port)
		existingMap[key] = true
	}

	var imported, skipped int
	for _, b := range backups {
		key := fmt.Sprintf("%s@%s:%d", b.Username, b.Address, b.Port)
		if existingMap[key] {
			skipped++
			continue
		}

		h := &Host{
			Name:      b.Name,
			Address:   b.Address,
			Port:      b.Port,
			Username:  b.Username,
			GroupName: b.GroupName,
		}
		if _, err := s.CreateHost(h); err != nil {
			return imported, skipped, fmt.Errorf("failed to import host %s: %w", b.Name, err)
		}
		existingMap[key] = true
		imported++
	}

	return imported, skipped, nil
}
