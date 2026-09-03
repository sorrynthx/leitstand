package ssh

import (
	"fmt"
	"path"
	"strings"
	"time"
)

func (c *Client) RemoveRemotePath(remotePath string) error {
	sftpClient, err := c.GetSFTPClient()
	if err != nil {
		return fmt.Errorf("failed to get SFTP subsystem: %w", err)
	}

	fi, err := sftpClient.Stat(remotePath)
	if err != nil {
		c.ResetSFTPClient()
		return err
	}

	if fi.IsDir() {
		return sftpClient.RemoveDirectory(remotePath)
	}
	return sftpClient.Remove(remotePath)
}

func (c *Client) MkdirRemote(remotePath string) error {
	sftpClient, err := c.GetSFTPClient()
	if err != nil {
		return fmt.Errorf("failed to get SFTP subsystem: %w", err)
	}

	err = sftpClient.MkdirAll(path.Clean(remotePath))
	if err != nil {
		c.ResetSFTPClient()
	}
	return err
}

func (c *Client) RenameRemote(oldPath, newPath string) error {
	sftpClient, err := c.GetSFTPClient()
	if err != nil {
		return fmt.Errorf("failed to get SFTP subsystem: %w", err)
	}

	err = sftpClient.Rename(oldPath, newPath)
	if err != nil {
		c.ResetSFTPClient()
	}
	return err
}

func (c *Client) DeleteRemote(remotePath string) error {
	sftpClient, err := c.GetSFTPClient()
	if err != nil {
		return fmt.Errorf("failed to get SFTP subsystem: %w", err)
	}

	stat, err := sftpClient.Stat(remotePath)
	if err != nil {
		c.ResetSFTPClient()
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
