package tui

import (
	"leitstand/internal/i18n"
)

// handleNavigationIMEWarning checks if an unhandled key corresponds to
// Korean IME input or German Umlaut/special characters, and updates the status message.
func (m *Model) handleNavigationIMEWarning(keyStr string) bool {
	for _, r := range keyStr {
		if isHangul(r) {
			m.statusMessage = i18n.T("warn_ime_hangul")
			m.updateViewportContent()
			return true
		}
		if isGermanSpecial(r) {
			m.statusMessage = i18n.T("warn_ime_german")
			m.updateViewportContent()
			return true
		}
	}
	return false
}

func isHangul(r rune) bool {
	return (r >= 0x1100 && r <= 0x11FF) ||
		(r >= 0x3130 && r <= 0x318F) ||
		(r >= 0xAC00 && r <= 0xD7AF)
}

func isGermanSpecial(r rune) bool {
	switch r {
	case 'ä', 'ö', 'ü', 'ß', 'Ä', 'Ö', 'Ü':
		return true
	default:
		return false
	}
}
