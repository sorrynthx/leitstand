package tui

import (
	"context"
	"fmt"
	"leitstand/internal/i18n"
	"leitstand/internal/storage"
	"path"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) tryHandleSFTPMessage(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch msg := msg.(type) {

	case TransferActionMsg:
		if m.fileManager != nil {
			m.fileManager.IsTransferring = true
			m.fileManager.TransferIsUpload = msg.IsUpload
			m.fileManager.TransferDoneMsg = ""
		}
		return m, m.startFileTransferCmd(msg), true

	case FileTransferProgressMsg:
		if m.fileManager != nil {
			if m.fileManager.IsTransferCanceled {
				m.fileManager.IsTransferring = false
				m.fileManager.IsTransferBackground = false
				return m, nil, true
			}
			m.fileManager.IsTransferring = !msg.IsDone
			m.fileManager.CurrentFileName = msg.FileName
			m.fileManager.FileIndex = msg.FileIndex
			m.fileManager.FileTotal = msg.FileTotal
			m.fileManager.CurrentBytes = msg.CurrentBytes
			m.fileManager.CurrentTotal = msg.TotalBytes
			m.fileManager.BytesPerSec = msg.BytesPerSec

			if m.fileManager.IsTransferBackground && !msg.IsDone && msg.Err == nil {
				pct := 0.0
				if msg.TotalBytes > 0 {
					pct = (float64(msg.CurrentBytes) / float64(msg.TotalBytes)) * 100.0
				}
				speedStr := fmt.Sprintf("%.1f KB/s", msg.BytesPerSec/1024.0)
				if msg.BytesPerSec >= 1024*1024 {
					speedStr = fmt.Sprintf("%.1f MB/s", msg.BytesPerSec/(1024*1024))
				}
				m.statusMessage = fmt.Sprintf("⬆️ [%s] %.1f%% (%s) [f] 키로 복귀", msg.FileName, pct, speedStr)
			}

			if msg.Err != nil {
				m.fileManager.IsTransferring = false
				m.fileManager.IsTransferBackground = false
				m.fileManager.TransferDoneMsg = ""
				if msg.Err == context.Canceled || strings.Contains(msg.Err.Error(), "canceled") || strings.Contains(msg.Err.Error(), "closed") {
					m.fileManager.StatusMessage = "⚠️ 전송이 사용자에 의해 취소되었습니다."
					m.statusMessage = "⚠️ 파일 전송이 취소되었습니다."
				} else {
					m.fileManager.StatusMessage = fmt.Sprintf("❌ %v", msg.Err)
					m.statusMessage = fmt.Sprintf("❌ 파일 전송 실패: %v", msg.Err)
				}
				return m, nil, true
			} else if msg.IsDone {
				m.fileManager.IsTransferring = false
				m.fileManager.IsTransferBackground = false
				doneText := fmt.Sprintf(i18n.T("sftp_transfer_done"), msg.FileTotal)
				if msg.IsMove {
					doneText = fmt.Sprintf(i18n.T("sftp_move_done"), msg.FileTotal)
				}
				m.fileManager.TransferDoneMsg = doneText
				m.fileManager.StatusMessage = doneText
				m.statusMessage = doneText
				m.fileManager.TransferDoneTime = time.Now()
				m.fileManager.LocalSelected = make(map[string]bool)
				m.fileManager.RemoteSelected = make(map[string]bool)
				m.fileManager.ClipboardPaths = nil
				m.fileManager.ClipboardIsCut = false
				var curHost *storage.Host
				if len(m.hosts) > 0 && m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) {
					curHost = m.hosts[m.selectedIndex]
				}
				return m, tea.Batch(
					m.fileManager.RefreshLocalCmd(),
					m.fetchRemoteFilesCmd(curHost, m.fileManager.RemotePath, m.fileManager.RemotePath, m.fileManager.ShowHidden),
				), true
			}
		}
		if msg.msgChan != nil {
			return m, listenStreamCmd(msg.msgChan), true
		}
		return m, nil, true

	case FileOpActionMsg:
		return m, m.executeFileOpCmd(msg), true

	case FileOpResultMsg:
		if m.fileManager != nil {
			if msg.Err != nil {
				m.fileManager.StatusMessage = fmt.Sprintf("⚠️ %v", msg.Err)
			} else {
				m.fileManager.StatusMessage = msg.Msg
			}
			m.fileManager.LocalSelected = make(map[string]bool)
			m.fileManager.RemoteSelected = make(map[string]bool)
			var curHost *storage.Host
			if len(m.hosts) > 0 && m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) {
				curHost = m.hosts[m.selectedIndex]
			}
			if msg.IsLocal {
				return m, m.fileManager.RefreshLocalCmd(), true
			}
			return m, m.fetchRemoteFilesCmd(curHost, m.fileManager.RemotePath, m.fileManager.RemotePath, m.fileManager.ShowHidden), true
		}
		return m, nil, true

	case LocalFileListMsg:
		if m.fileManager != nil && m.fileManager.HostID == msg.HostID {
			if msg.Err != nil {
				m.fileManager.LocalPath = msg.OldPath
				m.fileManager.StatusMessage = fmt.Sprintf(i18n.T("sftp_perm_denied"), filepath.Base(msg.Path))
			} else {
				m.fileManager.LocalPath = msg.Path
				m.fileManager.LocalItems = msg.Items
				if m.fileManager.LocalCursor >= len(msg.Items) {
					if len(msg.Items) > 0 {
						m.fileManager.LocalCursor = len(msg.Items) - 1
					} else {
						m.fileManager.LocalCursor = 0
					}
				}
				if m.fileManager.LocalCursor < 0 {
					m.fileManager.LocalCursor = 0
				}
			}
		}
		return m, nil, true

	case RemoteFileListMsg:
		if msg.Err != nil {
			m.hostStatus[msg.HostID] = HostStatusOffline
			m.errors[msg.HostID] = msg.Err
		} else {
			m.hostStatus[msg.HostID] = HostStatusOnline
			delete(m.errors, msg.HostID)
		}
		if m.fileManager != nil && m.fileManager.HostID == msg.HostID {
			if msg.Err != nil {
				m.fileManager.RemotePath = msg.OldPath
				m.fileManager.StatusMessage = fmt.Sprintf(i18n.T("sftp_perm_denied"), path.Base(msg.Path))
			} else {
				m.fileManager.RemoteItems = msg.Items
				if msg.Path != "" {
					m.fileManager.RemotePath = msg.Path
				}
				if m.fileManager.RemoteCursor >= len(msg.Items) {
					if len(msg.Items) > 0 {
						m.fileManager.RemoteCursor = len(msg.Items) - 1
					} else {
						m.fileManager.RemoteCursor = 0
					}
				}
				if m.fileManager.RemoteCursor < 0 {
					m.fileManager.RemoteCursor = 0
				}
			}
		}
		return m, nil, true

	case NavigateRemoteMsg:
		if m.fileManager != nil && m.fileManager.HostID == msg.HostID {
			var curHost *storage.Host
			for _, h := range m.hosts {
				if h.ID == msg.HostID {
					curHost = h
					break
				}
			}
			if curHost != nil {
				return m, m.fetchRemoteFilesCmd(curHost, msg.NewPath, msg.OldPath, m.fileManager.ShowHidden), true
			}
		}
		return m, nil, true

	case FileManagerQuickCmdMsg:
		return m, m.executeFileManagerQuickCmd(msg), true

	case FileManagerQuickCmdResultMsg:
		if m.fileManager != nil && m.fileManager.HostID == msg.HostID {
			outText := strings.TrimSpace(msg.Output)
			if msg.Err == nil && outText == "" {
				outText = "✨ (명령어가 성공적으로 실행되었으나 출력 결과가 없습니다. 조건에 일치하는 대상이 0건입니다.)"
			} else if msg.Err != nil && outText == "" {
				outText = fmt.Sprintf("⚠️ (오류 발생: %v)", msg.Err)
			}

			if msg.Err != nil {
				m.fileManager.StatusMessage = fmt.Sprintf("⚠️ %v", msg.Err)
			} else {
				m.fileManager.StatusMessage = fmt.Sprintf("✨ Executed: %s", msg.Command)
			}
			if msg.NewCWD != "" && msg.NewCWD != msg.OldCWD {
				if msg.IsLocal {
					m.fileManager.LocalPath = msg.NewCWD
				} else {
					m.fileManager.RemotePath = msg.NewCWD
				}
			}
			m.fileManager.CmdOutputTitle = fmt.Sprintf("%s (in %s)", msg.Command, msg.OldCWD)
			m.fileManager.CmdOutputContent = outText
			m.fileManager.CmdOutputScroll = 0
			m.fileManager.ShowCmdOutput = true

			if msg.IsLocal {
				return m, m.fileManager.RefreshLocalCmd(), true
			}
			var curHost *storage.Host
			for _, h := range m.hosts {
				if h.ID == msg.HostID {
					curHost = h
					break
				}
			}
			if curHost != nil {
				return m, m.fetchRemoteFilesCmd(curHost, m.fileManager.RemotePath, m.fileManager.RemotePath, m.fileManager.ShowHidden), true
			}
		}
		return m, nil, true

	case FileManagerRefreshMsg:
		if m.fileManager != nil && m.fileManager.HostID == msg.HostID {
			if msg.RefreshRemote && len(m.hosts) > 0 && m.selectedIndex >= 0 && m.selectedIndex < len(m.hosts) {
				curHost := m.hosts[m.selectedIndex]
				return m, m.fetchRemoteFilesCmd(curHost, m.fileManager.RemotePath, m.fileManager.RemotePath, m.fileManager.ShowHidden), true
			}
		}
		return m, nil, true
	}

	return m, nil, false
}

func (m *Model) handleSFTPMessage(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd, _ := m.tryHandleSFTPMessage(msg)
	return model, cmd
}
