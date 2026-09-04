package tui

import (
	"leitstand/internal/i18n"
	"testing"
)

func TestIsHangul(t *testing.T) {
	tests := []struct {
		char rune
		want bool
	}{
		{'ㄱ', true},
		{'ㅎ', true},
		{'ㅏ', true},
		{'가', true},
		{'한', true},
		{'글', true},
		{'a', false},
		{'z', false},
		{'1', false},
		{'ä', false},
	}

	for _, tt := range tests {
		if got := isHangul(tt.char); got != tt.want {
			t.Errorf("isHangul(%q) = %v, want %v", tt.char, got, tt.want)
		}
	}
}

func TestIsGermanSpecial(t *testing.T) {
	tests := []struct {
		char rune
		want bool
	}{
		{'ä', true},
		{'ö', true},
		{'ü', true},
		{'ß', true},
		{'Ä', true},
		{'Ö', true},
		{'Ü', true},
		{'a', false},
		{'s', false},
		{'한', false},
		{'1', false},
	}

	for _, tt := range tests {
		if got := isGermanSpecial(tt.char); got != tt.want {
			t.Errorf("isGermanSpecial(%q) = %v, want %v", tt.char, got, tt.want)
		}
	}
}

func TestHandleNavigationIMEWarning(t *testing.T) {
	i18n.SetLang(i18n.LangKO)
	m := &Model{}

	// Test Hangul detection
	if !m.handleNavigationIMEWarning("ㅁ") {
		t.Errorf("Expected handleNavigationIMEWarning('ㅁ') to return true")
	}
	if m.statusMessage != i18n.T("warn_ime_hangul") {
		t.Errorf("Expected status message for Hangul IME, got %q", m.statusMessage)
	}

	// Test German Umlaut detection
	if !m.handleNavigationIMEWarning("ä") {
		t.Errorf("Expected handleNavigationIMEWarning('ä') to return true")
	}
	if m.statusMessage != i18n.T("warn_ime_german") {
		t.Errorf("Expected status message for German special key, got %q", m.statusMessage)
	}

	// Test standard ASCII navigation key
	m.statusMessage = ""
	if m.handleNavigationIMEWarning("s") {
		t.Errorf("Expected handleNavigationIMEWarning('s') to return false")
	}
	if m.statusMessage != "" {
		t.Errorf("Expected empty status message for ASCII key, got %q", m.statusMessage)
	}
}
