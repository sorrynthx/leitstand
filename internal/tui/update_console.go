package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"leitstand/internal/highlight"
	"leitstand/internal/logger"
	"leitstand/internal/ssh"
	"leitstand/internal/storage"
	"leitstand/internal/vault"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	gossh "golang.org/x/crypto/ssh"
)

// IsStreamingCommand checks if a command produces continuous background streaming output.
func IsStreamingCommand(cmdText string) bool {
	trimmed := strings.TrimSpace(cmdText)
	if strings.Contains(trimmed, "tail -f") || strings.Contains(trimmed, "tail -F") ||
		strings.Contains(trimmed, "docker logs -f") || strings.Contains(trimmed, "docker-compose logs -f") ||
		strings.Contains(trimmed, "journalctl -f") || strings.HasPrefix(trimmed, "ping ") ||
		strings.HasPrefix(trimmed, "watch ") ||
		trimmed == "top" || strings.HasPrefix(trimmed, "top ") ||
		trimmed == "htop" || trimmed == "btop" || trimmed == "iotop" || trimmed == "iftop" {
		return true
	}
	return false
}

// IsScreenCommand checks if a streaming command is a full-screen frame refresher (like top, htop, watch).
func IsScreenCommand(cmdText string) bool {
	trimmed := strings.TrimSpace(cmdText)
	return trimmed == "top" || strings.HasPrefix(trimmed, "top ") ||
		trimmed == "htop" || trimmed == "btop" || trimmed == "iotop" || trimmed == "iftop" ||
		strings.HasPrefix(trimmed, "watch ")
}

func listenStreamCmd(msgChan <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-msgChan
		if !ok {
			return nil
		}
		return msg
	}
}

// execStreamingCmdInTab initiates an asynchronous, continuous streaming command for a specific tab.
func (m *Model) execStreamingCmdInTab(host *storage.Host, tab *ConsoleTab, cmdText string) tea.Cmd {
	tabID := tab.ID
	hostID := host.ID
	cwd := tab.CWD
	if cwd == "" {
		cwd = "~"
	}

	ctx, cancel := context.WithCancel(m.ctx)
	tab.IsStreaming = true
	tab.IsScreenApp = IsScreenCommand(cmdText)
	tab.StreamCmd = cmdText
	tab.StreamCancel = cancel

	msgChan := make(chan tea.Msg, 50)

	if m.isDemo {
		go StartDemoStream(ctx, hostID, tabID, host.Name, tab.IsScreenApp, msgChan)
		return listenStreamCmd(msgChan)
	}

	client, err := m.getSSHClient(host)
	if err != nil {
		tab.IsStreaming = false
		tab.StreamCancel = nil
		return func() tea.Msg {
			return StreamFinishedMsg{HostID: hostID, TabID: tabID, Err: err}
		}
	}

	go func() {
		defer close(msgChan)
		rawClient := client.RawClient()
		if rawClient == nil {
			msgChan <- StreamFinishedMsg{HostID: hostID, TabID: tabID, Err: fmt.Errorf("ssh client closed")}
			return
		}

		session, err := rawClient.NewSession()
		if err != nil {
			msgChan <- StreamFinishedMsg{HostID: hostID, TabID: tabID, Err: err}
			return
		}
		defer session.Close()

		stdoutPipe, err := session.StdoutPipe()
		if err != nil {
			msgChan <- StreamFinishedMsg{HostID: hostID, TabID: tabID, Err: err}
			return
		}
		stderrPipe, _ := session.StderrPipe()

		var safeCdTarget string
		if cwd == "~" || cwd == "" {
			safeCdTarget = "$HOME"
		} else {
			safeCdTarget = fmt.Sprintf("\"%s\"", cwd)
		}

		actualCmd := cmdText
		if tab.IsScreenApp {
			trimmed := strings.TrimSpace(cmdText)
			if trimmed == "top" || strings.HasPrefix(trimmed, "top ") {
				actualCmd = "top -b -d 1 -c"
			}
		}

		wrappedCmd := fmt.Sprintf("cd %s 2>/dev/null || cd \"$HOME\" ; %s", safeCdTarget, actualCmd)

		if err := session.Start(wrappedCmd); err != nil {
			msgChan <- StreamFinishedMsg{HostID: hostID, TabID: tabID, Err: err}
			return
		}

		go func() {
			<-ctx.Done()
			_ = session.Signal(gossh.SIGKILL)
			_ = session.Close()
		}()

		reader := io.MultiReader(stdoutPipe, stderrPipe)
		scanner := bufio.NewScanner(reader)

		if tab.IsScreenApp {
			var frameLines []string
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "top - ") && len(frameLines) > 0 {
					frame := strings.Join(frameLines, "\n")
					frameLines = []string{line}
					select {
					case <-ctx.Done():
						return
					case msgChan <- StreamChunkMsg{HostID: hostID, TabID: tabID, Chunk: frame, msgChan: msgChan}:
					}
				} else {
					frameLines = append(frameLines, line)
				}
			}
			if len(frameLines) > 0 {
				msgChan <- StreamChunkMsg{HostID: hostID, TabID: tabID, Chunk: strings.Join(frameLines, "\n"), msgChan: msgChan}
			}
		} else {
			for scanner.Scan() {
				line := scanner.Text()
				select {
				case <-ctx.Done():
					break
				case msgChan <- StreamChunkMsg{HostID: hostID, TabID: tabID, Chunk: line, msgChan: msgChan}:
				}
			}
		}

		_ = session.Wait()
		msgChan <- StreamFinishedMsg{HostID: hostID, TabID: tabID, Err: nil}
	}()

	return listenStreamCmd(msgChan)
}

