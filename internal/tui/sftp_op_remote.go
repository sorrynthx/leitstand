package tui

import (
	"fmt"
	"leitstand/internal/storage"
	"path"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) executeRemoteFileOp(op FileOpActionMsg) tea.Msg {
	var curHost *storage.Host
	for _, h := range m.hosts {
		if h.ID == op.HostID {
			curHost = h
			break
		}
	}
	if curHost == nil {
		return FileOpResultMsg{HostID: op.HostID, IsLocal: false, Err: fmt.Errorf("host not found")}
	}

	if m.isDemo {
		return FileOpResultMsg{HostID: op.HostID, IsLocal: false, Msg: fmt.Sprintf("✨ Demo Remote operation '%s' completed", op.OpType)}
	}

	sftpClient, err := m.getSFTPClient(curHost)
	if err != nil {
		return FileOpResultMsg{HostID: op.HostID, IsLocal: false, Err: fmt.Errorf("SFTP client error: %w", err)}
	}

	switch op.OpType {
	case "mkdir":
		targetPath := path.Join(op.DirPath, op.NewName)
		err := sftpClient.MkdirAll(targetPath)
		if err != nil {
			return FileOpResultMsg{HostID: op.HostID, IsLocal: false, Err: fmt.Errorf("remote mkdir failed: %w", err)}
		}
		return FileOpResultMsg{HostID: op.HostID, IsLocal: false, Msg: fmt.Sprintf("✨ Remote folder '%s' created", op.NewName)}

	case "touch":
		targetPath := path.Join(op.DirPath, op.NewName)
		f, err := sftpClient.Create(targetPath)
		if err != nil {
			return FileOpResultMsg{HostID: op.HostID, IsLocal: false, Err: fmt.Errorf("remote touch failed: %w", err)}
		}
		_ = f.Close()
		return FileOpResultMsg{HostID: op.HostID, IsLocal: false, Msg: fmt.Sprintf("✨ Remote file '%s' created", op.NewName)}

	case "rename":
		oldPath := path.Join(op.DirPath, op.OldName)
		newPath := path.Join(op.DirPath, op.NewName)
		err := sftpClient.Rename(oldPath, newPath)
		if err != nil {
			return FileOpResultMsg{HostID: op.HostID, IsLocal: false, Err: fmt.Errorf("remote rename failed: %w", err)}
		}
		return FileOpResultMsg{HostID: op.HostID, IsLocal: false, Msg: fmt.Sprintf("✨ Remote file renamed to '%s'", op.NewName)}

	case "delete":
		deletedCount := 0
		for _, p := range op.TargetPaths {
			err := sftpClient.RemoveAll(p)
			if err == nil {
				deletedCount++
			}
		}
		return FileOpResultMsg{HostID: op.HostID, IsLocal: false, Msg: fmt.Sprintf("🗑️ Deleted %d remote item(s)", deletedCount)}
	}

	return FileOpResultMsg{HostID: op.HostID, IsLocal: false, Err: fmt.Errorf("unknown remote op type: %s", op.OpType)}
}

func (m *Model) executeRemoteQuickCmd(cmdMsg FileManagerQuickCmdMsg) tea.Msg {
	var curHost *storage.Host
	for _, h := range m.hosts {
		if h.ID == cmdMsg.HostID {
			curHost = h
			break
		}
	}
	if curHost == nil {
		return FileManagerQuickCmdResultMsg{HostID: cmdMsg.HostID, IsLocal: false, Command: cmdMsg.Command, Err: fmt.Errorf("host not found")}
	}

	if m.isDemo {
		output := SimulateDemoCmd(cmdMsg.Command, curHost.Name)
		return FileManagerQuickCmdResultMsg{
			HostID:  cmdMsg.HostID,
			IsLocal: false,
			Command: cmdMsg.Command,
			OldCWD:  cmdMsg.DirPath,
			Output:  output,
		}
	}

	sshClient, err := m.getSSHClient(curHost)
	if err != nil {
		return FileManagerQuickCmdResultMsg{HostID: cmdMsg.HostID, IsLocal: false, Command: cmdMsg.Command, Err: fmt.Errorf("SSH connection failed: %w", err)}
	}

	fullCmd := fmt.Sprintf("cd %q 2>/dev/null; (%s); cd %q 2>/dev/null && pwd", cmdMsg.DirPath, cmdMsg.Command, cmdMsg.DirPath)
	stdoutBytes, _, err := sshClient.ExecWithTimeout(fullCmd, 15*time.Second)
	output := string(stdoutBytes)
	if err != nil {
		return FileManagerQuickCmdResultMsg{
			HostID:  cmdMsg.HostID,
			IsLocal: false,
			Command: cmdMsg.Command,
			OldCWD:  cmdMsg.DirPath,
			Output:  output,
			Err:     err,
		}
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	newCWD := cmdMsg.DirPath
	if len(lines) > 0 {
		possibleCWD := strings.TrimSpace(lines[len(lines)-1])
		if strings.HasPrefix(possibleCWD, "/") {
			newCWD = possibleCWD
			lines = lines[:len(lines)-1]
			output = strings.Join(lines, "\n")
		}
	}

	return FileManagerQuickCmdResultMsg{
		HostID:  cmdMsg.HostID,
		IsLocal: false,
		Command: cmdMsg.Command,
		OldCWD:  cmdMsg.DirPath,
		NewCWD:  newCWD,
		Output:  output,
	}
}
