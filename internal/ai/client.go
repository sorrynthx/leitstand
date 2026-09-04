package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client executes streaming chat completion calls to OpenAI-compatible endpoints.
type Client struct {
	Endpoint   string
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

// NewClient creates a new unified AI streaming client.
func NewClient(endpoint, apiKey, model string) *Client {
	if endpoint == "" {
		endpoint = DefaultGroqEndpoint
	}
	if model == "" {
		model = DefaultGroqModel
	}
	return &Client{
		Endpoint: strings.TrimRight(endpoint, "/"),
		APIKey:   apiKey,
		Model:    model,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// StreamChat streams completion chunks asynchronously, executing onChunk for each incoming token
// and invoking onDone with the aggregated full text when the generation finishes.
func (c *Client) StreamChat(ctx context.Context, messages []ChatMessage, onChunk func(string), onDone func(string, error)) {
	go func() {
		targetURL := c.Endpoint
		if !strings.HasSuffix(targetURL, "/chat/completions") {
			targetURL += "/chat/completions"
		}

		reqBody := ChatRequest{
			Model:       c.Model,
			Messages:    messages,
			Stream:      true,
			Temperature: 0.3,
		}

		jsonBytes, err := json.Marshal(reqBody)
		if err != nil {
			onDone("", fmt.Errorf("failed to marshal chat request: %w", err))
			return
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", targetURL, bytes.NewReader(jsonBytes))
		if err != nil {
			onDone("", fmt.Errorf("failed to create http request: %w", err))
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		if c.APIKey != "" {
			httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
		}

		resp, err := c.HTTPClient.Do(httpReq)
		if err != nil {
			onDone("", fmt.Errorf("failed to connect to AI provider: %w", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errBytes, _ := io.ReadAll(resp.Body)
			onDone("", fmt.Errorf("AI provider error (status %d): %s", resp.StatusCode, strings.TrimSpace(string(errBytes))))
			return
		}

		reader := bufio.NewReader(resp.Body)
		var fullBuilder strings.Builder

		for {
			select {
			case <-ctx.Done():
				onDone(fullBuilder.String(), ctx.Err())
				return
			default:
			}

			line, err := reader.ReadString('\n')
			if len(line) > 0 {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "data: ") {
					dataContent := strings.TrimPrefix(line, "data: ")
					if dataContent == "[DONE]" {
						break
					}

					var chunk ChatResponseChunk
					if err := json.Unmarshal([]byte(dataContent), &chunk); err == nil {
						if len(chunk.Choices) > 0 {
							content := chunk.Choices[0].Delta.Content
							if content != "" {
								fullBuilder.WriteString(content)
								if onChunk != nil {
									onChunk(content)
								}
							}
						}
					}
				}
			}

			if err != nil {
				if err == io.EOF {
					break
				}
				onDone(fullBuilder.String(), fmt.Errorf("error reading stream: %w", err))
				return
			}
		}

		onDone(fullBuilder.String(), nil)
	}()
}
