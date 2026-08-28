package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// updateRunbookDrawer handles events when the Quick Command / Runbook Drawer is open.
// Returns (updatedModel, teaCmd, wasDrawerActive).
func (m *Model) updateRunbookDrawer(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	if !m.showDrawer || m.drawer == nil {
		return m, nil, false
	}

	done, chosenCmd, cmd := m.drawer.Update(msg)
	if done {
		m.showDrawer = false
		m.drawer = nil
		if chosenCmd != "" {
			m.consoleInput.SetValue(chosenCmd)
			m.consoleInput.SetCursor(len(chosenCmd))
			m.activePane = PaneConsole
			m.consoleInput.Focus()
			m.statusMessage = fmt.Sprintf("⌨️ Inserted: %s (Press Enter to run, or edit)", chosenCmd)
		}
		return m, nil, true
	}
	return m, cmd, true
}
