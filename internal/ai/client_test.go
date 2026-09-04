package ai

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestStreamChatSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected flusher")
		}

		chunks := []string{"Hello", " ", "from", " ", "AI!"}
		for _, c := range chunks {
			fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":%q}}]}\n\n", c)
			flusher.Flush()
			time.Sleep(5 * time.Millisecond)
		}
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key", "test-model")
	var receivedChunks []string
	var mu sync.Mutex
	doneChan := make(chan string)
	errChan := make(chan error)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client.StreamChat(
		ctx,
		[]ChatMessage{{Role: "user", Content: "Hi"}},
		func(chunk string) {
			mu.Lock()
			receivedChunks = append(receivedChunks, chunk)
			mu.Unlock()
		},
		func(full string, err error) {
			if err != nil {
				errChan <- err
			} else {
				doneChan <- full
			}
		},
	)

	select {
	case fullText := <-doneChan:
		expected := "Hello from AI!"
		if fullText != expected {
			t.Errorf("expected fullText '%s', got '%s'", expected, fullText)
		}
		mu.Lock()
		if len(receivedChunks) != 5 {
			t.Errorf("expected 5 chunks, got %d", len(receivedChunks))
		}
		mu.Unlock()
	case err := <-errChan:
		t.Fatalf("unexpected error: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for stream to finish")
	}
}
