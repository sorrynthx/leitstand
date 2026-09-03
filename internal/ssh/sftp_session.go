package ssh

import (
	"context"
	"fmt"
	"io"
	"leitstand/internal/logger"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// DedicatedSFTPSession encapsulates an isolated SFTP subsystem session with explicit termination.
type DedicatedSFTPSession struct {
	Session *ssh.Session
	Client  *sftp.Client
	stdin   io.WriteCloser
}

// Close terminates the SFTP subsystem and immediately kills the remote sftp-server process.
func (s *DedicatedSFTPSession) Close() error {
	logger.Infof("[SFTP] Closing dedicated transfer session -> sending EOF & SIGTERM")
	if s.stdin != nil {
		_ = s.stdin.Close()
	}
	if s.Client != nil {
		_ = s.Client.Close()
	}
	if s.Session != nil {
		_ = s.Session.Signal(ssh.SIGTERM)
		_ = s.Session.Close()
	}
	return nil
}

func (c *Client) createDedicatedSFTPSession(ctx context.Context) (*DedicatedSFTPSession, error) {
	c.mu.Lock()
	raw := c.rawClient
	c.mu.Unlock()
	if raw == nil {
		return nil, fmt.Errorf("ssh client is closed")
	}

	session, err := raw.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create ssh session: %w", err)
	}

	r, err := session.StdoutPipe()
	if err != nil {
		_ = session.Close()
		return nil, err
	}
	w, err := session.StdinPipe()
	if err != nil {
		_ = session.Close()
		return nil, err
	}
	if err := session.RequestSubsystem("sftp"); err != nil {
		_ = session.Close()
		return nil, fmt.Errorf("failed to request sftp subsystem: %w", err)
	}

	sc, err := sftp.NewClientPipe(r, w)
	if err != nil {
		_ = session.Close()
		return nil, fmt.Errorf("failed to create sftp client pipe: %w", err)
	}

	logger.Infof("[SFTP] Dedicated transfer session created successfully")
	return &DedicatedSFTPSession{
		Session: session,
		Client:  sc,
		stdin:   w,
	}, nil
}
