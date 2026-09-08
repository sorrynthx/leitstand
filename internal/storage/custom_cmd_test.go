package storage_test

import (
	"database/sql"
	"errors"
	"leitstand/internal/storage"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStorageCustomCommands(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_custom_cmd.db")

	store, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	// 1. Initial list should be empty
	cmds, err := store.GetCustomCommands()
	if err != nil {
		t.Fatalf("failed to get custom commands: %v", err)
	}
	if len(cmds) != 0 {
		t.Fatalf("expected empty commands, got %d", len(cmds))
	}

	// 2. Add first command
	c1, err := store.AddCustomCommand(
		"Docker Logs",
		"docker logs -f --tail 100",
		"Docker",
		"Follow recent container logs",
	)
	if err != nil {
		t.Fatalf("failed to add custom command: %v", err)
	}
	if c1.ID == 0 {
		t.Fatalf("expected non-zero ID, got %d", c1.ID)
	}
	if c1.Title != "Docker Logs" || c1.Command != "docker logs -f --tail 100" {
		t.Fatalf("unexpected custom command fields: %+v", c1)
	}

	// 3. Add second command
	c2, err := store.AddCustomCommand(
		"Nginx Reload",
		"nginx -s reload",
		"Web",
		"Graceful reload nginx config",
	)
	if err != nil {
		t.Fatalf("failed to add second command: %v", err)
	}

	// 4. Query list
	list, err := store.GetCustomCommands()
	if err != nil {
		t.Fatalf("failed to list commands: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(list))
	}

	// 5. Update c1
	err = store.UpdateCustomCommand(
		c1.ID,
		"Docker Live Logs",
		"docker logs -f --tail 200",
		"Docker",
		"Follow 200 lines of container logs",
	)
	if err != nil {
		t.Fatalf("failed to update command: %v", err)
	}

	listAfterUpdate, err := store.GetCustomCommands()
	if err != nil {
		t.Fatalf("failed to list after update: %v", err)
	}
	var updatedC1 *storage.CustomCommand
	for _, cmd := range listAfterUpdate {
		if cmd.ID == c1.ID {
			updatedC1 = cmd
			break
		}
	}
	if updatedC1 == nil || updatedC1.Title != "Docker Live Logs" || updatedC1.Command != "docker logs -f --tail 200" {
		t.Fatalf("update was not persisted correctly: %+v", updatedC1)
	}

	// 6. Delete c2
	err = store.DeleteCustomCommand(c2.ID)
	if err != nil {
		t.Fatalf("failed to delete command: %v", err)
	}

	listAfterDelete, err := store.GetCustomCommands()
	if err != nil {
		t.Fatalf("failed to list after delete: %v", err)
	}
	if len(listAfterDelete) != 1 || listAfterDelete[0].ID != c1.ID {
		t.Fatalf("expected only c1 to remain, got %d", len(listAfterDelete))
	}

	// 7. Delete non-existent
	err = store.DeleteCustomCommand(99999)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected ErrNoRows on non-existent delete, got: %v", err)
	}
}

func TestStorageCustomCommandsExportImport(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_export_import.db")
	jsonPath := filepath.Join(tempDir, "runbook_backup.json")

	store, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	// 1. Add tricky commands with quotes, awk, redirects, and ampersands
	specialCmd := `awk '{print $1, $2}' | grep "foo\"bar" 2>&1 > /dev/null && echo '<tag>&ok'`
	c1, err := store.AddCustomCommand("Special AWK & Redirection", specialCmd, "Script", "Tricky chars test")
	if err != nil {
		t.Fatalf("failed to add special cmd: %v", err)
	}

	c2, err := store.AddCustomCommand("⭐️ 한글과 이모지 테스트", "df -hT -x tmpfs", "System", "Emoji and Korean")
	if err != nil {
		t.Fatalf("failed to add unicode cmd: %v", err)
	}

	// 2. Export to JSON
	exportedCount, err := store.ExportCustomCommandsJSON(jsonPath)
	if err != nil {
		t.Fatalf("ExportCustomCommandsJSON failed: %v", err)
	}
	if exportedCount != 2 {
		t.Fatalf("expected 2 exported commands, got %d", exportedCount)
	}

	// 3. Verify that raw JSON does not escape '<', '>', '&' to unicode entities
	content, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}
	if !strings.Contains(string(content), "<tag>&ok") {
		t.Fatalf("expected raw '<tag>&ok' in JSON, got: %s", string(content))
	}

	// 4. Delete commands from DB
	if err := store.DeleteCustomCommand(c1.ID); err != nil {
		t.Fatalf("delete c1 failed: %v", err)
	}
	if err := store.DeleteCustomCommand(c2.ID); err != nil {
		t.Fatalf("delete c2 failed: %v", err)
	}
	remaining, _ := store.GetCustomCommands()
	if len(remaining) != 0 {
		t.Fatalf("expected 0 commands, got %d", len(remaining))
	}

	// 5. Import back from JSON (Surround path with quotes to test path sanitization)
	quotedPath := ` "` + jsonPath + `" `
	imported, skipped, err := store.ImportCustomCommandsJSON(quotedPath)
	if err != nil {
		t.Fatalf("ImportCustomCommandsJSON failed: %v", err)
	}
	if imported != 2 || skipped != 0 {
		t.Fatalf("expected imported=2, skipped=0, got imported=%d, skipped=%d", imported, skipped)
	}

	// 6. Verify restored contents match exact special characters
	restored, err := store.GetCustomCommands()
	if err != nil || len(restored) != 2 {
		t.Fatalf("expected 2 restored commands, got %d (err: %v)", len(restored), err)
	}
	var foundSpecial bool
	for _, cmd := range restored {
		if cmd.Command == specialCmd {
			foundSpecial = true
			break
		}
	}
	if !foundSpecial {
		t.Fatalf("special command was not restored accurately: %+v", restored)
	}

	// 7. Re-import same file to verify deduplication
	imported2, skipped2, err := store.ImportCustomCommandsJSON(jsonPath)
	if err != nil {
		t.Fatalf("Second ImportCustomCommandsJSON failed: %v", err)
	}
	if imported2 != 0 || skipped2 != 2 {
		t.Fatalf("expected imported=0, skipped=2 on second import, got imported=%d, skipped=%d", imported2, skipped2)
	}
}

