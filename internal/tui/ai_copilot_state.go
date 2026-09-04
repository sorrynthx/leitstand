package tui

import (
	"context"
	"fmt"
	"leitstand/internal/ai"
	"leitstand/internal/config"
	"leitstand/internal/storage"
	"leitstand/internal/vault"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

var codeBlockRegex = regexp.MustCompile("(?s)```(?:bash|sh|shell)?\n?(.*?)```")

// AICopilotModal manages the state, context, and conversation history of the AI assistant.
type AICopilotModal struct {
	HostID           int64
	HostName         string
	OSDistro         string
	ActiveTabCWD     string
	LastCommand      string
	LastExitCode     string
	LastError        string
	Messages         []*storage.AIChatMessage
	Input            textinput.Model
	Viewport         viewport.Model
	IsStreaming      bool
	StreamingContent string
	StreamCancel     context.CancelFunc
	ExtractedCommand string
	Explanation      string
	IsDangerous      bool
	StatusMessage    string
	cfg              *config.AppConfig
	store            *storage.Storage
	vault            *vault.Vault
	aiClient         *ai.Client
}

// NewAICopilotModal creates a new AI Copilot modal for a specific host.
func NewAICopilotModal(hostID int64, hostName, distro string, cfg *config.AppConfig, store *storage.Storage, v *vault.Vault) *AICopilotModal {
	ti := textinput.New()
	ti.Prompt = "🤖 ❯ "
	ti.Focus()

	vp := viewport.New(60, 15)

	m := &AICopilotModal{
		HostID:   hostID,
		HostName: hostName,
		OSDistro: distro,
		cfg:      cfg,
		store:    store,
		vault:    v,
		Input:    ti,
		Viewport: vp,
	}


	m.initClient()
	m.LoadHistory()
	return m
}

func (m *AICopilotModal) initClient() {
	if m.cfg == nil {
		return
	}
	endpoint := m.cfg.AI.Endpoint
	model := m.cfg.AI.Model
	var apiKey string

	if m.store != nil {
		if ep, err := m.store.GetSetting("ai_endpoint"); err == nil && ep != "" {
			endpoint = ep
			m.cfg.AI.Endpoint = ep
		}
		if md, err := m.store.GetSetting("ai_model"); err == nil && md != "" {
			model = md
			m.cfg.AI.Model = md
		}
		if rawKey, err := m.store.GetSetting("ai_api_key"); err == nil && rawKey != "" {
			apiKey = rawKey
		}
	}

	m.aiClient = ai.NewClient(endpoint, apiKey, model)
}

// UpdateHostContext updates the live terminal context from the active console tab.
func (m *AICopilotModal) UpdateHostContext(tab *ConsoleTab, distro string) {
	if distro != "" {
		m.OSDistro = distro
	}
	if tab != nil {
		m.ActiveTabCWD = tab.CWD
		if tab.LastCommand != "" {
			m.LastCommand = tab.LastCommand
		}
		if tab.LastExitCode != 0 {
			m.LastExitCode = fmt.Sprintf("%d", tab.LastExitCode)
		} else {
			m.LastExitCode = "0"
		}
		m.LastError = tab.LastError
	}
}

// LoadHistory loads recent conversation history within the configured maxHistory ceiling.
func (m *AICopilotModal) LoadHistory() {
	if m.store == nil || m.cfg == nil || m.cfg.AI.RetentionDays <= 0 {
		return
	}
	maxH := m.cfg.AI.MaxHistory
	if maxH <= 0 {
		maxH = 20
	}
	hist, err := m.store.GetAIChatHistory(m.HostID, maxH)
	if err == nil {
		m.Messages = hist
		m.refreshExtractedCommand()
	}
}

// ResetForNewQuery clears current input and recommendation for a fresh query.
func (m *AICopilotModal) ResetForNewQuery() {
	m.Input.SetValue("")
	m.ExtractedCommand = ""
	m.Explanation = ""
	m.IsDangerous = false
	m.StatusMessage = ""
	m.IsStreaming = false
	m.StreamingContent = ""
	m.Input.Focus()
}

func sanitizeCommand(cmd string) string {
	cmd = strings.TrimSpace(cmd)
	if (strings.HasPrefix(cmd, "/home/") || strings.HasPrefix(cmd, "/root/") || strings.HasPrefix(cmd, "~/")) && !strings.Contains(cmd, " ") {
		return "pwd"
	}
	return cmd
}

func (m *AICopilotModal) refreshExtractedCommand() {
	m.ExtractedCommand = ""
	m.Explanation = ""
	m.IsDangerous = false

	if len(m.Messages) == 0 {
		return
	}
	lastMsg := m.Messages[len(m.Messages)-1]
	if lastMsg.Role != "assistant" {
		return
	}

	if matches := codeBlockRegex.FindStringSubmatch(lastMsg.Content); len(matches) > 1 {
		cmd := sanitizeCommand(matches[1])
		if cmd != "" {
			m.ExtractedCommand = cmd
			m.IsDangerous = CheckCommandSafety(cmd)
			clean := codeBlockRegex.ReplaceAllString(lastMsg.Content, "")
			clean = strings.TrimSpace(strings.ReplaceAll(clean, "\n", " "))
			m.Explanation = clean
			return
		}
	}
	inlineRegex := regexp.MustCompile("`([^`]+)`")
	if matches := inlineRegex.FindStringSubmatch(lastMsg.Content); len(matches) > 1 {
		cmd := sanitizeCommand(matches[1])
		if cmd != "" && !strings.Contains(cmd, "\n") {
			m.ExtractedCommand = cmd
			m.IsDangerous = CheckCommandSafety(cmd)
			clean := inlineRegex.ReplaceAllString(lastMsg.Content, "")
			clean = strings.TrimSpace(strings.ReplaceAll(clean, "\n", " "))
			m.Explanation = clean
			return
		}
	}

	// No executable code block in latest response (e.g. safety refusal or pure explanation)
	clean := strings.TrimSpace(strings.ReplaceAll(lastMsg.Content, "\n", " "))
	m.Explanation = clean
}
