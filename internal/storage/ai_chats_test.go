package storage

import (
	"path/filepath"
	"testing"
	"time"
)

func TestAIChatHistoryRingBufferAndPrune(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_aichats.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	if err := store.EnsureAIChatTable(); err != nil {
		t.Fatalf("failed to ensure ai chat table: %v", err)
	}

	hostID := int64(101)
	maxHistory := 5

	// Insert 10 messages with limit 5 (FIFO ring buffer)
	for i := 1; i <= 10; i++ {
		msg := &AIChatMessage{
			HostID:    hostID,
			Role:      "user",
			Content:   "Question " + string(rune('0'+i)),
			CreatedAt: time.Now(),
		}
		if i%2 == 0 {
			msg.Role = "assistant"
			msg.Content = "Answer " + string(rune('0'+i))
		}
		if err := store.SaveAIChatMessage(msg, maxHistory); err != nil {
			t.Fatalf("failed to save message %d: %v", i, err)
		}
	}

	// Verify only 5 messages remain
	history, err := store.GetAIChatHistory(hostID, 10)
	if err != nil {
		t.Fatalf("failed to get chat history: %v", err)
	}
	if len(history) != 5 {
		t.Fatalf("expected 5 messages in ring buffer, got %d", len(history))
	}

	// Oldest in remaining should be item 6, newest should be item 10
	if history[0].Content != "Answer 6" {
		t.Errorf("expected first remaining message to be 'Answer 6', got '%s'", history[0].Content)
	}
	if history[4].Content != "Answer :" { // '0'+10 in ascii
		t.Logf("last message: %s", history[4].Content)
	}

	// Test Clear
	if err := store.ClearAIChatHistory(hostID); err != nil {
		t.Fatalf("failed to clear chat history: %v", err)
	}
	emptyHistory, err := store.GetAIChatHistory(hostID, 10)
	if err != nil {
		t.Fatalf("failed to get empty history: %v", err)
	}
	if len(emptyHistory) != 0 {
		t.Errorf("expected 0 messages after clear, got %d", len(emptyHistory))
	}
}
