package tui

import (
	"leitstand/internal/config"
	"leitstand/internal/storage"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAICopilotTelemetryContext(t *testing.T) {
	cfg := config.NewDefaultConfig()
	modal := NewAICopilotModal(1, "test-server", "Ubuntu 24.04", cfg, nil, nil)

	if modal.TelemetrySummary != "" {
		t.Errorf("expected empty telemetry initially, got %q", modal.TelemetrySummary)
	}

	record := &storage.MetricRecord{
		HostID:      1,
		Timestamp:   time.Now(),
		CPUPercent:  42.5,
		MemoryTotal: 8 * 1024 * 1024 * 1024,
		MemoryUsed:  4 * 1024 * 1024 * 1024,
		DiskTotal:   100 * 1024 * 1024 * 1024,
		DiskUsed:    50 * 1024 * 1024 * 1024,
	}

	modal.UpdateTelemetry(record)

	if !strings.Contains(modal.TelemetrySummary, "CPU: 42.5%") {
		t.Errorf("expected CPU 42.5%% in summary, got %q", modal.TelemetrySummary)
	}
	if !strings.Contains(modal.TelemetrySummary, "RAM: 4.0GB/8.0GB") {
		t.Errorf("expected RAM in summary, got %q", modal.TelemetrySummary)
	}
	if !strings.Contains(modal.TelemetrySummary, "Disk: 50.0GB/100.0GB") {
		t.Errorf("expected Disk in summary, got %q", modal.TelemetrySummary)
	}

	prompt := modal.buildSystemPrompt()
	if !strings.Contains(prompt, "Live Server Telemetry:") {
		t.Errorf("expected Live Server Telemetry in prompt, got:\n%s", prompt)
	}
	if !strings.Contains(prompt, "CPU: 42.5%") {
		t.Errorf("expected CPU: 42.5%% in prompt, got:\n%s", prompt)
	}
}

func TestSettingsModalAIProviderPresets(t *testing.T) {
	sm := NewSettingsModal("en", 5*time.Second, 85, 90, 90, "", nil, nil)
	sm.activeTab = TabAI
	sm.focusField = FieldAIProvider

	// 0: groq
	sm.aiProviderIndex = 0
	sm.applyAIProviderPreset("")
	if sm.inputs[4].Value() != "https://api.groq.com/openai/v1" {
		t.Errorf("expected groq endpoint, got %q", sm.inputs[4].Value())
	}
	if sm.inputs[6].Value() != "openai/gpt-oss-20b" {
		t.Errorf("expected groq model, got %q", sm.inputs[6].Value())
	}

	// 1: ollama
	sm.aiProviderIndex = 1
	sm.applyAIProviderPreset("groq")
	if sm.inputs[4].Value() != "http://127.0.0.1:11434/v1" {
		t.Errorf("expected ollama endpoint, got %q", sm.inputs[4].Value())
	}
	if sm.inputs[6].Value() != "llama3" {
		t.Errorf("expected ollama model, got %q", sm.inputs[6].Value())
	}

	// 2: openai
	sm.aiProviderIndex = 2
	sm.applyAIProviderPreset("ollama")
	if sm.inputs[4].Value() != "https://api.openai.com/v1" {
		t.Errorf("expected openai endpoint, got %q", sm.inputs[4].Value())
	}
	if sm.inputs[6].Value() != "gpt-4o-mini" {
		t.Errorf("expected openai model, got %q", sm.inputs[6].Value())
	}
}

func TestSettingsModalAPIKeyPasteNoProviderChange(t *testing.T) {
	sm := NewSettingsModal("en", 5*time.Second, 85, 90, 90, "", nil, nil)
	sm.switchTab(TabAI)
	sm.focusField = FieldAIKey
	sm.focusCurrent()

	// Simulate typing/pasting characters with 'g', 's', 'k', '_', 'l', 'h', 'j'
	for _, r := range "gsk_live_key_123" {
		sm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	if sm.aiProviders[sm.aiProviderIndex] != "groq" {
		t.Errorf("expected provider to stay 'groq', but changed to %q", sm.aiProviders[sm.aiProviderIndex])
	}
	if sm.focusField != FieldAIKey {
		t.Errorf("expected focusField to stay FieldAIKey, got %v", sm.focusField)
	}
	if sm.inputs[5].Value() != "gsk_live_key_123" {
		t.Errorf("expected input value 'gsk_live_key_123', got %q", sm.inputs[5].Value())
	}
}
