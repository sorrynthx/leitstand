package tui

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type FileOpResultMsg struct {
	HostID  int64
	IsLocal bool
	Msg     string
	Err     error
}

func (m *Model) executeFileOpCmd(op FileOpActionMsg) tea.Cmd {
	return func() tea.Msg {
		if op.IsLocal {
			return m.executeLocalFileOp(op)
		}
		return m.executeRemoteFileOp(op)
	}
}

func (m *Model) executeLocalFileOp(op FileOpActionMsg) tea.Msg {
	switch op.OpType {
	case "mkdir":
		targetPath := filepath.Join(op.DirPath, op.NewName)
		err := os.MkdirAll(targetPath, 0755)
		if err != nil {
			return FileOpResultMsg{HostID: op.HostID, IsLocal: true, Err: fmt.Errorf("local mkdir failed: %w", err)}
		}
		return FileOpResultMsg{HostID: op.HostID, IsLocal: true, Msg: fmt.Sprintf("✨ Local folder '%s' created", op.NewName)}

	case "touch":
		targetPath := filepath.Join(op.DirPath, op.NewName)
		f, err := os.Create(targetPath)
		if err != nil {
			return FileOpResultMsg{HostID: op.HostID, IsLocal: true, Err: fmt.Errorf("local touch failed: %w", err)}
		}
		_ = f.Close()
		return FileOpResultMsg{HostID: op.HostID, IsLocal: true, Msg: fmt.Sprintf("✨ Local file '%s' created", op.NewName)}

	case "rename":
		oldPath := filepath.Join(op.DirPath, op.OldName)
		newPath := filepath.Join(op.DirPath, op.NewName)
		err := os.Rename(oldPath, newPath)
		if err != nil {
			return FileOpResultMsg{HostID: op.HostID, IsLocal: true, Err: fmt.Errorf("local rename failed: %w", err)}
		}
		return FileOpResultMsg{HostID: op.HostID, IsLocal: true, Msg: fmt.Sprintf("✨ Local file renamed to '%s'", op.NewName)}

	case "delete":
		deletedCount := 0
		for _, p := range op.TargetPaths {
			err := os.RemoveAll(p)
			if err == nil {
				deletedCount++
			}
		}
		return FileOpResultMsg{HostID: op.HostID, IsLocal: true, Msg: fmt.Sprintf("🗑️ Deleted %d local item(s)", deletedCount)}
	}

	return FileOpResultMsg{HostID: op.HostID, IsLocal: true, Err: fmt.Errorf("unknown local op type: %s", op.OpType)}
}

func (m *Model) executeFileManagerQuickCmd(cmdMsg FileManagerQuickCmdMsg) tea.Cmd {
	return func() tea.Msg {
		if cmdMsg.IsLocal {
			return m.executeLocalQuickCmd(cmdMsg)
		}
		return m.executeRemoteQuickCmd(cmdMsg)
	}
}

func (m *Model) executeLocalQuickCmd(cmdMsg FileManagerQuickCmdMsg) tea.Msg {
	parts := strings.Fields(cmdMsg.Command)
	if len(parts) > 0 && parts[0] == "cd" {
		targetDir := cmdMsg.DirPath
		if len(parts) > 1 {
			targetDir = parts[1]
			if !filepath.IsAbs(targetDir) {
				targetDir = filepath.Join(cmdMsg.DirPath, targetDir)
			}
		}
		cleanPath := filepath.Clean(targetDir)
		if info, err := os.Stat(cleanPath); err == nil && info.IsDir() {
			return FileManagerQuickCmdResultMsg{
				HostID:  cmdMsg.HostID,
				IsLocal: true,
				Command: cmdMsg.Command,
				OldCWD:  cmdMsg.DirPath,
				NewCWD:  cleanPath,
				Output:  fmt.Sprintf("Changed local directory to: %s", cleanPath),
			}
		}
		return FileManagerQuickCmdResultMsg{
			HostID:  cmdMsg.HostID,
			IsLocal: true,
			Command: cmdMsg.Command,
			OldCWD:  cmdMsg.DirPath,
			Err:     fmt.Errorf("directory not found: %s", targetDir),
		}
	}

	return FileManagerQuickCmdResultMsg{
		HostID:  cmdMsg.HostID,
		IsLocal: true,
		Command: cmdMsg.Command,
		OldCWD:  cmdMsg.DirPath,
		Output:  fmt.Sprintf("Executed local command: %s (Simulated)", cmdMsg.Command),
	}
}

func copyLocalFileOrDir(src, dst string) error {
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return copyLocalDir(src, dst)
	}
	return copyLocalFile(src, dst)
}

func copyLocalFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyLocalDir(src, dst string) error {
	_ = os.MkdirAll(dst, 0755)
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		s := filepath.Join(src, entry.Name())
		d := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyLocalDir(s, d); err != nil {
				return err
			}
		} else {
			if err := copyLocalFile(s, d); err != nil {
				return err
			}
		}
	}
	return nil
}
