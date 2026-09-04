package tui

import (
	"regexp"
	"strings"
)

var dangerousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\b(reboot|poweroff|halt|init\s+[06])\b`),
	regexp.MustCompile(`\bshutdown\b`),
	regexp.MustCompile(`\brm\s+-[a-zA-Z]*r[a-zA-Z]*f?\s+([/~]|\/\*|\*|\.\/|\.\.)`),
	regexp.MustCompile(`\brm\s+-[a-zA-Z]*f[a-zA-Z]*r?\s+([/~]|\/\*|\*|\.\/|\.\.)`),
	regexp.MustCompile(`^rm(\s+-[a-zA-Z]+)?$`),
	regexp.MustCompile(`\bmkfs(\.[a-zA-Z0-9]+)?\b`),
	regexp.MustCompile(`\bdd\s+.*of=/dev/(sd|vd|nvme|hd)`),
	regexp.MustCompile(`\b(systemctl|service)\s+(stop|disable)\s+ssh(d)?\b`),
	regexp.MustCompile(`\bchmod\s+-[a-zA-Z]*R\s+(777|000)\s+/`),
	regexp.MustCompile(`\b(wipefs|fdisk|parted|swapoff)\b`),
}

// CheckCommandSafety returns true if the command is considered dangerous/destructive.
func CheckCommandSafety(cmd string) bool {
	lower := strings.ToLower(strings.TrimSpace(cmd))
	for _, p := range dangerousPatterns {
		if p.MatchString(lower) {
			return true
		}
	}
	return false
}
