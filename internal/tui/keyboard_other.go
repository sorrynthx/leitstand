//go:build !windows

package tui

// IsCapsLockOn returns false on non-Windows platforms as a fallback.
func IsCapsLockOn() bool {
	return false
}
