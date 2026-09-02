package tui

import (
	"fmt"
	"leitstand/internal/quickcmd"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) completeInputCmd() tea.Cmd {
	raw := m.consoleInput.Value()
	if len(m.hosts) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(m.hosts) {
		return nil
	}
	curHost := m.hosts[m.selectedIndex]
	if m.hostStatus[curHost.ID] == HostStatusOffline {
		return nil
	}
	tab := m.CurrentActiveTab()
	cwd := "~"
	if tab != nil && tab.CWD != "" {
		cwd = tab.CWD
	}

	distro := "linux"
	if si := m.sysInfos[curHost.ID]; si != nil && si.OSDistro != "" {
		distro = si.OSDistro
	}

	trimmed := strings.TrimSpace(raw)
	lastWord := trimmed
	if idx := strings.LastIndex(trimmed, " "); idx != -1 {
		lastWord = trimmed[idx+1:]
	}

	quickCmds := quickcmd.Catalog[quickcmd.DetectOSTab(distro)]
	var candidates []string

	for _, qc := range quickCmds {
		if strings.HasPrefix(qc.Command, raw) {
			candidates = append(candidates, qc.Command)
		}
	}

	if len(candidates) == 1 {
		return func() tea.Msg {
			return TabCompletionMsg{
				OriginalInput: raw,
				NewInput:      candidates[0],
				Candidates:    candidates,
			}
		}
	}

	if len(candidates) > 1 {
		prefix := commonPrefix(candidates)
		return func() tea.Msg {
			return TabCompletionMsg{
				OriginalInput: raw,
				NewInput:      prefix,
				Candidates:    candidates,
			}
		}
	}

	return func() tea.Msg {
		if m.isDemo {
			demoFiles := []string{"systemd", "sysctl.conf", "nginx", "docker.sock", "app.log", "config.yaml"}
			var matches []string
			for _, f := range demoFiles {
				if strings.HasPrefix(f, lastWord) {
					matches = append(matches, f)
				}
			}
			if len(matches) == 1 {
				newRaw := raw[:len(raw)-len(lastWord)] + matches[0]
				return TabCompletionMsg{OriginalInput: raw, NewInput: newRaw, Candidates: matches}
			}
			return TabCompletionMsg{OriginalInput: raw, NewInput: raw, Candidates: matches}
		}

		client, err := m.getSSHClient(curHost)
		if err != nil {
			return TabCompletionMsg{OriginalInput: raw, NewInput: raw, Candidates: nil}
		}

		var safeCdTarget string
		if cwd == "~" || cwd == "" {
			safeCdTarget = "$HOME"
		} else {
			safeCdTarget = fmt.Sprintf("%q", cwd)
		}

		searchPattern := lastWord + "*"
		if lastWord == "" {
			searchPattern = "*"
		}

		wrappedCmd := fmt.Sprintf("cd %s 2>/dev/null || cd \"$HOME\" ; ls -1d %s 2>/dev/null | head -n 20", safeCdTarget, searchPattern)
		stdout, _, err := client.ExecWithTimeout(wrappedCmd, 4)
		if err != nil || len(stdout) == 0 {
			return TabCompletionMsg{OriginalInput: raw, NewInput: raw, Candidates: nil}
		}

		lines := strings.Split(strings.TrimSpace(string(stdout)), "\n")
		var validLines []string
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l != "" {
				validLines = append(validLines, l)
			}
		}

		if len(validLines) == 1 {
			newRaw := raw[:len(raw)-len(lastWord)] + validLines[0]
			return TabCompletionMsg{OriginalInput: raw, NewInput: newRaw, Candidates: validLines}
		}

		if len(validLines) > 1 {
			prefix := commonPrefix(validLines)
			newRaw := raw[:len(raw)-len(lastWord)] + prefix
			return TabCompletionMsg{OriginalInput: raw, NewInput: newRaw, Candidates: validLines}
		}

		return TabCompletionMsg{OriginalInput: raw, NewInput: raw, Candidates: nil}
	}
}

func commonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	prefix := strs[0]
	for _, s := range strs[1:] {
		for !strings.HasPrefix(s, prefix) {
			prefix = prefix[:len(prefix)-1]
			if prefix == "" {
				return ""
			}
		}
	}
	return prefix
}

type TabCompletionMsg struct {
	HostID        int64
	TabID         string
	OriginalInput string
	NewInput      string
	Candidates    []string
}

func (m *Model) initOrResizeViewport() {
	vpWidth := m.width - int(float64(m.width)*0.28) - 6
	if vpWidth < 30 {
		vpWidth = 30
	}

	vpHeight := m.height - 15
	if m.cfg != nil && m.cfg.Telemetry.PollingInterval <= 0 {
		vpHeight = m.height - 6
	}

	if m.fullScreenConsole {
		vpWidth = m.width - 4
		vpHeight = m.height - 5
	}
	if vpHeight < 3 {
		vpHeight = 3
	}

	if !m.viewportReady {
		m.viewport = viewport.New(vpWidth, vpHeight)
		m.viewportReady = true
	} else {
		m.viewport.Width = vpWidth
		m.viewport.Height = vpHeight
	}

	for _, hts := range m.hostTabs {
		for _, tab := range hts.Tabs {
			tab.Viewport.Width = vpWidth
			tab.Viewport.Height = vpHeight
		}
	}

	m.updateViewportContent()
}

func (m *Model) updateViewportContent() {
	if len(m.hosts) == 0 {
		m.viewport.SetContent("No host selected.")
		return
	}

	tab := m.CurrentActiveTab()
	if tab == nil {
		m.viewport.SetContent("No active tab.")
		return
	}

	tab.UpdateViewportContent()
}
