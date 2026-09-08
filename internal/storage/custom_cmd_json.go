package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CustomCommandBackup represents a portable, human-readable custom command definition.
type CustomCommandBackup struct {
	Title       string `json:"title"`
	Command     string `json:"command"`
	Category    string `json:"category,omitempty"`
	Description string `json:"description,omitempty"`
}

// ExportCustomCommandsJSON exports all custom commands into a structured JSON file.
// Uses SetEscapeHTML(false) so shell special characters like &, <, > are not distorted.
func (s *Storage) ExportCustomCommandsJSON(targetPath string) (int, error) {
	cmds, err := s.GetCustomCommands()
	if err != nil {
		return 0, fmt.Errorf("failed to list custom commands: %w", err)
	}

	var backups []CustomCommandBackup
	for _, c := range cmds {
		backups = append(backups, CustomCommandBackup{
			Title:       c.Title,
			Command:     c.Command,
			Category:    c.Category,
			Description: c.Description,
		})
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(backups); err != nil {
		return 0, fmt.Errorf("failed to encode custom commands: %w", err)
	}

	cleanedPath := filepath.Clean(strings.Trim(targetPath, " \"'"))
	if err := os.WriteFile(cleanedPath, buf.Bytes(), 0600); err != nil {
		return 0, fmt.Errorf("failed to write export file: %w", err)
	}

	return len(backups), nil
}

// ImportCustomCommandsJSON imports custom commands from a JSON backup file.
// Non-duplicate commands are added, and duplicates (by title + command) are skipped.
func (s *Storage) ImportCustomCommandsJSON(sourcePath string) (int, int, error) {
	cleanedPath := filepath.Clean(strings.Trim(sourcePath, " \"'"))
	data, err := os.ReadFile(cleanedPath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read import file: %w", err)
	}

	var backups []CustomCommandBackup
	if err := json.Unmarshal(data, &backups); err != nil {
		return 0, 0, fmt.Errorf("invalid runbook JSON format: %w", err)
	}

	existing, err := s.GetCustomCommands()
	if err != nil {
		return 0, 0, err
	}

	existingMap := make(map[string]bool)
	for _, c := range existing {
		key := strings.TrimSpace(c.Title) + "::" + strings.TrimSpace(c.Command)
		existingMap[key] = true
	}

	var imported, skipped int
	for _, b := range backups {
		t := strings.TrimSpace(b.Title)
		cmd := strings.TrimSpace(b.Command)
		if t == "" || cmd == "" {
			continue
		}

		key := t + "::" + cmd
		if existingMap[key] {
			skipped++
			continue
		}

		cat := strings.TrimSpace(b.Category)
		if cat == "" {
			cat = "General"
		}

		if _, err := s.AddCustomCommand(t, cmd, cat, strings.TrimSpace(b.Description)); err != nil {
			return imported, skipped, fmt.Errorf("failed to import command %q: %w", t, err)
		}
		existingMap[key] = true
		imported++
	}

	return imported, skipped, nil
}
