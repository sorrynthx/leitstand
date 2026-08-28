package ssh

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pkg/sftp"
)

// FileItem represents a local or remote file metadata entry.
type FileItem struct {
	Name    string
	Path    string
	Size    int64
	IsDir   bool
	ModTime time.Time
	Mode    os.FileMode
}

// FormatSize formats byte size into human readable string (KB, MB, GB).
func (f *FileItem) FormatSize() string {
	if f.IsDir {
		return "<DIR>"
	}
	val := float64(f.Size)
	switch {
	case val >= 1024*1024*1024:
		return fmt.Sprintf("%.1f GB", val/(1024*1024*1024))
	case val >= 1024*1024:
		return fmt.Sprintf("%.1f MB", val/(1024*1024))
	case val >= 1024:
		return fmt.Sprintf("%.1f KB", val/1024)
	default:
		return fmt.Sprintf("%d B", f.Size)
	}
}

// ListLocalDirectory lists files in a local directory path.
func ListLocalDirectory(dirPath string, showHidden bool) ([]*FileItem, error) {
	if dirPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			dirPath = "."
		} else {
			dirPath = home
		}
	}

	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		absPath = dirPath
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	var items []*FileItem

	// Add parent directory item if not root
	parent := filepath.Dir(absPath)
	if parent != absPath {
		items = append(items, &FileItem{
			Name:  "..",
			Path:  parent,
			IsDir: true,
		})
	}

	for _, entry := range entries {
		name := entry.Name()
		if !showHidden && strings.HasPrefix(name, ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		items = append(items, &FileItem{
			Name:    name,
			Path:    filepath.Join(absPath, name),
			Size:    info.Size(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime(),
			Mode:    info.Mode(),
		})
	}

	sortFileItems(items)
	return items, nil
}

// ListRemoteDirectory lists files on the remote server using SFTP.
func (c *Client) ListRemoteDirectory(dirPath string, showHidden bool) ([]*FileItem, string, error) {
	c.mu.Lock()
	raw := c.rawClient
	c.mu.Unlock()
	if raw == nil {
		return nil, "", fmt.Errorf("ssh client not connected")
	}

	sftpClient, err := sftp.NewClient(raw)
	if err != nil {
		return nil, "", fmt.Errorf("failed to initialize sftp subsystem: %w", err)
	}
	defer sftpClient.Close()

	if dirPath == "" || dirPath == "~" || dirPath == "." {
		home, err := sftpClient.Getwd()
		if err == nil && home != "" {
			dirPath = home
		} else {
			dirPath = "."
		}
	}

	realPath, err := sftpClient.RealPath(dirPath)
	if err != nil {
		realPath = dirPath
	}

	entries, err := sftpClient.ReadDir(realPath)
	if err != nil {
		return nil, realPath, err
	}

	var items []*FileItem

	// Add parent directory item
	if realPath != "/" && realPath != "." {
		parent := path.Dir(realPath)
		if parent == "" {
			parent = "/"
		}
		items = append(items, &FileItem{
			Name:  "..",
			Path:  parent,
			IsDir: true,
		})
	}

	for _, info := range entries {
		name := info.Name()
		if !showHidden && strings.HasPrefix(name, ".") {
			continue
		}

		itemPath := path.Join(realPath, name)
		items = append(items, &FileItem{
			Name:    name,
			Path:    itemPath,
			Size:    info.Size(),
			IsDir:   info.IsDir(),
			ModTime: info.ModTime(),
			Mode:    info.Mode(),
		})
	}

	sortFileItems(items)
	return items, realPath, nil
}

// UploadFile transfers a local file to the remote server with progress tracking.
func (c *Client) UploadFile(localPath, remotePath string, onProgress func(transferred, total int64)) error {
	c.mu.Lock()
	raw := c.rawClient
	c.mu.Unlock()
	if raw == nil {
		return fmt.Errorf("ssh client not connected")
	}

	sftpClient, err := sftp.NewClient(raw)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	srcFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	stat, err := srcFile.Stat()
	if err != nil {
		return err
	}
	totalSize := stat.Size()

	// Ensure destination directory exists
	destDir := filepath.Dir(remotePath)
	_ = sftpClient.MkdirAll(destDir)

	dstFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	buf := make([]byte, 64*1024) // 64KB buffer for high throughput
	var transferred int64

	for {
		n, readErr := srcFile.Read(buf)
		if n > 0 {
			_, writeErr := dstFile.Write(buf[:n])
			if writeErr != nil {
				return writeErr
			}
			transferred += int64(n)
			if onProgress != nil {
				onProgress(transferred, totalSize)
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return readErr
		}
	}

	return nil
}

// DownloadFile transfers a remote file to the local PC with progress tracking.
func (c *Client) DownloadFile(remotePath, localPath string, onProgress func(transferred, total int64)) error {
	c.mu.Lock()
	raw := c.rawClient
	c.mu.Unlock()
	if raw == nil {
		return fmt.Errorf("ssh client not connected")
	}

	sftpClient, err := sftp.NewClient(raw)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	srcFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	stat, err := srcFile.Stat()
	if err != nil {
		return err
	}
	totalSize := stat.Size()

	// Ensure local parent directory exists
	localDir := filepath.Dir(localPath)
	_ = os.MkdirAll(localDir, 0755)

	dstFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	buf := make([]byte, 64*1024)
	var transferred int64

	for {
		n, readErr := srcFile.Read(buf)
		if n > 0 {
			_, writeErr := dstFile.Write(buf[:n])
			if writeErr != nil {
				return writeErr
			}
			transferred += int64(n)
			if onProgress != nil {
				onProgress(transferred, totalSize)
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return readErr
		}
	}

	return nil
}

// MkdirRemote creates a new directory on remote server.
func (c *Client) MkdirRemote(remotePath string) error {
	c.mu.Lock()
	raw := c.rawClient
	c.mu.Unlock()
	if raw == nil {
		return fmt.Errorf("ssh client not connected")
	}

	sftpClient, err := sftp.NewClient(raw)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	return sftpClient.MkdirAll(remotePath)
}

// CreateRemoteFile creates an empty file on remote server.
func (c *Client) CreateRemoteFile(remotePath string) error {
	c.mu.Lock()
	raw := c.rawClient
	c.mu.Unlock()
	if raw == nil {
		return fmt.Errorf("ssh client not connected")
	}

	sftpClient, err := sftp.NewClient(raw)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	f, err := sftpClient.Create(remotePath)
	if err != nil {
		return err
	}
	return f.Close()
}

// RenameRemote renames a remote file or directory.
func (c *Client) RenameRemote(oldPath, newPath string) error {
	c.mu.Lock()
	raw := c.rawClient
	c.mu.Unlock()
	if raw == nil {
		return fmt.Errorf("ssh client not connected")
	}

	sftpClient, err := sftp.NewClient(raw)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	return sftpClient.Rename(oldPath, newPath)
}

// DeleteRemote removes a remote file or directory recursively.
func (c *Client) DeleteRemote(remotePath string) error {
	c.mu.Lock()
	raw := c.rawClient
	c.mu.Unlock()
	if raw == nil {
		return fmt.Errorf("ssh client not connected")
	}

	sftpClient, err := sftp.NewClient(raw)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	stat, err := sftpClient.Stat(remotePath)
	if err != nil {
		return err
	}

	if stat.IsDir() {
		_, _, execErr := c.ExecWithTimeout(fmt.Sprintf("rm -rf \"%s\"", remotePath), 10*time.Second)
		if execErr == nil {
			return nil
		}
		return sftpClient.RemoveDirectory(remotePath)
	}
	return sftpClient.Remove(remotePath)
}

func sortFileItems(items []*FileItem) {
	sort.Slice(items, func(i, j int) bool {
		// ".." always on top
		if items[i].Name == ".." {
			return true
		}
		if items[j].Name == ".." {
			return false
		}
		// Directories first, then files alphabetically
		if items[i].IsDir && !items[j].IsDir {
			return true
		}
		if !items[i].IsDir && items[j].IsDir {
			return false
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
}
