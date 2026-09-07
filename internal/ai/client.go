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
		targetURL := resolveChatURL(c.Endpoint)

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
			onDone("", formatAPIError(resp.StatusCode, errBytes, c.Model))
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

func formatAPIError(statusCode int, body []byte, model string) error {
	trimmed := strings.TrimSpace(string(body))
	var apiErr struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    any    `json:"code"`
		} `json:"error"`
	}
	msg := trimmed
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error.Message != "" {
		msg = apiErr.Error.Message
	}

	switch statusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("AI 인증 실패 (401): API 키가 올바르지 않거나 만료되었습니다. 설정([p] ➔ AI)을 확인해 주세요. (%s)", msg)
	case http.StatusForbidden:
		return fmt.Errorf("AI 접근 거부 (403): 권한이 없습니다. (%s)", msg)
	case http.StatusNotFound:
		return fmt.Errorf("AI 모델/경로 없음 (404): 지정한 모델(%s) 또는 엔드포인트를 찾을 수 없습니다. (%s)", model, msg)
	case http.StatusTooManyRequests:
		return fmt.Errorf("AI 요청 한도 초과 (429): Rate Limit에 도달했습니다. 잠시 후 다시 시도해 주세요. (%s)", msg)
	default:
		return fmt.Errorf("AI 공급자 오류 (상태코드 %d): %s", statusCode, msg)
	}
}

// resolveChatURL normalizes various endpoint formats (Groq, OpenAI, Ollama)
// into the full /chat/completions URL.
func resolveChatURL(endpoint string) string {
	ep := strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if strings.HasSuffix(ep, "/chat/completions") {
		return ep
	}
	// Groq endpoint normalization
	if strings.Contains(ep, "api.groq.com") {
		if strings.HasSuffix(ep, "/v1") && !strings.HasSuffix(ep, "/openai/v1") {
			ep = strings.TrimSuffix(ep, "/v1") + "/openai/v1"
		} else if !strings.Contains(ep, "/openai/v1") {
			ep = ep + "/openai/v1"
		}
		return ep + "/chat/completions"
	}
	// OpenAI endpoint normalization
	if strings.Contains(ep, "api.openai.com") && !strings.Contains(ep, "/v1") {
		ep = ep + "/v1"
		return ep + "/chat/completions"
	}
	// General: add /v1 if missing
	if !strings.HasSuffix(ep, "/v1") {
		ep = ep + "/v1"
	}
	return ep + "/chat/completions"
}
