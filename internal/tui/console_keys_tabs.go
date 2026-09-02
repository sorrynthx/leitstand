package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleConsoleTabKeys(keyStr string, hts *HostTabs) (tea.Model, tea.Cmd, bool) {
	saveCurrentInput := func() {
		if curTab := hts.ActiveTab(); curTab != nil {
			curTab.InputText = m.consoleInput.Value()
		}
	}
	restoreNewInput := func() {
		if newTab := hts.ActiveTab(); newTab != nil {
			m.consoleInput.SetValue(newTab.InputText)
			m.consoleInput.SetCursor(len(newTab.InputText))
		} else {
			m.consoleInput.SetValue("")
		}
	}

	if keyStr == "ctrl+t" || keyStr == "ctrl+n" {
		saveCurrentInput()
		hts.AddNewTab(m.viewport.Width, m.viewport.Height)
		restoreNewInput()
		m.updateViewportContent()
		m.statusMessage = fmt.Sprintf(i18n.T("tab_new_msg"), len(hts.Tabs))
		return m, nil, true
	}

	if keyStr == "ctrl+w" {
		closed := hts.CloseActiveTab()
		if closed {
			restoreNewInput()
			m.statusMessage = i18n.T("tab_closed_msg")
			m.updateViewportContent()
		}
		return m, nil, true
	}

	if keyStr == "alt+left" || keyStr == "alt+p" || keyStr == "ctrl+pgup" {
		saveCurrentInput()
		hts.PrevTab()
		restoreNewInput()
		m.updateViewportContent()
		m.statusMessage = fmt.Sprintf("📌 Tab #%d (%s)", hts.ActiveIndex+1, hts.ActiveTab().Title)
		return m, nil, true
	}

	if keyStr == "alt+right" || keyStr == "alt+n" || keyStr == "ctrl+pgdown" {
		saveCurrentInput()
		hts.NextTab()
		restoreNewInput()
		m.updateViewportContent()
		m.statusMessage = fmt.Sprintf("📌 Tab #%d (%s)", hts.ActiveIndex+1, hts.ActiveTab().Title)
		return m, nil, true
	}

	if strings.HasPrefix(keyStr, "alt+") && len(keyStr) == 5 {
		digitRune := keyStr[4]
		if digitRune == '9' {
			lastIdx := len(hts.Tabs) - 1
			saveCurrentInput()
			if hts.SwitchTab(lastIdx) {
				restoreNewInput()
				m.updateViewportContent()
				m.statusMessage = fmt.Sprintf("📌 Switched to Last Tab #%d", lastIdx+1)
				return m, nil, true
			}
		} else if digitRune >= '1' && digitRune <= '8' {
			idx, err := strconv.Atoi(string(digitRune))
			if err == nil && idx >= 1 {
				targetIdx := idx - 1
				saveCurrentInput()
				if hts.SwitchTab(targetIdx) {
					restoreNewInput()
					m.updateViewportContent()
					m.statusMessage = fmt.Sprintf("📌 Switched to Tab #%d (%s)", idx, hts.ActiveTab().Title)
					return m, nil, true
				}
			}
		}
	}

	if keyStr == "up" {
		tab := hts.ActiveTab()
		if tab != nil && len(tab.CmdHistory) > 0 {
			if tab.HistoryIndex == -1 {
				tab.HistoryIndex = len(tab.CmdHistory) - 1
			} else if tab.HistoryIndex > 0 {
				tab.HistoryIndex--
			}
			m.consoleInput.SetValue(tab.CmdHistory[tab.HistoryIndex])
			m.consoleInput.SetCursor(len(tab.CmdHistory[tab.HistoryIndex]))
			return m, nil, true
		}
	}

	if keyStr == "down" {
		tab := hts.ActiveTab()
		if tab != nil && len(tab.CmdHistory) > 0 && tab.HistoryIndex != -1 {
			if tab.HistoryIndex < len(tab.CmdHistory)-1 {
				tab.HistoryIndex++
				m.consoleInput.SetValue(tab.CmdHistory[tab.HistoryIndex])
				m.consoleInput.SetCursor(len(tab.CmdHistory[tab.HistoryIndex]))
			} else {
				tab.HistoryIndex = -1
				m.consoleInput.SetValue("")
			}
			return m, nil, true
		}
	}

	return m, nil, false
}
