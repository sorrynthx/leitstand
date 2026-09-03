package ssh

import (
	"context"
	"io"
	"leitstand/internal/logger"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/sftp"
)

func (c *Client) UploadFile(ctx context.Context, localPath, remotePath string, onProgress func(transferred, total int64)) error {
	sess, err := c.createDedicatedSFTPSession(ctx)
	if err != nil {
		return err
	}
	defer sess.Close()
	defer watchCancel(ctx, sess)()
	sftpClient := sess.Client

	srcFile, err := os.Open(localPath)
	if err != nil {
		return err
	}

	fi, err := srcFile.Stat()
	srcFile.Close()
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return c.uploadDirRecursive(ctx, sftpClient, localPath, remotePath, onProgress)
	}

	return uploadSingleFile(ctx, sftpClient, localPath, remotePath, onProgress)
}

func uploadSingleFile(ctx context.Context, sftpClient *sftp.Client, localPath, remotePath string, onProgress func(transferred, total int64)) error {
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
		select {
		case <-ctx.Done():
			logger.Warnf("[SFTP] uploadSingleFile interrupted by cancel context: %v", ctx.Err())
			return ctx.Err()
		default:
		}

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

func (c *Client) uploadDirRecursive(ctx context.Context, sftpClient *sftp.Client, localDir, remoteDir string, onProgress func(transferred, total int64)) error {
	_ = sftpClient.MkdirAll(remoteDir)
	entries, err := os.ReadDir(localDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		subLocal := filepath.Join(localDir, entry.Name())
		subRemote := path.Join(remoteDir, entry.Name())
		if entry.IsDir() {
			if err := c.uploadDirRecursive(ctx, sftpClient, subLocal, subRemote, onProgress); err != nil {
				return err
			}
		} else {
			if err := uploadSingleFile(ctx, sftpClient, subLocal, subRemote, onProgress); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) DownloadFile(ctx context.Context, remotePath, localPath string, onProgress func(transferred, total int64)) error {
	sess, err := c.createDedicatedSFTPSession(ctx)
	if err != nil {
		return err
	}
	defer sess.Close()
	defer watchCancel(ctx, sess)()
	sftpClient := sess.Client

	srcFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return err
	}

	fi, err := srcFile.Stat()
	srcFile.Close()
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return c.downloadDirRecursive(ctx, sftpClient, remotePath, localPath, onProgress)
	}

	return downloadSingleFile(ctx, sftpClient, remotePath, localPath, onProgress)
}

func downloadSingleFile(ctx context.Context, sftpClient *sftp.Client, remotePath, localPath string, onProgress func(transferred, total int64)) error {
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
		select {
		case <-ctx.Done():
			logger.Warnf("[SFTP] downloadSingleFile interrupted by cancel context: %v", ctx.Err())
			return ctx.Err()
		default:
		}

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

func (c *Client) downloadDirRecursive(ctx context.Context, sftpClient *sftp.Client, remoteDir, localDir string, onProgress func(transferred, total int64)) error {
	_ = os.MkdirAll(localDir, 0755)
	entries, err := sftpClient.ReadDir(remoteDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		subRemote := path.Join(remoteDir, entry.Name())
		subLocal := filepath.Join(localDir, entry.Name())
		if entry.IsDir() {
			if err := c.downloadDirRecursive(ctx, sftpClient, subRemote, subLocal, onProgress); err != nil {
				return err
			}
		} else {
			if err := downloadSingleFile(ctx, sftpClient, subRemote, subLocal, onProgress); err != nil {
				return err
			}
		}
	}
	return nil
}

func watchCancel(ctx context.Context, closer io.Closer) func() {
	ch := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = closer.Close()
		case <-ch:
		}
	}()
	return func() { close(ch) }
}
