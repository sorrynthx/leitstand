package i18n

import (
	"testing"
)

func TestI18n(t *testing.T) {
	SetLang(LangKO)
	if T("btn_save") != "저장" {
		t.Fatalf("Expected '저장', got '%s'", T("btn_save"))
	}

	SetLang(LangEN)
	if T("btn_save") != "Save" {
		t.Fatalf("Expected 'Save', got '%s'", T("btn_save"))
	}

	SetLang(LangDE)
	if T("btn_save") != "Speichern" {
		t.Fatalf("Expected 'Speichern', got '%s'", T("btn_save"))
	}

	// Test Fallback
	if T("non_existent_key_xyz") != "non_existent_key_xyz" {
		t.Fatalf("Expected raw key fallback")
	}

	// Test Formatting
	SetLang(LangKO)
	formatted := Tf("modal_delete_warn", "api-server", "10.0.0.1")
	if formatted == "" {
		t.Fatalf("Expected formatted string")
	}
}
