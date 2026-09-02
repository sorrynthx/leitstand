package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"leitstand/internal/logger"
	"path"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleConsoleAutoCompletion() (tea.Model, tea.Cmd, bool) {
	if len(m.hosts) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(m.hosts) {
		return m, nil, true
	}
	curHost := m.hosts[m.selectedIndex]
	hts := m.GetOrCreateHostTabs(curHost.ID, curHost.Name)
	activeTab := hts.ActiveTab()
	if activeTab == nil {
		return m, nil, true
	}

	val := m.consoleInput.Value()
	if strings.TrimSpace(val) == "" {
		return m, nil, true
	}

	fields := strings.Fields(val)
	if len(fields) == 0 {
		return m, nil, true
	}

	lastToken := fields[len(fields)-1]
	prefix := val[:len(val)-len(lastToken)]

	var searchDir string
	var searchPrefix string

	if strings.Contains(lastToken, "/") || strings.Contains(lastToken, "\\") {
		lastSlash := strings.LastIndexAny(lastToken, "/\\")
		searchDir = lastToken[:lastSlash+1]
		searchPrefix = lastToken[lastSlash+1:]
	} else {
		searchDir = ""
		searchPrefix = lastToken
	}

	cleanSearchDir := strings.ReplaceAll(searchDir, "\\", "/")
	cwd := activeTab.CWD
	if cwd == "" || cwd == "~" {
		cwd = "."
	}

	var remoteDir string
	if strings.HasPrefix(cleanSearchDir, "/") {
		remoteDir = path.Clean(cleanSearchDir)
	} else if cleanSearchDir != "" {
		remoteDir = path.Clean(cwd + "/" + cleanSearchDir)
	} else {
		remoteDir = cwd
	}

	logger.Infof("AutoCompletion trigger: val=%q, searchDir=%q, searchPrefix=%q, remoteDir=%q, CWD=%q", val, searchDir, searchPrefix, remoteDir, activeTab.CWD)

	var candidates []string

	if m.isDemo {
		demoItems := []string{"config.yaml", "data/", "logs/", "scripts/", "main.go", "server.log", "test.sh", "kgkim/", "interpass/"}
		for _, item := range demoItems {
			if strings.HasPrefix(item, searchPrefix) {
				candidates = append(candidates, item)
			}
		}
	} else {
		// 1. Try SFTP ReadDir
		sftpClient, err := m.getSFTPClient(curHost)
		if err == nil && sftpClient != nil {
			entries, err := sftpClient.ReadDir(remoteDir)
			if err == nil {
				for _, entry := range entries {
					name := entry.Name()
					if name == "." || name == ".." {
						continue
					}
					if entry.IsDir() {
						name += "/"
					}
					if strings.HasPrefix(name, searchPrefix) {
						candidates = append(candidates, name)
					}
				}
			}
		}

		// 2. Fallback to SSH exec `ls -1a -p` if SFTP ReadDir returned empty or failed
		if len(candidates) == 0 {
			sshClient, err := m.getSSHClient(curHost)
			if err == nil && sshClient != nil {
				rawClient := sshClient.RawClient()
				if rawClient != nil {
					session, err := rawClient.NewSession()
					if err == nil {
						defer session.Close()
						cmd := fmt.Sprintf("cd \"%s\" 2>/dev/null && ls -1 -p -a", remoteDir)
						out, err := session.Output(cmd)
						if err == nil {
							lines := strings.Split(string(out), "\n")
							for _, line := range lines {
								name := strings.TrimSpace(line)
								if name == "" || name == "." || name == "./" || name == ".." || name == "../" {
									continue
								}
								if strings.HasPrefix(name, searchPrefix) {
									candidates = append(candidates, name)
								}
							}
						}
					}
				}
			}
		}
	}

	logger.Infof("AutoCompletion result: candidates=%v", candidates)

	if len(candidates) == 0 {
		m.statusMessage = "ℹ️ " + i18n.T("no_completion_matches")
		return m, nil, true
	}

	if len(candidates) == 1 {
		completedToken := searchDir + candidates[0]
		newVal := prefix + completedToken
		m.consoleInput.SetValue(newVal)
		m.consoleInput.SetCursor(len(newVal))
		m.statusMessage = "✨ Auto-completed: " + candidates[0]
		return m, nil, true
	}

	commonPrefix := findCommonPrefix(candidates)
	if len(commonPrefix) > len(searchPrefix) {
		completedToken := searchDir + commonPrefix
		newVal := prefix + completedToken
		m.consoleInput.SetValue(newVal)
		m.consoleInput.SetCursor(len(newVal))
	}

	suggestionsStr := "💡 [Auto-Complete Suggestions]: " + strings.Join(candidates, "   ")
	activeTab.AppendLog(suggestionsStr)
	m.statusMessage = fmt.Sprintf("💡 %d matching suggestions found.", len(candidates))
	m.updateViewportContent()

	return m, nil, true
}

func findCommonPrefix(strs []string) string {
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
