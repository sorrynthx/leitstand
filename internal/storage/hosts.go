package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrHostNotFound is returned when a requested host record does not exist.
var ErrHostNotFound = errors.New("host not found")

// CreateHost inserts a new host record and sets its generated ID.
func (s *Storage) CreateHost(h *Host) (int64, error) {
	now := time.Now().Unix()
	query := `
		INSERT INTO hosts (name, address, port, username, group_name, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?);
	`
	res, err := s.db.Exec(query, h.Name, h.Address, h.Port, h.Username, h.GroupName, now, now)
	if err != nil {
		return 0, fmt.Errorf("failed to insert host: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get inserted host id: %w", err)
	}

	h.ID = id
	h.CreatedAt = time.Unix(now, 0)
	h.UpdatedAt = time.Unix(now, 0)
	return id, nil
}

// GetHost retrieves a host by its ID.
func (s *Storage) GetHost(id int64) (*Host, error) {
	query := `
		SELECT id, name, address, port, username, group_name, created_at, updated_at
		FROM hosts
		WHERE id = ?;
	`
	var h Host
	var createdAt, updatedAt int64
	err := s.db.QueryRow(query, id).Scan(
		&h.ID, &h.Name, &h.Address, &h.Port, &h.Username, &h.GroupName, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrHostNotFound
		}
		return nil, fmt.Errorf("failed to query host %d: %w", id, err)
	}

	h.CreatedAt = time.Unix(createdAt, 0)
	h.UpdatedAt = time.Unix(updatedAt, 0)
	return &h, nil
}

// ListHosts returns all registered hosts ordered by group and name.
func (s *Storage) ListHosts() ([]*Host, error) {
	query := `
		SELECT id, name, address, port, username, group_name, created_at, updated_at
		FROM hosts
		ORDER BY group_name ASC, name ASC;
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list hosts: %w", err)
	}
	defer rows.Close()

	var hosts []*Host
	for rows.Next() {
		var h Host
		var createdAt, updatedAt int64
		if err := rows.Scan(
			&h.ID, &h.Name, &h.Address, &h.Port, &h.Username, &h.GroupName, &createdAt, &updatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan host row: %w", err)
		}
		h.CreatedAt = time.Unix(createdAt, 0)
		h.UpdatedAt = time.Unix(updatedAt, 0)
		hosts = append(hosts, &h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return hosts, nil
}

// UpdateHost updates host metadata.
func (s *Storage) UpdateHost(h *Host) error {
	now := time.Now().Unix()
	query := `
		UPDATE hosts
		SET name = ?, address = ?, port = ?, username = ?, group_name = ?, updated_at = ?
		WHERE id = ?;
	`
	res, err := s.db.Exec(query, h.Name, h.Address, h.Port, h.Username, h.GroupName, now, h.ID)
	if err != nil {
		return fmt.Errorf("failed to update host %d: %w", h.ID, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrHostNotFound
	}

	h.UpdatedAt = time.Unix(now, 0)
	return nil
}

// DeleteHost deletes a host and cascades deletion to host_secrets and metrics_raw.
func (s *Storage) DeleteHost(id int64) error {
	query := `DELETE FROM hosts WHERE id = ?;`
	res, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete host %d: %w", id, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrHostNotFound
	}

	return nil
}

// SaveHostSecret inserts or replaces encrypted credentials for a host.
func (s *Storage) SaveHostSecret(secret *HostSecret) error {
	query := `
		INSERT INTO host_secrets (host_id, auth_method, nonce, ciphertext)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(host_id) DO UPDATE SET
			auth_method = excluded.auth_method,
			nonce = excluded.nonce,
			ciphertext = excluded.ciphertext;
	`
	_, err := s.db.Exec(query, secret.HostID, secret.AuthMethod, secret.Nonce, secret.Ciphertext)
	if err != nil {
		return fmt.Errorf("failed to save host secret for host %d: %w", secret.HostID, err)
	}
	return nil
}

// GetHostSecret retrieves the encrypted secret record for a host.
func (s *Storage) GetHostSecret(hostID int64) (*HostSecret, error) {
	query := `
		SELECT host_id, auth_method, nonce, ciphertext
		FROM host_secrets
		WHERE host_id = ?;
	`
	var secret HostSecret
	err := s.db.QueryRow(query, hostID).Scan(
		&secret.HostID, &secret.AuthMethod, &secret.Nonce, &secret.Ciphertext,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("host secret not found")
		}
		return nil, fmt.Errorf("failed to query host secret %d: %w", hostID, err)
	}
	return &secret, nil
}
