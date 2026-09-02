package tui

import (
	"context"
	"fmt"
	"leitstand/internal/ssh"
	"leitstand/internal/storage"
	"os"
	"path"
	"path/filepath"

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
	_ = cancel

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

			progressCb := func(transferred, total int64) {
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
					transferErr = client.UploadFile(srcPath, destPath, progressCb)
					if transferErr == nil && action.IsMove {
						_ = os.RemoveAll(srcPath)
					}
				} else {
					destPath := filepath.Join(action.DestDirPath, fileName)
					transferErr = client.DownloadFile(srcPath, destPath, progressCb)
					if transferErr == nil && action.IsMove {
						_ = client.DeleteRemote(srcPath)
					}
				}
			}

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
				Err:          transferErr,
				msgChan:      msgChan,
			}

			if transferErr != nil {
				return
			}
		}
	}()

	return listenStreamCmd(msgChan)
}
