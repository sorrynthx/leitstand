package tui

import (
	"context"
	"fmt"
	"io"
	"leitstand/internal/i18n"
	"leitstand/internal/ssh"
	"leitstand/internal/storage"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// RemoteFileListMsg carries directory entries retrieved from remote server.
type RemoteFileListMsg struct {
	HostID  int64
	Path    string
	OldPath string
	Items   []*ssh.FileItem
	Err     error
}

// FileOpResultMsg carries the outcome of file operations (mkdir, touch, rename, delete).
type FileOpResultMsg struct {
	HostID  int64
	IsLocal bool
	Msg     string
	Err     error
}

// openFileManagerCmd launches the dual-pane SFTP manager.
func (m *Model) openFileManagerCmd(host *storage.Host) tea.Cmd {
	m.showFileManager = true
	m.fileManager = NewFileManagerModal(host.ID, host.Name, "", ".")

	return tea.Batch(
		m.fileManager.RefreshLocalCmd(),
		m.fetchRemoteFilesCmd(host, ".", ".", m.fileManager.ShowHidden),
	)
}

// fetchRemoteFilesCmd retrieves remote server directory contents.
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

// startFileTransferCmd executes background batch upload/download or cross-pane move.
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
			fileName := filepath.Base(srcPath)
			destPath := filepath.ToSlash(filepath.Join(action.DestDirPath, fileName))

			startTime := time.Now()
			lastTime := startTime
			var lastBytes int64

			progressCb := func(transferred, total int64) {
				now := time.Now()
				elapsed := now.Sub(lastTime).Seconds()
				var speed float64
				if elapsed >= 0.2 {
					speed = float64(transferred-lastBytes) / elapsed
					lastTime = now
					lastBytes = transferred
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
					BytesPerSec:  speed,
					IsDone:       false,
					IsMove:       action.IsMove,
					Err:          nil,
					msgChan:      msgChan,
				}:
				}
			}

			var transferErr error
			if action.IsUpload {
				transferErr = client.UploadFile(srcPath, destPath, progressCb)
			} else {
				transferErr = client.DownloadFile(srcPath, destPath, progressCb)
			}

			// If move is requested and transfer succeeded, delete source file
			if transferErr == nil && action.IsMove {
				if action.IsUpload {
					_ = os.RemoveAll(srcPath)
				} else {
					_ = client.DeleteRemote(srcPath)
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

// executeFileOpCmd performs local or remote file operations (mkdir, touch, rename, delete).
func (m *Model) executeFileOpCmd(action FileOpActionMsg) tea.Cmd {
	return func() tea.Msg {
		var host *storage.Host
		for _, h := range m.hosts {
			if h.ID == action.HostID {
				host = h
				break
			}
		}

		if action.IsLocal {
			// Local Operations
			switch action.OpType {
			case "mkdir":
				target := filepath.Join(action.DirPath, action.NewName)
				err := os.MkdirAll(target, 0755)
				if err != nil {
					return FileOpResultMsg{HostID: action.HostID, IsLocal: true, Err: err}
				}
				return FileOpResultMsg{HostID: action.HostID, IsLocal: true, Msg: fmt.Sprintf(i18n.T("sftp_mkdir_done"), action.NewName)}

			case "touch":
				target := filepath.Join(action.DirPath, action.NewName)
				f, err := os.Create(target)
				if err != nil {
					return FileOpResultMsg{HostID: action.HostID, IsLocal: true, Err: err}
				}
				_ = f.Close()
				return FileOpResultMsg{HostID: action.HostID, IsLocal: true, Msg: fmt.Sprintf(i18n.T("sftp_touch_done"), action.NewName)}

			case "rename":
				oldPath := filepath.Join(action.DirPath, action.OldName)
				newPath := filepath.Join(action.DirPath, action.NewName)
				err := os.Rename(oldPath, newPath)
				if err != nil {
					return FileOpResultMsg{HostID: action.HostID, IsLocal: true, Err: err}
				}
				return FileOpResultMsg{HostID: action.HostID, IsLocal: true, Msg: fmt.Sprintf(i18n.T("sftp_rename_done"), action.NewName)}

			case "delete":
				for _, p := range action.TargetPaths {
					_ = os.RemoveAll(p)
				}
				return FileOpResultMsg{HostID: action.HostID, IsLocal: true, Msg: fmt.Sprintf(i18n.T("sftp_delete_done"), filepath.Base(action.TargetPaths[0]))}

			case "move":
				destDir := action.NewName
				_ = os.MkdirAll(destDir, 0755)
				for _, p := range action.TargetPaths {
					destFile := filepath.Join(destDir, filepath.Base(p))
					err := os.Rename(p, destFile)
					if err != nil {
						return FileOpResultMsg{HostID: action.HostID, IsLocal: true, Err: err}
					}
				}
				return FileOpResultMsg{HostID: action.HostID, IsLocal: true, Msg: fmt.Sprintf(i18n.T("sftp_paste_done_move"), len(action.TargetPaths))}

			case "copy_same":
				destDir := action.NewName
				_ = os.MkdirAll(destDir, 0755)
				for _, p := range action.TargetPaths {
					destFile := filepath.Join(destDir, filepath.Base(p))
					srcF, err := os.Open(p)
					if err == nil {
						destF, cErr := os.Create(destFile)
						if cErr == nil {
							_, _ = io.Copy(destF, srcF)
							_ = destF.Close()
						}
						_ = srcF.Close()
					}
				}
				return FileOpResultMsg{HostID: action.HostID, IsLocal: true, Msg: fmt.Sprintf(i18n.T("sftp_paste_done_copy"), len(action.TargetPaths))}
			}
			return nil
		}

		// Remote Operations
		if m.isDemo {
			return FileOpResultMsg{HostID: action.HostID, IsLocal: false, Msg: fmt.Sprintf("[Demo] %s %s completed.", action.OpType, action.NewName)}
		}

		if host == nil {
			return FileOpResultMsg{HostID: action.HostID, IsLocal: false, Err: fmt.Errorf("host not found")}
		}

		client, err := m.getSSHClient(host)
		if err != nil {
			return FileOpResultMsg{HostID: action.HostID, IsLocal: false, Err: err}
		}

		switch action.OpType {
		case "mkdir":
			target := filepath.ToSlash(filepath.Join(action.DirPath, action.NewName))
			err := client.MkdirRemote(target)
			if err != nil {
				return FileOpResultMsg{HostID: action.HostID, IsLocal: false, Err: err}
			}
			return FileOpResultMsg{HostID: action.HostID, IsLocal: false, Msg: fmt.Sprintf(i18n.T("sftp_mkdir_done"), action.NewName)}

		case "touch":
			target := filepath.ToSlash(filepath.Join(action.DirPath, action.NewName))
			err := client.CreateRemoteFile(target)
			if err != nil {
				return FileOpResultMsg{HostID: action.HostID, IsLocal: false, Err: err}
			}
			return FileOpResultMsg{HostID: action.HostID, IsLocal: false, Msg: fmt.Sprintf(i18n.T("sftp_touch_done"), action.NewName)}

		case "rename":
			oldPath := filepath.ToSlash(filepath.Join(action.DirPath, action.OldName))
			newPath := filepath.ToSlash(filepath.Join(action.DirPath, action.NewName))
			err := client.RenameRemote(oldPath, newPath)
			if err != nil {
				return FileOpResultMsg{HostID: action.HostID, IsLocal: false, Err: err}
			}
			return FileOpResultMsg{HostID: action.HostID, IsLocal: false, Msg: fmt.Sprintf(i18n.T("sftp_rename_done"), action.NewName)}

		case "delete":
			for _, p := range action.TargetPaths {
				_ = client.DeleteRemote(p)
			}
			return FileOpResultMsg{HostID: action.HostID, IsLocal: false, Msg: fmt.Sprintf(i18n.T("sftp_delete_done"), filepath.Base(action.TargetPaths[0]))}

		case "move":
			destDir := action.NewName
			_ = client.MkdirRemote(destDir)
			for _, p := range action.TargetPaths {
				destFile := path.Join(destDir, path.Base(p))
				err := client.RenameRemote(p, destFile)
				if err != nil {
					return FileOpResultMsg{HostID: action.HostID, IsLocal: false, Err: err}
				}
			}
			return FileOpResultMsg{HostID: action.HostID, IsLocal: false, Msg: fmt.Sprintf(i18n.T("sftp_paste_done_move"), len(action.TargetPaths))}

		case "copy_same":
			destDir := action.NewName
			_ = client.MkdirRemote(destDir)
			for _, p := range action.TargetPaths {
				destFile := path.Join(destDir, path.Base(p))
				cpCmd := fmt.Sprintf("cp -r %q %q", p, destFile)
				_, _, _ = client.Exec(cpCmd)
			}
			return FileOpResultMsg{HostID: action.HostID, IsLocal: false, Msg: fmt.Sprintf(i18n.T("sftp_paste_done_copy"), len(action.TargetPaths))}
		}

		return nil
	}
}

// executeFileManagerQuickCmd executes an inline shell command directly in the active folder.
func (m *Model) executeFileManagerQuickCmd(action FileManagerQuickCmdMsg) tea.Cmd {
	return func() tea.Msg {
		var host *storage.Host
		for _, h := range m.hosts {
			if h.ID == action.HostID {
				host = h
				break
			}
		}

		trimmedCmd := strings.TrimSpace(action.Command)
		if action.IsLocal {
			// Local Command Execution
			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = exec.Command("powershell", "-NoProfile", "-Command", fmt.Sprintf("Set-Location '%s'; %s; (Get-Location).Path", action.DirPath, trimmedCmd))
			} else {
				cmd = exec.Command("sh", "-c", fmt.Sprintf("cd %s 2>/dev/null || cd \"$HOME\" ; %s ; pwd", action.DirPath, trimmedCmd))
			}
			outBytes, err := cmd.CombinedOutput()
			outStr := string(outBytes)
			lines := strings.Split(strings.TrimSpace(outStr), "\n")
			newCWD := action.DirPath
			if len(lines) > 0 {
				lastLine := strings.TrimSpace(lines[len(lines)-1])
				if filepath.IsAbs(lastLine) {
					newCWD = lastLine
					outStr = strings.TrimSpace(strings.Join(lines[:len(lines)-1], "\n"))
				}
			}
			return FileManagerQuickCmdResultMsg{
				HostID:  action.HostID,
				IsLocal: true,
				Command: trimmedCmd,
				OldCWD:  action.DirPath,
				NewCWD:  newCWD,
				Output:  outStr,
				Err:     err,
			}
		}

		// Remote Command Execution
		if m.isDemo {
			out := SimulateDemoCmd(trimmedCmd, host.Name)
			return FileManagerQuickCmdResultMsg{
				HostID:  action.HostID,
				IsLocal: false,
				Command: trimmedCmd,
				OldCWD:  action.DirPath,
				NewCWD:  action.DirPath,
				Output:  out,
			}
		}

		if host == nil {
			return FileManagerQuickCmdResultMsg{HostID: action.HostID, IsLocal: false, Command: trimmedCmd, Err: fmt.Errorf("host not found")}
		}

		client, err := m.getSSHClient(host)
		if err != nil {
			return FileManagerQuickCmdResultMsg{HostID: action.HostID, IsLocal: false, Command: trimmedCmd, Err: err}
		}

		wrappedCmd := fmt.Sprintf("cd %s 2>/dev/null || cd \"$HOME\" ; %s ; echo '__LEITSTAND_PWD__' ; pwd", action.DirPath, trimmedCmd)
		stdout, stderr, err := client.Exec(wrappedCmd)

		outStr := string(stdout)
		errStr := string(stderr)
		if errStr != "" {
			if outStr != "" {
				outStr += "\n"
			}
			outStr += errStr
		}

		newCWD := action.DirPath
		displayOutput := outStr
		if idx := strings.Index(outStr, "__LEITSTAND_PWD__"); idx != -1 {
			displayOutput = strings.TrimSpace(outStr[:idx])
			pwdPart := strings.TrimSpace(outStr[idx+len("__LEITSTAND_PWD__"):])
			if pwdPart != "" && strings.HasPrefix(pwdPart, "/") {
				newCWD = pwdPart
			}
		}

		return FileManagerQuickCmdResultMsg{
			HostID:  action.HostID,
			IsLocal: false,
			Command: trimmedCmd,
			OldCWD:  action.DirPath,
			NewCWD:  newCWD,
			Output:  displayOutput,
			Err:     err,
		}
	}
}
