package tui

// renderConsoleOrCopilotSplit renders the console pane (full width, inline copilot mode).
func (m *Model) renderConsoleOrCopilotSplit(width, height int) string {
	return m.renderRemoteConsole(width, height)
}

