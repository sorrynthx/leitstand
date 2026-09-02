package tui

import (
	"fmt"
	"io"
	"leitstand/internal/logger"
	"leitstand/internal/ssh"
	"leitstand/internal/storage"
	"leitstand/internal/vault"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/sftp"
)

type RootElevateResultMsg struct {
	HostID   int64
	HostName string
	TabID    string
	Success  bool
	Password string
	Mode     int
	Remember bool
	Err      error
}

func listenStreamCmd(msgChan <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-msgChan
		if !ok {
			return nil
		}
		return msg
	}
}

type ptyExec struct {
	client *ssh.Client
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func (p ptyExec) SetStdin(r io.Reader)  { p.stdin = r }
func (p ptyExec) SetStdout(w io.Writer) { p.stdout = w }
func (p ptyExec) SetStderr(w io.Writer) { p.stderr = w }

func (p ptyExec) Run() error {
	in := p.stdin
	if in == nil {
		in = os.Stdin
	}
	out := p.stdout
	if out == nil {
		out = os.Stdout
	}
	errOut := p.stderr
	if errOut == nil {
		errOut = os.Stderr
	}
	return p.client.InteractiveSession(in, out, errOut)
}

func (m *Model) launchInteractiveTerminalCmd(host *storage.Host) tea.Cmd {
	client, err := m.getSSHClient(host)
	if err != nil {
		m.hostStatus[host.ID] = HostStatusOffline
		m.errors[host.ID] = err
		m.statusMessage = fmt.Sprintf("❌ PTY session failed: %v", err)
		return nil
	}
	m.hostStatus[host.ID] = HostStatusOnline
	delete(m.errors, host.ID)
	return tea.Exec(ptyExec{client: client}, func(err error) tea.Msg {
		return TerminalExitedMsg{}
	})
}

func (m *Model) getSSHClient(host *storage.Host) (*ssh.Client, error) {
	if m.store == nil || m.vault == nil {
		return nil, fmt.Errorf("storage or vault not available")
	}

	if m.sshPool == nil {
		m.sshPool = ssh.NewPool(10 * time.Second)
	}

	secret, err := m.store.GetHostSecret(host.ID)
	if err != nil {
		return nil, fmt.Errorf("secret not found: %w", err)
	}

	decrypted, err := m.vault.Decrypt(secret.Nonce, secret.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("vault decrypt failed: %w", err)
	}
	defer vault.ZeroBytes(decrypted)

	payload, err := storage.ParseSecretPayload(decrypted, secret.AuthMethod)
	if err != nil {
		return nil, fmt.Errorf("parse secret failed: %w", err)
	}

	logger.Debugf("getSSHClient: connecting to host %s (%s:%d)", host.Name, host.Address, host.Port)
	return m.sshPool.GetOrCreateFromPayload(host, secret, payload)
}

func (m *Model) getSFTPClient(host *storage.Host) (*sftp.Client, error) {
	client, err := m.getSSHClient(host)
	if err != nil {
		return nil, err
	}
	return client.GetSFTPClient()
}

func padToWidth(str string, width int) string {
	if width <= 0 {
		return str
	}
	runes := []rune(str)
	if len(runes) >= width {
		return string(runes[:width])
	}
	return str + strings.Repeat(" ", width-len(runes))
}
