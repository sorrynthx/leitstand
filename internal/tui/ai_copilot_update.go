package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles keyboard inputs and streaming events for the AI Copilot.
// Returns: (closeModal bool, injectedCmd string, runNow bool, cmd tea.Cmd)
func (m *AICopilotModal) Update(msg tea.Msg) (bool, string, bool, tea.Cmd) {
	switch msg := msg.(type) {
	case AICopilotChunkMsg:
		m.StreamingContent += msg.Text
		if msg.ChunkChan != nil {
			cmd := ReadNextAIChunk(msg.ChunkChan, msg.DoneChan)
			return false, "", false, cmd
		}
		return false, "", false, nil

	case AICopilotDoneMsg:
		m.FinalizeStream(msg.Err)
		return false, "", false, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.IsStreaming && m.StreamCancel != nil {
				m.StreamCancel()
				m.FinalizeStream(nil)
				return false, "", false, nil
			}
			return true, "", false, nil

		case tea.KeyTab:
			// Tab: Copy to console input for editing!
			if m.ExtractedCommand != "" && !m.IsStreaming {
				return true, m.ExtractedCommand, false, nil
			}
			return false, "", false, nil

		case tea.KeyCtrlL:
			if m.store != nil {
				_ = m.store.ClearAIChatHistory(m.HostID)
			}
			m.ResetForNewQuery()
			m.StatusMessage = i18n.T("ai_cleared_msg")
			return false, "", false, nil

		case tea.KeyEnter:
			// If we already have an extracted command, Enter means EXECUTE NOW on server!
			if m.ExtractedCommand != "" && !m.IsStreaming {
				if m.IsDangerous {
					m.StatusMessage = i18n.T("ai_blocked_dangerous_run")
					return false, "", false, nil
				}
				injected := m.ExtractedCommand
				m.StatusMessage = fmt.Sprintf(i18n.T("ai_injected_msg"), injected)
				return true, injected, true, nil
			}

			// If user typed a question, submit to start streaming!
			val := strings.TrimSpace(m.Input.Value())
			if val != "" {
				m.Input.SetValue("")
				m.ExtractedCommand = ""
				m.Explanation = ""
				cmd := m.StartStream(val)
				return false, "", false, cmd
			}
			return false, "", false, nil

		default:
			var tiCmd tea.Cmd
			m.Input, tiCmd = m.Input.Update(msg)
			return false, "", false, tiCmd
		}
	}

	return false, "", false, nil
}
