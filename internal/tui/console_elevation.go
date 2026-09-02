package tui

import (
	"fmt"
	"leitstand/internal/logger"
	"leitstand/internal/storage"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type ElevationMode int

const (
	ElevationNone ElevationMode = iota
	ElevationSudo
	ElevationSuRoot
)

func (m *Model) execSudoValidateAndElevateCmd(host *storage.Host, tab *ConsoleTab, password string, remember bool) tea.Cmd {
	hostID := host.ID
	hostName := host.Name
	tabID := tab.ID

	return func() tea.Msg {
		logger.Infof("execSudoValidateAndElevateCmd: testing credentials for host '%s'", hostName)

		client, err := m.getSSHClient(host)
		if err != nil {
			return RootElevateResultMsg{HostID: hostID, HostName: hostName, TabID: tabID, Success: false, Err: err}
		}

		stdinData := []byte(password + "\n")

		// Probe 1: Try sudo -S -k -p '' true
		sudoTestCmd := "sudo -S -k -p '' true"
		_, sudoErrOut, sudoErr := client.ExecWithStdin(sudoTestCmd, stdinData, 8*time.Second)

		if sudoErr == nil {
			logger.Infof("execSudoValidateAndElevateCmd: sudo succeeded on %s", hostName)
			return RootElevateResultMsg{
				HostID:   hostID,
				HostName: hostName,
				TabID:    tabID,
				Success:  true,
				Password: password,
				Remember: remember,
				Mode:     int(ElevationSudo),
			}
		}

		logger.Warnf("execSudoValidateAndElevateCmd: sudo failed (%v: %s), falling back to su root on %s", sudoErr, strings.TrimSpace(string(sudoErrOut)), hostName)

		// Probe 2: Try su root -c true
		suTestCmd := "su root -c true"
		_, suOut, suErr := client.ExecWithStdin(suTestCmd, stdinData, 8*time.Second)

		if suErr == nil {
			logger.Infof("execSudoValidateAndElevateCmd: su root succeeded on %s", hostName)
			return RootElevateResultMsg{
				HostID:   hostID,
				HostName: hostName,
				TabID:    tabID,
				Success:  true,
				Password: password,
				Remember: remember,
				Mode:     int(ElevationSuRoot),
			}
		}

		errDetail := strings.TrimSpace(string(suOut))
		if errDetail == "" {
			errDetail = fmt.Sprintf("sudo: %v / su: %v", sudoErr, suErr)
		}

		return RootElevateResultMsg{
			HostID:   hostID,
			HostName: hostName,
			TabID:    tabID,
			Success:  false,
			Err:      fmt.Errorf("%s", errDetail),
		}
	}
}

func (m *Model) execSudoCmd(host *storage.Host, cmdText string, password string) tea.Cmd {
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

	elevMode := ElevationMode(m.sudoModeCache[host.ID])

	return func() tea.Msg {
		logger.Infof("execSudoCmd: host='%s', cwd='%s', cmd='%s', mode=%d", host.Name, cwd, cmdText, elevMode)

		if m.isDemo {
			return CmdResultMsg{
				HostID:  host.ID,
				TabID:   tabID,
				Command: cmdText,
				CWD:     cwd,
				NewCWD:  cwd,
				Stdout:  fmt.Sprintf("✨ [SUDO Success] Executed as root: %s\nuid=0(root) gid=0(root) groups=0(root)", cmdText),
				Err:     nil,
			}
		}

		client, err := m.getSSHClient(host)
		if err != nil {
			return CmdResultMsg{HostID: host.ID, TabID: tabID, Command: cmdText, CWD: cwd, Err: err}
		}

		var safeCdTarget string
		if cwd == "~" || cwd == "" {
			safeCdTarget = "$HOME"
		} else {
			safeCdTarget = fmt.Sprintf("%q", cwd)
		}

		trimmed := strings.TrimSpace(cmdText)
		actualCmd := trimmed
		if strings.HasPrefix(trimmed, "sudo ") {
			actualCmd = strings.TrimPrefix(trimmed, "sudo ")
		}

		escapedCmd := strings.ReplaceAll(actualCmd, "'", "'\"'\"'")
		stdinData := []byte(password + "\n")
		var wrappedCmd string
		innerScript := fmt.Sprintf("cd %s 2>/dev/null || cd \"$HOME\" ; %s ; __LEITSTAND_RET=$? ; echo '' ; echo '___LEITSTAND_PWD___' ; pwd ; exit $__LEITSTAND_RET", safeCdTarget, escapedCmd)

		if elevMode == ElevationSuRoot {
			wrappedCmd = fmt.Sprintf("su root -c '%s'", innerScript)
		} else if password != "" {
			wrappedCmd = fmt.Sprintf("sudo -S -p '' sh -c '%s'", innerScript)
		} else {
			wrappedCmd = fmt.Sprintf("sh -c '%s'", innerScript)
		}

		stdout, stderr, err := client.ExecWithStdin(wrappedCmd, stdinData, 30*time.Second)

		stdoutStr := string(stdout)
		newCWD := cwd
		if idx := strings.Index(stdoutStr, "___LEITSTAND_PWD___"); idx != -1 {
			outPart := strings.TrimRight(stdoutStr[:idx], "\r\n ")
			pwdRaw := stdoutStr[idx+len("___LEITSTAND_PWD___"):]
			pwdLines := strings.Split(strings.TrimSpace(pwdRaw), "\n")
			if len(pwdLines) > 0 {
				pwdLine := strings.TrimSpace(pwdLines[0])
				if pwdLine != "" {
					newCWD = pwdLine
				}
			}
			stdoutStr = outPart
		}

		stderrStr := string(stderr)
		stderrStr = strings.ReplaceAll(stderrStr, "[sudo] password for ", "")

		return CmdResultMsg{
			HostID:  host.ID,
			TabID:   tabID,
			Command: cmdText,
			CWD:     cwd,
			NewCWD:  newCWD,
			Stdout:  stdoutStr,
			Stderr:  stderrStr,
			Err:     err,
		}
	}
}
