package tui

import (
	"regexp"
	"strings"
)

// AIChatTurn represents a completed user-question + assistant-response pair.
type AIChatTurn struct {
	UserPrompt       string
	AssistantContent string
	ExtractedCommand string
	Explanation      string
	IsDangerous      bool
}

// GetChatTurns groups the linear Messages history into Q&A turns.
func (m *AICopilotModal) GetChatTurns() []AIChatTurn {
	var turns []AIChatTurn
	var currTurn *AIChatTurn

	for _, msg := range m.Messages {
		if msg.Role == "user" {
			if currTurn != nil {
				turns = append(turns, *currTurn)
			}
			currTurn = &AIChatTurn{
				UserPrompt: msg.Content,
			}
		} else if msg.Role == "assistant" && currTurn != nil {
			currTurn.AssistantContent = msg.Content
			currTurn.parseAssistantResponse()
			turns = append(turns, *currTurn)
			currTurn = nil
		}
	}
	if currTurn != nil {
		turns = append(turns, *currTurn)
	}
	return turns
}

func (t *AIChatTurn) parseAssistantResponse() {
	raw := t.AssistantContent
	if matches := codeBlockRegex.FindStringSubmatch(raw); len(matches) > 1 {
		cmd := sanitizeCommand(matches[1])
		if cmd != "" {
			t.ExtractedCommand = cmd
			t.IsDangerous = CheckCommandSafety(cmd)
			t.Explanation = strings.TrimSpace(codeBlockRegex.ReplaceAllString(raw, ""))
			return
		}
	}
	inlineRegex := regexp.MustCompile("`([^`]+)`")
	if matches := inlineRegex.FindStringSubmatch(raw); len(matches) > 1 {
		cmd := sanitizeCommand(matches[1])
		if cmd != "" && !strings.Contains(cmd, "\n") {
			t.ExtractedCommand = cmd
			t.IsDangerous = CheckCommandSafety(cmd)
			t.Explanation = strings.TrimSpace(raw)
			return
		}
	}
	t.Explanation = strings.TrimSpace(raw)
}

// NavigateHistory moves history pointer by delta (-1 = older, +1 = newer).
func (m *AICopilotModal) NavigateHistory(delta int) bool {
	turns := m.GetChatTurns()
	if len(turns) == 0 {
		return false
	}

	newIdx := m.historyNavIndex + delta
	if newIdx < -1 {
		newIdx = -1
	}
	if newIdx >= len(turns) {
		newIdx = len(turns) - 1
	}
	if newIdx == m.historyNavIndex {
		return false
	}

	m.historyNavIndex = newIdx
	if m.historyNavIndex == -1 {
		// Back to draft input
		m.Input.SetValue(m.draftInput)
		m.refreshExtractedCommand()
		return true
	}

	// Inverting index: turns[0] is oldest, turns[len-1] is newest
	turn := turns[len(turns)-1-m.historyNavIndex]
	m.Input.SetValue(turn.UserPrompt)
	m.ExtractedCommand = turn.ExtractedCommand
	m.Explanation = turn.Explanation
	m.IsDangerous = turn.IsDangerous
	m.Input.CursorEnd()
	return true
}