func (m *Model) execRemoteCmd(host *storage.Host, cmdText string) tea.Cmd {
	hts := m.GetOrCreateHostTabs(host.ID, host.Name)
	tab := hts.ActiveTab()
	if tab == nil {
		tab = hts.AddNewTab(m.viewport.Width, m.viewport.Height)
	}

	tab.SetAutoTitle(hts.ActiveIndex, cmdText)

	// Check if this is a continuous streaming command
	if IsStreamingCommand(cmdText) {
		tab.AppendLog(fmt.Sprintf("❯ %s  [🔴 LIVE STREAMING - Press Ctrl+C to stop]", cmdText))
		return m.execStreamingCmdInTab(host, tab, cmdText)
	}

	return m.execRemoteCmdInTab(host, tab, cmdText)
}

func (m *Model) execRemoteCmdInTab(host *storage.Host, tab *ConsoleTab, cmdText string) tea.Cmd {
	tabID := tab.ID
	cwd := tab.CWD
	if cwd == "" {
		cwd = "~"
	}

	return func() tea.Msg {
		logger.Infof("execRemoteCmd: host='%s', tab='%s', cwd='%s', input='%s'", host.Name, tabID, cwd, cmdText)

		// Normalize commands that require TTY in non-interactive mode
		actualCmd := cmdText
		trimmed := strings.TrimSpace(cmdText)
		if trimmed == "top" {
			actualCmd = "top -b -n 1 | head -n 35"
		} else if trimmed == "htop" {
			actualCmd = "echo '💡 [Notice] htop requires interactive terminal. Showing top snapshot:' ; top -b -n 1 | head -n 30"
		} else if strings.HasPrefix(trimmed, "vi ") || strings.HasPrefix(trimmed, "vim ") || strings.HasPrefix(trimmed, "nano ") || strings.HasPrefix(trimmed, "view ") {
			targetFile := strings.TrimSpace(trimmed[strings.Index(trimmed, " "):])
			actualCmd = fmt.Sprintf("echo '💡 [Notice] Interactive editors (vi/nano) require full terminal. Displaying file contents with cat:' ; cat %s", targetFile)
		} else if strings.HasPrefix(trimmed, "less ") || strings.HasPrefix(trimmed, "more ") {
			targetFile := strings.TrimSpace(trimmed[strings.Index(trimmed, " "):])
			actualCmd = fmt.Sprintf("cat %s", targetFile)
		}

		if m.isDemo {
			// Demo mode CWD simulation
			newCwd := cwd
			var out string
			if strings.HasPrefix(trimmed, "cd ") || trimmed == "cd" {
				target := strings.TrimSpace(strings.TrimPrefix(trimmed, "cd"))
				if target == "" || target == "~" {
					newCwd = DemoDefaultHome
				} else if target == "/" {
					newCwd = "/"
				} else if target == ".." {
					if newCwd == DemoDefaultHome {
						newCwd = "/home"
					} else if newCwd == "/home" {
						newCwd = "/"
					} else {
						newCwd = DemoDefaultHome
					}
				} else if strings.HasPrefix(target, "/") {
					newCwd = target
				} else {
					newCwd = DemoDefaultHome + "/" + target
				}
				out = ""
			} else if trimmed == "pwd" {
				if cwd == "~" {
					out = DemoDefaultHome
				} else {
					out = cwd
				}
			} else {
				out = SimulateDemoCmd(actualCmd, host.Name)
			}

			return CmdResultMsg{
				HostID:  host.ID,
				TabID:   tabID,
				Command: cmdText,
				CWD:     cwd,
				NewCWD:  newCwd,
				Stdout:  out,
			}
		}

		if m.collector == nil || m.collector.Pool() == nil || m.vault == nil {
			logger.Errorf("execRemoteCmd: collector or vault not ready")
			return CmdResultMsg{
				HostID:  host.ID,
				TabID:   tabID,
				Command: cmdText,
				CWD:     cwd,
				Err:     fmt.Errorf("collector or vault not ready"),
			}
		}

		if m.hostStatus[host.ID] == HostStatusOffline || m.pausedHosts[host.ID] {
			return CmdResultMsg{
				HostID:  host.ID,
				TabID:   tabID,
				Command: cmdText,
				CWD:     cwd,
				Err:     fmt.Errorf("server is currently offline (VPN required?). Press [r] on the host to connect"),
			}
		}

		client, err := m.getSSHClient(host)
		if err != nil {
			logger.Errorf("execRemoteCmd: failed to get ssh client for host %s: %v", host.Name, err)
			return CmdResultMsg{HostID: host.ID, TabID: tabID, Command: cmdText, CWD: cwd, Err: err}
		}

		// Execute command in current working directory and track new CWD with a safe shell wrapper
		var safeCdTarget string
		if cwd == "~" || cwd == "" {
			safeCdTarget = "$HOME"
		} else {
			safeCdTarget = fmt.Sprintf("\"%s\"", cwd)
		}

		wrappedCmd := fmt.Sprintf("(cd %s 2>/dev/null || cd \"$HOME\") ; ( %s ) ; __LEITSTAND_RET=$? ; echo '' ; echo '___LEITSTAND_PWD___' ; pwd ; exit $__LEITSTAND_RET", safeCdTarget, actualCmd)
		logger.Debugf("execRemoteCmd: executing wrapped command on %s: %s", host.Name, wrappedCmd)

		stdout, stderr, err := client.ExecWithTimeout(wrappedCmd, 15*time.Second)

		stdoutStr := string(stdout)
		newCWD := cwd
		if idx := strings.Index(stdoutStr, "___LEITSTAND_PWD___"); idx != -1 {
			outPart := strings.TrimRight(stdoutStr[:idx], "\r\n")
			pwdPart := strings.TrimSpace(stdoutStr[idx+len("___LEITSTAND_PWD___"):])
			if pwdPart != "" {
				newCWD = pwdPart
			}
			stdoutStr = outPart
		}

		// Apply syntax highlighting if displaying file content (cat, head, tail)
		if (strings.HasPrefix(trimmed, "cat ") || strings.HasPrefix(trimmed, "head ") || strings.HasPrefix(trimmed, "tail ")) && err == nil {
			targetFile := strings.TrimSpace(trimmed[strings.Index(trimmed, " "):])
			stdoutStr = highlight.Highlight(targetFile, stdoutStr)
		}

		return CmdResultMsg{
			HostID:  host.ID,
			TabID:   tabID,
			Command: cmdText,
			CWD:     cwd,
			NewCWD:  newCWD,
			Stdout:  stdoutStr,
			Stderr:  string(stderr),
			Err:     err,
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

	return func() tea.Msg {
		logger.Infof("execSudoCmd: host='%s', cwd='%s', cmd='%s'", host.Name, cwd, cmdText)

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
			safeCdTarget = fmt.Sprintf("\"%s\"", cwd)
		}

		escapedPass := strings.ReplaceAll(password, "'", "'\\''")
		trimmed := strings.TrimSpace(cmdText)
		var actualSudoCmd string
		if strings.HasPrefix(trimmed, "sudo ") {
			actualSudoCmd = fmt.Sprintf("echo '%s' | sudo -S -p '' %s", escapedPass, strings.TrimPrefix(trimmed, "sudo "))
		} else if strings.HasPrefix(trimmed, "su ") || trimmed == "su" {
			actualSudoCmd = fmt.Sprintf("echo '%s' | sudo -S -p '' -i", escapedPass)
		} else {
			actualSudoCmd = fmt.Sprintf("echo '%s' | sudo -S -p '' %s", escapedPass, trimmed)
		}

		wrappedCmd := fmt.Sprintf("(cd %s 2>/dev/null || cd \"$HOME\") ; ( %s ) ; __LEITSTAND_RET=$? ; echo '' ; echo '___LEITSTAND_PWD___' ; pwd ; exit $__LEITSTAND_RET", safeCdTarget, actualSudoCmd)

		stdout, stderr, err := client.ExecWithTimeout(wrappedCmd, 20*time.Second)

		stdoutStr := string(stdout)
		newCWD := cwd
		if idx := strings.Index(stdoutStr, "___LEITSTAND_PWD___"); idx != -1 {
			outPart := strings.TrimRight(stdoutStr[:idx], "\r\n")
			pwdPart := strings.TrimSpace(stdoutStr[idx+len("___LEITSTAND_PWD___"):])
			if pwdPart != "" {
				newCWD = pwdPart
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

func (m *Model) openRemoteFileCmd(host *storage.Host, filePath string) tea.Cmd {
	return func() tea.Msg {
		if m.isDemo {
			return OpenFileMsg{
				HostID:   host.ID,
				HostName: host.Name,
				FilePath: filePath,
				Content:  GetDemoFileContent(filePath),
			}
		}

		client, err := m.getSSHClient(host)
		if err != nil {
			return OpenFileMsg{HostID: host.ID, HostName: host.Name, FilePath: filePath, Err: err}
		}

		tab := m.CurrentActiveTab()
		cwd := "~"
		if tab != nil && tab.CWD != "" {
			cwd = tab.CWD
		}

		var cdTarget string
		if cwd == "~" || cwd == "" {
			cdTarget = "$HOME"
		} else {
			cdTarget = fmt.Sprintf("\"%s\"", cwd)
		}

		readCmd := fmt.Sprintf("cd %s 2>/dev/null || cd \"$HOME\" ; if [ -f \"%s\" ]; then cat \"%s\" ; else touch \"%s\" && cat \"%s\" ; fi", cdTarget, filePath, filePath, filePath, filePath)
		stdout, stderr, err := client.ExecWithTimeout(readCmd, 5*time.Second)
		if err != nil {
			return OpenFileMsg{HostID: host.ID, HostName: host.Name, FilePath: filePath, Err: fmt.Errorf("%v (stderr: %s)", err, string(stderr))}
		}

		return OpenFileMsg{
			HostID:   host.ID,
			HostName: host.Name,
			FilePath: filePath,
			Content:  string(stdout),
		}
	}
}

func (m *Model) saveRemoteFileCmd(hostID int64, filePath string, content string) tea.Cmd {
	return func() tea.Msg {
		var host *storage.Host
		for _, h := range m.hosts {
			if h.ID == hostID {
				host = h
				break
			}
		}
		if host == nil {
			return FileSavedMsg{HostID: hostID, FilePath: filePath, Err: fmt.Errorf("host not found")}
		}

		if m.isDemo {
			return FileSavedMsg{HostID: hostID, FilePath: filePath}
		}

		client, err := m.getSSHClient(host)
		if err != nil {
			return FileSavedMsg{HostID: hostID, FilePath: filePath, Err: err}
		}

		tab := m.CurrentActiveTab()
		cwd := "~"
		if tab != nil && tab.CWD != "" {
			cwd = tab.CWD
		}

		var cdTarget string
		if cwd == "~" || cwd == "" {
			cdTarget = "$HOME"
		} else {
			cdTarget = fmt.Sprintf("\"%s\"", cwd)
		}

		saveCmd := fmt.Sprintf("cd %s 2>/dev/null || cd \"$HOME\" ; cat << 'EOF_LEITSTAND_SAVE' > \"%s\"\n%s\nEOF_LEITSTAND_SAVE", cdTarget, filePath, content)
		_, stderr, err := client.ExecWithTimeout(saveCmd, 10*time.Second)
		if err != nil {
			return FileSavedMsg{HostID: hostID, FilePath: filePath, Err: fmt.Errorf("%v (stderr: %s)", err, string(stderr))}
		}

		return FileSavedMsg{HostID: hostID, FilePath: filePath}
	}
}

func (m *Model) completeInputCmd() tea.Cmd {
	raw := m.consoleInput.Value()
	if len(m.hosts) == 0 {
		return nil
	}
	curHost := m.hosts[m.selectedIndex]
	tab := m.CurrentActiveTab()
	cwd := "~"
	if tab != nil && tab.CWD != "" {
		cwd = tab.CWD
	}

	return func() tea.Msg {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return nil
		}

		commonCmds := []string{
			"docker", "docker-compose", "systemctl", "journalctl", "nginx", "python3", "node",
			"git", "tail", "cat", "grep", "curl", "chmod", "chown", "mkdir", "clear",
			"uname", "free", "df", "top", "htop", "kill", "find", "sudo", "apt", "yum", "ps",
			"ls", "ll", "la", "tar", "gzip", "sed", "awk", "ping", "netstat", "ss", "uptime", "history", "whoami", "pwd",
		}

		parts := strings.Split(raw, " ")
		lastToken := parts[len(parts)-1]

		if len(parts) == 1 && !strings.Contains(raw, "/") {
			var matches []string
			for _, cmd := range commonCmds {
				if strings.HasPrefix(cmd, lastToken) {
					matches = append(matches, cmd)
				}
			}
			if len(matches) == 1 {
				return TabCompletionMsg{
					HostID:        curHost.ID,
					OriginalInput: raw,
					NewInput:      matches[0] + " ",
					Candidates:    matches,
				}
			} else if len(matches) > 1 {
				return TabCompletionMsg{
					HostID:        curHost.ID,
					OriginalInput: raw,
					NewInput:      longestCommonPrefix(matches),
					Candidates:    matches,
				}
			}
		}

		if m.isDemo {
			demoFiles := GetDemoCompletionFiles()
			var matches []string
			for _, f := range demoFiles {
				if strings.HasPrefix(f, lastToken) {
					matches = append(matches, f)
				}
			}
			if len(matches) > 0 {
				prefix := strings.Join(parts[:len(parts)-1], " ")
				if prefix != "" {
					prefix += " "
				}
				newInput := prefix + matches[0]
				if !strings.HasSuffix(matches[0], "/") {
					newInput += " "
				}
				return TabCompletionMsg{
					HostID:        curHost.ID,
					OriginalInput: raw,
					NewInput:      newInput,
					Candidates:    matches,
				}
			}
			return nil
		}

		if m.hostStatus[curHost.ID] == HostStatusOffline {
			return nil
		}

		client, err := m.getSSHClient(curHost)
		if err != nil {
			return nil
		}

		var cdTarget string
		if cwd == "~" || cwd == "" {
			cdTarget = "$HOME"
		} else {
			cdTarget = fmt.Sprintf("\"%s\"", cwd)
		}

		queryCmd := fmt.Sprintf("cd %s 2>/dev/null || cd \"$HOME\" ; ls -dp %s* 2>/dev/null | head -n 25", cdTarget, lastToken)
		stdout, _, err := client.ExecWithTimeout(queryCmd, 2*time.Second)
		if err != nil || len(stdout) == 0 {
			return TabCompletionMsg{HostID: curHost.ID, OriginalInput: raw, Candidates: nil}
		}

		lines := strings.Split(strings.TrimSpace(string(stdout)), "\n")
		var validMatches []string
		for _, l := range lines {
			t := strings.TrimSpace(l)
			if t != "" {
				validMatches = append(validMatches, t)
			}
		}

		if len(validMatches) == 0 {
			return TabCompletionMsg{HostID: curHost.ID, OriginalInput: raw, Candidates: nil}
		}

		prefixParts := parts[:len(parts)-1]
		prefixStr := strings.Join(prefixParts, " ")
		if prefixStr != "" {
			prefixStr += " "
		}

		if len(validMatches) == 1 {
			completed := validMatches[0]
			newInput := prefixStr + completed
			if !strings.HasSuffix(completed, "/") {
				newInput += " "
			}
			return TabCompletionMsg{
				HostID:        curHost.ID,
				OriginalInput: raw,
				NewInput:      newInput,
				Candidates:    validMatches,
			}
		}

		lcp := longestCommonPrefix(validMatches)
		newInput := prefixStr + lcp
		return TabCompletionMsg{
			HostID:        curHost.ID,
			OriginalInput: raw,
			NewInput:      newInput,
			Candidates:    validMatches,
		}
	}
}

func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	prefix := strs[0]
	for _, s := range strs[1:] {
		for !strings.HasPrefix(s, prefix) {
			if len(prefix) == 0 {
				return ""
			}
			prefix = prefix[:len(prefix)-1]
		}
	}
	return prefix
}

func (m *Model) initOrResizeViewport() {
	vpWidth := m.width - int(float64(m.width)*0.30) - 8
	if vpWidth < 30 {
		vpWidth = 30
	}

	availHeight := m.height - 5
	if availHeight < 10 {
		availHeight = 10
	}

	vpHeight := availHeight - 8 - 5
	if m.fullScreenConsole {
		vpWidth = m.width - 6
		vpHeight = m.height - 8
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

func (m *Model) getSSHClient(host *storage.Host) (*ssh.Client, error) {
	if m.collector == nil || m.collector.Pool() == nil || m.vault == nil {
		return nil, fmt.Errorf("collector or vault not ready")
	}

	secret, err := m.store.GetHostSecret(host.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret for host %d: %w", host.ID, err)
	}

	decrypted, err := m.vault.Decrypt(secret.Nonce, secret.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret for host %d: %w", host.ID, err)
	}
	defer vault.ZeroBytes(decrypted)

	payload, err := storage.ParseSecretPayload(decrypted, secret.AuthMethod)
	if err != nil {
		return nil, fmt.Errorf("failed to parse secret payload for host %d: %w", host.ID, err)
	}

	return m.collector.Pool().GetOrCreateFromPayload(host, secret, payload)
}

// launchInteractiveTerminalCmd starts a full interactive SSH terminal session.
func (m *Model) launchInteractiveTerminalCmd(host *storage.Host) tea.Cmd {
	if m.isDemo {
		runner := ssh.NewDemoShellRunner(host.Name, host.Address)
		return tea.Exec(runner, func(err error) tea.Msg {
			return TerminalExitedMsg{HostID: host.ID, Err: err}
		})
	}

	client, err := m.getSSHClient(host)
	if err != nil {
		m.statusMessage = fmt.Sprintf("⚠️ Failed to establish SSH terminal: %v", err)
		return nil
	}

	runner := ssh.NewShellRunner(client)
	return tea.Exec(runner, func(err error) tea.Msg {
		return TerminalExitedMsg{HostID: host.ID, Err: err}
	})
}
