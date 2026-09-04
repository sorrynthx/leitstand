package storage

import (
	"fmt"
	"strings"
	"time"
)

// EnsureAIChatTable creates the ai_chat_history table and indexing if not already present.
func (s *Storage) EnsureAIChatTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS ai_chat_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		host_id INTEGER NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_ai_chat_host ON ai_chat_history(host_id, created_at);
	`
	_, err := s.db.Exec(query)
	return err
}

// SaveAIChatMessage inserts a message into the ring buffer, truncating oversized content
// and maintaining the maxHistory sliding-window ceiling for the given host.
func (s *Storage) SaveAIChatMessage(msg *AIChatMessage, maxHistory int) error {
	if err := s.EnsureAIChatTable(); err != nil {
		return err
	}

	// Guard: Truncate single message content to 10KB to prevent runaway DB bloat
	content := msg.Content
	if len(content) > 10240 {
		content = content[:10240] + "\n...(truncated)"
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insertQ := `INSERT INTO ai_chat_history (host_id, role, content, created_at) VALUES (?, ?, ?, ?);`
	t := msg.CreatedAt
	if t.IsZero() {
		t = time.Now()
	}
	res, err := tx.Exec(insertQ, msg.HostID, msg.Role, content, t)
	if err != nil {
		return fmt.Errorf("failed to insert ai chat: %w", err)
	}
	msg.ID, _ = res.LastInsertId()

	// FIFO Ring Buffer: Automatically prune messages beyond maxHistory for this host
	if maxHistory > 0 {
		pruneQ := `
		DELETE FROM ai_chat_history
		WHERE host_id = ? AND id NOT IN (
			SELECT id FROM ai_chat_history
			WHERE host_id = ?
			ORDER BY id DESC
			LIMIT ?
		);`
		if _, err := tx.Exec(pruneQ, msg.HostID, msg.HostID, maxHistory); err != nil {
			return fmt.Errorf("failed to enforce ai chat ring buffer: %w", err)
		}
	}

	return tx.Commit()
}

// GetAIChatHistory retrieves the most recent chat messages for a host in chronological order.
func (s *Storage) GetAIChatHistory(hostID int64, limit int) ([]*AIChatMessage, error) {
	if err := s.EnsureAIChatTable(); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 20
	}

	query := `
	SELECT id, host_id, role, content, created_at
	FROM (
		SELECT id, host_id, role, content, created_at
		FROM ai_chat_history
		WHERE host_id = ?
		ORDER BY id DESC
		LIMIT ?
	)
	ORDER BY id ASC;`

	rows, err := s.db.Query(query, hostID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query ai chat history: %w", err)
	}
	defer rows.Close()

	var messages []*AIChatMessage
	for rows.Next() {
		var m AIChatMessage
		var ct string
		if err := rows.Scan(&m.ID, &m.HostID, &m.Role, &m.Content, &ct); err != nil {
			return nil, err
		}
		m.CreatedAt = parseSQLTime(ct)
		messages = append(messages, &m)
	}
	return messages, nil
}

// PruneAIChatsOlderThan removes all chat messages older than retentionDays.
func (s *Storage) PruneAIChatsOlderThan(days int) (int64, error) {
	if err := s.EnsureAIChatTable(); err != nil {
		return 0, err
	}

	query := `DELETE FROM ai_chat_history WHERE created_at < datetime('now', '-' || ? || ' days');`
	res, err := s.db.Exec(query, days)
	if err != nil {
		return 0, fmt.Errorf("failed to prune old ai chat history: %w", err)
	}
	return res.RowsAffected()
}

// ClearAIChatHistory deletes all chat messages for a specific host.
func (s *Storage) ClearAIChatHistory(hostID int64) error {
	if err := s.EnsureAIChatTable(); err != nil {
		return err
	}

	query := `DELETE FROM ai_chat_history WHERE host_id = ?;`
	_, err := s.db.Exec(query, hostID)
	return err
}

func parseSQLTime(s string) time.Time {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z07:00",
		time.RFC3339,
	}
	s = strings.TrimSpace(s)
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Now()
}
