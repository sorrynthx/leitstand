package ssh

import (
	"fmt"
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

	return items, realPath, nil
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
