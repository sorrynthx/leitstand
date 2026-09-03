package sessionlog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStripANSI(t *testing.T) {
	colored := "\x1b[31mRed Text\x1b[0m and \x1b[1;32mGreen Bold\x1b[0m"
	plain := StripANSI(colored)
	expected := "Red Text and Green Bold"
	if plain != expected {
		t.Errorf("expected %q, got %q", expected, plain)
	}
}

func TestSanitizeFileName(t *testing.T) {
	raw := "prod-server:3306/web*app?"
	safe := SanitizeFileName(raw)
	if strings.ContainsAny(safe, ":/*?") {
		t.Errorf("filename contains invalid characters: %s", safe)
	}
}

func TestSaveSessionLog(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "leitstand_log_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	content := "\x1b[32m[~] ❯ ls -la\x1b[0m\ntotal 0\n"
	savedPath, err := SaveSessionLog(tempDir, "web-server-01", content)
	if err != nil {
		t.Fatalf("SaveSessionLog failed: %v", err)
	}

	if !strings.HasPrefix(filepath.Base(savedPath), "session_web-server-01_") {
		t.Errorf("unexpected filename: %s", savedPath)
	}

	data, err := os.ReadFile(savedPath)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	savedStr := string(data)
	if !strings.Contains(savedStr, "Leitstand Console Session Audit Log") {
		t.Error("expected audit header in log file")
	}
	if !strings.Contains(savedStr, "[~] ❯ ls -la") {
		t.Error("expected cleaned content in log file")
	}
	if strings.Contains(savedStr, "\x1b[32m") {
		t.Error("expected ANSI escape codes to be stripped")
	}
}
