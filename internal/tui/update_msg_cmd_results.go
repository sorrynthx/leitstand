package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleCommandResultMessage(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case CmdResultMsg:
		if len(m.hosts) > 0 {
			for _, h := range m.hosts {
				if h.ID == msg.HostID {
					hts := m.GetOrCreateHostTabs(h.ID, h.Name)
					var targetTab *ConsoleTab
					for _, t := range hts.Tabs {
						if t.ID == msg.TabID {
							targetTab = t
							break
						}
					}
					if targetTab == nil {
						targetTab = hts.ActiveTab()
					}

					if targetTab != nil {
						cwdDisplay := msg.CWD
						if cwdDisplay == "" {
							cwdDisplay = "~"
						}

						if msg.NewCWD != "" {
							targetTab.CWD = msg.NewCWD
						}

						if msg.Err != nil {
							targetTab.AppendLog(fmt.Sprintf("[%s] ❯ %s\n❌ Error: %v\n%s", cwdDisplay, msg.Command, msg.Err, msg.Stderr))
							m.statusMessage = fmt.Sprintf("⚠️ Error executing '%s'", msg.Command)
						} else {
							out := msg.Stdout
							if out != "" {
								targetTab.AppendLog(fmt.Sprintf("[%s] ❯ %s\n%s", cwdDisplay, msg.Command, out))
							} else {
								targetTab.AppendLog(fmt.Sprintf("[%s] ❯ %s\n(no output)", cwdDisplay, msg.Command))
							}
							m.statusMessage = fmt.Sprintf("✅ Executed '%s' successfully (%s)", msg.Command, time.Now().Format("15:04:05"))
						}
					}
					break
				}
			}
		}
		m.updateViewportContent()
		return m, nil

	case RootElevateResultMsg:
		if len(m.hosts) > 0 {
			for _, h := range m.hosts {
				if h.ID == msg.HostID {
					hts := m.GetOrCreateHostTabs(h.ID, h.Name)
					var targetTab *ConsoleTab
					for _, t := range hts.Tabs {
						if t.ID == msg.TabID {
							targetTab = t
							break
						}
					}
					if targetTab == nil {
						targetTab = hts.ActiveTab()
					}

					if targetTab != nil {
						if msg.Success {
							targetTab.IsRoot = true
							m.sudoCache[msg.HostID] = msg.Password
							m.sudoModeCache[msg.HostID] = msg.Mode
							modeStr := "sudo"
							if msg.Mode == int(ElevationSuRoot) {
								modeStr = "su root"
							}
							targetTab.SetAutoTitle(hts.ActiveIndex, "")
							targetTab.AppendLog(fmt.Sprintf("👑 [ROOT Mode Activated] Authenticated via %s on %s.\nCommands in this tab will execute with root privileges. Type 'exit' to log out.", modeStr, h.Name))
							m.statusMessage = fmt.Sprintf("👑 Root session activated (%s).", modeStr)
						} else {
							targetTab.IsRoot = false
							delete(m.sudoCache, msg.HostID)
							delete(m.sudoModeCache, msg.HostID)
							errMsg := "Incorrect password or user lacks sudo/su privileges."
							if msg.Err != nil {
								errMsg = msg.Err.Error()
							}
							targetTab.AppendLog(fmt.Sprintf("❌ [Authentication Failed] Root elevation failed on %s: %s", h.Name, errMsg))
							m.statusMessage = "❌ Root authentication failed."
						}
					}
					break
				}
			}
		}
		m.updateViewportContent()
		return m, nil

	case TabCompletionMsg:
		if msg.NewInput != "" && msg.NewInput != m.consoleInput.Value() {
			m.consoleInput.SetValue(msg.NewInput)
			m.consoleInput.SetCursor(len(msg.NewInput))
		}
		if len(msg.Candidates) > 1 {
			m.statusMessage = fmt.Sprintf("🔍 Matches: %s", strings.Join(msg.Candidates, "  "))
		} else if len(msg.Candidates) == 1 {
			m.statusMessage = "✨ Completed"
		} else {
			m.statusMessage = "(no completion matches)"
		}
		return m, nil

	case TickMsg:
		return m, tea.Batch(
			m.pollActiveHostsCmd(),
			m.tickCmd(),
		)
	}

	return m, nil
}
