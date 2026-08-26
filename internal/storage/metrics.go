package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// SaveMetric writes a single raw telemetry record.
func (s *Storage) SaveMetric(m *MetricRecord) error {
	ts := m.Timestamp.Unix()
	if ts <= 0 {
		ts = time.Now().Unix()
	}

	query := `
		INSERT INTO metrics_raw (
			host_id, timestamp, cpu_percent, memory_total, memory_used,
			disk_used, disk_total, net_rx_bytes, net_tx_bytes
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(host_id, timestamp) DO UPDATE SET
			cpu_percent = excluded.cpu_percent,
			memory_total = excluded.memory_total,
			memory_used = excluded.memory_used,
			disk_used = excluded.disk_used,
			disk_total = excluded.disk_total,
			net_rx_bytes = excluded.net_rx_bytes,
			net_tx_bytes = excluded.net_tx_bytes;
	`
	_, err := s.db.Exec(
		query,
		m.HostID, ts, m.CPUPercent, m.MemoryTotal, m.MemoryUsed,
		m.DiskUsed, m.DiskTotal, m.NetRxBytes, m.NetTxBytes,
	)
	if err != nil {
		return fmt.Errorf("failed to save metric for host %d: %w", m.HostID, err)
	}
	return nil
}

// GetLatestMetric retrieves the most recent telemetry record for a host.
func (s *Storage) GetLatestMetric(hostID int64) (*MetricRecord, error) {
	query := `
		SELECT host_id, timestamp, cpu_percent, memory_total, memory_used,
		       disk_used, disk_total, net_rx_bytes, net_tx_bytes
		FROM metrics_raw
		WHERE host_id = ?
		ORDER BY timestamp DESC
		LIMIT 1;
	`
	var m MetricRecord
	var ts int64
	err := s.db.QueryRow(query, hostID).Scan(
		&m.HostID, &ts, &m.CPUPercent, &m.MemoryTotal, &m.MemoryUsed,
		&m.DiskUsed, &m.DiskTotal, &m.NetRxBytes, &m.NetTxBytes,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No metrics yet
		}
		return nil, fmt.Errorf("failed to query latest metric for host %d: %w", hostID, err)
	}

	m.Timestamp = time.Unix(ts, 0)
	return &m, nil
}

// GetMetricsRange retrieves metrics for a host within a specific time window.
func (s *Storage) GetMetricsRange(hostID int64, from, to time.Time) ([]*MetricRecord, error) {
	query := `
		SELECT host_id, timestamp, cpu_percent, memory_total, memory_used,
		       disk_used, disk_total, net_rx_bytes, net_tx_bytes
		FROM metrics_raw
		WHERE host_id = ? AND timestamp >= ? AND timestamp <= ?
		ORDER BY timestamp ASC;
	`
	rows, err := s.db.Query(query, hostID, from.Unix(), to.Unix())
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics range: %w", err)
	}
	defer rows.Close()

	var records []*MetricRecord
	for rows.Next() {
		var m MetricRecord
		var ts int64
		if err := rows.Scan(
			&m.HostID, &ts, &m.CPUPercent, &m.MemoryTotal, &m.MemoryUsed,
			&m.DiskUsed, &m.DiskTotal, &m.NetRxBytes, &m.NetTxBytes,
		); err != nil {
			return nil, fmt.Errorf("failed to scan metric row: %w", err)
		}
		m.Timestamp = time.Unix(ts, 0)
		records = append(records, &m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return records, nil
}

// PruneMetricsOlderThan deletes raw metrics older than the retention threshold.
func (s *Storage) PruneMetricsOlderThan(retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		return 0, nil
	}
	cutoff := time.Now().AddDate(0, 0, -retentionDays).Unix()
	res, err := s.db.Exec(`DELETE FROM metrics_raw WHERE timestamp < ?;`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to prune old metrics: %w", err)
	}
	return res.RowsAffected()
}
