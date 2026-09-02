package tui

import (
	"fmt"
	"strings"
)

// AppendLog adds a log line to a tab and updates its viewport.
func (tab *ConsoleTab) AppendLog(entry string) {
	wasAtBottom := tab.Viewport.AtBottom() || len(tab.Logs) == 0
	tab.Logs = append(tab.Logs, entry)
	if len(tab.Logs) > 200 {
		tab.Logs = tab.Logs[len(tab.Logs)-200:]
	}
	tab.Viewport.SetContent(strings.Join(tab.Logs, "\n"))
	if wasAtBottom {
		tab.Viewport.GotoBottom()
	}
}

// SetFrame replaces the tab content with a fresh live screen frame (top/htop/watch).
func (tab *ConsoleTab) SetFrame(frame string) {
	wasAtTop := tab.Viewport.YOffset == 0
	tab.Logs = []string{frame}
	tab.Viewport.SetContent(frame)
	if wasAtTop {
		tab.Viewport.GotoTop()
	}
}

// StopScreenApp clears screen-app frames and returns the tab to standard console mode.
func (tab *ConsoleTab) StopScreenApp(cmdName string) {
	tab.IsStreaming = false
	tab.IsScreenApp = false
	if tab.StreamCancel != nil {
		tab.StreamCancel()
		tab.StreamCancel = nil
	}
	msg := fmt.Sprintf("🛑 [Stopped screen app: %s]", cmdName)
	tab.Logs = []string{msg}
	tab.Viewport.SetContent(msg)
	tab.Viewport.GotoBottom()
}

// UpdateViewportContent refreshes the tab's viewport content while respecting current scroll position.
func (tab *ConsoleTab) UpdateViewportContent() {
	if len(tab.Logs) == 0 {
		welcomeMsg := fmt.Sprintf("Terminal Session Tab [%s]\nType remote commands below and press Enter to execute.\n[Ctrl+N] New Tab  [Ctrl+W] Close Tab  [Alt+1~9] Switch Tab", tab.Title)
		tab.Viewport.SetContent(welcomeMsg)
		return
	}
	wasAtBottom := tab.Viewport.AtBottom()
	tab.Viewport.SetContent(strings.Join(tab.Logs, "\n"))
	if wasAtBottom {
		tab.Viewport.GotoBottom()
	}
}

// SetAutoTitle automatically updates tab title based on running command or cwd.
func (tab *ConsoleTab) SetAutoTitle(tabIndex int, cmdText string) {
	trimmed := strings.TrimSpace(cmdText)
	if trimmed == "" {
		if tab.IsRoot {
			tab.Title = fmt.Sprintf("%d: root#", tabIndex+1)
		} else {
			tab.Title = fmt.Sprintf("%d: bash", tabIndex+1)
		}
		return
	}

	fields := strings.Fields(trimmed)
	shortCmd := fields[0]
	if len(fields) > 1 && (shortCmd == "sudo" || shortCmd == "tail" || shortCmd == "docker" || shortCmd == "journalctl" || shortCmd == "su") {
		shortCmd = fmt.Sprintf("%s %s", fields[0], fields[1])
	}

	if tab.IsRoot && !strings.HasPrefix(shortCmd, "root") {
		shortCmd = fmt.Sprintf("root: %s", shortCmd)
	}

	if len(shortCmd) > 18 {
		shortCmd = shortCmd[:18] + "…"
	}

	tab.Title = fmt.Sprintf("%d: %s", tabIndex+1, shortCmd)
}
