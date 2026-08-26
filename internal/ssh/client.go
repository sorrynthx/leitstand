package ssh

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client wraps an active ssh.Client connection.
type Client struct {
	mu        sync.Mutex
	rawClient *ssh.Client
	address   string
	createdAt time.Time
}

// NewClient connects to the remote SSH server using the provided configuration.
func NewClient(address string, config *ssh.ClientConfig) (*Client, error) {
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("ssh dial failed to %s: %w", address, err)
	}

	return &Client{
		rawClient: client,
		address:   address,
		createdAt: time.Now(),
	}, nil
}

// Exec runs a non-interactive remote command and returns stdout, stderr with a 15-second default timeout.
func (c *Client) Exec(cmd string) (stdout []byte, stderr []byte, err error) {
	return c.ExecWithTimeout(cmd, 15*time.Second)
}

// ExecWithTimeout runs a non-interactive remote command with a strict timeout to prevent hangs.
func (c *Client) ExecWithTimeout(cmd string, timeout time.Duration) (stdout []byte, stderr []byte, err error) {
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

	done := make(chan error, 1)
	go func() {
		done <- session.Run(cmd)
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

// Close closes the underlying SSH connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.rawClient != nil {
		err := c.rawClient.Close()
		c.rawClient = nil
		return err
	}
	return nil
}
