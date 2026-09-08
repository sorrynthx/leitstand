package tui

import (
	"context"
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

func TestAICopilotSaveToRunbook(t *testing.T) {
	tempDB := t.TempDir() + "/test_leitstand.db"
	store, err := storage.Open(tempDB)
	if err != nil {
		t.Fatalf("failed to create test storage: %v", err)
	}
	defer store.Close()

	cfg := config.NewDefaultConfig()
	modal := NewAICopilotModal(1, "test-server", "Ubuntu 24.04", cfg, store, nil)
	modal.Messages = []*storage.AIChatMessage{
		{HostID: 1, Role: "user", Content: "How do I check memory usage?"},
		{HostID: 1, Role: "assistant", Content: "Use free command:\n```bash\nfree -h\n```"},
	}
	modal.refreshExtractedCommand()

	if modal.ExtractedCommand != "free -h" {
		t.Fatalf("expected extracted command 'free -h', got %q", modal.ExtractedCommand)
	}

	// Trigger Ctrl+S key to save into runbook
	closeModal, injectedCmd, runNow, _ := modal.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if closeModal || injectedCmd != "" || runNow {
		t.Errorf("expected modal to stay open on save, got closeModal=%v, injectedCmd=%q", closeModal, injectedCmd)
	}

	cmds, err := store.GetCustomCommands()
	if err != nil {
		t.Fatalf("failed to query custom commands: %v", err)
	}
	if len(cmds) != 1 {
		t.Fatalf("expected 1 custom command saved, got %d", len(cmds))
	}
	if cmds[0].Command != "free -h" {
		t.Errorf("expected command 'free -h', got %q", cmds[0].Command)
	}
	if cmds[0].Category != "AI Copilot" {
		t.Errorf("expected category 'AI Copilot', got %q", cmds[0].Category)
	}
	if cmds[0].Title != "How do I check memory usage?" {
		t.Errorf("expected title from user query, got %q", cmds[0].Title)
	}

	// Trigger Ctrl+S again -> should NOT duplicate
	modal.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	cmdsAfter, _ := store.GetCustomCommands()
	if len(cmdsAfter) != 1 {
		t.Fatalf("expected command not to be duplicated, got %d", len(cmdsAfter))
	}
	if !strings.Contains(modal.StatusMessage, "already saved") && !strings.Contains(modal.StatusMessage, "이미") {
		t.Errorf("expected already in runbook status message, got %q", modal.StatusMessage)
	}
}

func TestQuitConfirmationModal(t *testing.T) {
	m := &Model{
		activePane: PaneHostList,
	}

	// Press 'q' -> should open quit confirmation modal instead of quitting
	m.updateHostListNavigation("q")
	if !m.showQuitModal {
		t.Errorf("expected showQuitModal to be true when 'q' is pressed")
	}

	// Press 'n' -> should cancel quit confirmation modal
	_, _, handled := m.updateActiveModals(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if !handled || m.showQuitModal {
		t.Errorf("expected quit modal to be closed on 'n'")
	}

	// Press 'ctrl+c' -> should open quit confirmation modal
	m.updateHostListNavigation("ctrl+c")
	if !m.showQuitModal {
		t.Errorf("expected showQuitModal to be true when 'ctrl+c' is pressed")
	}

	// Press 'enter' -> should quit
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	_ = ctx
	_, cmd, handled := m.updateActiveModals(tea.KeyMsg{Type: tea.KeyEnter})
	if !handled || cmd == nil {
		t.Errorf("expected enter to be handled with tea.Quit cmd")
	}
}
