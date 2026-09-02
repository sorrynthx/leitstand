package tui

import (
	"fmt"
	"leitstand/internal/highlight"
	"leitstand/internal/storage"
	"path"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type OpenFileMsg struct {
	HostID   int64
	HostName string
	FilePath string
	Content  string
	Err      error
}

type FileSavedMsg struct {
	HostID   int64
	HostName string
	FilePath string
	Err      error
}

func (m *Model) openRemoteFileCmd(host *storage.Host, filePath string) tea.Cmd {
	tab := m.GetOrCreateHostTabs(host.ID, host.Name).ActiveTab()
	if tab != nil && tab.CWD != "" && !path.IsAbs(filePath) {
		filePath = path.Join(tab.CWD, filePath)
	}

	return func() tea.Msg {
		if m.isDemo {
			sampleContent := "# Demo configuration file\nserver {\n  listen 80;\n  server_name demo.local;\n  location / {\n    proxy_pass http://localhost:8080;\n  }\n}"
			return OpenFileMsg{
				HostID:   host.ID,
				HostName: host.Name,
				FilePath: filePath,
				Content:  sampleContent,
			}
		}

		sftpClient, err := m.getSFTPClient(host)
		if err != nil {
			return OpenFileMsg{HostID: host.ID, HostName: host.Name, FilePath: filePath, Err: err}
		}

		f, err := sftpClient.Open(filePath)
		if err != nil {
			return OpenFileMsg{HostID: host.ID, HostName: host.Name, FilePath: filePath, Err: err}
		}
		defer f.Close()

		stat, err := f.Stat()
		if err == nil && stat.Size() > 2*1024*1024 {
			return OpenFileMsg{HostID: host.ID, HostName: host.Name, FilePath: filePath, Err: fmt.Errorf("file size (%.1f MB) exceeds 2MB edit limit", float64(stat.Size())/(1024*1024))}
		}

		buf := make([]byte, 2*1024*1024)
		n, err := f.Read(buf)
		if err != nil && err.Error() != "EOF" {
			return OpenFileMsg{HostID: host.ID, HostName: host.Name, FilePath: filePath, Err: err}
		}

		content := string(buf[:n])
		if !highlight.IsTextContent(content) {
			return OpenFileMsg{HostID: host.ID, HostName: host.Name, FilePath: filePath, Err: fmt.Errorf("file appears to be binary and cannot be edited in text editor")}
		}

		return OpenFileMsg{
			HostID:   host.ID,
			HostName: host.Name,
			FilePath: filePath,
			Content:  content,
		}
	}
}

func (m *Model) saveRemoteFileCmd(hostID int64, filePath, newContent string) tea.Cmd {
	var curHost *storage.Host
	for _, h := range m.hosts {
		if h.ID == hostID {
			curHost = h
			break
		}
	}
	if curHost != nil {
		tab := m.GetOrCreateHostTabs(hostID, curHost.Name).ActiveTab()
		if tab != nil && tab.CWD != "" && !path.IsAbs(filePath) {
			filePath = path.Join(tab.CWD, filePath)
		}
	}

	return func() tea.Msg {
		var curHost *storage.Host
		for _, h := range m.hosts {
			if h.ID == hostID {
				curHost = h
				break
			}
		}
		if curHost == nil {
			return FileSavedMsg{HostID: hostID, FilePath: filePath, Err: fmt.Errorf("host not found")}
		}

		if m.isDemo {
			return FileSavedMsg{HostID: hostID, HostName: curHost.Name, FilePath: filePath, Err: nil}
		}

		sftpClient, err := m.getSFTPClient(curHost)
		if err != nil {
			return FileSavedMsg{HostID: hostID, HostName: curHost.Name, FilePath: filePath, Err: err}
		}

		f, err := sftpClient.Create(filePath)
		if err != nil {
			if pass, ok := m.sudoCache[hostID]; ok && pass != "" {
				elevMode := ElevationMode(m.sudoModeCache[hostID])
				client, sshErr := m.getSSHClient(curHost)
				if sshErr == nil {
					cmdScript := fmt.Sprintf("cat << 'LEITSTAND_EOF' > %q\n%s\nLEITSTAND_EOF", filePath, newContent)
					var wrappedCmd string
					if elevMode == ElevationSuRoot {
						wrappedCmd = fmt.Sprintf("su root -c %q", cmdScript)
					} else {
						wrappedCmd = fmt.Sprintf("sudo -S -p '' sh -c %q", cmdScript)
					}
					_, stderr, execErr := client.ExecWithStdin(wrappedCmd, []byte(pass+"\n"), 10*time.Second)
					if execErr == nil {
						return FileSavedMsg{HostID: hostID, HostName: curHost.Name, FilePath: filePath, Err: nil}
					}
					err = fmt.Errorf("%v (%s)", err, strings.TrimSpace(string(stderr)))
				}
			}
			return FileSavedMsg{HostID: hostID, HostName: curHost.Name, FilePath: filePath, Err: err}
		}
		defer f.Close()

		_, err = f.Write([]byte(newContent))
		if err != nil {
			return FileSavedMsg{HostID: hostID, HostName: curHost.Name, FilePath: filePath, Err: err}
		}

		return FileSavedMsg{
			HostID:   hostID,
			HostName: curHost.Name,
			FilePath: filePath,
			Err:      nil,
		}
	}
}
