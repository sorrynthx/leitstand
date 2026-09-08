package storage

import (
	"database/sql"
	"fmt"
	"time"
)

// EnsureCustomCommandsTable creates the custom_commands table if it does not exist.
func (s *Storage) EnsureCustomCommandsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS custom_commands (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		title       TEXT NOT NULL,
		command     TEXT NOT NULL,
		category    TEXT NOT NULL DEFAULT '',
		description TEXT NOT NULL DEFAULT '',
		created_at  INTEGER NOT NULL,
		updated_at  INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_custom_commands_category ON custom_commands(category);
	`
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create custom_commands table: %w", err)
	}
	return nil
}

// GetCustomCommands retrieves all user-defined custom commands.
func (s *Storage) GetCustomCommands() ([]*CustomCommand, error) {
	if err := s.EnsureCustomCommandsTable(); err != nil {
		return nil, err
	}

	query := `
	SELECT id, title, command, category, description, created_at, updated_at
	FROM custom_commands
	ORDER BY category ASC, id ASC;
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query custom commands: %w", err)
	}
	defer rows.Close()

	var cmds []*CustomCommand
	for rows.Next() {
		var (
			c         CustomCommand
			createdAt int64
			updatedAt int64
		)
		if err := rows.Scan(&c.ID, &c.Title, &c.Command, &c.Category, &c.Description, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan custom command: %w", err)
		}
		c.CreatedAt = time.Unix(createdAt, 0)
		c.UpdatedAt = time.Unix(updatedAt, 0)
		cmds = append(cmds, &c)
	}
	return cmds, rows.Err()
}

// AddCustomCommand inserts a new custom command into the database.
func (s *Storage) AddCustomCommand(title, command, category, description string) (*CustomCommand, error) {
	if err := s.EnsureCustomCommandsTable(); err != nil {
		return nil, err
	}

	now := time.Now()
	nowUnix := now.Unix()

	query := `
	INSERT INTO custom_commands (title, command, category, description, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?);
	`
	res, err := s.db.Exec(query, title, command, category, description, nowUnix, nowUnix)
	if err != nil {
		return nil, fmt.Errorf("failed to insert custom command: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return &CustomCommand{
		ID:          id,
		Title:       title,
		Command:     command,
		Category:    category,
		Description: description,
		CreatedAt:   time.Unix(nowUnix, 0),
		UpdatedAt:   time.Unix(nowUnix, 0),
	}, nil
}

// UpdateCustomCommand updates an existing custom command.
func (s *Storage) UpdateCustomCommand(id int64, title, command, category, description string) error {
	if err := s.EnsureCustomCommandsTable(); err != nil {
		return err
	}

	nowUnix := time.Now().Unix()
	query := `
	UPDATE custom_commands
	SET title = ?, command = ?, category = ?, description = ?, updated_at = ?
	WHERE id = ?;
	`
	res, err := s.db.Exec(query, title, command, category, description, nowUnix, id)
	if err != nil {
		return fmt.Errorf("failed to update custom command %d: %w", id, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteCustomCommand removes a custom command by ID.
func (s *Storage) DeleteCustomCommand(id int64) error {
	if err := s.EnsureCustomCommandsTable(); err != nil {
		return err
	}

	query := `DELETE FROM custom_commands WHERE id = ?;`
	res, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete custom command %d: %w", id, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
