package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// SpinnerFrames defines the Braille dots spinner sequence for loading feedback.
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// ExecTickMsg is triggered every 100ms while a remote command is executing.
type ExecTickMsg struct {
	HostID int64
	TabID  string
}

// tickExecCommand schedules the next 100ms tick for a running command.
func tickExecCommand(hostID int64, tabID string) tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return ExecTickMsg{
			HostID: hostID,
			TabID:  tabID,
		}
	})
}

// handleExecTickMessage updates the spinner frame and refreshes the console viewport.
func (m *Model) handleExecTickMessage(msg ExecTickMsg) (tea.Model, tea.Cmd) {
	for _, h := range m.hosts {
		if h.ID == msg.HostID {
			hts := m.GetOrCreateHostTabs(h.ID, h.Name)
			tab := hts.ActiveTab()
			if tab != nil && tab.ID == msg.TabID && tab.IsRunning {
				tab.SpinnerFrame = (tab.SpinnerFrame + 1) % len(SpinnerFrames)
				elapsed := time.Since(tab.RunningStart).Seconds()
				spinner := SpinnerFrames[tab.SpinnerFrame]

				// Update dynamic status bar message
				cmdTrunc := tab.RunningCmd
				if len(cmdTrunc) > 30 {
					cmdTrunc = cmdTrunc[:27] + "..."
				}
				m.statusMessage = fmt.Sprintf("%s ⏳ [%s] %s (⏱️ %.1fs)", spinner, h.Name, cmdTrunc, elapsed)

				m.updateViewportContent()
				return m, tickExecCommand(msg.HostID, msg.TabID)
			}
			break
		}
	}
	return m, nil
}
