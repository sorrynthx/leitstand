package ssh

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Client wraps an active ssh.Client connection.
type Client struct {
	mu            sync.Mutex
	rawClient     *ssh.Client
	sftpClient    *sftp.Client
	address       string
	createdAt     time.Time
	keepAliveStop chan struct{}
}

// NewClient connects to the remote SSH server using the provided configuration.
func NewClient(address string, config *ssh.ClientConfig) (*Client, error) {
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("ssh dial failed to %s: %w", address, err)
	}

	c := &Client{
		rawClient: client,
		address:   address,
		createdAt: time.Now(),
	}
	c.StartKeepAlive(30 * time.Second)
	return c, nil
}

// Exec runs a non-interactive remote command and returns stdout, stderr with a 15-second default timeout.
func (c *Client) Exec(cmd string) (stdout []byte, stderr []byte, err error) {
	return c.ExecWithTimeout(cmd, 15*time.Second)
}

// ExecWithTimeout runs a non-interactive remote command with a strict timeout to prevent hangs.
func (c *Client) ExecWithTimeout(cmd string, timeout time.Duration) (stdout []byte, stderr []byte, err error) {
	return c.ExecWithStdin(cmd, nil, timeout)
}

// ExecWithStdin runs a command while piping stdinData directly into the SSH session's stdin channel without EOF race.
func (c *Client) ExecWithStdin(cmd string, stdinData []byte, timeout time.Duration) (stdout []byte, stderr []byte, err error) {
	c.mu.Lock()
	if c.rawClient == nil {
		c.mu.Unlock()
		return nil, nil, fmt.Errorf("ssh client is closed")
	}
	client := c.rawClient
	c.mu.Unlock()

	session, err := client.NewSession()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create ssh session: %w", err)
	}
	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	var stdinPipe io.WriteCloser
	if len(stdinData) > 0 {
		stdinPipe, err = session.StdinPipe()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open stdin pipe: %w", err)
		}
	}

	if err := session.Start(cmd); err != nil {
		return nil, nil, fmt.Errorf("failed to start command: %w", err)
	}

	if stdinPipe != nil {
		go func() {
			defer stdinPipe.Close()
			_, _ = stdinPipe.Write(stdinData)
		}()
	}

	done := make(chan error, 1)
	go func() {
		done <- session.Wait()
	}()

	select {
	case <-time.After(timeout):
		_ = session.Signal(ssh.SIGKILL)
		_ = session.Close()
		return stdoutBuf.Bytes(), stderrBuf.Bytes(), fmt.Errorf("command execution timed out after %v (possible interactive prompt or long-running task)", timeout)
	case runErr := <-done:
		if runErr != nil {
			return stdoutBuf.Bytes(), stderrBuf.Bytes(), fmt.Errorf("command execution error: %w", runErr)
		}
		return stdoutBuf.Bytes(), stderrBuf.Bytes(), nil
	}
}

// ExecWithPtyAndStdin runs a command inside an allocated pseudo-terminal (PTY) while feeding stdinData.
// This allows commands like `su` and `sudo` that strictly require a TTY to authenticate seamlessly.
func (c *Client) ExecWithPtyAndStdin(cmd string, stdinData []byte, timeout time.Duration) (stdout []byte, stderr []byte, err error) {
	c.mu.Lock()
	if c.rawClient == nil {
		c.mu.Unlock()
		return nil, nil, fmt.Errorf("ssh client is closed")
	}
	client := c.rawClient
	c.mu.Unlock()

	session, err := client.NewSession()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create ssh session: %w", err)
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0, // Disable echo so password is not displayed
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", 30, 120, modes); err != nil {
		// Fallback to standard Stdin execution if PTY request fails
		return c.ExecWithStdin(cmd, stdinData, timeout)
	}

	var combinedBuf bytes.Buffer
	session.Stdout = &combinedBuf
	session.Stderr = &combinedBuf

	stdinPipe, err := session.StdinPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open stdin pipe: %w", err)
	}

	if err := session.Start(cmd); err != nil {
		return nil, nil, fmt.Errorf("failed to start command: %w", err)
	}

	if len(stdinData) > 0 {
		go func() {
			defer stdinPipe.Close()
			time.Sleep(120 * time.Millisecond) // Wait briefly for remote prompt to be ready
			_, _ = stdinPipe.Write(stdinData)
		}()
	} else {
		_ = stdinPipe.Close()
	}

	done := make(chan error, 1)
	go func() {
		done <- session.Wait()
	}()

	select {
	case <-time.After(timeout):
		_ = session.Signal(ssh.SIGKILL)
		_ = session.Close()
		return combinedBuf.Bytes(), nil, fmt.Errorf("command execution timed out after %v", timeout)
	case runErr := <-done:
		if runErr != nil {
			return combinedBuf.Bytes(), nil, fmt.Errorf("command execution error: %w", runErr)
		}
		return combinedBuf.Bytes(), nil, nil
	}
}

// Ping checks if the SSH connection is still responsive and returns RTT latency.
func (c *Client) Ping() (time.Duration, error) {
	start := time.Now()
	_, _, err := c.Exec("echo 1")
	if err != nil {
		return 0, fmt.Errorf("ssh ping failed: %w", err)
	}
	return time.Since(start), nil
}

// IsAlive checks if the underlying connection is still open.
func (c *Client) IsAlive() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.rawClient == nil {
		return false
	}
	// Attempt a lightweight keep-alive request
	_, _, err := c.rawClient.SendRequest("keepalive@leitstand", true, nil)
	return err == nil
}

// RawClient returns the underlying *ssh.Client (useful for PTY / SFTP in later phases).
func (c *Client) RawClient() *ssh.Client {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.rawClient
}

// GetSFTPClient returns a thread-safe, cached *sftp.Client for this SSH connection.
func (c *Client) GetSFTPClient() (*sftp.Client, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.rawClient == nil {
		return nil, fmt.Errorf("ssh client is closed")
	}
	if c.sftpClient != nil {
		return c.sftpClient, nil
	}
	sc, err := sftp.NewClient(c.rawClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create sftp subsystem: %w", err)
	}
	c.sftpClient = sc
	return c.sftpClient, nil
}

// ResetSFTPClient resets the cached SFTP client instance so next call reconnects cleanly.
func (c *Client) ResetSFTPClient() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sftpClient != nil {
		_ = c.sftpClient.Close()
		c.sftpClient = nil
	}
}

// Close closes the underlying SSH connection and its cached SFTP subsystem.
func (c *Client) Close() error {
	c.StopKeepAlive()
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sftpClient != nil {
		_ = c.sftpClient.Close()
		c.sftpClient = nil
	}

	if c.rawClient != nil {
		err := c.rawClient.Close()
		c.rawClient = nil
		return err
	}
	return nil
}
