package tui

import (
	"fmt"
	"leitstand/internal/storage"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) execRemoteCmd(host *storage.Host, cmdText string) tea.Cmd {
	hts := m.GetOrCreateHostTabs(host.ID, host.Name)
	tab := hts.ActiveTab()
	if tab == nil {
		tab = hts.AddNewTab(m.viewport.Width, m.viewport.Height)
	}

	tab.SetAutoTitle(hts.ActiveIndex, cmdText)
	tabID := tab.ID
	cwd := tab.CWD
	if cwd == "" {
		cwd = "~"
	}

	if IsStreamingCommand(cmdText) {
		tab.AppendLog(fmt.Sprintf("❯ %s  [🔴 LIVE STREAMING - Press Ctrl+C to stop]", cmdText))
		return m.execStreamingCmdInTab(host, tab, cmdText)
	}

	return func() tea.Msg {
		if m.isDemo {
			out := SimulateDemoCmd(cmdText, host.Name)
			return CmdResultMsg{
				HostID:  host.ID,
				TabID:   tabID,
				Command: cmdText,
				CWD:     cwd,
				Stdout:  out,
			}
		}

		client, err := m.getSSHClient(host)
		if err != nil {
			return CmdResultMsg{
				HostID:  host.ID,
				TabID:   tabID,
				Command: cmdText,
				CWD:     cwd,
				Err:     err,
			}
		}

		var wrappedCmd string
		if cwd != "" && cwd != "~" {
			wrappedCmd = fmt.Sprintf("cd %q 2>/dev/null || cd \"$HOME\" ; %s ; __LEITSTAND_RET=$? ; echo '' ; echo '___LEITSTAND_PWD___' ; pwd ; exit $__LEITSTAND_RET", cwd, cmdText)
		} else {
			wrappedCmd = fmt.Sprintf("cd $HOME 2>/dev/null || cd \"$HOME\" ; %s ; __LEITSTAND_RET=$? ; echo '' ; echo '___LEITSTAND_PWD___' ; pwd ; exit $__LEITSTAND_RET", cmdText)
		}

		stdoutBytes, stderrBytes, err := client.ExecWithTimeout(wrappedCmd, 30*time.Second)
		stdout := string(stdoutBytes)
		stderr := string(stderrBytes)

		newCWD := cwd
		if idx := strings.Index(stdout, "___LEITSTAND_PWD___"); idx != -1 {
			outPart := strings.TrimRight(stdout[:idx], "\r\n ")
			pwdRaw := stdout[idx+len("___LEITSTAND_PWD___"):]
			pwdLines := strings.Split(strings.TrimSpace(pwdRaw), "\n")
			if len(pwdLines) > 0 {
				pwdLine := strings.TrimSpace(pwdLines[0])
				if pwdLine != "" {
					newCWD = pwdLine
				}
			}
			stdout = outPart
		}

		return CmdResultMsg{
			HostID:  host.ID,
			TabID:   tabID,
			Command: cmdText,
			CWD:     cwd,
			NewCWD:  newCWD,
			Stdout:  stdout,
			Stderr:  stderr,
			Err:     err,
		}
	}
}
