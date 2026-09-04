package ai

// Provider type constants
const (
	ProviderGroq   = "groq"
	ProviderOllama = "ollama"
	ProviderOpenAI = "openai"
	ProviderCustom = "custom"
)

// Default endpoint constants
const (
	DefaultGroqEndpoint   = "https://api.groq.com/openai/v1"
	DefaultGroqModel      = "llama-3.3-70b-versatile"
	DefaultOllamaEndpoint = "http://127.0.0.1:11434/v1"
	DefaultOllamaModel    = "llama3"
	DefaultOpenAIEndpoint = "https://api.openai.com/v1"
	DefaultOpenAIModel    = "gpt-4o-mini"
)

// ChatMessage represents a single message in an LLM conversation.
type ChatMessage struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"` // Text body
}

// ChatRequest represents an OpenAI-compatible /chat/completions request body.
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	Temperature float64       `json:"temperature,omitempty"`
}

// ChatResponseChunk represents a single chunk streamed via Server-Sent Events (SSE).
type ChatResponseChunk struct {
	ID      string `json:"id"`
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}
