package tui

import (
	"context"
	"fmt"
	"leitstand/internal/ai"
	"leitstand/internal/config"
	"leitstand/internal/i18n"
	"leitstand/internal/storage"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// AICopilotChunkMsg represents an incoming streaming token.
type AICopilotChunkMsg struct {
	Text      string
	ChunkChan <-chan string
	DoneChan  <-chan error
}

// AICopilotDoneMsg represents completion of streaming or an error.
type AICopilotDoneMsg struct {
	Err error
}

// StartStream begins streaming the response for a user prompt.
func (m *AICopilotModal) StartStream(userPrompt string) tea.Cmd {
	if m.IsStreaming {
		return nil
	}
	m.IsStreaming = true
	m.StreamingContent = ""
	m.StatusMessage = i18n.T("ai_streaming_status")

	// 1. Record User Message
	userMsg := &storage.AIChatMessage{
		HostID:  m.HostID,
		Role:    "user",
		Content: userPrompt,
	}
	if m.store != nil && m.cfg != nil {
		_ = m.store.SaveAIChatMessage(userMsg, m.cfg.AI.MaxHistory)
	}
	m.Messages = append(m.Messages, userMsg)

	// 2. Prepare context messages (limit to recent turns to avoid stale command bleed)
	sysPrompt := m.buildSystemPrompt()
	reqMessages := []ai.ChatMessage{
		{Role: "system", Content: sysPrompt},
	}
	history := m.Messages
	if len(history) > 4 {
		history = history[len(history)-4:]
	}
	for _, hist := range history {
		reqMessages = append(reqMessages, ai.ChatMessage{
			Role:    hist.Role,
			Content: hist.Content,
		})
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.StreamCancel = cancel

	if m.aiClient == nil {
		m.initClient()
	}

	client := m.aiClient
	chunkChan := make(chan string, 100)
	doneChan := make(chan error, 1)

	client.StreamChat(
		ctx,
		reqMessages,
		func(chunk string) {
			chunkChan <- chunk
		},
		func(full string, err error) {
			close(chunkChan)
			doneChan <- err
		},
	)

	return ReadNextAIChunk(chunkChan, doneChan)
}

func (m *AICopilotModal) loadBasePrompt() string {
	promptPath := filepath.Join(config.DefaultDataDir(), "copilot_system_prompt.txt")
	if data, err := os.ReadFile(promptPath); err == nil && len(strings.TrimSpace(string(data))) > 0 {
		return strings.TrimSpace(string(data)) + "\n"
	}

	defaultPrompt := `You are Leitstand AI Terminal Copilot, an expert Linux engineer.
Guidelines:
1. Respond in the language of the user's prompt (Korean if Korean, etc.).
2. Provide a 1-sentence concise explanation of what the command does.
3. Inside the single ` + "```bash" + ` code block, provide ONLY the exact executable terminal command (e.g. ls -la, df -h, systemctl status nginx).
4. NEVER output directory paths, file contents, or command outputs inside the ` + "```bash" + ` block. If asked 'pwd 실행해줘', the command is pwd, NOT the path.
5. Keep the entire response under 3 lines.
CRITICAL SAFETY RESTRICTIONS:
- Strictly REFUSE to provide destructive commands: reboot, shutdown, poweroff, halt, init 0/6, rm -rf /, rm -rf *, mkfs, dd to block devices, iptables -F, systemctl stop ssh.
- NEVER recommend a bare 'rm' without specific target arguments. If user asks which files can be deleted (e.g. '어떤 파일 삭제 가능하니?'), do NOT recommend 'rm'. Recommend safe inspection commands like 'find /tmp -maxdepth 2 -type f -size +10M', 'du -sh * | sort -h', or 'journalctl --disk-usage'.
- If asked for destructive operations, explain politely that dangerous commands cannot be generated for safety. Do NOT provide a ` + "```bash" + ` code block in that case.`

	_ = config.EnsureDataDirExists("")
	_ = os.WriteFile(promptPath, []byte(defaultPrompt+"\n"), 0644)
	return defaultPrompt + "\n"
}

func (m *AICopilotModal) buildSystemPrompt() string {
	var sb strings.Builder
	sb.WriteString(m.loadBasePrompt())
	sb.WriteString(fmt.Sprintf("Target Host: %s\n", m.HostName))
	if m.OSDistro != "" {
		sb.WriteString(fmt.Sprintf("OS / Distro: %s\n", m.OSDistro))
	}
	if m.ActiveTabCWD != "" {
		sb.WriteString(fmt.Sprintf("Working Directory: %s\n", m.ActiveTabCWD))
	}
	if m.LastCommand != "" {
		sb.WriteString(fmt.Sprintf("Last Executed Command: %s\n", m.LastCommand))
	}
	if m.LastExitCode != "" && m.LastExitCode != "0" {
		sb.WriteString(fmt.Sprintf("Last Exit Code: %s\n", m.LastExitCode))
	}
	if m.LastError != "" {
		sb.WriteString(fmt.Sprintf("Last Error Output: %s\n", m.LastError))
	}
	return sb.String()
}

// ReadNextAIChunk waits for the next chunk from the channel.
func ReadNextAIChunk(chunkChan <-chan string, doneChan <-chan error) tea.Cmd {
	return func() tea.Msg {
		chunk, ok := <-chunkChan
		if !ok {
			var err error
			if doneChan != nil {
				err = <-doneChan
			}
			return AICopilotDoneMsg{Err: err}
		}
		return AICopilotChunkMsg{Text: chunk, ChunkChan: chunkChan, DoneChan: doneChan}
	}
}


// FinalizeStream stores the assistant's complete response to database and updates state.
func (m *AICopilotModal) FinalizeStream(err error) {
	m.IsStreaming = false
	m.StreamCancel = nil

	if err != nil {
		m.StatusMessage = fmt.Sprintf("%s: %v", i18n.T("ai_status_error"), err)
	} else {
		m.StatusMessage = ""
	}

	if strings.TrimSpace(m.StreamingContent) != "" {
		asstMsg := &storage.AIChatMessage{
			HostID:  m.HostID,
			Role:    "assistant",
			Content: m.StreamingContent,
		}
		if m.store != nil && m.cfg != nil {
			_ = m.store.SaveAIChatMessage(asstMsg, m.cfg.AI.MaxHistory)
		}
		m.Messages = append(m.Messages, asstMsg)
		m.refreshExtractedCommand()
	}
	m.StreamingContent = ""
}
