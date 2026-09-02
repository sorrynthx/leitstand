package ssh

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/sftp"
)

func (c *Client) UploadFile(localPath, remotePath string, onProgress func(transferred, total int64)) error {
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

	srcFile, err := os.Open(localPath)
	if err != nil {
		return err
	}

	fi, err := srcFile.Stat()
	if err != nil {
		srcFile.Close()
		return err
	}

	if fi.IsDir() {
		srcFile.Close()
		return c.uploadDirRecursive(sftpClient, localPath, remotePath, onProgress)
	}
	srcFile.Close()

	return uploadSingleFile(sftpClient, localPath, remotePath, onProgress)
}

func uploadSingleFile(sftpClient *sftp.Client, localPath, remotePath string, onProgress func(transferred, total int64)) error {
	srcFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	fi, err := srcFile.Stat()
	if err != nil {
		return err
	}
	totalSize := fi.Size()

	destDir := path.Dir(remotePath)
	_ = sftpClient.MkdirAll(destDir)

	dstFile, err := sftpClient.Create(remotePath)
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

func (c *Client) uploadDirRecursive(sftpClient *sftp.Client, localDir, remoteDir string, onProgress func(transferred, total int64)) error {
	_ = sftpClient.MkdirAll(remoteDir)
	entries, err := os.ReadDir(localDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		subLocal := filepath.Join(localDir, entry.Name())
		subRemote := path.Join(remoteDir, entry.Name())
		if entry.IsDir() {
			if err := c.uploadDirRecursive(sftpClient, subLocal, subRemote, onProgress); err != nil {
				return err
			}
		} else {
			if err := uploadSingleFile(sftpClient, subLocal, subRemote, onProgress); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) DownloadFile(remotePath, localPath string, onProgress func(transferred, total int64)) error {
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

	srcFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return err
	}

	fi, err := srcFile.Stat()
	if err != nil {
		srcFile.Close()
		return err
	}

	if fi.IsDir() {
		srcFile.Close()
		return c.downloadDirRecursive(sftpClient, remotePath, localPath, onProgress)
	}
	srcFile.Close()

	return downloadSingleFile(sftpClient, remotePath, localPath, onProgress)
}

func downloadSingleFile(sftpClient *sftp.Client, remotePath, localPath string, onProgress func(transferred, total int64)) error {
	srcFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	fi, err := srcFile.Stat()
	if err != nil {
		return err
	}
	totalSize := fi.Size()

	destDir := filepath.Dir(localPath)
	_ = os.MkdirAll(destDir, 0755)

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

func (c *Client) downloadDirRecursive(sftpClient *sftp.Client, remoteDir, localDir string, onProgress func(transferred, total int64)) error {
	_ = os.MkdirAll(localDir, 0755)
	entries, err := sftpClient.ReadDir(remoteDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		subRemote := path.Join(remoteDir, entry.Name())
		subLocal := filepath.Join(localDir, entry.Name())
		if entry.IsDir() {
			if err := c.downloadDirRecursive(sftpClient, subRemote, subLocal, onProgress); err != nil {
				return err
			}
		} else {
			if err := downloadSingleFile(sftpClient, subRemote, subLocal, onProgress); err != nil {
				return err
			}
		}
	}
	return nil
}
