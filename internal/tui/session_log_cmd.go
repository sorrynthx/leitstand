package tui

import (
	"leitstand/internal/logger"
	"leitstand/internal/sessionlog"
	"leitstand/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
)

// SessionLogExportedMsg notifies the model that a session log file was written.
type SessionLogExportedMsg struct {
	FilePath string
	HostName string
	Err      error
}

func (m *Model) exportCurrentSessionLogCmd(host *storage.Host, tab *ConsoleTab) tea.Cmd {
	if host == nil || tab == nil {
		return nil
	}

	content := tab.ExportLogs()
	hostName := host.Name
	logDir := ""
	if m.cfg != nil && m.cfg.Logging.SessionLogDir != "" {
		logDir = m.cfg.Logging.SessionLogDir
	}

	return func() tea.Msg {
		savedPath, err := sessionlog.SaveSessionLog(logDir, hostName, content)
		if err != nil {
			logger.Warnf("[SessionLog] Failed to save session log for host %s: %v", hostName, err)
		} else {
			logger.Infof("[SessionLog] Session log exported for host %s -> %s", hostName, savedPath)
		}
		return SessionLogExportedMsg{
			FilePath: savedPath,
			HostName: hostName,
			Err:      err,
		}
	}
}
