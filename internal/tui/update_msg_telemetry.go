package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) tryHandleTelemetryMessage(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case MetricResultMsg:
		if msg.Err != nil {
			m.errors[msg.HostID] = msg.Err
			m.hostStatus[msg.HostID] = HostStatusOffline
			m.statusMessage = "❌ SSH Connection failed: " + msg.Err.Error()
		} else {
			delete(m.errors, msg.HostID)
			m.hostStatus[msg.HostID] = HostStatusOnline
			if msg.Record != nil {
				m.metrics[msg.HostID] = msg.Record
			}
			if msg.SysInfo != nil {
				m.sysInfos[msg.HostID] = msg.SysInfo
			}
			for _, h := range m.hosts {
				if h.ID == msg.HostID {
					m.statusMessage = fmt.Sprintf("✨ Connected to '%s' (%s) | Telemetry Active", h.Name, h.Address)
					break
				}
			}
		}
		m.updateViewportContent()
		return m, nil, true

	case OpenFileMsg:
		if msg.Err != nil {
			m.statusMessage = "❌ " + msg.Err.Error()
			return m, nil, true
		}
		m.editorModal = NewEditorModal(msg.HostID, msg.HostName, msg.FilePath, msg.Content, m.width, m.height)
		m.showEditorModal = true
		return m, nil, true

	case FileSavedMsg:
		if msg.Err != nil {
			if m.editorModal != nil {
				m.editorModal.isSaving = false
				m.editorModal.err = msg.Err
			} else {
				m.statusMessage = fmt.Sprintf("❌ File save error: %v", msg.Err)
			}
			return m, nil, true
		}
		saveSuccessMsg := fmt.Sprintf("✅ File saved successfully at %s! (Remote file updated)", time.Now().Format("15:04:05"))
		if m.editorModal != nil {
			m.editorModal.isSaving = false
			m.editorModal.err = nil
			m.editorModal.StatusMsg = saveSuccessMsg
		}
		m.statusMessage = "✨ " + saveSuccessMsg
		m.updateViewportContent()

		if m.fileManager != nil && len(m.hosts) > 0 && m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) {
			curHost := m.hosts[m.selectedIndex]
			return m, m.fetchRemoteFilesCmd(curHost, m.fileManager.RemotePath, m.fileManager.RemotePath, m.fileManager.ShowHidden), true
		}
		return m, nil, true

	case StreamChunkMsg:
		if len(m.hosts) > 0 {
			for _, h := range m.hosts {
				if h.ID == msg.HostID {
					hts := m.GetOrCreateHostTabs(h.ID, h.Name)
					var targetTab *ConsoleTab
					for _, t := range hts.Tabs {
						if t.ID == msg.TabID {
							targetTab = t
							break
						}
					}
					if targetTab == nil {
						targetTab = hts.ActiveTab()
					}
					if targetTab != nil {
						if targetTab.IsScreenApp {
							targetTab.SetFrame(msg.Chunk)
						} else {
							targetTab.AppendLog(msg.Chunk)
						}
					}
					break
				}
			}
		}
		m.updateViewportContent()
		return m, listenStreamCmd(msg.msgChan), true

	case StreamFinishedMsg:
		if len(m.hosts) > 0 {
			for _, h := range m.hosts {
				if h.ID == msg.HostID {
					hts := m.GetOrCreateHostTabs(h.ID, h.Name)
					var targetTab *ConsoleTab
					for _, t := range hts.Tabs {
						if t.ID == msg.TabID {
							targetTab = t
							break
						}
					}
					if targetTab == nil {
						targetTab = hts.ActiveTab()
					}
					if targetTab != nil {
						if targetTab.IsScreenApp {
							targetTab.StopScreenApp(targetTab.StreamCmd)
						} else {
							targetTab.IsStreaming = false
							if msg.Err != nil {
								targetTab.AppendLog("❌ [Stream Error] " + msg.Err.Error())
							} else {
								targetTab.AppendLog("🏁 [Stream Completed]")
							}
						}
					}
					break
				}
			}
		}
		m.updateViewportContent()
		return m, nil, true
	}

	return m, nil, false
}

func (m *Model) handleTelemetryMessage(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd, _ := m.tryHandleTelemetryMessage(msg)
	return model, cmd
}
