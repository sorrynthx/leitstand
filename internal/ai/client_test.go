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

func TestFormatAPIError(t *testing.T) {
	err401 := formatAPIError(http.StatusUnauthorized, []byte(`{"error":{"message":"Invalid API key"}}`), "llama3")
	if err401 == nil || !testingContains(err401.Error(), "401") || !testingContains(err401.Error(), "Invalid API key") {
		t.Errorf("unexpected 401 error format: %v", err401)
	}

	err429 := formatAPIError(http.StatusTooManyRequests, []byte(`{"error":{"message":"Rate limit reached"}}`), "llama3")
	if err429 == nil || !testingContains(err429.Error(), "429") {
		t.Errorf("unexpected 429 error format: %v", err429)
	}

	err404 := formatAPIError(http.StatusNotFound, []byte(`Not found`), "unknown-model")
	if err404 == nil || !testingContains(err404.Error(), "404") || !testingContains(err404.Error(), "unknown-model") {
		t.Errorf("unexpected 404 error format: %v", err404)
	}
}

func TestProviderDefaults(t *testing.T) {
	if ep := GetDefaultEndpoint(ProviderGroq); ep != DefaultGroqEndpoint {
		t.Errorf("expected Groq endpoint %s, got %s", DefaultGroqEndpoint, ep)
	}
	if m := GetDefaultModel(ProviderGroq); m != DefaultGroqModel {
		t.Errorf("expected Groq model %s, got %s", DefaultGroqModel, m)
	}

	if ep := GetDefaultEndpoint(ProviderOllama); ep != DefaultOllamaEndpoint {
		t.Errorf("expected Ollama endpoint %s, got %s", DefaultOllamaEndpoint, ep)
	}
	if m := GetDefaultModel(ProviderOllama); m != DefaultOllamaModel {
		t.Errorf("expected Ollama model %s, got %s", DefaultOllamaModel, m)
	}

	if ep := GetDefaultEndpoint(ProviderOpenAI); ep != DefaultOpenAIEndpoint {
		t.Errorf("expected OpenAI endpoint %s, got %s", DefaultOpenAIEndpoint, ep)
	}
	if m := GetDefaultModel(ProviderOpenAI); m != DefaultOpenAIModel {
		t.Errorf("expected OpenAI model %s, got %s", DefaultOpenAIModel, m)
	}
}

func testingContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && (s[:len(substr)] == substr || testingContains(s[1:], substr))))
}

func TestResolveChatURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://api.groq.com", "https://api.groq.com/openai/v1/chat/completions"},
		{"https://api.groq.com/", "https://api.groq.com/openai/v1/chat/completions"},
		{"https://api.groq.com/v1", "https://api.groq.com/openai/v1/chat/completions"},
		{"https://api.groq.com/openai/v1", "https://api.groq.com/openai/v1/chat/completions"},
		{"https://api.groq.com/openai/v1/chat/completions", "https://api.groq.com/openai/v1/chat/completions"},
		{"http://127.0.0.1:11434", "http://127.0.0.1:11434/v1/chat/completions"},
		{"http://127.0.0.1:11434/v1", "http://127.0.0.1:11434/v1/chat/completions"},
		{"https://api.openai.com", "https://api.openai.com/v1/chat/completions"},
		{"https://api.openai.com/v1", "https://api.openai.com/v1/chat/completions"},
	}

	for _, tc := range tests {
		got := resolveChatURL(tc.input)
		if got != tc.expected {
			t.Errorf("resolveChatURL(%q) = %q; expected %q", tc.input, got, tc.expected)
		}
	}
}
