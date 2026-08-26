package tui

import (
	"fmt"
	"leitstand/internal/highlight"
	"leitstand/internal/logger"
	"leitstand/internal/storage"
	"leitstand/internal/vault"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) execRemoteCmd(host *storage.Host, cmdText string) tea.Cmd {
	cwd := m.hostCWD[host.ID]
	if cwd == "" {
		cwd = "~"
	}

	return func() tea.Msg {
		logger.Infof("execRemoteCmd: host='%s', cwd='%s', input='%s'", host.Name, cwd, cmdText)

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
					newCwd = "/home/ubuntu"
				} else if target == "/" {
					newCwd = "/"
				} else if target == ".." {
					if newCwd == "/home/ubuntu" {
						newCwd = "/home"
					} else if newCwd == "/home" {
						newCwd = "/"
					} else {
						newCwd = "/home/ubuntu"
					}
				} else if strings.HasPrefix(target, "/") {
					newCwd = target
				} else {
					newCwd = "/home/ubuntu/" + target
				}
				out = ""
			} else if trimmed == "pwd" {
				if cwd == "~" {
					out = "/home/ubuntu"
				} else {
					out = cwd
				}
			} else {
				out = simulateDemoCmd(actualCmd, host.Name)
			}

			return CmdResultMsg{
				HostID:  host.ID,
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
				Command: cmdText,
				CWD:     cwd,
				Err:     fmt.Errorf("collector or vault not ready"),
			}
		}

		if m.hostStatus[host.ID] == HostStatusOffline || m.pausedHosts[host.ID] {
			return CmdResultMsg{
				HostID:  host.ID,
				Command: cmdText,
				CWD:     cwd,
				Err:     fmt.Errorf("server is currently offline (VPN required?). Press [r] on the host to connect"),
			}
		}

		secret, err := m.store.GetHostSecret(host.ID)
		if err != nil {
			logger.Errorf("execRemoteCmd: failed to get secret for host %d: %v", host.ID, err)
			return CmdResultMsg{HostID: host.ID, Command: cmdText, CWD: cwd, Err: err}
		}

		decrypted, err := m.vault.Decrypt(secret.Nonce, secret.Ciphertext)
		if err != nil {
			logger.Errorf("execRemoteCmd: failed to decrypt secret for host %d: %v", host.ID, err)
			return CmdResultMsg{HostID: host.ID, Command: cmdText, CWD: cwd, Err: err}
		}
		defer vault.ZeroBytes(decrypted)

		client, err := m.collector.Pool().GetOrCreate(host, secret, decrypted, nil)
		if err != nil {
			logger.Errorf("execRemoteCmd: failed to connect to host %s: %v", host.Name, err)
			return CmdResultMsg{HostID: host.ID, Command: cmdText, CWD: cwd, Err: err}
		}

		// Execute command in current working directory and track new CWD with a safe shell wrapper
		var safeCdTarget string
		if cwd == "~" || cwd == "" {
			safeCdTarget = "$HOME"
		} else {
			safeCdTarget = fmt.Sprintf("\"%s\"", cwd)
		}

		wrappedCmd := fmt.Sprintf("cd %s 2>/dev/null || cd \"$HOME\" ; %s ; __LEITSTAND_RET=$? ; echo '' ; echo -n '___LEITSTAND_PWD___' ; pwd ; exit $__LEITSTAND_RET", safeCdTarget, actualCmd)
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

		logger.Infof("execRemoteCmd: host='%s' cmd='%s' completed (stdout_len=%d, stderr_len=%d, new_cwd='%s', err=%v)",
			host.Name, cmdText, len(stdoutStr), len(stderr), newCWD, err)

		return CmdResultMsg{
			HostID:  host.ID,
			Command: cmdText,
			CWD:     cwd,
			NewCWD:  newCWD,
			Stdout:  stdoutStr,
			Stderr:  string(stderr),
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
				Content:  "# Demo configuration file\nversion: '3.8'\nservices:\n  web:\n    image: nginx:alpine\n    ports:\n      - '80:80'\n",
			}
		}

		if m.collector == nil || m.collector.Pool() == nil || m.vault == nil {
			return OpenFileMsg{HostID: host.ID, HostName: host.Name, FilePath: filePath, Err: fmt.Errorf("collector or vault not ready")}
		}

		secret, err := m.store.GetHostSecret(host.ID)
		if err != nil {
			return OpenFileMsg{HostID: host.ID, HostName: host.Name, FilePath: filePath, Err: err}
		}

		decrypted, err := m.vault.Decrypt(secret.Nonce, secret.Ciphertext)
		if err != nil {
			return OpenFileMsg{HostID: host.ID, HostName: host.Name, FilePath: filePath, Err: err}
		}
		defer vault.ZeroBytes(decrypted)

		client, err := m.collector.Pool().GetOrCreate(host, secret, decrypted, nil)
		if err != nil {
			return OpenFileMsg{HostID: host.ID, HostName: host.Name, FilePath: filePath, Err: err}
		}

		cwd := m.hostCWD[host.ID]
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

		if m.collector == nil || m.collector.Pool() == nil || m.vault == nil {
			return FileSavedMsg{HostID: hostID, FilePath: filePath, Err: fmt.Errorf("collector or vault not ready")}
		}

		secret, err := m.store.GetHostSecret(host.ID)
		if err != nil {
			return FileSavedMsg{HostID: hostID, FilePath: filePath, Err: err}
		}

		decrypted, err := m.vault.Decrypt(secret.Nonce, secret.Ciphertext)
		if err != nil {
			return FileSavedMsg{HostID: hostID, FilePath: filePath, Err: err}
		}
		defer vault.ZeroBytes(decrypted)

		client, err := m.collector.Pool().GetOrCreate(host, secret, decrypted, nil)
		if err != nil {
			return FileSavedMsg{HostID: hostID, FilePath: filePath, Err: err}
		}

		cwd := m.hostCWD[host.ID]
		var cdTarget string
		if cwd == "~" || cwd == "" {
			cdTarget = "$HOME"
		} else {
			cdTarget = fmt.Sprintf("\"%s\"", cwd)
		}

		// Save file safely using heredoc
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
	cwd := m.hostCWD[curHost.ID]
	if cwd == "" {
		cwd = "~"
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
		}

		parts := strings.Split(raw, " ")
		lastToken := parts[len(parts)-1]

		// Command completion if typing first token without slash
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

		// Demo mode completion
		if m.isDemo {
			demoFiles := []string{"app/", "app.py", "logs/", "config.yaml", "README.md", "docker-compose.yml"}
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

		// Remote SSH path completion
		if m.hostStatus[curHost.ID] == HostStatusOffline || m.collector == nil || m.collector.Pool() == nil || m.vault == nil {
			return nil
		}

		secret, err := m.store.GetHostSecret(curHost.ID)
		if err != nil {
			return nil
		}
		decrypted, err := m.vault.Decrypt(secret.Nonce, secret.Ciphertext)
		if err != nil {
			return nil
		}
		defer vault.ZeroBytes(decrypted)

		client, err := m.collector.Pool().GetOrCreate(curHost, secret, decrypted, nil)
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

func (m *Model) appendConsoleLog(hostID int64, logEntry string) {
	logs := m.consoleLogs[hostID]
	logs = append(logs, logEntry)
	// Keep last 50 lines per host
	if len(logs) > 50 {
		logs = logs[len(logs)-50:]
	}
	m.consoleLogs[hostID] = logs
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

	vpHeight := availHeight - 8 - 4
	if m.fullScreenConsole {
		vpWidth = m.width - 6
		vpHeight = m.height - 7
	}
	if vpHeight < 4 {
		vpHeight = 4
	}

	if !m.viewportReady {
		m.viewport = viewport.New(vpWidth, vpHeight)
		m.viewportReady = true
	} else {
		m.viewport.Width = vpWidth
		m.viewport.Height = vpHeight
	}

	m.updateViewportContent()
}

func (m *Model) updateViewportContent() {
	if len(m.hosts) == 0 {
		m.viewport.SetContent("No host selected.")
		return
	}

	selectedHost := m.hosts[m.selectedIndex]
	logs := m.consoleLogs[selectedHost.ID]

	if len(logs) == 0 {
		welcomeMsg := fmt.Sprintf("Connected to %s (%s)\nType remote commands below and press Enter to execute.\nUse [PageUp/PageDown] or [Ctrl+U/Ctrl+D] to scroll output.", selectedHost.Name, selectedHost.Address)
		m.viewport.SetContent(welcomeMsg)
		return
	}

	content := strings.Join(logs, "\n\n")
	m.viewport.SetContent(content)
}
