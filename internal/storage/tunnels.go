package storage

import (
	"database/sql"
	"fmt"
	"time"
)

// EnsureTunnelsTable creates the ssh_tunnels table if it does not exist.
func (s *Storage) EnsureTunnelsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS ssh_tunnels (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		host_id     INTEGER NOT NULL,
		name        TEXT NOT NULL,
		local_port  INTEGER NOT NULL,
		remote_host TEXT NOT NULL DEFAULT '127.0.0.1',
		remote_port INTEGER NOT NULL,
		auto_start  INTEGER NOT NULL DEFAULT 0,
		created_at  INTEGER NOT NULL,
		FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_ssh_tunnels_host_id ON ssh_tunnels(host_id);
	`
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create ssh_tunnels table: %w", err)
	}
	return nil
}

// GetTunnelsByHost retrieves all configured tunnels for a specific host.
func (s *Storage) GetTunnelsByHost(hostID int64) ([]*SSHTunnel, error) {
	if err := s.EnsureTunnelsTable(); err != nil {
		return nil, err
	}

	query := `
	SELECT id, host_id, name, local_port, remote_host, remote_port, auto_start, created_at
	FROM ssh_tunnels
	WHERE host_id = ?
	ORDER BY id ASC;
	`
	rows, err := s.db.Query(query, hostID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tunnels for host %d: %w", hostID, err)
	}
	defer rows.Close()

	return scanTunnels(rows)
}

// GetAllTunnels retrieves all tunnels configured across all hosts.
func (s *Storage) GetAllTunnels() ([]*SSHTunnel, error) {
	if err := s.EnsureTunnelsTable(); err != nil {
		return nil, err
	}

	query := `
	SELECT id, host_id, name, local_port, remote_host, remote_port, auto_start, created_at
	FROM ssh_tunnels
	ORDER BY id ASC;
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all tunnels: %w", err)
	}
	defer rows.Close()

	return scanTunnels(rows)
}

// SaveTunnel inserts a new tunnel or updates an existing one.
func (s *Storage) SaveTunnel(t *SSHTunnel) error {
	if err := s.EnsureTunnelsTable(); err != nil {
		return err
	}

	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	autoStartInt := 0
	if t.AutoStart {
		autoStartInt = 1
	}

	if t.ID == 0 {
		query := `
		INSERT INTO ssh_tunnels (host_id, name, local_port, remote_host, remote_port, auto_start, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?);
		`
		res, err := s.db.Exec(query, t.HostID, t.Name, t.LocalPort, t.RemoteHost, t.RemotePort, autoStartInt, t.CreatedAt.Unix())
		if err != nil {
			return fmt.Errorf("failed to insert tunnel: %w", err)
		}
		id, err := res.LastInsertId()
		if err == nil {
			t.ID = id
		}
		return nil
	}

	query := `
	UPDATE ssh_tunnels
	SET name = ?, local_port = ?, remote_host = ?, remote_port = ?, auto_start = ?
	WHERE id = ?;
	`
	_, err := s.db.Exec(query, t.Name, t.LocalPort, t.RemoteHost, t.RemotePort, autoStartInt, t.ID)
	if err != nil {
		return fmt.Errorf("failed to update tunnel %d: %w", t.ID, err)
	}
	return nil
}

// DeleteTunnel removes a tunnel record by its ID.
func (s *Storage) DeleteTunnel(id int64) error {
	if err := s.EnsureTunnelsTable(); err != nil {
		return err
	}

	query := `DELETE FROM ssh_tunnels WHERE id = ?;`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tunnel %d: %w", id, err)
	}
	return nil
}

func scanTunnels(rows *sql.Rows) ([]*SSHTunnel, error) {
	var tunnels []*SSHTunnel
	for rows.Next() {
		var t SSHTunnel
		var autoStartInt int
		var createdAtUnix int64
		if err := rows.Scan(
			&t.ID,
			&t.HostID,
			&t.Name,
			&t.LocalPort,
			&t.RemoteHost,
			&t.RemotePort,
			&autoStartInt,
			&createdAtUnix,
		); err != nil {
			return nil, fmt.Errorf("failed to scan tunnel: %w", err)
		}
		t.AutoStart = autoStartInt == 1
		t.CreatedAt = time.Unix(createdAtUnix, 0)
		tunnels = append(tunnels, &t)
	}
	return tunnels, rows.Err()
}
