package ssh

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/pkg/sftp"
)

func (c *Client) RemoveRemotePath(remotePath string) error {
	c.mu.Lock()
	raw := c.rawClient
	c.mu.Unlock()
	if raw == nil {
		return fmt.Errorf("client is closed")
	}

	sftpClient, err := sftp.NewClient(raw)
	if err != nil {
		return fmt.Errorf("failed to create SFTP subsystem: %w", err)
	}
	defer sftpClient.Close()

	fi, err := sftpClient.Stat(remotePath)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return sftpClient.RemoveDirectory(remotePath)
	}
	return sftpClient.Remove(remotePath)
}

func (c *Client) MkdirRemote(remotePath string) error {
	c.mu.Lock()
	raw := c.rawClient
	c.mu.Unlock()
	if raw == nil {
		return fmt.Errorf("client is closed")
	}

	sftpClient, err := sftp.NewClient(raw)
	if err != nil {
		return fmt.Errorf("failed to create SFTP subsystem: %w", err)
	}
	defer sftpClient.Close()

	return sftpClient.MkdirAll(path.Clean(remotePath))
}

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

func (c *Client) CopyRemote(srcPath, destPath string) error {
	cmd := fmt.Sprintf("cp -r %q %q", srcPath, destPath)
	_, stderr, err := c.ExecWithTimeout(cmd, 30*time.Second)
	if err != nil && len(stderr) > 0 {
		return fmt.Errorf("%v (%s)", err, strings.TrimSpace(string(stderr)))
	}
	return err
}
