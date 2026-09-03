package sessionlog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\([a-zA-Z]|\x1b\][^\x07\x1b]*(\x07|\x1b\\)`)

// StripANSI strips all ANSI escape codes from string for clean, readable plain-text logging.
func StripANSI(str string) string {
	return ansiRegex.ReplaceAllString(str, "")
}

// SanitizeFileName ensures hostname can be used as a safe filename component across all OSes.
func SanitizeFileName(name string) string {
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			return r
		}
		return '_'
	}, name)
	if safe == "" {
		return "session"
	}
	return safe
}

// SaveSessionLog exports the raw terminal viewport/log content to a clean, timestamped file.
func SaveSessionLog(dir, hostName, content string) (string, error) {
	if dir == "" {
		dir = filepath.Join(".", "logs")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create session log directory %s: %w", dir, err)
	}

	safeHost := SanitizeFileName(hostName)
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("session_%s_%s.log", safeHost, timestamp)
	fullPath := filepath.Join(dir, fileName)

	cleanContent := StripANSI(content)

	header := fmt.Sprintf(
		"================================================================================\n"+
			" Leitstand Console Session Audit Log\n"+
			" Host: %s\n"+
			" Exported At: %s\n"+
			"================================================================================\n\n",
		hostName,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	payload := header + cleanContent + "\n"

	if err := os.WriteFile(fullPath, []byte(payload), 0644); err != nil {
		return "", fmt.Errorf("failed to write session log file: %w", err)
	}

	return fullPath, nil
}
