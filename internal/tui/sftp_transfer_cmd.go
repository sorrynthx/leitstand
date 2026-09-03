package tui

import (
	"context"
	"fmt"
	"leitstand/internal/logger"
	"leitstand/internal/ssh"
	"leitstand/internal/storage"
	"os"
	"path"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type RemoteFileListMsg struct {
	HostID  int64
	Path    string
	OldPath string
	Items   []*ssh.FileItem
	Err     error
}

func (m *Model) openFileManagerCmd(host *storage.Host) tea.Cmd {
	m.showFileManager = true
	if m.fileManager != nil && m.fileManager.HostID == host.ID && m.fileManager.IsTransferring {
		m.fileManager.IsTransferBackground = false
		return nil
	}

	m.fileManager = NewFileManagerModal(host.ID, host.Name, "", ".")

	return tea.Batch(
		m.fileManager.RefreshLocalCmd(),
		m.fetchRemoteFilesCmd(host, ".", ".", m.fileManager.ShowHidden),
	)
}

func (m *Model) fetchRemoteFilesCmd(host *storage.Host, remotePath, oldPath string, showHidden bool) tea.Cmd {
	return func() tea.Msg {
		if m.isDemo {
			items := GetDemoRemoteFiles(remotePath)
			return RemoteFileListMsg{
				HostID:  host.ID,
				Path:    remotePath,
				OldPath: oldPath,
				Items:   items,
				Err:     nil,
			}
		}

		client, err := m.getSSHClient(host)
		if err != nil {
			return RemoteFileListMsg{HostID: host.ID, Path: remotePath, OldPath: oldPath, Err: err}
		}

		items, actualPath, err := client.ListRemoteDirectory(remotePath, showHidden)
		return RemoteFileListMsg{
			HostID:  host.ID,
			Path:    actualPath,
			OldPath: oldPath,
			Items:   items,
			Err:     err,
		}
	}
}

func (m *Model) startFileTransferCmd(action TransferActionMsg) tea.Cmd {
	msgChan := make(chan tea.Msg, 50)
	ctx, cancel := context.WithCancel(m.ctx)
	logger.Infof("[SFTP] Starting transfer (upload=%v, files=%d, dest=%s)", action.IsUpload, len(action.SrcPaths), action.DestDirPath)
	if m.fileManager != nil {
		m.fileManager.TransferCancel = cancel
		m.fileManager.IsTransferCanceled = false
		m.fileManager.ShowTransferCancelPrompt = false
		m.fileManager.IsTransferBackground = false
	}

	if m.isDemo {
		go SimulateDemoFileTransfer(ctx, action.HostID, action.IsUpload, action.SrcPaths, action.DestDirPath, msgChan)
		return listenStreamCmd(msgChan)
	}

	var host *storage.Host
	for _, h := range m.hosts {
		if h.ID == action.HostID {
			host = h
			break
		}
	}
	if host == nil {
		return func() tea.Msg {
			return FileTransferProgressMsg{HostID: action.HostID, IsDone: true, Err: fmt.Errorf("host not found")}
		}
	}

	client, err := m.getSSHClient(host)
	if err != nil {
		return func() tea.Msg {
			return FileTransferProgressMsg{HostID: action.HostID, IsDone: true, Err: err}
		}
	}

	go func() {
		defer close(msgChan)
		totalFiles := len(action.SrcPaths)

		for i, srcPath := range action.SrcPaths {
			fileName := path.Base(srcPath)
			if action.IsUpload {
				fileName = filepath.Base(srcPath)
			}
			var transferErr error

			startTime := time.Now()
			lastTime := startTime
			lastBytes := int64(0)
			var currentBps float64

			progressCb := func(transferred, total int64) {
				now := time.Now()
				elapsed := now.Sub(lastTime)
				if elapsed >= 200*time.Millisecond {
					currentBps = float64(transferred-lastBytes) / elapsed.Seconds()
					lastTime = now
					lastBytes = transferred
				} else if currentBps <= 0 && now.Sub(startTime).Seconds() > 0.1 {
					currentBps = float64(transferred) / now.Sub(startTime).Seconds()
				}

				select {
				case <-ctx.Done():
					return
				case msgChan <- FileTransferProgressMsg{
					HostID:       action.HostID,
					FileName:     fileName,
					FileIndex:    i + 1,
					FileTotal:    totalFiles,
					CurrentBytes: transferred,
					TotalBytes:   total,
					BytesPerSec:  currentBps,
					IsDone:       false,
					IsMove:       action.IsMove,
					msgChan:      msgChan,
				}:
				}
			}

			if action.IsSameHost {
				if action.IsLocalOp {
					destPath := filepath.Join(action.DestDirPath, fileName)
					if action.IsMove {
						transferErr = os.Rename(srcPath, destPath)
					} else {
						transferErr = copyLocalFileOrDir(srcPath, destPath)
					}
				} else {
					destPath := path.Join(action.DestDirPath, fileName)
					if action.IsMove {
						transferErr = client.RenameRemote(srcPath, destPath)
					} else {
						transferErr = client.CopyRemote(srcPath, destPath)
					}
				}
			} else {
				if action.IsUpload {
					destPath := path.Join(action.DestDirPath, fileName)
					transferErr = client.UploadFile(ctx, srcPath, destPath, progressCb)
					if transferErr == nil && action.IsMove {
						_ = os.RemoveAll(srcPath)
					}
				} else {
					destPath := filepath.Join(action.DestDirPath, fileName)
					transferErr = client.DownloadFile(ctx, srcPath, destPath, progressCb)
					if transferErr == nil && action.IsMove {
						_ = client.DeleteRemote(srcPath)
					}
				}
			}

			if transferErr != nil {
				logger.Warnf("[SFTP] Transfer error/cancel on %s: %v", fileName, transferErr)
				msgChan <- FileTransferProgressMsg{
					HostID:   action.HostID,
					FileName: fileName,
					IsDone:   true,
					Err:      transferErr,
				}
				return
			}

			logger.Infof("[SFTP] Completed transfer for %s (%d/%d)", fileName, i+1, totalFiles)
			isLast := (i == totalFiles-1)
			msgChan <- FileTransferProgressMsg{
				HostID:       action.HostID,
				FileName:     fileName,
				FileIndex:    i + 1,
				FileTotal:    totalFiles,
				CurrentBytes: 100,
				TotalBytes:   100,
				IsDone:       isLast,
				IsMove:       action.IsMove,
				Err:          nil,
				msgChan:      msgChan,
			}
		}
	}()

	return listenStreamCmd(msgChan)
}
